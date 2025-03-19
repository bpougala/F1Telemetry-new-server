package livetiming

import (
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/graph/model"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func CreateTimingSubscribeMessage() SubscribeMessage {
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
	topics = append(topics, "SessionInfo", "SessionData", "TimingData", "TimingStats", "TimingAppData", "LapCount", "DriverList", "CarData.z", "RaceControlMessages")
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
		I: 1,
	}
}

func ProcessSessionDataAndInfo(connection *websocket.Conn, dbClient *mongo.Client, ctx context.Context, sessionKeyChan chan int, resolver *graph.Resolver) {
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("unexpected close error: %v", err)
			} else {
				fmt.Println("read:", err)
			}
			return
		}
		isError := false
		meetingData, err := BuildMeetingData(message)
		if err != nil {
			isError = true
		}
		sessionInfo, err := BuildSessionInfo(message)
		if err != nil {
			isError = true
		}
		drivers, err := BuildDriverList(message)
		if err != nil {
			isError = true
		}
		positions, err := BuildPositions(message)
		if err == nil && !isError {
			var positionsInterface []interface{}
			var modelPositions []*model.Position
			for _, position := range positions {
				var modelPosition model.Position
				modelPosition.Position = position.Position
				modelPosition.RacingNumber = position.RacingNumber
				modelPosition.SessionKey = sessionInfo.Key
				modelPositions = append(modelPositions, &modelPosition)
				position.SessionKey = sessionInfo.Key
				positionsInterface = append(positionsInterface, position)
			}
			dbClient.Database("f1").Collection("positions").InsertMany(ctx, positionsInterface)
			resolver.NotifyPositionSubscribers(modelPositions)
		}
		timingData, err := BuildTimingData(message)
		if err == nil && !isError {
			var laptimes []*model.LapTime
			for _, timing := range timingData {
				var lapTime model.LapTime
				time := &model.Time{
					Value:           timing.LastLapTime.Value,
					OverallFastest:  timing.LastLapTime.OverallFastest,
					PersonalFastest: timing.LastLapTime.PersonalFastest,
				}
				lapTime.SessionKey = sessionInfo.Key
				lapTime.BestLapTime = timing.BestLapTime.Value
				lapTime.NumberOfLaps = timing.NumberOfLaps
				lapTime.RacingNumber = timing.RacingNumber
				lapTime.LastLapTime = time
				lapTime.GapToLeader = &timing.GapToLeader
				lapTime.IntervalToPositionAhead = &timing.IntervalToPositionAhead
				laptimes = append(laptimes, &lapTime)
				resolver.NotifyLapTimeSubscribers(laptimes)
			}
		}
		carData, err := BuildCarData(message)
		if err == nil && !isError {
			var carDataModel model.CarData
			carDataModel.Compressed = carData.Compressed
			resolver.NotifyCarDataSubscribers(&carDataModel)
		}
		raceControlMessages, err := BuildRaceControl(message)
		if err == nil && !isError {
			var raceControlModel []*model.RaceControl
			for _, message := range raceControlMessages {
				var raceControl model.RaceControl
				raceControl.Message = message.Message
				raceControl.Category = &message.Category
				raceControl.Date = message.Utc
				raceControl.Flag = &message.Flag
				raceControlModel = append(raceControlModel, &raceControl)
			}
			resolver.NotifyRaceControlSubscribers(raceControlModel)
		}
		stints, err := BuildStints(message)
		if err == nil && !isError {
			var stintsModel []*model.Stint
			for _, stint := range stints {
				var stintModel model.Stint
				stintModel.Compound = stint.Compound
				stintModel.LapFlags = stint.LapFlags
				stintModel.RacingNumber = stint.RacingNumber
				stintModel.New = stint.New
				stintModel.StartLaps = stint.StartLaps
				stintModel.StintNumber = stint.StintNumber
				stintModel.Timestamp = stint.Timestamp
				stintModel.TotalLaps = stint.TotalLaps
				stintModel.TyresNotChanged = stint.TyresNotChanged
				stintsModel = append(stintsModel, &stintModel)
			}
			resolver.NotifyStintSubscribers(stintsModel)
		}
		sectors, err := BuildSectors(message)
		if err == nil && !isError {
			var sectorsModel []*model.Sector
			for _, sectorTime := range sectors {
				for _, sector := range sectorTime.Sectors {
					var sectorModel model.Sector
					sectorModel.RacingNumber = sectorTime.RacingNumber
					sectorModel.SectorNumber = sector.SectorNumber
					sectorModel.Value = sector.Value
					sectorModel.OverallFastest = sector.OverallFastest
					sectorModel.PersonalFastest = sector.PersonalFastest
					if sectorTime.Utc.Year() != 1 {
						time := sectorTime.Utc.String()
						sectorModel.Utc = &time
					}
					sectorsModel = append(sectorsModel, &sectorModel)
				}
			}
			resolver.NotifySectorTimeSubscribers(sectorsModel)
		}
		if !isError {
			meetingDB := convertMeetingToDB(meetingData)
			var driverInterface []interface{}
			for _, driver := range drivers {
				driver.SessionKey = sessionInfo.Key
				driverInterface = append(driverInterface, driver)
			}
			dbClient.Database("f1").Collection("meetingdata").FindOneAndReplace(ctx, bson.D{{"_id", meetingDB.Key}}, meetingDB, options.FindOneAndReplace().SetUpsert(true))
			dbClient.Database("f1").Collection("sessioninfo").FindOneAndReplace(ctx, bson.D{{"_id", sessionInfo.Key}}, sessionInfo, options.FindOneAndReplace().SetUpsert(true))
			_, err = dbClient.Database("f1").Collection("drivers").InsertMany(ctx, driverInterface)
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
	fmt.Printf("subscribe message: %v\n", subscribeMessage)
	err := connection.WriteJSON(subscribeMessage)
	if err != nil {
		fmt.Println("An error has occurred:", err)
		return
	}
	fmt.Println("subscribing to driver data")
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("unexpected close error: %v", err)
			} else {
				fmt.Println("read:", err)
			}
		}
		fmt.Printf("driver message: %s\n", string(message))
		drivers, err := BuildDriverList(message)
		if err == nil {
			var driverInterface []interface{}
			for _, driver := range drivers {
				driver.SessionKey = sessionKey
				driverInterface = append(driverInterface, driver)
			}
			//_, err = dbClient.Database("f1").Collection("drivers").InsertMany(ctx, driverInterface)
			if err != nil {
				fmt.Println(err)
			}
			break
		} else {
			fmt.Printf("error: %v\n", err)
		}
	}
}

func ProcessTimingData(connection *websocket.Conn, dbClient *mongo.Client, ctx context.Context, sessionKey int) {
	subscribeMessage := CreateTimingSubscribeMessage()
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
		positions, err := BuildPositions(message)
		if err == nil {
			var positionsInterface []interface{}
			for _, position := range positions {
				position.SessionKey = sessionKey
				positionsInterface = append(positionsInterface, position)
			}
			//_, err = dbClient.Database("f1").Collection("positions").InsertMany(ctx, positionsInterface)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
