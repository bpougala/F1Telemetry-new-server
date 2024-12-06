package data_ingestor

import (
	"F1Telemetry-new-server/graph/model"
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
	var meetings []model.Meeting
	err = cursor.All(ctx, &meetings)
	if err != nil {
		return nil, err
	}
	return meetings, nil
}

func FetchSessions(dbClient *mongo.Client, ctx context.Context, meetingKey int) ([]model.Session, error) {
	cursor, err := dbClient.Database("f1").Collection("sessions").Find(ctx, bson.D{{"meetingkey", meetingKey}})
	if err != nil {
		return nil, err
	}
	var sessions []model.Session
	err = cursor.All(ctx, &sessions)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func FetchDrivers(dbClient *mongo.Client, ctx context.Context, sessionKey int) ([]model.Driver, error) {
	cursor, err := dbClient.Database("f1").Collection("drivers").Find(ctx, bson.D{{"sessionkey", sessionKey}})
	if err != nil {
		return nil, err
	}
	var drivers []model.Driver
	err = cursor.All(ctx, &drivers)
	if err != nil {
		return nil, err
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
