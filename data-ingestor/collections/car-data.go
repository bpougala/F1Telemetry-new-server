package collections

type CarData struct {
	MeetingKey   int    `json:"meeting_key"`
	SessionKey   int    `json:"session_key"`
	DriverNumber int    `json:"driver_number"`
	Date         string `json:"date"`
	Rpm          int    `json:"rpm"`
	Speed        int    `json:"speed"`
	NGear        int    `json:"n_gear"`
	Throttle     int    `json:"throttle"`
	Drs          int    `json:"drs"`
	Brake        int    `json:"brake"`
}
