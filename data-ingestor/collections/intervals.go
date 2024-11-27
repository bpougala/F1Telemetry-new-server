package collections

type Interval struct {
	Date         string `json:"date"`
	DriverNumber int    `json:"driver_number"`
	MeetingKey   int    `json:"meeting_key"`
	SessionKey   int    `json:"session_key"`
	GapToLeader  *any   `json:"gap_to_leader,omitempty"`
	Interval     *any   `json:"interval,omitempty"`
}
