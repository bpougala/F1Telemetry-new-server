package data_ingestor

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

var uri = os.Getenv("MONGO_CONNECTION_URL")

func GetMongoClient(ctx *context.Context) (*mongo.Client, error) {
	fmt.Println(uri)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetAuth(options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
	}).SetServerAPIOptions(serverAPI)
	fmt.Println("Connecting to MongoDB")
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
