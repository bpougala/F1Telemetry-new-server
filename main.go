package main

import (
	"F1Telemetry-new-server/data-ingestor"
	"F1Telemetry-new-server/data-ingestor/collections"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
func createIndex() error {
	if err != nil {
		return err
	}
	_, err = dbClient.Database("f1").Collection("sessions").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"SessionKey", 1}},
	})
	if err != nil {
		return err
	}
	_, err = dbClient.Database("f1").Collection("meetings").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"MeetingKey", 1}},
	})
	return nil
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

func main() {
	/*err = saveMeetings()
	if err != nil {
		panic(err)
	}
	err = createIndex()
	if err != nil {
		panic(err)
	}*/
	meetings, err := FetchMeetings()
	if err != nil {
		panic(err)
	}
	jsonStr, err := json.Marshal(meetings)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonStr))
}

/*func IngestSession(year int, meetingKey string, sessionKey string) error {
	if err != nil {
		return err
	}
	dbClient.Database("openf1").Collection("sessions").InsertOne(ctx, Session{
		Key: sessionKey,


	}
	return nil
}*/
