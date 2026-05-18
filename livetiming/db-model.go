package livetiming

import "time"

type SessionInfoDB struct {
	ArchiveStatus string `json:"ArchiveStatus"`
	StartDate     string `json:"StartDate"`
	EndDate       string `json:"EndDate"`
	Type          string `json:"Type"`
	GmtOffset     string `json:"GmtOffset"`
	Key           int    `json:"_id"`
	MeetingKey    int    `json:"MeetingKey"`
	Name          string `json:"Name"`
	Path          string `json:"Path"`
}

type MeetingDataDB struct {
	Key          int    `json:"_id"`
	Name         string `json:"Name"`
	OfficialName string `json:"OfficialName"`
	Location     string `json:"Location"`
	Number       int    `json:"Number"`
	CountryName  string `json:"CountryName"`
	CountryCode  string `json:"CountryCode"`
	Circuit      string `json:"Circuit"`
	CircuitKey   int    `json:"CircuitKey"`
}

type IntervalsDB struct {
	SessionKey   int    `json:"sessionkey"`
	DriverNumber int    `json:"drivernumber"`
	Interval     string `json:"interval"`
	GapToLeader  string `json:"gaptoleader"`
	Timestamp    string `json:"timestamp"`
}

type PositionsDB struct {
	SessionKey   int    `json:"sessionkey"`
	RacingNumber int    `json:"racingnumber"`
	Position     int    `json:"position"`
	Timestamp    string `json:"timestamp"`
}

type StintDB struct {
	SessionKey            int    `json:"session"`
	RacingNumber          int    `json:"driver"`
	Compound              string `json:"compound"`
	Is_new                bool   `json:"is_new"`
	Are_tyres_not_changed bool   `json:"are_tyres_not_changed"`
	TotalLaps             int    `json:"total_laps"`
	StartLaps             int    `json:"start_laps"`
	Timestamp             string `json:"timestamp"`
}

type SectorDB struct {
	SessionKey      int       `json:"session_key"`
	RacingNumber    int       `json:"racing_number"`
	SectorNumber    int       `json:"sector_number"`
	Value           string    `json:"value"`
	OverallFastest  bool      `json:"overall_fastest"`
	PersonalFastest bool      `json:"personal_fastest"`
	Utc             time.Time `json:"utc,omitempty"`
}
