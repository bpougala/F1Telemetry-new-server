package livetiming

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
	Country      string `json:"Country"`
	Circuit      string `json:"Circuit"`
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
	DriverNumber int    `json:"drivernumber"`
	Position     int    `json:"position"`
	Timestamp    string `json:"timestamp"`
}

type StintDB struct {
	SessionKey            int    `json:"session"`
	DriverNumber          int    `json:"driver"`
	Compound              string `json:"compound"`
	Is_new                bool   `json:"is_new"`
	Are_tyres_not_changed bool   `json:"are_tyres_not_changed"`
	TotalLaps             int    `json:"total_laps"`
	StartLaps             int    `json:"start_laps"`
	Timestamp             string `json:"timestamp"`
}
