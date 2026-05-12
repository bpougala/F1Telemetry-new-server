package ws

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 4096
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	topics map[string]bool
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				fmt.Println("ws read error:", err)
			}
			return
		}
		var msg InboundMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			errMsg := OutboundMessage{Type: "error", Message: "invalid JSON"}
			jsonBytes, _ := json.Marshal(errMsg)
			select {
			case c.send <- jsonBytes:
			default:
			}
			continue
		}
		switch msg.Type {
		case "subscribe":
			c.hub.Subscribe(c, msg.Topics)
		case "unsubscribe":
			c.hub.Unsubscribe(c, msg.Topics)
		case "pong":
			// Client-level pong, no action needed (WebSocket-level pong handled above)
		default:
			errMsg := OutboundMessage{Type: "error", Message: fmt.Sprintf("unknown message type: %s", msg.Type)}
			jsonBytes, _ := json.Marshal(errMsg)
			select {
			case c.send <- jsonBytes:
			default:
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
