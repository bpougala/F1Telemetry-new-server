package main

import (
	dataingestor "F1Telemetry-new-server/data-ingestor"
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/livetiming"
	"bufio"
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const defaultPort = "8080"

func readText(dbClient *mongo.Client, ctx context.Context) error {
	file, err := os.Open("race-qatar.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	//var intervalsForDB []interface{}
	var stintsForDB []interface{}
	for scanner.Scan() {
		line := scanner.Text()
		stint, err := livetiming.Scanner([]byte(line))
		if err == nil {
			var stintDB livetiming.StintDB
			stintDB.SessionKey = 9655
			stintDB.DriverNumber = stint.DriverNumber
			stintDB.Compound = stint.Compound
			if stint.New == "0" {
				stintDB.Is_new = false
			} else {
				stintDB.Is_new = true
			}
			if stint.TyresNotChanged == "0" {
				stintDB.Are_tyres_not_changed = false
			} else {
				stintDB.Are_tyres_not_changed = true
			}
			stintDB.TotalLaps = stint.TotalLaps
			stintDB.StartLaps = stint.StartLaps
			stintDB.Timestamp = stint.Timestamp
			fmt.Printf("Stint: %+v\n", stintDB)
			stintsForDB = append(stintsForDB, stintDB)
		}
		/*var interval livetiming.WebSocketMessage
		if err := json.Unmarshal([]byte(line), &interval); err != nil {
			fmt.Printf("Skipping line: %s\nError: %v\n", line, err)
			continue
		}
		if interval.Control == "" {
			continue
		}
		for _, interval := range interval.DataBlock {
			if len(interval.DataFeed) < 3 {
				continue
			}
			var intervalData livetiming.IntervalData
			jsonData, err := json.Marshal(interval.DataFeed[1])
			if err != nil {
				fmt.Printf("Failed to marshal interval data: %v\n", err)
				continue
			}
			err = json.Unmarshal(jsonData, &intervalData)
			if err != nil {
				fmt.Printf("Failed to unmarshal interval data: %v\n", err)
				continue
			}
			timeStamp := interval.DataFeed[2]
			data := intervalData.Lines
			var intervalsDB livetiming.IntervalsDB
			intervalsDB.SessionKey = 9655
			for key, value := range data {
				if value.IntervalToPositionAhead.Value != "" {
					driverNumber, _ := strconv.Atoi(key)
					intervalsDB.DriverNumber = driverNumber
					intervalsDB.Interval = value.IntervalToPositionAhead.Value
					intervalsDB.GapToLeader = value.GapToLeader
					intervalsDB.Timestamp = timeStamp.(string)
				}
			}
			if intervalsDB.Interval != "" {
				intervalsForDB = append(intervalsForDB, intervalsDB)
			}
		}*/
	}
	/*_, err = dbClient.Database("f1").Collection("intervals").InsertMany(ctx, intervalsForDB)
	if err != nil {
		return err
	}
	if err := scanner.Err(); err != nil {
		return err
	}*/
	_, err = dbClient.Database("f1").Collection("stints").InsertMany(ctx, stintsForDB)
	if err != nil {
		fmt.Println("oh oh")
		return err
	}
	return nil
}

func main() {
	ctx := context.Background()
	dbClient, err := dataingestor.GetMongoClient(&ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	err = readText(dbClient, ctx)
	if err != nil {
		fmt.Printf("Failed to read file: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	ctx := context.Background()
	dbClient, err := dataingestor.GetMongoClient(&ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	cookies, connObject, err := livetiming.Negotiate()
	if err != nil {
		panic(err)
	}
	retries := 10
	connection, resp, err := livetiming.SetWebSocket(connObject.ConnectionToken, cookies)
	for retries > 0 && err != nil {
		connection, resp, err = livetiming.SetWebSocket(connObject.ConnectionToken, cookies)
		retries--
	}
	if err != nil {
		if resp != nil {
			fmt.Printf("handshake failed with status %d", resp.StatusCode)
		}
		panic(err)
	}
	defer connection.Close()

	sessionKeyChan := make(chan int)
	go livetiming.ProcessSessionDataAndInfo(connection, dbClient, ctx, sessionKeyChan)
	sessionKey := <-sessionKeyChan
	go livetiming.ProcessTimingData(connection, dbClient, ctx, sessionKey)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Shutting down")
}
