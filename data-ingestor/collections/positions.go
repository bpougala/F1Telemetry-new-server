package collections

type Position struct {
	SessionKey   int    `json:"session_key"`
	MeetingKey   int    `json:"meeting_key"`
	DriverNumber int    `json:"driver_number"`
	Date         string `json:"date"`
	Position     int    `json:"position"`
}

type PositionsDB struct {
	SessionKey   int    `json:"sessionkey"`
	RacingNumber int    `json:"racingnumber"`
	Position     int    `json:"position"`
	Timestamp    string `json:"timestamp"`
}
