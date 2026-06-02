package livetiming

import (
	"F1Telemetry-new-server/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// connectWithRetry mirrors the retry logic in server.go
func connectWithRetry(t *testing.T) *websocket.Conn {
	t.Helper()

	cfg := &config.Config{
		NegotiateBaseURL: "https://livetiming.formula1.com",
		ConnectBaseURL:   "wss://livetiming.formula1.com",
	}
	cookies, connObject, err := Negotiate(cfg)
	if err != nil {
		t.Fatalf("Negotiate failed: %v", err)
	}
	t.Logf("Negotiated connection token: %s", connObject.ConnectionToken[:20]+"...")

	var connection *websocket.Conn
	for attempt := 1; attempt <= 10; attempt++ {
		connection, _, err = SetWebSocket(cfg, connObject.ConnectionToken, cookies)
		if err == nil {
			t.Logf("WebSocket connected on attempt %d", attempt)
			return connection
		}
		t.Logf("SetWebSocket attempt %d failed: %v", attempt, err)
		time.Sleep(time.Second)
	}
	t.Fatalf("SetWebSocket failed after 10 attempts: %v", err)
	return nil
}

func TestPositionZFromLiveTiming(t *testing.T) {
	connection := connectWithRetry(t)
	defer connection.Close()

	subscribeMessage := CreateOriginalSessionMessage()
	err := connection.WriteJSON(subscribeMessage)
	if err != nil {
		t.Fatalf("Failed to send subscribe message: %v", err)
	}

	topicCounts := make(map[string]int)
	positionZReceived := false
	timeout := time.After(30 * time.Second)

	for i := 0; i < 50; i++ {
		select {
		case <-timeout:
			t.Logf("Timeout after 30s. Topic counts: %v", topicCounts)
			if !positionZReceived {
				t.Fatal("Never received a Position.z message within 30 seconds")
			}
			return
		default:
		}

		connection.SetReadDeadline(time.Now().Add(10 * time.Second))
		_, message, err := connection.ReadMessage()
		if err != nil {
			t.Fatalf("ReadMessage failed on message %d: %v", i, err)
		}

		var raw map[string]json.RawMessage
		if err := json.Unmarshal(message, &raw); err != nil {
			continue
		}

		if _, hasR := raw["R"]; hasR {
			var initialData struct {
				R map[string]json.RawMessage `json:"R"`
			}
			json.Unmarshal(message, &initialData)
			for key, val := range initialData.R {
				size := len(val)
				if size > 100 {
					t.Logf("Initial snapshot key: %s (size: %d bytes)", key, size)
				}
				if key == "Position.z" {
					positionZReceived = true
					t.Logf("Position.z in initial snapshot: %d bytes", size)
				}
			}
			continue
		}

		if _, hasM := raw["M"]; hasM {
			var updateData UpdateData
			json.Unmarshal(message, &updateData)
			for _, msg := range updateData.M {
				if len(msg.A) < 1 {
					continue
				}
				topic, ok := msg.A[0].(string)
				if !ok {
					continue
				}
				topicCounts[topic]++
				if topic == "Position.z" {
					positionZReceived = true
					if len(msg.A) >= 2 {
						switch v := msg.A[1].(type) {
						case string:
							t.Logf("Position.z update: string payload (%d chars)", len(v))
						case map[string]interface{}:
							t.Logf("Position.z update: map payload with keys %v", mapKeys(v))
						default:
							t.Logf("Position.z update: unexpected type %T", v)
						}
					}
				}
			}
		}
	}

	t.Logf("Topic counts after %d messages: %v", 50, topicCounts)
	if !positionZReceived {
		t.Fatal("Never received a Position.z message in 50 messages")
	}
	t.Log("Position.z data IS being sent by the SignalR endpoint")
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestPositionZExtraction(t *testing.T) {
	connection := connectWithRetry(t)
	defer connection.Close()

	subscribeMessage := CreateOriginalSessionMessage()
	err := connection.WriteJSON(subscribeMessage)
	if err != nil {
		t.Fatalf("Failed to send subscribe message: %v", err)
	}

	timeout := time.After(30 * time.Second)

	for i := 0; i < 50; i++ {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for messages")
		default:
		}

		connection.SetReadDeadline(time.Now().Add(10 * time.Second))
		_, message, err := connection.ReadMessage()
		if err != nil {
			t.Fatalf("ReadMessage failed: %v", err)
		}

		compressed, err := ExtractPositionZCompressed(message)
		if err == nil && compressed != "" {
			t.Logf("ExtractPositionZCompressed succeeded: %d chars", len(compressed))
			fmt.Printf("First 50 chars: %s\n", compressed[:min(50, len(compressed))])
			return
		}
	}
	t.Fatal("Could not extract Position.z from any of 50 messages")
}

// TestPositionZArchive fetches archived Position.z data from F1's static API
// and verifies the full decompression pipeline works.
func TestPositionZArchive(t *testing.T) {
	archiveURL := "https://livetiming.formula1.com/static/2026/2026-05-24_Canadian_Grand_Prix/2026-05-24_Race/Position.z.jsonStream"

	req, err := http.NewRequest("GET", archiveURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "BestHTTP")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to fetch archive: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Archive returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	// Strip UTF-8 BOM
	raw := strings.TrimPrefix(string(body), "\xef\xbb\xbf")
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	t.Logf("Archive has %d lines", len(lines))

	if len(lines) == 0 {
		t.Fatal("Archive is empty")
	}

	decompressedCount := 0
	totalEntries := 0

	// Skip deep into the race to get actual coordinates
	startIdx := min(len(lines)/2, len(lines)-10)
	for i, line := range lines[startIdx : startIdx+min(10, len(lines)-startIdx)] {
		line = strings.TrimRight(line, "\r")

		// Each line: <timestamp>"<base64data>"
		quoteIdx := strings.Index(line, "\"")
		if quoteIdx == -1 {
			t.Logf("Line %d: no quote found, skipping", i)
			continue
		}

		compressed := strings.Trim(line[quoteIdx:], "\"")

		posRoot, err := DecompressPositionData(compressed)
		if err != nil {
			t.Logf("Line %d: decompress failed: %v", i, err)
			continue
		}

		decompressedCount++
		for _, entry := range posRoot.Position {
			totalEntries += len(entry.Entries)
			if i == 0 {
				t.Logf("Timestamp: %s, %d cars", entry.Timestamp.Format(time.RFC3339Nano), len(entry.Entries))
				for carNum, pos := range entry.Entries {
					t.Logf("  Car %s: X=%d Y=%d Z=%d Status=%s", carNum, pos.X, pos.Y, pos.Z, pos.Status)
				}
			}
		}
	}

	t.Logf("Successfully decompressed %d/10 lines, %d total car position entries", decompressedCount, totalEntries)

	if decompressedCount == 0 {
		t.Fatal("Could not decompress any Position.z data")
	}
}
