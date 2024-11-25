package collections

type Driver struct {
	SessionKey    int     `json:"session_key"`
	MeetingKey    int     `json:"meeting_key"`
	BroadcastName string  `json:"broadcast_name"`
	CountryCode   string  `json:"country_code"`
	FirstName     string  `json:"first_name"`
	FullName      string  `json:"full_name"`
	HeadshotURL   string  `json:"headshot_url"`
	LastName      string  `json:"last_name"`
	DriverNumber  int     `json:"driver_number"`
	TeamColour    *string `json:"team_colour,omitempty"`
	TeamName      *string `json:"team_name,omitempty"`
	NameAcronym   *string `json:"name_acronym,omitempty"`
}
