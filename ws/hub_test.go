package ws

import (
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/graph/model"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func setupTestHub() *Hub {
	hub := NewHub(nil)
	go hub.Run()
	return hub
}

func connectTestClient(t *testing.T, hub *Hub) (*websocket.Conn, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	}))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not connect to test server: %v", err)
	}
	return conn, func() {
		conn.Close()
		server.Close()
	}
}

func readMessage(t *testing.T, conn *websocket.Conn) OutboundMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message: %v", err)
	}
	var msg OutboundMessage
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		t.Fatalf("Could not unmarshal message: %v", err)
	}
	return msg
}

func TestSubscribeSendsConfirmation(t *testing.T) {
	hub := setupTestHub()
	conn, cleanup := connectTestClient(t, hub)
	defer cleanup()

	subscribeMsg := InboundMessage{Type: "subscribe", Topics: []string{"positions", "lapTimes"}}
	if err := conn.WriteJSON(subscribeMsg); err != nil {
		t.Fatalf("Could not send subscribe: %v", err)
	}

	msg := readMessage(t, conn)
	if msg.Type != "subscribed" {
		t.Fatalf("Expected type 'subscribed', got '%s'", msg.Type)
	}
	if len(msg.Topics) != 2 {
		t.Fatalf("Expected 2 topics, got %d", len(msg.Topics))
	}
}

func TestSubscribeToUnknownTopicReturnsError(t *testing.T) {
	hub := setupTestHub()
	conn, cleanup := connectTestClient(t, hub)
	defer cleanup()

	subscribeMsg := InboundMessage{Type: "subscribe", Topics: []string{"bogus"}}
	if err := conn.WriteJSON(subscribeMsg); err != nil {
		t.Fatalf("Could not send subscribe: %v", err)
	}

	msg := readMessage(t, conn)
	if msg.Type != "error" {
		t.Fatalf("Expected type 'error', got '%s'", msg.Type)
	}
	if msg.Message != "unknown topic: bogus" {
		t.Fatalf("Expected error about unknown topic, got '%s'", msg.Message)
	}
}

func TestLiveDataBroadcastToSubscribers(t *testing.T) {
	hub := setupTestHub()
	resolver := graph.NewResolver()
	hub.RegisterOnResolver(resolver)

	conn, cleanup := connectTestClient(t, hub)
	defer cleanup()

	// Subscribe to positions
	subscribeMsg := InboundMessage{Type: "subscribe", Topics: []string{"positions"}}
	if err := conn.WriteJSON(subscribeMsg); err != nil {
		t.Fatalf("Could not send subscribe: %v", err)
	}

	// Read subscription confirmation
	msg := readMessage(t, conn)
	if msg.Type != "subscribed" {
		t.Fatalf("Expected 'subscribed', got '%s'", msg.Type)
	}

	// Give the hub time to process subscription
	time.Sleep(50 * time.Millisecond)

	// Send live position data through resolver
	pos := 1
	positions := []*model.Position{
		{RacingNumber: 44, Position: &pos, SessionKey: 9999},
	}
	resolver.NotifyPositionSubscribers(positions)

	// Read the data message
	dataMsg := readMessage(t, conn)
	if dataMsg.Type != "data" {
		t.Fatalf("Expected type 'data', got '%s'", dataMsg.Type)
	}
	if dataMsg.Topic != "positions" {
		t.Fatalf("Expected topic 'positions', got '%s'", dataMsg.Topic)
	}
}

func TestUnsubscribeStopsReceivingData(t *testing.T) {
	hub := setupTestHub()
	resolver := graph.NewResolver()
	hub.RegisterOnResolver(resolver)

	conn, cleanup := connectTestClient(t, hub)
	defer cleanup()

	// Subscribe
	if err := conn.WriteJSON(InboundMessage{Type: "subscribe", Topics: []string{"positions"}}); err != nil {
		t.Fatalf("Could not send subscribe: %v", err)
	}
	readMessage(t, conn) // subscribed confirmation

	time.Sleep(50 * time.Millisecond)

	// Unsubscribe
	if err := conn.WriteJSON(InboundMessage{Type: "unsubscribe", Topics: []string{"positions"}}); err != nil {
		t.Fatalf("Could not send unsubscribe: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Send data — should NOT be received
	pos := 1
	resolver.NotifyPositionSubscribers([]*model.Position{
		{RacingNumber: 44, Position: &pos},
	})

	// Try to read — should timeout
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _, err := conn.ReadMessage()
	if err == nil {
		t.Fatalf("Expected no message after unsubscribe, but got one")
	}
}

func TestClientDisconnectCleansUp(t *testing.T) {
	hub := setupTestHub()
	conn, cleanup := connectTestClient(t, hub)

	// Subscribe
	if err := conn.WriteJSON(InboundMessage{Type: "subscribe", Topics: []string{"positions"}}); err != nil {
		t.Fatalf("Could not send subscribe: %v", err)
	}
	readMessage(t, conn)

	time.Sleep(50 * time.Millisecond)

	// Verify client is registered
	hub.mu.RLock()
	clientCount := len(hub.clients)
	subscriberCount := len(hub.topics["positions"])
	hub.mu.RUnlock()

	if clientCount != 1 {
		t.Fatalf("Expected 1 client, got %d", clientCount)
	}
	if subscriberCount != 1 {
		t.Fatalf("Expected 1 subscriber to positions, got %d", subscriberCount)
	}

	// Disconnect
	cleanup()
	time.Sleep(200 * time.Millisecond)

	// Verify cleanup
	hub.mu.RLock()
	clientCount = len(hub.clients)
	subscriberCount = len(hub.topics["positions"])
	hub.mu.RUnlock()

	if clientCount != 0 {
		t.Fatalf("Expected 0 clients after disconnect, got %d", clientCount)
	}
	if subscriberCount != 0 {
		t.Fatalf("Expected 0 subscribers after disconnect, got %d", subscriberCount)
	}
}

func TestSessionChangedBroadcast(t *testing.T) {
	hub := setupTestHub()
	conn, cleanup := connectTestClient(t, hub)
	defer cleanup()

	time.Sleep(50 * time.Millisecond)

	hub.SetSessionKey(9999)

	msg := readMessage(t, conn)
	if msg.Type != "session_changed" {
		t.Fatalf("Expected 'session_changed', got '%s'", msg.Type)
	}
	if msg.SessionKey != 9999 {
		t.Fatalf("Expected sessionKey 9999, got %d", msg.SessionKey)
	}
}

func TestInvalidMessageReturnsError(t *testing.T) {
	hub := setupTestHub()
	conn, cleanup := connectTestClient(t, hub)
	defer cleanup()

	// Send invalid JSON
	if err := conn.WriteMessage(websocket.TextMessage, []byte("not json")); err != nil {
		t.Fatalf("Could not send message: %v", err)
	}

	msg := readMessage(t, conn)
	if msg.Type != "error" {
		t.Fatalf("Expected 'error', got '%s'", msg.Type)
	}
}

func TestUnknownMessageTypeReturnsError(t *testing.T) {
	hub := setupTestHub()
	conn, cleanup := connectTestClient(t, hub)
	defer cleanup()

	if err := conn.WriteJSON(InboundMessage{Type: "foobar"}); err != nil {
		t.Fatalf("Could not send message: %v", err)
	}

	msg := readMessage(t, conn)
	if msg.Type != "error" {
		t.Fatalf("Expected 'error', got '%s'", msg.Type)
	}
	if msg.Message != "unknown message type: foobar" {
		t.Fatalf("Unexpected error message: %s", msg.Message)
	}
}
