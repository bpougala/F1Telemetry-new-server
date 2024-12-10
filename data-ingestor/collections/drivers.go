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

type DriverDB struct {
	RacingNumber  int    `json:"RacingNumber"`
	BroadcastName string `json:"BroadcastName"`
	FullName      string `json:"FullName"`
	Abbreviation  string `json:"Tla"`
	TeamName      string `json:"TeamName"`
	TeamColour    string `json:"TeamColour"`
	FirstName     string `json:"FirstName"`
	LastName      string `json:"LastName"`
	HeadshotUrl   string `json:"HeadshotUrl"`
	CountryCode   string `json:"CountryCode"`
	SessionKey    int    `json:"SessionKey"`
}
