package collections

type Sessions struct {
	MeetingKey       int    `json:"meeting_key"`
	SessionKey       int    `json:"session_key"`
	Location         string `json:"location"`
	DateStart        string `json:"date_start"`
	DateEnd          string `json:"date_end"`
	SessionType      string `json:"session_type"`
	SessionName      string `json:"session_name"`
	CountryKey       int    `json:"country_key"`
	CountryCode      string `json:"country_code"`
	CountryName      string `json:"country_name"`
	CircuitKey       int    `json:"circuit_key"`
	CircuitShortName string `json:"circuit_short_name"`
	GmtOffset        string `json:"gmt_offset"`
	Year             int    `json:"year"`
}

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
