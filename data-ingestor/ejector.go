package data_ingestor

import (
	"F1Telemetry-new-server/graph/model"
	"F1Telemetry-new-server/livetiming"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func FetchMeetings(dbClient *mongo.Client, ctx context.Context) ([]model.Meeting, error) {
	cursor, err := dbClient.Database("f1").Collection("meetingdata").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var rawMeetings []livetiming.MeetingDataDB
	var meetings []model.Meeting
	err = cursor.All(ctx, &rawMeetings)
	if err != nil {
		return nil, err
	}
	for _, rawMeeting := range rawMeetings {
		meeting := model.Meeting{
			MeetingKey:          rawMeeting.Key,
			MeetingName:         rawMeeting.Name,
			MeetingOfficialName: rawMeeting.OfficialName,
			Location:            rawMeeting.Location,
			CountryName:         rawMeeting.CountryName,
			CountryCode:         rawMeeting.CountryCode,
			CircuitShortName:    rawMeeting.Circuit,
		}
		meetings = append(meetings, meeting)
	}
	return meetings, nil
}

func FetchSessions(dbClient *mongo.Client, ctx context.Context, meetingKey int) ([]model.Session, error) {
	cursor, err := dbClient.Database("f1").Collection("sessioninfo").Find(ctx, bson.D{{"meetingkey", meetingKey}})
	if err != nil {
		return nil, err
	}
	var rawSessions []livetiming.SessionInfoDB
	var sessions []model.Session
	err = cursor.All(ctx, &rawSessions)
	if err != nil {
		return nil, err
	}
	for _, rawSession := range rawSessions {
		session := model.Session{
			SessionKey:  rawSession.Key,
			MeetingKey:  rawSession.MeetingKey,
			SessionName: rawSession.Name,
			DateStart:   rawSession.StartDate,
			DateEnd:     rawSession.EndDate,
			SessionType: rawSession.Type,
			GmtOffset:   rawSession.GmtOffset,
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func FetchDrivers(dbClient *mongo.Client, ctx context.Context, sessionKey int) ([]model.Driver, error) {
	cursor, err := dbClient.Database("f1").Collection("drivers").Find(ctx, bson.D{{"sessionkey", sessionKey}})
	if err != nil {
		return nil, err
	}
	var rawDrivers []livetiming.Driver
	var drivers []model.Driver
	err = cursor.All(ctx, &rawDrivers)
	if err != nil {
		return nil, err
	}
	for _, rawDriver := range rawDrivers {
		driver := model.Driver{
			SessionKey:   rawDriver.SessionKey,
			RacingNumber: rawDriver.RacingNumber,
			FirstName:    rawDriver.FirstName,
			LastName:     rawDriver.LastName,
			TeamColour:   &rawDriver.TeamColour,
			TeamName:     &rawDriver.TeamName,
			Abbreviation: &rawDriver.Abbreviation,
			FullName:     rawDriver.FullName,
		}
		drivers = append(drivers, driver)
	}
	return drivers, nil
}

func FetchPositions(dbClient *mongo.Client, ctx context.Context, sessionKey int) ([]model.Position, error) {
	cursor, err := dbClient.Database("f1").Collection("positions").Find(ctx, bson.D{{"sessionkey", sessionKey}})
	if err != nil {
		return nil, err
	}
	var positions []model.Position
	err = cursor.All(ctx, &positions)
	if err != nil {
		return nil, err
	}
	return positions, nil
}

func FetchCarData(dbClient *mongo.Client, ctx context.Context, sessionKey int, driverNumber int) ([]model.CarData, error) {
	cursor, err := dbClient.Database("f1").Collection("cardata").Find(ctx, bson.D{{"sessionkey", sessionKey}, {"drivernumber", driverNumber}})
	if err != nil {
		return nil, err
	}
	var carData []model.CarData
	err = cursor.All(ctx, &carData)
	if err != nil {
		return nil, err
	}
	return carData, nil
}

func FetchLaps(dbClient *mongo.Client, ctx context.Context, sessionKey int) ([]model.Lap, error) {
	cursor, err := dbClient.Database("f1").Collection("laps").Find(ctx, bson.D{{"sessionkey", sessionKey}})
	if err != nil {
		return nil, err
	}
	var laps []model.Lap
	err = cursor.All(ctx, &laps)
	if err != nil {
		return nil, err
	}
	return laps, nil
}

func FetchIntervals(dbClient *mongo.Client, ctx context.Context, sessionKey int) ([]model.Interval, error) {
	cursor, err := dbClient.Database("f1").Collection("intervals").Find(ctx, bson.D{{"sessionkey", sessionKey}})
	if err != nil {
		return nil, err
	}
	var intervals []model.Interval
	err = cursor.All(ctx, &intervals)
	if err != nil {
		fmt.Println("error")
		return nil, err
	}
	return intervals, nil
}
