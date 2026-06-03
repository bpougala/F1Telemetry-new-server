// Package mocksignalr emulates the F1 SignalR negotiate/connect endpoints for local
// development and tests. On connect it streams a recorded session file line-by-line
// over the websocket, exactly as the real service would, then closes the connection.
package mocksignalr

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Handler returns an http.Handler exposing the two SignalR endpoints used by the
// client (Negotiate and SetWebSocket), streaming the given recorded session file.
func Handler(filePath string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/signalr/negotiate", negotiateHandler)
	mux.HandleFunc("/signalr/connect", func(w http.ResponseWriter, r *http.Request) {
		connectHandler(w, r, filePath)
	})
	return mux
}

func negotiateHandler(w http.ResponseWriter, r *http.Request) {
	// SetWebSocket indexes cookies[0] and cookies[1], so emit both AWS ALB cookies.
	http.SetCookie(w, &http.Cookie{Name: "AWSALB", Value: "mock-awsalb"})
	http.SetCookie(w, &http.Cookie{Name: "AWSALBCORS", Value: "mock-awsalbcors"})
	resp := map[string]interface{}{
		"Url":                     "/signalr",
		"ConnectionToken":         "mock-token",
		"ConnectionId":            "mock-connection-id",
		"KeepAliveTimeout":        20.0,
		"DisconnectTimeout":       30.0,
		"ConnectionTimeout":       110.0,
		"TryWebSockets":           true,
		"ProtocolVersion":         "1.5",
		"TransportConnectTimeout": 10.0,
		"LongPollDelay":           0.0,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func connectHandler(w http.ResponseWriter, r *http.Request, filePath string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Drain inbound frames (the client's subscribe message and ping control frames,
	// so gorilla's default handler auto-responds with pongs and keeps the link alive).
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// MOCK_REPLAY_DELAY paces the stream (e.g. "5ms") so broadcasts occur over a
	// sustained window for load testing. Unset/0 streams as fast as possible (default).
	var delay time.Duration
	if d := os.Getenv("MOCK_REPLAY_DELAY"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			delay = parsed
		}
	}

	scanner := bufio.NewScanner(file)
	// Recorded snapshot lines exceed bufio's 64KB default; allow up to 8MB per line.
	scanner.Buffer(make([]byte, 0, 1024*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, line); err != nil {
			return
		}
		if delay > 0 {
			time.Sleep(delay)
		}
	}

	// All lines sent — close gracefully so the client's read loop returns.
	_ = conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "replay complete"),
		time.Now().Add(time.Second),
	)
}
