package main

import (
	"F1Telemetry-new-server/data-ingestor"
	"F1Telemetry-new-server/data-ingestor/collections"
	"F1Telemetry-new-server/livetiming"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

var ctx = context.Background()

var dbClient, err = data_ingestor.GetMongoClient(&ctx)

/*type MeetingDB struct {
	SessionKeys  []int  `json:"SessionKeys"`
	Year         int    `json:"Year"`
	Key          int    `json:"Key"`
	Code         string `json:"Code"`
	Number       int    `json:"Number"`
	Location     string `json:"Location"`
	OfficialName string `json:"OfficialName"`
	Name         string `json:"Name"`
	Country      struct {
		Key  int    `json:"Key"`
		Code string `json:"Code"`
		Name string `json:"Name"`
	} `json:"Country"`
	Circuit struct {
		Key       int    `json:"Key"`
		ShortName string `json:"ShortName"`
	} `json:"Circuit"`
}

func SaveMeetings() error {
	if err != nil {
		return err
	}
	meetings, err_ := data_ingestor.IngestMeetings(2024)
	if err_ != nil {
		return err_
	}
	var meetingsInterface []interface{}
	for _, meeting := range meetings.Meetings {
		err = SaveSessions(meeting)
		if err != nil {
			return err
		}
		meetingDB := MeetingDB{
			SessionKeys:  data_ingestor.GetSessionKeys(meetings, meeting.Key),
			Key:          meeting.Key,
			Year:         meetings.Year,
			Code:         meeting.Code,
			Number:       meeting.Number,
			Location:     meeting.Location,
			OfficialName: meeting.OfficialName,
			Name:         meeting.Name,
			Country:      meeting.Country,
			Circuit:      meeting.Circuit,
		}
		meetingsInterface = append(meetingsInterface, meetingDB)
	}
	_, err = dbClient.Database("f1").Collection("meetings").InsertMany(ctx, meetingsInterface)
	if err != nil {
		return err
	}
	return nil
}*/

func saveMeetings() error {
	if err != nil {
		return err
	}
	currentYear := time.Now().Year()
	meetings, err := data_ingestor.IngestMeetings(currentYear)
	if err != nil {
		return err
	}
	var meetingsInterface []interface{}
	for _, meeting := range meetings {
		sessions, err := data_ingestor.IngestSession(meeting.MeetingKey)
		if err != nil {
			return err
		}
		err = saveSessions(sessions)
		if err != nil {
			return err
		}
		meetingsInterface = append(meetingsInterface, meeting)
	}
	_, err = dbClient.Database("f1").Collection("meetings").InsertMany(ctx, meetingsInterface)
	if err != nil {
		return err
	}
	return nil
}

func saveSessions(sessions []collections.Sessions) error {
	if err != nil {
		return err
	}
	var sessionsInterface []interface{}
	for _, session := range sessions {
		sessionsInterface = append(sessionsInterface, session)
	}
	_, err = dbClient.Database("f1").Collection("sessions").InsertMany(ctx, sessionsInterface)
	return err
}

func savePositions(positions []collections.Position) error {
	if err != nil {
		return err
	}
	var positionsInterface []interface{}
	for _, position := range positions {
		positionsInterface = append(positionsInterface, position)
	}
	_, err = dbClient.Database("f1").Collection("positions").InsertMany(ctx, positionsInterface)
	if err != nil {
		return err
	}
	return err
}

func saveDrivers(drivers []collections.Driver) error {
	var driversInterface []interface{}
	for _, driver := range drivers {
		driversInterface = append(driversInterface, driver)
	}
	if len(driversInterface) > 0 {
		_, err = dbClient.Database("f1").Collection("drivers").InsertMany(ctx, driversInterface)
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

func SaveRaceControl(raceControl []collections.RaceControl) error {
	if err != nil {
		return err
	}
	var raceControlInt []interface{}
	for _, raceC := range raceControl {
		raceControlInt = append(raceControlInt, raceC)
	}
	_, err = dbClient.Database("f1").Collection("racecontrol").InsertMany(ctx, raceControlInt)
	return err
}

func SaveCarData(carData []collections.CarData) error {
	if err != nil {
		return err
	}
	var carDataInt []interface{}
	for _, carD := range carData {
		carDataInt = append(carDataInt, carD)
	}
	_, err = dbClient.Database("f1").Collection("cardata").InsertMany(ctx, carDataInt)
	return err
}

func SaveLaps(laps []collections.Laps) error {
	if err != nil {
		return err
	}
	var lapsInt []interface{}
	if len(laps) == 0 {
		return nil
	}
	for _, lap := range laps {
		lapsInt = append(lapsInt, lap)
	}
	_, err = dbClient.Database("f1").Collection("laps").InsertMany(ctx, lapsInt)
	return err
}

func SaveIntervals(intervals []collections.Interval) error {
	if err != nil {
		return err
	}
	var intervalsInt []interface{}
	if len(intervals) == 0 {
		return nil
	}
	for _, interval := range intervals {
		intervalsInt = append(intervalsInt, interval)
	}
	_, err = dbClient.Database("f1").Collection("intervals").InsertMany(ctx, intervalsInt)
	return err
}

/*func SaveSessions(meeting data_ingestor.Meeting) error {
	if err != nil {
		return err
	}
	var sessionsInterface []interface{}
	for _, session := range meeting.Sessions {
		sessionsInterface = append(sessionsInterface, session)
	}
	_, err = dbClient.Database("f1").Collection("sessions").InsertMany(ctx, sessionsInterface)
	if err != nil {
		return err
	}
	return nil
}

func FetchMeetings() ([]MeetingDB, error) {
	if err != nil {
		return nil, err
	}
	cursor, err := dbClient.Database("f1").Collection("meetings").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var meetings []MeetingDB
	err = cursor.All(ctx, &meetings)
	if err != nil {
		return nil, err
	}
	return meetings, nil
}*/

func FetchMeetings() ([]collections.Meeting, error) {
	if err != nil {
		return nil, err
	}
	cursor, err := dbClient.Database("f1").Collection("meetings").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var meetings []collections.Meeting
	err = cursor.All(ctx, &meetings)
	if err != nil {
		return nil, err
	}
	return meetings, nil
}

func FetchSessions(meetingKey int) ([]collections.Sessions, error) {
	if err != nil {
		return nil, err
	}
	cursor, err := dbClient.Database("f1").Collection("sessions").Find(ctx, bson.D{{"meetingkey", meetingKey}})
	if err != nil {
		return nil, err
	}
	var sessions []collections.Sessions
	err = cursor.All(ctx, &sessions)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

type StintData struct {
	Lines map[string]struct {
		RacingNumber string                `json:"RacingNumber"`
		Stints       []livetiming.RawStint `json:"Stints"`
		Line         int                   `json:"Line"`
		GridPos      string                `json:"GridPos"`
	} `json:"Lines"`
}

type DriverData struct {
	Drivers map[string]collections.Driver
}

func mainBisBis() {
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}
	drivers, err := data_ingestor.IngestRaceControl(9685)
	if err != nil {
		log.Fatalf("Failed to ingest drivers: %s", err)
	}
	_, err = dbClient.Database("f1").Collection("racecontrol").InsertMany(ctx, drivers)
	if err != nil {
		log.Fatalf("Failed to save drivers: %s", err)
	}
	/*file, err := os.Open("test.json")
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	// Read the file content
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}
	var initialData map[string]collections.DriverInput
	err = json.Unmarshal(byteValue, &initialData)
	if err != nil {
		panic(err)
	}
	var drivers []interface{}
	for key, value := range initialData {
		var driver collections.DriverDB
		driver.RacingNumber, _ = strconv.Atoi(key)
		driver.BroadcastName = value.BroadcastName
		driver.FullName = value.FullName
		driver.Abbreviation = value.Abbreviation
		driver.CountryCode = value.CountryCode
		driver.FirstName = value.FirstName
		driver.LastName = value.LastName
		driver.HeadshotUrl = value.HeadshotUrl
		driver.TeamName = value.TeamName
		driver.TeamColour = value.TeamColour
		driver.SessionKey = 9644
		drivers = append(drivers, driver)
	}
	_, err = dbClient.Database("f1").Collection("drivers").InsertMany(ctx, drivers)
	if err != nil {
		panic(err)
	}*/
	/*var stints []interface{}
	for _, value := range initialData.Lines {
		var stint livetiming.Stint
		stint.RacingNumber, _ = strconv.Atoi(value.RacingNumber)
		stint.SessionKey = 9644
		for count, rawStint := range value.Stints {
			stint.StintNumber = count
			stint.Compound = rawStint.Compound
			stint.LapFlags = rawStint.LapFlags
			stint.New, _ = strconv.ParseBool(rawStint.New)
			stint.StartLaps = rawStint.StartLaps
			stint.TotalLaps = rawStint.TotalLaps
			stint.TyresNotChanged, _ = strconv.Atoi(rawStint.TyresNotChanged)
			stint.Timestamp = rawStint.Timestamp
			stints = append(stints, stint)
		}
	}*/

	/*	var raceControlInt []interface{}
			for _, raceControlMessage := range data.Messages {
				raceControlMessage.SessionKey = 9662
				raceControlInt = append(raceControlInt, raceControlMessage)
			}
			_, err = dbClient.Database("f1").Collection("racecontrol").InsertMany(ctx, raceControlInt)
		_, err = dbClient.Database("f1").Collection("stints").InsertMany(ctx, stints)
		if err != nil {
			panic(err)
		}
		_, err = dbClient.Database("f1").Collection("cardata").Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{"sessionkey", 1}},
		})
		if err != nil {
			panic(err)
		}
		/*_, err = dbClient.Database("f1").Collection("cardata").Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{"sessionkey", 1}},
		})
		if err != nil {
			panic(err)
		}*/
	fmt.Println("done!")
}

/*sessions, err := FetchSessions(1229)
	if err != nil {
		fmt.Printf("error fetching sessions: %s, meeting key: %d\n", err.Error(), 1229)
	}
	sessionKey := 9468
	fmt.Printf("session key: %d\n", sessionKey)
	drivers, err := data_ingestor.FetchDrivers(dbClient, ctx, sessionKey)
	if err != nil {
		fmt.Printf("error fetching drivers: %s, session key: %d\n", err.Error(), sessionKey)
	}
	for _, driver := range drivers {
		carDataMessages, err := data_ingestor.IngestCarData(sessionKey, driver.RacingNumber)
		if err != nil {
			fmt.Printf("error fetching car data: %s, session key: %d, driver number: %d\n", err.Error(), sessionKey, driver.RacingNumber)
		}
		//err = SaveRaceControl(raceControlMessages)
		err = SaveCarData(carDataMessages)
		if err != nil {
			fmt.Printf("error in SaveCarData: %s, session key: %d, driver number: %d\n", err.Error(), sessionKey, driver.RacingNumber)
		}
		time.Sleep(8 * time.Second)
		//raceControlMessages, err := data_ingestor.IngestRaceControl(9466)

	}
	/*var indexes []mongo.IndexModel
	indexes = append(indexes, mongo.IndexModel{
		Keys: bson.D{{"sessionkey", 1}},
	})
	indexes = append(indexes, mongo.IndexModel{
		Keys: bson.D{{"drivernumber", 1}},
	})
	_, err = dbClient.Database("f1").Collection("cardata").Indexes().CreateMany(ctx, indexes)
	if err != nil {
		panic(err)
	}
}
*/
/*err = saveMeetings()
	if err != nil {
		panic(err)
	}
	err = createIndex()
	if err != nil {
		panic(err)
	}
	meetings, err := FetchMeetings()
	if err != nil {
		panic(err)
	}
	for _, meeting := range meetings {
		sessions, err := FetchSessions(meeting.MeetingKey)
		if err != nil {
			panic(err)
		}
		for _, session := range sessions {
			drivers, err := data_ingestor.IngestDrivers(9466)
			if err != nil {
				panic(err)
			}
			err = saveDrivers(drivers)
			if err != nil {
				panic(err)
			}
			positions, err := data_ingestor.IngestPositions(9466)
			if err != nil {
				panic(err)
			}
			err = savePositions(positions)
			time.Sleep(2 * time.Second)
		}
	}
	err = createIndex()
	if err != nil {
		panic(err)
	}
}
*/
/*func IngestSession(year int, meetingKey string, sessionKey string) error {
	if err != nil {
		return err
	}
	dbClient.Database("openf1").Collection("sessions").InsertOne(ctx, Session{
		Key: sessionKey,


	}
	return nil
}*/
