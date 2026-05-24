package ws

// ValidTopics defines the set of topics clients can subscribe to.
var ValidTopics = map[string]bool{
	"lapTimes":    true,
	"positions":   true,
	"carData":     true,
	"raceControl": true,
	"stints":      true,
	"sectors":     true,
	"segments":    true,
	"trackStatus": true,
}

// InboundMessage represents a message from the client.
type InboundMessage struct {
	Type   string   `json:"type"`
	Topics []string `json:"topics,omitempty"`
}

// OutboundMessage represents a message sent to the client.
type OutboundMessage struct {
	Type       string      `json:"type"`
	Topic      string      `json:"topic,omitempty"`
	Payload    interface{} `json:"payload,omitempty"`
	Message    string      `json:"message,omitempty"`
	SessionKey int         `json:"sessionKey,omitempty"`
	Topics     []string    `json:"topics,omitempty"`
}
