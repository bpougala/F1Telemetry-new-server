package data_ingestor

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var uri = os.Getenv("MONGO_CONNECTION_URL")

func GetMongoClient(ctx *context.Context) (*mongo.Client, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(*ctx, opts)
	if err != nil {
		return nil, err
	}
	var result bson.M
	err = client.Database("admin").RunCommand(*ctx, bson.D{{"ping", 1}}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return client, nil
}
