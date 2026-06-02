package main

import (
	"F1Telemetry-new-server/auth"
	"F1Telemetry-new-server/backfill"
	appconfig "F1Telemetry-new-server/config"
	dataingestor "F1Telemetry-new-server/data-ingestor"
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/livetiming"
	"F1Telemetry-new-server/mocksignalr"
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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/websocket"
)

const defaultPort = "8080"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "backfill" {
		backfill.Run()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "backfill-circuits" {
		backfill.RunCircuits()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "mockserver" {
		file := "australian-fp2-1.txt"
		if len(os.Args) > 2 {
			file = os.Args[2]
		}
		addr := "127.0.0.1:9999"
		log.Printf("mock SignalR server streaming %s on %s", file, addr)
		if err := http.ListenAndServe(addr, mocksignalr.Handler(file)); err != nil {
			log.Fatalf("mock server failed: %v", err)
		}
		return
	}

	ctx := context.Background()
	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	dbClient, err := dataingestor.GetDynamoClient(&ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config for S3: %v", err)
	}
	s3Client := s3.NewFromConfig(awsCfg)
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	resolver := graph.NewResolver()

	hub := ws.NewHub(dbClient)
	go hub.Run()
	hub.RegisterOnResolver(resolver)

	valkeyClient := auth.NewValkeyClient()
	if err := valkeyClient.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: Valkey not reachable at startup: %v", err)
	}

	server := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	server.AddTransport(&transport.Websocket{
		InitFunc: auth.WebsocketInitFunc(valkeyClient),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", auth.Middleware(valkeyClient, server))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r, valkeyClient)
	})

	log.Printf("connect to http://localhost:%s/ for lots of fun!\n", port)
	go func() {
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	fmt.Println("connected to websocket")

	fmt.Println("REPLAY_POSITION_Z enabled — replaying archived Position.z data")
	go livetiming.ReplayPositionZ(resolver)

	go func() {
		backoff := 2 * time.Second
		maxBackoff := 60 * time.Second

		for {
			cookies, connObject, err := livetiming.Negotiate(cfg)
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
				connection, resp, err = livetiming.SetWebSocket(cfg, connObject.ConnectionToken, cookies)
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

			err = livetiming.ProcessSessionDataAndInfo(connection, dbClient, s3Client, ctx, resolver, hub, connObject.KeepAliveTimeout)
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
