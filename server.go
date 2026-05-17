package main

import (
	"F1Telemetry-new-server/backfill"
	dataingestor "F1Telemetry-new-server/data-ingestor"
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/livetiming"
	"F1Telemetry-new-server/ws"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
)

const defaultPort = "8080"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "backfill" {
		backfill.Run()
		return
	}

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
		backoff := 2 * time.Second
		maxBackoff := 60 * time.Second

		for {
			cookies, connObject, err := livetiming.Negotiate()
			if err != nil {
				fmt.Println("negotiate failed, retrying:", err)
				time.Sleep(backoff)
				backoff = min(backoff*2, maxBackoff)
				continue
			}

			var connection *websocket.Conn
			retries := 10
			for retries > 0 {
				var resp *http.Response
				connection, resp, err = livetiming.SetWebSocket(connObject.ConnectionToken, cookies)
				if err == nil {
					break
				}
				if resp != nil {
					fmt.Printf("handshake failed with status %d\n", resp.StatusCode)
				}
				retries--
				time.Sleep(time.Second)
			}

			if err != nil {
				fmt.Println("websocket setup failed:", err)
				time.Sleep(backoff)
				backoff = min(backoff*2, maxBackoff)
				continue
			}

			// Reset backoff on successful connection
			backoff = 2 * time.Second
			fmt.Println("connected to F1 live timing")

			err = livetiming.ProcessSessionDataAndInfo(connection, dbClient, ctx, resolver, hub)
			connection.Close()

			if err != nil {
				fmt.Println("connection dropped, reconnecting:", err)
				time.Sleep(backoff)
				continue
			}
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Shutting down")
}
