package ws

import (
	dataingestor "F1Telemetry-new-server/data-ingestor"
	"context"
	"encoding/json"
	"fmt"
)

func (h *Hub) sendRecap(client *Client, topic string) {
	sessionKey := h.GetSessionKey()
	if sessionKey == 0 {
		return
	}

	ctx := context.Background()
	var payload interface{}
	var err error

	switch topic {
	case "positions":
		payload, err = dataingestor.FetchPositions(h.dbClient, ctx, sessionKey)
	case "lapTimes":
		payload, err = dataingestor.FetchTimings(h.dbClient, ctx, sessionKey)
	case "sectors":
		payload, err = dataingestor.FetchSectors(h.dbClient, ctx, sessionKey)
	case "stints":
		payload, err = dataingestor.FetchStints(h.dbClient, ctx, sessionKey)
	case "trackStatus":
		payload, err = dataingestor.FetchTrackStatus(h.dbClient, ctx, sessionKey)
	case "raceControl":
		payload, err = dataingestor.FetchRaceControl(h.dbClient, ctx, sessionKey)
	case "carData":
		return // streaming-only, no historical recap
	default:
		return
	}

	if err != nil {
		fmt.Printf("error fetching recap for %s: %v\n", topic, err)
		return
	}
	if payload == nil {
		return
	}

	msg := OutboundMessage{
		Type:    "recap",
		Topic:   topic,
		Payload: payload,
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("error marshaling recap for %s: %v\n", topic, err)
		return
	}
	select {
	case client.send <- jsonBytes:
	default:
		// client buffer full, skip recap
	}
}
