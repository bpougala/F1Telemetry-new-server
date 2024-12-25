package data_ingestor

import (
	"F1Telemetry-new-server/data-ingestor/collections"
	"F1Telemetry-new-server/graph/model"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func FetchMeetings(dbClient *mongo.Client, ctx context.Context) ([]model.Meeting, error) {
	cursor, err := dbClient.Database("f1").Collection("meetingdata").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var rawMeetings []collections.MeetingDataDB
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
	cursor, err := dbClient.Database("f1").Collection("sessioninfo").Find(ctx, bson.D{{Key: "meetingkey", Value: meetingKey}})
	if err != nil {
		return nil, err
	}
	var rawSessions []collections.SessionInfoDB
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
			Status:      rawSession.ArchiveStatus,
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
	var rawDrivers []collections.DriverDB
	var drivers []model.Driver
	err = cursor.All(ctx, &rawDrivers)
	if err != nil {
		return nil, err
	}
	for _, rawDriver := range rawDrivers {
		driver := model.Driver{
			SessionKey:    rawDriver.SessionKey,
			RacingNumber:  rawDriver.RacingNumber,
			FirstName:     rawDriver.FirstName,
			LastName:      rawDriver.LastName,
			TeamColour:    &rawDriver.TeamColour,
			TeamName:      &rawDriver.TeamName,
			Abbreviation:  &rawDriver.Abbreviation,
			FullName:      rawDriver.FullName,
			HeadshotURL:   rawDriver.HeadshotUrl,
			BroadcastName: rawDriver.BroadcastName,
			CountryCode:   rawDriver.CountryCode,
		}
		drivers = append(drivers, driver)
	}
	return drivers, nil
}

func FetchStints(dbClient *mongo.Client, ctx context.Context, sessionKey int) ([]model.Stint, error) {
	cursor, err := dbClient.Database("f1").Collection("stints").Find(ctx, bson.D{{"sessionkey", sessionKey}})
	if err != nil {
		return nil, err
	}
	var rawStints []collections.StintDB
	var stints []model.Stint
	err = cursor.All(ctx, &rawStints)
	if err != nil {
		return nil, err
	}
	for _, rawStint := range rawStints {
		stint := model.Stint{
			RacingNumber:    rawStint.RacingNumber,
			LapFlags:        rawStint.LapFlags,
			Compound:        rawStint.Compound,
			New:             rawStint.New,
			TyresNotChanged: rawStint.TyresNotChanged,
			TotalLaps:       rawStint.TotalLaps,
			StartLaps:       rawStint.StartLaps,
			Timestamp:       rawStint.Timestamp,
		}
		stints = append(stints, stint)
	}
	return stints, nil
}

func FetchPositions(dbClient *mongo.Client, ctx context.Context, sessionKey int) ([]model.Position, error) {
	cursor, err := dbClient.Database("f1").Collection("positions").Find(ctx, bson.D{{"sessionkey", sessionKey}})
	if err != nil {
		return nil, err
	}
	var rawPositions []collections.PositionsDB
	var positions []model.Position
	err = cursor.All(ctx, &rawPositions)
	if err != nil {
		return nil, err
	}
	for _, rawPosition := range rawPositions {
		position := model.Position{
			SessionKey:   rawPosition.SessionKey,
			Position:     rawPosition.Position,
			RacingNumber: rawPosition.RacingNumber,
		}
		positions = append(positions, position)
	}
	return positions, nil
}

func FetchTimings(dbClient *mongo.Client, ctx context.Context, sessionKey int) ([]model.Timing, error) {
	cursor, err := dbClient.Database("f1").Collection("timings").Find(ctx, bson.D{{"sessionkey", sessionKey}})
	if err != nil {
		return nil, err
	}
	var rawTimings []collections.Timing
	var timings []model.Timing
	err = cursor.All(ctx, &rawTimings)
	if err != nil {
		return nil, err
	}
	for _, rawTiming := range rawTimings {
		timing := model.Timing{
			SessionKey:              rawTiming.SessionKey,
			Stopped:                 rawTiming.Stopped,
			Status:                  rawTiming.Status,
			Sectors:                 rawTiming.Sectors,
			Retired:                 rawTiming.Retired,
			RacingNumber:            rawTiming.RacingNumber,
			Position:                rawTiming.Position,
			PitOut:                  rawTiming.PitOut,
			NumberOfLaps:            rawTiming.NumberOfLaps,
			LastLapTime:             rawTiming.LastLapTime,
			IntervalToPositionAhead: rawTiming.IntervalToPositionAhead,
			InPit:                   rawTiming.InPit,
			GapToLeader:             rawTiming.GapToLeader,
			BestLapTime:             rawTiming.BestLapTime,
		}
		timings = append(timings, timing)
	}
	return timings, nil
}

func FetchRaceControl(dbClient *mongo.Client, ctx context.Context, sessionKey int) ([]model.RaceControl, error) {
	cursor, err := dbClient.Database("f1").Collection("racecontrol").Find(ctx, bson.D{{"sessionkey", sessionKey}})
	if err != nil {
		return nil, err
	}
	var rawRaceControl []collections.RaceControl
	err = cursor.All(ctx, &rawRaceControl)
	if err != nil {
		return nil, err
	}
	var raceControlMessages []model.RaceControl
	for _, rawRaceControl := range rawRaceControl {
		raceControl := model.RaceControl{
			Date:     rawRaceControl.Utc,
			Category: &rawRaceControl.Category,
			Flag:     &rawRaceControl.Flag,
			Message:  rawRaceControl.Message,
			Scope:    &rawRaceControl.Scope,
		}
		raceControlMessages = append(raceControlMessages, raceControl)
	}
	return raceControlMessages, nil
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
