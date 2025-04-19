package livetiming

import (
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/graph/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
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
	websocketUrl := fmt.Sprintf("https://livetiming.formula1.com/signalr/negotiate?connectionData=%s&clientProtocol=1.5", string(jsonData))
	resp, err := http.Get(websocketUrl)
	if err != nil {
		return nil, connection, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			fmt.Println("error defer func", err)
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

		go func(msg []byte) {
			meetingData, err := BuildMeetingData(msg)
			if err == nil {
				meetingDB := convertMeetingToDB(meetingData)
				dbClient.Database("f1").Collection("meetingdata").FindOneAndReplace(ctx, bson.D{{"_id", meetingDB.Key}}, meetingDB, options.FindOneAndReplace().SetUpsert(true))
			}
			sessionInfo, err := BuildSessionInfo(msg)
			if err == nil {
				if sessionInfo.Key == 0 {
					dbClient.Database("f1").Collection("sessioninfo").FindOneAndUpdate(
						ctx,
						bson.D{{"archiveStatus", "Generating"}},
						bson.D{{"$set", bson.D{{"archiveStatus", "Complete"}}}},
					)
				} else {
					dbClient.Database("f1").Collection("sessioninfo").FindOneAndReplace(ctx, bson.D{{"_id", sessionInfo.Key}}, sessionInfo, options.FindOneAndReplace().SetUpsert(true))
				}
			}
			drivers, err := BuildDriverList(msg)
			if err == nil {
				saveDrivers(dbClient, ctx, drivers, sessionInfo.Key)
			}
			positions, err := BuildPositions(msg)
			if err == nil {
				var positionsInterface []interface{}
				var modelPositions []*model.Position
				for _, position := range positions {
					var modelPosition model.Position
					modelPosition.Position = position.Position
					modelPosition.RacingNumber = position.RacingNumber
					modelPosition.SessionKey = sessionInfo.Key
					modelPosition.Retired = position.Retired
					modelPosition.InPit = position.InPit
					modelPosition.Stopped = position.Stopped
					modelPosition.Status = position.Status
					modelPositions = append(modelPositions, &modelPosition)
					position.SessionKey = sessionInfo.Key
					positionsInterface = append(positionsInterface, position)
				}
				dbClient.Database("f1").Collection("positions").InsertMany(ctx, positionsInterface)
				resolver.NotifyPositionSubscribers(modelPositions)
			}
			timingData, err := BuildTimingData(msg)
			if err == nil {
				var laptimes []*model.LapTime
				for _, timing := range timingData {
					var lapTime model.LapTime
					time := &model.TimeRef{
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
					if sessionInfo.ArchiveStatus != "Generating" { // might not work if we stop getting data straight after race is marked as Complete
						timing.SessionKey = sessionInfo.Key
						_, err = dbClient.Database("f1").Collection("timings").InsertOne(ctx, timing)
					}
				}
			}
			carData, err := BuildCarData(msg)
			if err == nil {
				var carDataModel model.CarData
				carDataModel.Compressed = carData.Compressed
				resolver.NotifyCarDataSubscribers(&carDataModel)
			}
			raceControlMessages, err := BuildRaceControl(msg)
			if err == nil {
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
				saveRaceControlMessages(dbClient, ctx, raceControlModel, sessionInfo.Key)
			}
			stints, err := BuildStints(msg)
			if err == nil {
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
				saveStints(dbClient, ctx, stintsModel, sessionInfo.Key)
			}
			sectors, err := BuildSectors(msg)
			if err == nil {
				var sectorsModel []*model.Sector
				for _, sectorTime := range sectors {
					for _, sector := range sectorTime.Sectors {
						var sectorModel model.Sector
						sectorModel.RacingNumber = sectorTime.RacingNumber
						sectorModel.SectorNumber = sector.SectorNumber
						sectorModel.Value = sector.Value
						sectorModel.OverallFastest = sector.OverallFastest
						sectorModel.PersonalFastest = sector.PersonalFastest
						sectorModel.Utc = &sectorTime.Utc
						sectorsModel = append(sectorsModel, &sectorModel)
					}
				}
				resolver.NotifySectorTimeSubscribers(sectorsModel)
				if sessionInfo.ArchiveStatus != "Generating" {
					var sectorsInterface []interface{}
					for _, sector := range sectors {
						sectorsInterface = append(sectorsInterface, sector)
					}
					saveSectors(dbClient, ctx, sectors, sessionInfo.Key)
				}
			}
		}(message)
	}
}

func saveDrivers(dbClient *mongo.Client, ctx context.Context, drivers []Driver, sessionKey int) {
	var driversInterface []interface{}
	for _, driver := range drivers {
		driver.SessionKey = sessionKey
		driverDocToInsert := DriverDocument{
			ID: DriverID{
				SessionKey:   sessionKey,
				RacingNumber: driver.RacingNumber,
			},
			Driver: driver,
		}
		driversInterface = append(driversInterface, driverDocToInsert)
	}
	opts := options.InsertMany().SetOrdered(false)
	_, err := dbClient.Database("f1").Collection("drivers").InsertMany(ctx, driversInterface, opts)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func saveSectors(dbClient *mongo.Client, ctx context.Context, sectors []AllSectors, sessionKey int) {
	var sectorsInterface []interface{}
	for _, sector := range sectors {
		for _, sectorTime := range sector.Sectors {
			var sectorDoc SectorDB
			sectorDoc.SessionKey = sessionKey
			sectorDoc.RacingNumber = sector.RacingNumber
			sectorDoc.Utc = sector.Utc
			sectorDoc.SectorNumber = sectorTime.SectorNumber
			sectorDoc.Value = sectorTime.Value
			sectorDoc.OverallFastest = sectorTime.OverallFastest
			sectorDoc.PersonalFastest = sectorTime.PersonalFastest
			sectorsInterface = append(sectorsInterface, sectorDoc)
		}
	}
	_, err := dbClient.Database("f1").Collection("sectors").InsertMany(ctx, sectorsInterface)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func saveStints(dbClient *mongo.Client, ctx context.Context, stints []*model.Stint, sessionKey int) {
	var stintsInterface []interface{}
	for _, stint := range stints {
		var stintDoc StintDB
		stintDoc.SessionKey = sessionKey
		stintDoc.RacingNumber = stint.RacingNumber
		stintDoc.Compound = stint.Compound
		stintDoc.Is_new = stint.New
		stintDoc.Are_tyres_not_changed = stint.TyresNotChanged == 1
		stintDoc.TotalLaps = stint.TotalLaps
		stintDoc.StartLaps = stint.StartLaps
		stintDoc.Timestamp = stint.Timestamp
		stintsInterface = append(stintsInterface, stintDoc)
	}
	_, err := dbClient.Database("f1").Collection("stints").InsertMany(ctx, stintsInterface)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func saveRaceControlMessages(dbClient *mongo.Client, ctx context.Context, raceControlMessages []*model.RaceControl, sessionKey int) {
	var raceControlMessagesInterface []interface{}
	for _, raceControlMessage := range raceControlMessages {
		var raceControlMessageDoc RaceControl
		raceControlMessageDoc.SessionKey = sessionKey
		raceControlMessageDoc.Message = raceControlMessage.Message
		raceControlMessageDoc.Category = *raceControlMessage.Category
		raceControlMessageDoc.Utc = raceControlMessage.Date
		raceControlMessageDoc.Flag = *raceControlMessage.Flag
		raceControlMessageDoc.LapNumber = raceControlMessage.LapNumber
		raceControlMessagesInterface = append(raceControlMessagesInterface, raceControlMessageDoc)
	}
	_, err := dbClient.Database("f1").Collection("racecontrol").InsertMany(ctx, raceControlMessagesInterface)
	if err != nil {
		fmt.Printf("error: %v\n", err)
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
