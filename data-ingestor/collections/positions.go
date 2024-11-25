package collections

type Position struct {
	SessionKey   int    `json:"session_key"`
	MeetingKey   int    `json:"meeting_key"`
	DriverNumber int    `json:"driver_number"`
	Date         string `json:"date"`
	Position     int    `json:"position"`
}
