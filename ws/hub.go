package ws

import (
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/graph/model"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Hub struct {
	clients    map[*Client]bool
	topics     map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex

	sessionKey int
	sessionMu  sync.RWMutex

	segmentCounts   [3]int
	segmentCountsMu sync.RWMutex

	dbClient *dynamodb.Client
}

func NewHub(dbClient *dynamodb.Client) *Hub {
	topics := make(map[string]map[*Client]bool)
	for topic := range ValidTopics {
		topics[topic] = make(map[*Client]bool)
	}
	return &Hub{
		clients:    make(map[*Client]bool),
		topics:     topics,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		dbClient:   dbClient,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				for _, subscribers := range h.topics {
					delete(subscribers, client)
				}
				close(client.send)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) RegisterOnResolver(resolver *graph.Resolver) {
	lapTimesCh := make(chan []*model.LapTime, 16)
	resolver.RegisterLapTimeObserver("ws-hub", lapTimesCh)
	go h.forwardLapTimes(lapTimesCh)

	positionsCh := make(chan []*model.Position, 16)
	resolver.RegisterPositionObserver("ws-hub", positionsCh)
	go h.forwardPositions(positionsCh)

	carDataCh := make(chan *model.CarData, 16)
	resolver.RegisterCarDataObserver("ws-hub", carDataCh)
	go h.forwardCarData(carDataCh)

	raceControlCh := make(chan []*model.RaceControl, 16)
	resolver.RegisterRaceControlObserver("ws-hub", raceControlCh)
	go h.forwardRaceControl(raceControlCh)

	stintsCh := make(chan []*model.Stint, 16)
	resolver.RegisterStintObserver("ws-hub", stintsCh)
	go h.forwardStints(stintsCh)

	sectorsCh := make(chan []*model.Sector, 16)
	resolver.RegisterSectorTimeObserver("ws-hub", sectorsCh)
	go h.forwardSectors(sectorsCh)

	segmentsCh := make(chan []*model.Segment, 16)
	resolver.RegisterSegmentObserver("ws-hub", segmentsCh)
	go h.forwardSegments(segmentsCh)

	trackStatusCh := make(chan []*model.TrackStatus, 16)
	resolver.RegisterTrackStatusObserver("ws-hub", trackStatusCh)
	go h.forwardTrackStatus(trackStatusCh)

	qualifyingDataCh := make(chan *model.QualifyingData, 16)
	resolver.RegisterQualifyingDataObserver("ws-hub", qualifyingDataCh)
	go h.forwardQualifyingData(qualifyingDataCh)
}

func (h *Hub) forwardLapTimes(ch <-chan []*model.LapTime) {
	for data := range ch {
		h.broadcast("lapTimes", "data", data)
	}
}

func (h *Hub) forwardPositions(ch <-chan []*model.Position) {
	for data := range ch {
		h.broadcast("positions", "data", data)
	}
}

func (h *Hub) forwardCarData(ch <-chan *model.CarData) {
	for data := range ch {
		h.broadcast("carData", "data", data)
	}
}

func (h *Hub) forwardRaceControl(ch <-chan []*model.RaceControl) {
	for data := range ch {
		h.broadcast("raceControl", "data", data)
	}
}

func (h *Hub) forwardStints(ch <-chan []*model.Stint) {
	for data := range ch {
		h.broadcast("stints", "data", data)
	}
}

func (h *Hub) forwardSectors(ch <-chan []*model.Sector) {
	for data := range ch {
		h.broadcast("sectors", "data", data)
	}
}

func (h *Hub) forwardSegments(ch <-chan []*model.Segment) {
	for data := range ch {
		h.broadcast("segments", "data", data)
	}
}

func (h *Hub) forwardTrackStatus(ch <-chan []*model.TrackStatus) {
	for data := range ch {
		h.broadcast("trackStatus", "data", data)
	}
}

func (h *Hub) forwardQualifyingData(ch <-chan *model.QualifyingData) {
	for data := range ch {
		h.broadcast("qualifyingData", "data", data)
	}
}

func (h *Hub) broadcast(topic, msgType string, payload interface{}) {
	msg := OutboundMessage{
		Type:    msgType,
		Topic:   topic,
		Payload: payload,
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("error marshaling broadcast message:", err)
		return
	}
	h.mu.RLock()
	subscribers := h.topics[topic]
	for client := range subscribers {
		select {
		case client.send <- jsonBytes:
		default:
			// Client too slow, drop message
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) SetSessionKey(key int) {
	h.sessionMu.Lock()
	oldKey := h.sessionKey
	h.sessionKey = key
	h.sessionMu.Unlock()

	if oldKey != key && key != 0 {
		// Reset segment counts for the new session/circuit
		h.segmentCountsMu.Lock()
		h.segmentCounts = [3]int{}
		h.segmentCountsMu.Unlock()

		msg := OutboundMessage{
			Type:       "session_changed",
			SessionKey: key,
		}
		jsonBytes, err := json.Marshal(msg)
		if err != nil {
			return
		}
		h.mu.RLock()
		for client := range h.clients {
			select {
			case client.send <- jsonBytes:
			default:
			}
		}
		h.mu.RUnlock()
	}
}

func (h *Hub) UpdateSegmentCounts(segments []*model.Segment) {
	h.segmentCountsMu.Lock()
	wasZero := h.segmentCounts == [3]int{}
	for _, seg := range segments {
		idx := seg.SectorNumber - 1
		if idx >= 0 && idx < 3 && seg.SegmentNumber > h.segmentCounts[idx] {
			h.segmentCounts[idx] = seg.SegmentNumber
		}
	}
	counts := h.segmentCounts
	h.segmentCountsMu.Unlock()

	// Broadcast updated session_changed when counts are first populated
	if wasZero && counts != [3]int{} {
		h.broadcastSessionChanged()
	}
}

func (h *Hub) GetSegmentCounts() [3]int {
	h.segmentCountsMu.RLock()
	defer h.segmentCountsMu.RUnlock()
	return h.segmentCounts
}

func (h *Hub) broadcastSessionChanged() {
	key := h.GetSessionKey()
	if key == 0 {
		return
	}
	counts := h.GetSegmentCounts()
	msg := OutboundMessage{
		Type:       "session_changed",
		SessionKey: key,
	}
	if counts != [3]int{} {
		msg.SegmentCounts = counts[:]
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	for client := range h.clients {
		select {
		case client.send <- jsonBytes:
		default:
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) GetSessionKey() int {
	h.sessionMu.RLock()
	defer h.sessionMu.RUnlock()
	return h.sessionKey
}

func (h *Hub) Subscribe(client *Client, topics []string) {
	var subscribed []string
	var newTopics []string

	h.mu.Lock()
	for _, topic := range topics {
		if !ValidTopics[topic] {
			h.mu.Unlock()
			errMsg := OutboundMessage{
				Type:    "error",
				Message: fmt.Sprintf("unknown topic: %s", topic),
			}
			jsonBytes, _ := json.Marshal(errMsg)
			select {
			case client.send <- jsonBytes:
			default:
			}
			h.mu.Lock()
			continue
		}
		if !h.topics[topic][client] {
			h.topics[topic][client] = true
			newTopics = append(newTopics, topic)
		}
		subscribed = append(subscribed, topic)
	}
	h.mu.Unlock()

	if len(subscribed) > 0 {
		msg := OutboundMessage{
			Type:   "subscribed",
			Topics: subscribed,
		}
		jsonBytes, _ := json.Marshal(msg)
		select {
		case client.send <- jsonBytes:
		default:
		}
	}

	// Send recap for newly subscribed topics
	for _, topic := range newTopics {
		go h.sendRecap(client, topic)
	}

	// Send current session info with segment counts to newly connected client
	if len(newTopics) > 0 {
		key := h.GetSessionKey()
		counts := h.GetSegmentCounts()
		if key != 0 {
			msg := OutboundMessage{
				Type:       "session_changed",
				SessionKey: key,
			}
			if counts != [3]int{} {
				msg.SegmentCounts = counts[:]
			}
			jsonBytes, err := json.Marshal(msg)
			if err == nil {
				select {
				case client.send <- jsonBytes:
				default:
				}
			}
		}
	}
}

func (h *Hub) Unsubscribe(client *Client, topics []string) {
	h.mu.Lock()
	for _, topic := range topics {
		if subscribers, ok := h.topics[topic]; ok {
			delete(subscribers, client)
		}
	}
	h.mu.Unlock()
}
