package livetiming

import (
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/graph/model"
	"F1Telemetry-new-server/ws"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"
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
	req, err := http.NewRequest("GET", websocketUrl, nil)
	if err != nil {
		return nil, connection, err
	}
	req.Header.Set("User-Agent", "BestHTTP")
	req.Header.Set("Accept-Encoding", "gzip,identity")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, connection, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			fmt.Println("error defer func", err)
		}
	}(resp.Body)
	if resp.StatusCode != 200 {
		return nil, connection, fmt.Errorf("negotiate returned status %d", resp.StatusCode)
	}
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
	topics = append(topics, "SessionInfo", "SessionData", "TimingData", "DriverList", "TimingStats", "TimingAppData", "LapCount", "CarData.z", "RaceControlMessages", "Position.z", "TrackStatus", "WeatherData")
	var topicsList [][]string
	topicsList = append(topicsList, topics)
	return SubscribeMessage{
		H: "Streaming",
		M: "Subscribe",
		A: topicsList,
		I: 2,
	}
}

func ProcessSessionDataAndInfo(connection *websocket.Conn, dbClient *dynamodb.Client, s3Client *s3.Client, ctx context.Context, resolver *graph.Resolver, hub *ws.Hub) error {
	subscribeMessage := CreateOriginalSessionMessage()
	err := connection.WriteJSON(subscribeMessage)
	if err != nil {
		return err
	}
	connection.SetPongHandler(func(appData string) error {
		return nil
	})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := connection.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))
				if err != nil {
					fmt.Println("error sending ping:", err)
					return
				}
			}
		}
	}()
	logger := NewS3Logger(s3Client, ctx)
	var sessionKey int
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			return err
		}

		// Determine message type by checking for "R" (initial snapshot) or "M" (updates) key
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(message, &raw); err != nil {
			fmt.Println("error unmarshaling message:", err)
			continue
		}

		if _, hasR := raw["R"]; hasR {
			// Initial snapshot — dispatch to all initial parsers
			sessionKey = processInitialSnapshot(message, sessionKey, dbClient, s3Client, ctx, resolver, hub, logger)
		} else if _, hasM := raw["M"]; hasM {
			// Update messages — dispatch by topic
			sessionKey = processUpdateMessages(message, sessionKey, dbClient, ctx, resolver, hub, logger)
		}
	}
}

func processInitialSnapshot(msg []byte, sessionKey int, dbClient *dynamodb.Client, s3Client *s3.Client, ctx context.Context, resolver *graph.Resolver, hub *ws.Hub, logger *S3Logger) int {
	meetingData, err := BuildMeetingData(msg)
	raceName := ""
	if err == nil {
		raceName = meetingData.Name
		meetingDB := ConvertMeetingToDB(meetingData)
		err = SaveMeeting(dbClient, &ctx, meetingDB)
		if err != nil {
			fmt.Println("error saving meeting data:", err)
		}
		if meetingData.Circuit.Key != 0 {
			DownloadAndStoreCircuit(s3Client, ctx, meetingData.Circuit.Key, time.Now().Year())
		}
	}
	sessionInfo, err := BuildSessionInfo(msg)
	sessionName := ""
	if err == nil && sessionInfo.Key != 0 {
		sessionKey = sessionInfo.Key
		sessionName = sessionInfo.Name
		hub.SetSessionKey(sessionKey)
		err = SaveSession(dbClient, &ctx, sessionInfo)
		if err != nil {
			fmt.Println("error saving session info:", err)
		}
	}
	if raceName != "" && sessionName != "" {
		logger.SetSession(raceName, sessionName)
		logger.LogMessage("snapshot", msg)
	}
	drivers, err := BuildDriverList(msg)
	if err == nil {
		for i := range drivers {
			drivers[i].SessionKey = sessionKey
		}
		err = SaveDrivers(dbClient, &ctx, drivers, sessionKey)
		if err != nil {
			fmt.Println("error saving drivers:", err)
		}
	}
	positions, err := BuildPositions(msg)
	if err == nil {
		processPositions(positions, sessionKey, dbClient, ctx, resolver)
	}
	timingData, err := BuildTimingData(msg)
	if err == nil {
		processTimingData(timingData, sessionKey, dbClient, ctx, resolver)
	}
	qualifyingState, _ := BuildQualifyingState(msg)
	qualifyingParts, _ := BuildQualifyingParts(msg)
	if qualifyingState != nil {
		processQualifyingData(qualifyingState, qualifyingParts, resolver)
	}
	carData, err := BuildCarData(msg)
	if err == nil {
		var carDataModel model.CarData
		carDataModel.Compressed = carData.Compressed
		resolver.NotifyCarDataSubscribers(&carDataModel)
	}
	trackStatus, err := BuildTrackStatus(msg)
	if err == nil {
		processTrackStatus(trackStatus, sessionKey, dbClient, ctx, resolver)
	}
	weatherData, err := BuildWeatherData(msg)
	if err == nil {
		processWeather(weatherData, sessionKey, dbClient, ctx, resolver)
	}
	positionZCompressed, err := ExtractPositionZCompressed(msg)
	if err == nil {
		posRoot, err := DecompressPositionData(positionZCompressed)
		if err == nil {
			resolver.NotifyDriverLocationSubscribers(positionRootToDriverPositions(posRoot))
		}
	}
	raceControlMessages, err := BuildRaceControl(msg)
	if err == nil {
		processRaceControl(raceControlMessages, sessionKey, dbClient, ctx, resolver)
	}
	stints, err := BuildStints(msg)
	if err == nil {
		processStints(stints, sessionKey, dbClient, ctx, resolver)
	}
	sectors, err := BuildSectors(msg)
	if err == nil {
		processSectors(sectors, sessionKey, dbClient, ctx, resolver)
	}
	segments, err := BuildSegments(msg)
	if err == nil {
		processSegments(segments, resolver, hub)
	}
	return sessionKey
}

func processUpdateMessages(msg []byte, sessionKey int, dbClient *dynamodb.Client, ctx context.Context, resolver *graph.Resolver, hub *ws.Hub, logger *S3Logger) int {
	var updateData UpdateData
	if err := json.Unmarshal(msg, &updateData); err != nil {
		fmt.Println("error unmarshaling update data:", err)
		return sessionKey
	}
	for _, message := range updateData.M {
		if len(message.A) < 2 {
			continue
		}
		topic, ok := message.A[0].(string)
		if !ok {
			continue
		}
		// Log the topic's arguments to S3
		if topicData, err := json.Marshal(message.A); err == nil {
			logger.LogMessage(topic, topicData)
		}
		switch topic {
		case "SessionInfo":
			sessionInfo, err := BuildSessionInfo(msg)
			if err == nil {
				if sessionInfo.Key == 0 {
					_, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
						TableName: aws.String("sessions"),
						Key: map[string]types.AttributeValue{
							"SessionKey": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
						},
						UpdateExpression: aws.String("SET ArchiveStatus = :newStatus"),
						ExpressionAttributeValues: map[string]types.AttributeValue{
							":newStatus": &types.AttributeValueMemberS{Value: "Complete"},
						},
					})
					if err != nil {
						fmt.Println("error updating DynamoDB:", err)
					}
				} else {
					sessionKey = sessionInfo.Key
					hub.SetSessionKey(sessionKey)
					err = SaveSession(dbClient, &ctx, sessionInfo)
					if err != nil {
						fmt.Println("error saving session info:", err)
					}
				}
			}
		case "SessionData":
			sessionInfo, err := BuildSessionInfo(msg)
			if err == nil {
				if sessionInfo.Key == 0 {
					_, err := dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
						TableName: aws.String("sessions"),
						Key: map[string]types.AttributeValue{
							"SessionKey": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
						},
						UpdateExpression: aws.String("SET ArchiveStatus = :newStatus"),
						ExpressionAttributeValues: map[string]types.AttributeValue{
							":newStatus": &types.AttributeValueMemberS{Value: "Complete"},
						},
					})
					if err != nil {
						fmt.Println("error updating DynamoDB:", err)
					}
				}
			}
			qualifyingParts, _ := BuildQualifyingParts(msg)
			if qualifyingParts != nil {
				processQualifyingData(nil, qualifyingParts, resolver)
			}
		case "TimingData":
			positions, err := BuildPositions(msg)
			if err == nil {
				processPositions(positions, sessionKey, dbClient, ctx, resolver)
			}
			timingData, err := BuildTimingData(msg)
			if err == nil {
				processTimingData(timingData, sessionKey, dbClient, ctx, resolver)
			}
			qualifyingState, _ := BuildQualifyingState(msg)
			if qualifyingState != nil {
				processQualifyingData(qualifyingState, nil, resolver)
			}
			sectors, err := BuildSectors(msg)
			if err == nil {
				processSectors(sectors, sessionKey, dbClient, ctx, resolver)
			}
			segments, err := BuildSegments(msg)
			if err == nil {
				processSegments(segments, resolver, hub)
			}
		case "DriverList":
			positions, err := BuildPositions(msg)
			if err == nil {
				processPositions(positions, sessionKey, dbClient, ctx, resolver)
			}
		case "CarData.z":
			carData, err := BuildCarData(msg)
			if err == nil {
				var carDataModel model.CarData
				carDataModel.Compressed = carData.Compressed
				resolver.NotifyCarDataSubscribers(&carDataModel)
			}
		case "Position.z":
			positionZCompressed, err := ExtractPositionZCompressed(msg)
			if err != nil {
				fmt.Println("error extracting Position.z:", err)
			} else {
				posRoot, err := DecompressPositionData(positionZCompressed)
				if err != nil {
					fmt.Println("error decompressing Position.z:", err)
				} else {
					resolver.NotifyDriverLocationSubscribers(positionRootToDriverPositions(posRoot))
				}
			}
		case "TrackStatus":
			trackStatus, err := BuildTrackStatusUpdate(msg)
			if err == nil {
				processTrackStatus(trackStatus, sessionKey, dbClient, ctx, resolver)
			}
		case "RaceControlMessages":
			raceControlMessages, err := BuildRaceControl(msg)
			if err == nil {
				processRaceControl(raceControlMessages, sessionKey, dbClient, ctx, resolver)
			}
		case "TimingAppData":
			stints, err := BuildStints(msg)
			if err == nil {
				processStints(stints, sessionKey, dbClient, ctx, resolver)
			}
		case "WeatherData":
			weatherData, err := BuildWeatherDataUpdate(msg)
			if err == nil {
				processWeather(weatherData, sessionKey, dbClient, ctx, resolver)
			}
		}
	}
	return sessionKey
}

func processPositions(positions []Position, sessionKey int, dbClient *dynamodb.Client, ctx context.Context, resolver *graph.Resolver) {
	var positionsWithSessionKey []Position
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
		positionsWithSessionKey = append(positionsWithSessionKey, position)
	}
	err := SavePositions(dbClient, &ctx, positionsWithSessionKey)
	if err != nil {
		fmt.Println("error saving positions:", err)
	}
	resolver.NotifyPositionSubscribers(modelPositions)
}

func processTimingData(timingData []LapTimeMetric, sessionKey int, dbClient *dynamodb.Client, ctx context.Context, resolver *graph.Resolver) {
	var laptimes []*model.LapTime
	for _, timing := range timingData {
		var lapTime model.LapTime
		t := &model.TimeRef{
			Value:           timing.LastLapTime.Value,
			OverallFastest:  timing.LastLapTime.OverallFastest,
			PersonalFastest: timing.LastLapTime.PersonalFastest,
		}
		lapTime.SessionKey = sessionKey
		lapTime.BestLapTime = timing.BestLapTime.Value
		lapTime.NumberOfLaps = timing.NumberOfLaps
		lapTime.RacingNumber = timing.RacingNumber
		lapTime.LastLapTime = t
		if timing.GapToLeader != "" {
			lapTime.GapToLeader = &timing.GapToLeader
		}
		if timing.IntervalToPositionAhead.Value != "" {
			lapTime.IntervalToPositionAhead = &timing.IntervalToPositionAhead
		}
		lapTime.InPit = timing.InPit
		lapTime.PitOut = timing.PitOut
		lapTime.Retired = timing.Stopped

		// Qualifying-specific fields
		knockedOut := timing.KnockedOut
		lapTime.KnockedOut = &knockedOut
		cutoff := timing.Cutoff
		lapTime.Cutoff = &cutoff
		if len(timing.QualifyingBestLaps) > 0 {
			for _, qlt := range timing.QualifyingBestLaps {
				lapTime.BestLapTimes = append(lapTime.BestLapTimes, &model.QualifyingLapTime{
					Value: qlt.Value,
					Lap:   qlt.Lap,
				})
			}
		}
		if len(timing.QualifyingStats) > 0 {
			for _, qs := range timing.QualifyingStats {
				lapTime.QualifyingStats = append(lapTime.QualifyingStats, &model.QualifyingStats{
					TimeDiffToFastest:       qs.TimeDiffToFastest,
					TimeDiffToPositionAhead: qs.TimeDifftoPositionAhead,
				})
			}
		}

		timing.SessionKey = sessionKey
		err := SaveLapTime(dbClient, &ctx, timing)
		if err != nil {
			fmt.Println("error saving lap time:", err)
		}
		laptimes = append(laptimes, &lapTime)
	}
	resolver.NotifyLapTimeSubscribers(laptimes)
}

func processQualifyingData(state *QualifyingState, parts []QualifyingPart, resolver *graph.Resolver) {
	qd := &model.QualifyingData{}

	// Merge with current state if we only have partial data
	if current := resolver.CurrentQualifyingData; current != nil {
		*qd = *current
	}

	if state != nil {
		qd.SessionPart = state.SessionPart
		qd.NoEntries = state.NoEntries
		qd.CutOffTime = state.CutOffTime
		qd.CutOffPercentage = state.CutOffPercentage
	}

	if parts != nil {
		var modelParts []*model.QualifyingPart
		for _, p := range parts {
			modelParts = append(modelParts, &model.QualifyingPart{
				Part: p.QualifyingPart,
				Utc:  p.Utc,
			})
		}
		qd.Parts = modelParts
	}

	resolver.NotifyQualifyingDataSubscribers(qd)
}

func processTrackStatus(trackStatus []TrackStatus, sessionKey int, dbClient *dynamodb.Client, ctx context.Context, resolver *graph.Resolver) {
	var trackStatusModel []*model.TrackStatus
	for _, status := range trackStatus {
		var trackStatusEntry model.TrackStatus
		trackStatusEntry.Status = status.Status
		trackStatusEntry.Timestamp = status.Utc.Format(time.RFC3339)
		trackStatusEntry.SessionKey = sessionKey
		trackStatusModel = append(trackStatusModel, &trackStatusEntry)
	}
	resolver.NotifyTrackStatusSubscribers(trackStatusModel)
	err := SaveTrackStatus(dbClient, &ctx, trackStatusModel)
	if err != nil {
		fmt.Println("error saving track status:", err)
	}
}

func processRaceControl(raceControlMessages []RaceControl, sessionKey int, dbClient *dynamodb.Client, ctx context.Context, resolver *graph.Resolver) {
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
	err := SaveRaceControlMessages(dbClient, &ctx, sessionKey, raceControlModel)
	if err != nil {
		fmt.Printf("error: could not save race control messages: %v\n", err)
	}
}

func processStints(stints []Stint, sessionKey int, dbClient *dynamodb.Client, ctx context.Context, resolver *graph.Resolver) {
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
	err := SaveStints(dbClient, &ctx, stints, sessionKey)
	if err != nil {
		fmt.Println("error saving stints:", err)
	}
}

func processSectors(sectors []AllSectors, sessionKey int, dbClient *dynamodb.Client, ctx context.Context, resolver *graph.Resolver) {
	var sectorsModel []*model.Sector
	for _, sectorTime := range sectors {
		for _, sector := range sectorTime.Sectors {
			var sectorModel model.Sector
			sectorModel.LapNumber = sector.LapNumber
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
	err := SaveSectors(dbClient, &ctx, sessionKey, sectorsModel)
	if err != nil {
		fmt.Println("error saving sectors:", err)
	}
}

func processSegments(segments []AllSegments, resolver *graph.Resolver, hub *ws.Hub) {
	var segmentsModel []*model.Segment
	for _, driverSegments := range segments {
		for _, seg := range driverSegments.Segments {
			segmentsModel = append(segmentsModel, &model.Segment{
				RacingNumber:  driverSegments.RacingNumber,
				SectorNumber:  seg.SectorNumber,
				SegmentNumber: seg.SegmentNumber,
				Status:        seg.Status,
			})
		}
	}
	hub.UpdateSegmentCounts(segmentsModel)
	resolver.NotifySegmentSubscribers(segmentsModel)
}

func processWeather(weather *Weather, sessionKey int, dbClient *dynamodb.Client, ctx context.Context, resolver *graph.Resolver) {
	airTemp, _ := strconv.ParseFloat(weather.AirTemp, 64)
	trackTemp, _ := strconv.ParseFloat(weather.TrackTemp, 64)
	humidity, _ := strconv.ParseFloat(weather.Humidity, 64)
	pressure, _ := strconv.ParseFloat(weather.Pressure, 64)
	windSpeed, _ := strconv.ParseFloat(weather.WindSpeed, 64)
	windDirection, _ := strconv.Atoi(weather.WindDirection)
	rainfall := weather.Rainfall == "1"

	weatherModel := &model.Weather{
		SessionKey:       sessionKey,
		AirTemperature:   airTemp,
		TrackTemperature: trackTemp,
		Humidity:         humidity,
		Pressure:         pressure,
		Rainfall:         rainfall,
		WindSpeed:        windSpeed,
		WindDirection:    windDirection,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
	}
	resolver.NotifyWeatherSubscribers(weatherModel)
	err := SaveWeather(dbClient, &ctx, weatherModel)
	if err != nil {
		fmt.Println("error saving weather:", err)
	}
}

func positionRootToDriverPositions(posRoot PositionRoot) []*model.DriverPosition {
	var positions []*model.DriverPosition
	if len(posRoot.Position) == 0 {
		return positions
	}
	latest := posRoot.Position[len(posRoot.Position)-1]
	for numStr, car := range latest.Entries {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		positions = append(positions, &model.DriverPosition{
			RacingNumber: num,
			X:            car.X,
			Y:            car.Y,
		})
	}
	return positions
}

func ConvertMeetingToDB(meeting MeetingData) MeetingDataDB {
	return MeetingDataDB{
		Key:          meeting.Key,
		Name:         meeting.Name,
		OfficialName: meeting.OfficialName,
		Location:     meeting.Location,
		Number:       meeting.Number,
		CountryName:  meeting.Country.Name,
		CountryCode:  meeting.Country.Code,
		Circuit:      meeting.Circuit.ShortName,
		CircuitKey:   meeting.Circuit.Key,
	}
}
