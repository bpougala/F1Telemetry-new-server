package livetiming

type QMessage struct {
	R QualifyingMessage `json:"R"`
	I string            `json:"I"`
}

type SessionInfo struct {
	Meeting       Meeting       `json:"Meeting"`
	ArchiveStatus ArchiveStatus `json:"ArchiveStatus"`
	Key           int           `json:"Key"`
	Type          string        `json:"Type"`
	Name          string        `json:"Name"`
	StartDate     string        `json:"StartDate"`
	EndDate       string        `json:"EndDate"`
	GmtOffset     string        `json:"GmtOffset"`
	Path          string        `json:"Path"`
	Kf            bool          `json:"_kf"`
}

type Meeting struct {
	Key          int         `json:"Key"`
	Name         string      `json:"Name"`
	OfficialName string      `json:"OfficialName"`
	Location     string      `json:"Location"`
	Number       int         `json:"Number"`
	Country      CountryData `json:"Country"`
	Circuit      CircuitData `json:"Circuit"`
}

type QualifyingMessage struct {
	DriverList map[string]DriverData `json:"DriverList"`
	//TimingData          TimingData            `json:"TimingData"`
	SessionData SessionData `json:"SessionData"`
	//RaceControlMessages Messages              `json:"RaceControlMessages"`
	SessionInfo SessionInfo `json:"SessionInfo"`
}

type Messages struct {
	Messages []MessageData `json:"Messages"`
}

type MessageData struct {
	Utc      string `json:"Utc"`
	Category string `json:"Category"`
	Flag     string `json:"Flag,omitempty"`
	Scope    string `json:"Scope,omitempty"`
	Sector   int    `json:"Sector,omitempty"`
	Message  string `json:"Message"`
}

type CountryData struct {
	Key  int    `json:"Key"`
	Code string `json:"Code"`
	Name string `json:"Name"`
}

type CircuitData struct {
	Key       int    `json:"Key"`
	ShortName string `json:"ShortName"`
}

type MeetingData struct {
	Key          int         `json:"Key"`
	Name         string      `json:"Name"`
	OfficialName string      `json:"OfficialName"`
	Location     string      `json:"Location"`
	Number       int         `json:"Number"`
	Country      CountryData `json:"Country"`
	Circuit      CircuitData `json:"Circuit"`
}

type ArchiveStatus struct {
	Status string `json:"Status"`
}

type SessionData struct {
	Meeting       MeetingData   `json:"Meeting"`
	ArchiveStatus ArchiveStatus `json:"ArchiveStatus"`
	Key           int           `json:"Key"`
	Type          string        `json:"Type"`
	Name          string        `json:"Name"`
	StartDate     string        `json:"StartDate"`
	EndDate       string        `json:"EndDate"`
	GmtOffset     string        `json:"GmtOffset"`
	Path          string        `json:"Path"`
	Kf            bool          `json:"_kf"`
}

type DriverData struct {
	RacingNumber  string `json:"RacingNumber"`
	BroadcastName string `json:"BroadcastName"`
	FullName      string `json:"FullName"`
	Tla           string `json:"Tla"`
	Line          int    `json:"Line"`
	TeamName      string `json:"TeamName"`
	TeamColour    string `json:"TeamColour"`
	FirstName     string `json:"FirstName"`
	LastName      string `json:"LastName"`
	Reference     string `json:"Reference"`
	HeadshotUrl   string `json:"HeadshotUrl"`
	CountryCode   string `json:"CountryCode"`
}

type TimingData struct {
	NoEntries        []int               `json:"NoEntries"`
	SessionPart      int                 `json:"SessionPart"`
	CutOffTime       string              `json:"CutOffTime"`
	CutOffPercentage string              `json:"CutOffPercentage"`
	Lines            map[string]LineData `json:"Lines"`
}

type LineData struct {
	KnockedOut       bool      `json:"KnockedOut"`
	Cutoff           bool      `json:"Cutoff"`
	BestLapTimes     []LapTime `json:"BestLapTimes"`
	Stats            []Stat    `json:"Stats"`
	Line             int       `json:"Line"`
	Position         string    `json:"Position"`
	ShowPosition     bool      `json:"ShowPosition"`
	RacingNumber     string    `json:"RacingNumber"`
	Retired          bool      `json:"Retired"`
	InPit            bool      `json:"InPit"`
	PitOut           bool      `json:"PitOut"`
	Stopped          bool      `json:"Stopped"`
	Status           int       `json:"Status"`
	NumberOfLaps     int       `json:"NumberOfLaps"`
	NumberOfPitStops int       `json:"NumberOfPitStops"`
	Sectors          []Sector  `json:"Sectors"`
	Speeds           Speeds    `json:"Speeds"`
	BestLapTime      LapTime   `json:"BestLapTime"`
	LastLapTime      LapTime   `json:"LastLapTime"`
}

type LapTime struct {
	Value string `json:"Value"`
	Lap   int    `json:"Lap"`
}

type Stat struct {
	TimeDiffToFastest       string `json:"TimeDiffToFastest"`
	TimeDifftoPositionAhead string `json:"TimeDifftoPositionAhead"`
}

type Sector struct {
	Stopped         bool      `json:"Stopped"`
	PreviousValue   string    `json:"PreviousValue"`
	Segments        []Segment `json:"Segments"`
	Value           string    `json:"Value"`
	Status          int       `json:"Status"`
	OverallFastest  bool      `json:"OverallFastest"`
	PersonalFastest bool      `json:"PersonalFastest"`
}

type Segment struct {
	Status int `json:"Status"`
}

type Speeds struct {
	I1 Speed `json:"I1"`
	I2 Speed `json:"I2"`
	FL Speed `json:"FL"`
	ST Speed `json:"ST"`
}

type Speed struct {
	Value           string `json:"Value"`
	Status          int    `json:"Status"`
	OverallFastest  bool   `json:"OverallFastest"`
	PersonalFastest bool   `json:"PersonalFastest"`
}
