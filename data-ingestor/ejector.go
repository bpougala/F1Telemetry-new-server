package data_ingestor

import (
	"F1Telemetry-new-server/graph/model"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func FetchMeetings(dbClient *mongo.Client, ctx context.Context) ([]model.Meeting, error) {
	cursor, err := dbClient.Database("f1").Collection("meetings").Find(ctx, bson.D{})
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
