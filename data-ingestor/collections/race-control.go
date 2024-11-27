package collections

import "time"

type RaceControl struct {
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	Date         time.Time `json:"date"`
	Category     string    `json:"category"`
	Flag         string    `json:"flag"`
	LapNumber    int       `json:"lap_number"`
	Message      string    `json:"message"`
	DriverNumber int       `json:"driver_number"`
	Scope        string    `json:"scope"`
	Sector       int       `json:"sector"`
}
