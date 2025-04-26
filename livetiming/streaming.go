package livetiming

import (
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/graph/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/websocket"
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
	topics = append(topics, "SessionInfo", "SessionData", "TimingData", "TimingStats", "TimingAppData", "LapCount", "CarData.z", "RaceControlMessages")
	var topicsList [][]string
	topicsList = append(topicsList, topics)
	return SubscribeMessage{
		H: "Streaming",
		M: "Subscribe",
		A: topicsList,
		I: 2,
	}
}

func ProcessSessionDataAndInfo(connection *websocket.Conn, dbClient *dynamodb.Client, ctx context.Context, resolver *graph.Resolver) {
	subscribeMessage := CreateOriginalSessionMessage()
	err := connection.WriteJSON(subscribeMessage)
	if err != nil {
		fmt.Println("write ERROR:", err)
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
		var sessionKey int

		go func(msg []byte) {
			meetingData, err := BuildMeetingData(msg)
			if err == nil {
				meetingDB := convertMeetingToDB(meetingData)
				err = SaveMeeting(dbClient, &ctx, meetingDB)
				if err != nil {
					fmt.Println("error saving meeting data:", err)
				}
			}
			sessionInfo, err := BuildSessionInfo(msg)
			if err == nil {
				if sessionInfo.Key == 0 {
					_, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
						TableName: aws.String("sessions"),
						Key: map[string]types.AttributeValue{
							"archiveStatus": &types.AttributeValueMemberS{Value: "Generating"},
						},
						UpdateExpression: aws.String("SET archiveStatus = :newStatus"),
						ExpressionAttributeValues: map[string]types.AttributeValue{
							":newStatus": &types.AttributeValueMemberS{Value: "Complete"},
						},
					})
					if err != nil {
						fmt.Println("error updating DynamoDB:", err)
					}
				} else {
					sessionKey = sessionInfo.Key
					err = SaveSession(dbClient, &ctx, sessionInfo)
					if err != nil {
						fmt.Println("error saving session info:", err)
					}
				}
			}
			drivers, err := BuildDriverList(msg)
			if err == nil {
				err = SaveDrivers(dbClient, &ctx, drivers, sessionKey)
				if err != nil {
					fmt.Println("error saving drivers:", err)
				}
			}
			positions, err := BuildPositions(msg)
			if err == nil {
				var positionsInterface []interface{}
				var modelPositions []*model.Position
				for _, position := range positions {
					var modelPosition model.Position
					modelPosition.Position = position.Position
					modelPosition.RacingNumber = position.RacingNumber
					modelPosition.SessionKey = sessionKey
					modelPosition.Retired = position.Retired
					modelPosition.InPit = position.InPit
					modelPosition.Stopped = position.Stopped
					modelPosition.Status = position.Status
					modelPositions = append(modelPositions, &modelPosition)
					position.SessionKey = sessionKey
					positionsInterface = append(positionsInterface, position)
				}
				err = SavePositions(dbClient, &ctx, positions)
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
					lapTime.SessionKey = sessionKey
					lapTime.BestLapTime = timing.BestLapTime.Value
					lapTime.NumberOfLaps = timing.NumberOfLaps
					lapTime.RacingNumber = timing.RacingNumber
					lapTime.LastLapTime = time
					lapTime.GapToLeader = &timing.GapToLeader
					lapTime.IntervalToPositionAhead = &timing.IntervalToPositionAhead
					timing.SessionKey = sessionKey
					err = SaveLapTime(dbClient, &ctx, timing)
					if err != nil {
						fmt.Println("error saving lap time:", err)
					}
					laptimes = append(laptimes, &lapTime)
				}
				resolver.NotifyLapTimeSubscribers(laptimes)
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
				err = SaveRaceControlMessages(dbClient, &ctx, sessionKey, raceControlModel)
				if err != nil {
					fmt.Printf("error: could not save race control messages: %v\n", err)
				}
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
				err = SaveStints(dbClient, &ctx, stints, sessionKey)
				if err != nil {
					fmt.Println("error saving stints:", err)
				}
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
				err = SaveSectors(dbClient, &ctx, sessionKey, sectorsModel)
				if err != nil {
					fmt.Println("error saving sectors:", err)
				}
			}
		}(message)
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
