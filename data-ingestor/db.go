package data_ingestor

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var uri string

func init() {
	uri = os.Getenv("MONGO_CONNECTION_URL")
	if uri == "" {
		fmt.Println("default database used")
		uri = "mongodb://localhost:27017/f1"
	}
}

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
	err = createIndices(client, *ctx)
	return client, err
}

func createIndices(dbClient *mongo.Client, ctx context.Context) error {
	var _, err = dbClient.Database("f1").Collection("drivers").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"sessionkey", 1}},
	})
	if err != nil {
		return err
	}
	_, err = dbClient.Database("f1").Collection("positions").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"sessionkey", 1}},
	})
	if err != nil {
		return err
	}
	_, err = dbClient.Database("f1").Collection("racecontrol").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"sessionkey", 1}},
	})
	if err != nil {
		return err
	}
	_, err = dbClient.Database("f1").Collection("stints").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"sessionkey", 1}},
	})
	if err != nil {
		return err
	}
	_, err = dbClient.Database("f1").Collection("sectors").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"sessionkey", 1}},
	})
	if err != nil {
		return err
	}
	return err
}
