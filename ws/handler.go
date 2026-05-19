package ws

import (
	"F1Telemetry-new-server/auth"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, valkeyClient *redis.Client) {
	_, err := auth.ValidateToken(r.Context(), valkeyClient, r)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("ws upgrade error:", err)
		return
	}
	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		topics: make(map[string]bool),
	}
	hub.register <- client
	go client.writePump()
	go client.readPump()
}
