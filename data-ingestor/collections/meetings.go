package collections

type Meeting struct {
	CircuitKey          int    `json:"circuit_key"`
	CircuitShortName    string `json:"circuit_short_name"`
	MeetingKey          int    `json:"meeting_key"`
	MeetingCode         string `json:"meeting_code"`
	Location            string `json:"location"`
	CountryKey          int    `json:"country_key"`
	CountryCode         string `json:"country_code"`
	CountryName         string `json:"country_name"`
	MeetingName         string `json:"meeting_name"`
	MeetingOfficialName string `json:"meeting_official_name"`
	GmtOffset           string `json:"gmt_offset"`
	DateStart           string `json:"date_start"`
	Year                int    `json:"year"`
}
