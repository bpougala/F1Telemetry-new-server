package livetiming

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
)

type Connection struct {
	Url                     string
	ConnectionToken         string
	ConnectionId            string
	KeepAliveTimeout        float32
	DisconnectTimeout       float32
	ConnectionTimeout       float32
	TryWebSockets           bool
	ProtocolVersion         string
	TransportConnectTimeout float32
	LongPollDelay           float32
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	reader(ws)
}

func reader(socketConnection *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := socketConnection.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("message type: %d\nMessage: %s\n", messageType, string(p))
	}
}

func Negotiate() ([]*http.Cookie, Connection, error) {
	data := struct {
		Name string `json:"name"`
	}{Name: "Streaming"}
	var payload []interface{}
	payload = append(payload, data)
	jsonData, err := json.Marshal(payload)
	var connection Connection

	if err != nil {
		return nil, connection, err
	}
	url := fmt.Sprintf("https://livetiming.formula1.com/signalr/negotiate?connectionData=%s&clientProtocol=1.5", string(jsonData))
	resp, err := http.Get(url)
	if err != nil {
		return nil, connection, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &connection)
	if err != nil {
		return nil, connection, err
	}
	cookies := resp.Cookies()
	return cookies, connection, nil
}

type ConnectionData struct {
	Name string `json:"name"`
}

/*
{
	"H": "Streaming",
	"M": "Subscribe",
	"A": [["TimingData", "Position.z", "SessionData"]],
	"I": 1
}
*/

type SubscribeMessage struct {
	H string     `json:"H"`
	M string     `json:"M"`
	A [][]string `json:"A"`
	I int        `json:"I"`
}

func SetWebSocket(connectionToken string, cookies []*http.Cookie) (*websocket.Conn, *http.Response, error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	connectionDataObj := ConnectionData{Name: "Streaming"}
	connectionDataList := []ConnectionData{connectionDataObj}
	connectionData, err := json.Marshal(connectionDataList)
	if err != nil {
		return nil, nil, err
	}
	connectionDataStr := string(connectionData)
	path := fmt.Sprintf("wss://livetiming.formula1.com/signalr/connect?clientProtocol=1.5&transport=webSockets&connectionToken=%s&connectionData=%s", url.QueryEscape(connectionToken), url.QueryEscape(connectionDataStr))
	uri, err := url.Parse(path)
	if err != nil {
		return nil, nil, err
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, nil, err
	}
	client := &http.Client{
		Jar: jar,
	}
	headers := http.Header{}
	headers.Add("User-Agent", "BestHTTP")
	headers.Add("Accept-Encoding", "gzip,identity")
	headers.Add("Cookie", fmt.Sprintf("AWSALB=%s", url.QueryEscape(cookies[0].Value)))
	headers.Add("Cookie", fmt.Sprintf("AWSALBCORS=%s", url.QueryEscape(cookies[1].Value)))

	if len(cookies) > 0 {
		client.Jar.SetCookies(uri, cookies)
	}
	connection, resp, err := websocket.DefaultDialer.Dial(uri.String(), headers)
	if err != nil {
		fmt.Printf("handshake failed with status %d\nCookie 1: %s\nCookie 2: %s\n", resp.StatusCode, cookies[0].Value, cookies[1].Value)
		return nil, nil, err
	}

	return connection, resp, nil
}

func CreateOriginalMessage() SubscribeMessage {
	var topics []string
	topics = append(topics, "TimingData", "TimingStats", "TimingAppData", "LapCount")
	var topicsList [][]string
	topicsList = append(topicsList, topics)
	return SubscribeMessage{
		H: "Streaming",
		M: "Subscribe",
		A: topicsList,
		I: 1,
	}
}

func CreateOriginalSessionMessage() SubscribeMessage {
	var topics []string
	topics = append(topics, "SessionInfo", "SessionData")
	var topicsList [][]string
	topicsList = append(topicsList, topics)
	return SubscribeMessage{
		H: "Streaming",
		M: "Subscribe",
		A: topicsList,
		I: 2,
	}
}

func CreateDriverSubscribeMessage() SubscribeMessage {
	var topics []string
	topics = append(topics, "DriverList")
	var topicsList [][]string
	topicsList = append(topicsList, topics)
	return SubscribeMessage{
		H: "Streaming",
		M: "Subscribe",
		A: topicsList,
		I: 3,
	}
}

func ProcessSessionDataAndInfo(connection *websocket.Conn, dbClient *mongo.Client, ctx context.Context, sessionKeyChan chan int) {
	defer close(sessionKeyChan)
	fmt.Println("Processing session data and info")
	subscribeMessage := CreateOriginalSessionMessage()
	err := connection.WriteJSON(subscribeMessage)
	if err != nil {
		fmt.Println("write:", err)
		return
	}
	for {
		_, message, err := connection.ReadMessage()
		fmt.Println("received message")
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("unexpected close error: %v", err)
			} else {
				fmt.Println("read:", err)
			}
			return
		}
		meetingData, err := BuildMeetingData(message)
		if err != nil {
			fmt.Printf("Error building meeting data: %v\n", err)
		}
		sessionInfo, err := BuildSessionInfo(message)
		if err != nil {
			fmt.Printf("Error building session info: %v\n", err)
		}
		if err == nil {
			sessionKeyChan <- sessionInfo.Key
			meetingDB := convertMeetingToDB(meetingData)
			err := dbClient.Database("f1").Collection("meetingdata").FindOneAndReplace(ctx, bson.D{{"_id", meetingDB.Key}}, meetingDB, options.FindOneAndReplace().SetUpsert(true))
			if err != nil {
				fmt.Printf("Error inserting meeting data: %v\n", err)
			}
			err = dbClient.Database("f1").Collection("sessioninfo2").FindOneAndReplace(ctx, bson.D{{"_id", sessionInfo.Key}}, sessionInfo, options.FindOneAndReplace().SetUpsert(true))
			if err != nil {
				fmt.Printf("Error inserting session info: %v\n", err)
			}
		}
	}
}

func convertMeetingToDB(meeting MeetingData) MeetingDataDB {
	return MeetingDataDB{
		Key:          meeting.Key,
		Name:         meeting.Name,
		OfficialName: meeting.OfficialName,
		Location:     meeting.Location,
		Number:       meeting.Number,
		CountryName:  meeting.Country.Name,
		CountryCode:  meeting.Country.Code,
		Circuit:      meeting.Circuit.ShortName,
	}
}

func convertSessionToDB(session SessionInfo) SessionInfoDB {
	return SessionInfoDB{
		Key:        session.Key,
		MeetingKey: session.Meeting.Key,
		Name:       session.Name,
		StartDate:  session.StartDate,
		EndDate:    session.EndDate,
		Type:       session.Type,
		GmtOffset:  session.GmtOffset,
	}
}

func ProcessDriverData(connection *websocket.Conn, dbClient *mongo.Client, ctx context.Context, sessionKey int) {
	subscribeMessage := CreateDriverSubscribeMessage()
	err := connection.WriteJSON(subscribeMessage)
	if err != nil {
		fmt.Println("write:", err)
		return
	}
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("unexpected close error: %v", err)
			} else {
				fmt.Println("read:", err)
			}
			return
		}
		drivers, err := BuildDriverList(message)
		if err == nil {
			var driverInterface []interface{}
			for _, driver := range drivers {
				driver.SessionKey = sessionKey
				driverInterface = append(driverInterface, driver)
			}
			_, err = dbClient.Database("f1").Collection("drivers").InsertMany(ctx, driverInterface)
			break
		}
	}
}
