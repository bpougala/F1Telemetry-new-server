package main

import (
	dataingestor "F1Telemetry-new-server/data-ingestor"
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/livetiming"
	"F1Telemetry-new-server/ws"
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultPort = "8080"

func main() {
	ctx := context.Background()
	dbClient, err := dataingestor.GetDynamoClient(&ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	resolver := graph.NewResolver()

	hub := ws.NewHub(dbClient)
	go hub.Run()
	hub.RegisterOnResolver(resolver)

	server := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	server.AddTransport(&transport.Websocket{})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", server)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	log.Printf("connect to http://localhost:%s/ for lost of fun!\n", port)
	go func() {
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	fmt.Println("connected to websocket")

	go func() {
		for {
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
			err = livetiming.ProcessSessionDataAndInfo(connection, dbClient, ctx, resolver, hub)
			if err != nil {
				fmt.Println("error occurred, retrying:", err)
				time.Sleep(2 * time.Second)
				continue
			}
			break
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Shutting down")
}
