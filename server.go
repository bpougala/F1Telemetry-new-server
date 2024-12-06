package main

import (
	dataingestor "F1Telemetry-new-server/data-ingestor"
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/livetiming"
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const defaultPort = "8080"

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

/*_, err = dbClient.Database("f1").Collection("intervals").InsertMany(ctx, intervalsForDB)
if err != nil {
	return err
}
if err := scanner.Err(); err != nil {
	return err
}*/

func main() {
	ctx := context.Background()
	dbClient, err := dataingestor.GetMongoClient(&ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	go func() {
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	/*if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}*/
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

	fmt.Println("connected to websocket")
	sessionKeyChan := make(chan int)
	go livetiming.ProcessSessionDataAndInfo(connection, dbClient, ctx, sessionKeyChan)
	sessionKey := <-sessionKeyChan
	go livetiming.ProcessDriverData(connection, dbClient, ctx, sessionKey)
	//go livetiming.ProcessTimingData(connection, dbClient, ctx, sessionKey)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Shutting down")
}
