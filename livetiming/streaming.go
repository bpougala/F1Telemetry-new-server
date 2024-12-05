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

func ProcessSessionDataAndInfo(connection *websocket.Conn, dbClient *mongo.Client, ctx context.Context, sessionKeyChan chan int) {
	defer close(sessionKeyChan)
	subscribeMessage := CreateOriginalSessionMessage()
	err := connection.WriteJSON(subscribeMessage)
	if err != nil {
		fmt.Println("write:", err)
		return
	}
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		var qualifyingMessage QMessage
		if !isEmptyJSON(message) {
			if err = json.Unmarshal(message, &qualifyingMessage); err == nil {
				qualifyingMessageR := qualifyingMessage.R
				fmt.Println(qualifyingMessageR.SessionInfo)
				sessionKey, err := IngestSessionInfo(dbClient, ctx, qualifyingMessageR.SessionInfo)
				if err != nil {
					fmt.Printf("Failed to ingest session info: %v\n", err)
				} else {
					sessionKeyChan <- sessionKey
				}
			}
			break
		}
	}
}

func ProcessTimingData(connection *websocket.Conn, dbClient *mongo.Client, ctx context.Context, sessionKey int) {
	subscribeMessage := CreateOriginalMessage()
	err := connection.WriteJSON(subscribeMessage)
	if err != nil {
		fmt.Println("write:", err)
		return
	}
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			return
		}
	}

}

func IngestRaceControlMessages(dbClient *mongo.Client, ctx context.Context, messages Messages) error {
	var raceControlMessagesList []interface{}
	for _, message := range messages.Messages {
		raceControlMessagesList = append(raceControlMessagesList, message)
	}
	_, err := dbClient.Database("f1").Collection("racecontrol").InsertMany(ctx, raceControlMessagesList)
	if err != nil {
		return err
	}
	return nil
}

func IngestSessionInfo(dbClient *mongo.Client, ctx context.Context, sessionInfo SessionInfo) (int, error) {
	if sessionInfo.Name == "" {
		return 0, nil
	}
	sessionInfoDB := SessionInfoDB{
		ArchiveStatus: sessionInfo.ArchiveStatus.Status,
		StartDate:     sessionInfo.StartDate,
		EndDate:       sessionInfo.EndDate,
		Type:          sessionInfo.Type,
		GmtOffset:     sessionInfo.GmtOffset,
		Key:           sessionInfo.Key,
		MeetingKey:    sessionInfo.Meeting.Key,
		Name:          sessionInfo.Name,
		Path:          sessionInfo.Path,
	}
	dbClient.Database("f1").Collection("sessioninfo").FindOneAndReplace(ctx, bson.M{"_id": sessionInfo.Key}, sessionInfoDB, options.FindOneAndReplace().SetUpsert(true))
	meetingInfoDB := MeetingDataDB{
		Key:          sessionInfo.Meeting.Key,
		Name:         sessionInfo.Meeting.Name,
		OfficialName: sessionInfo.Meeting.OfficialName,
		Location:     sessionInfo.Meeting.Location,
		Number:       sessionInfo.Meeting.Number,
		Country:      sessionInfo.Meeting.Country.Name,
		Circuit:      sessionInfo.Meeting.Circuit.ShortName,
	}
	dbClient.Database("f1").Collection("meetingdata").FindOneAndReplace(ctx, bson.M{"_id": sessionInfo.Meeting.Key}, meetingInfoDB, options.FindOneAndReplace().SetUpsert(true))
	return sessionInfo.Key, nil
}

//
//func IngestSessionData(dbClient *mongo.Client, ctx context.Context, sessionData SessionData) error {
//	if sessionData.Name == "" {
//		return nil
//	}
//	_, err := dbClient.Database("f1").Collection("sessiondata").InsertOne(ctx, sessionData)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func Main() {
	cookies, connObject, err := Negotiate()
	if err != nil {
		panic(err)
	}
	retries := 10
	connection, resp, err := SetWebSocket(connObject.ConnectionToken, cookies)
	for retries > 0 && err != nil {
		connection, resp, err = SetWebSocket(connObject.ConnectionToken, cookies)
		retries--
	}
	done := make(chan struct{})
	if err != nil {
		log.Printf("handshake failed with status %d", resp.StatusCode)
		panic(err)
	}
	defer connection.Close()
	go func() {
		defer close(done)
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				return
			}
			fmt.Printf("recv: %s", message)
		}
	}()
}

func isEmptyJSON(data []byte) bool {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return true
	}
	return len(obj) == 0
}
