package livetiming

import (
	"encoding/json"
	"time"
)

type WebSocketMessage struct {
	Control   string `json:"C"`
	DataBlock []struct {
		SubscribeType string        `json:"H"`
		ResultType    string        `json:"M"`
		DataFeed      []interface{} `json:"A"`
	} `json:"M"`
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
type CountryData struct {
	Key  int    `json:"Key"`
	Code string `json:"Code"`
	Name string `json:"Name"`
}

type CircuitData struct {
	Key       int    `json:"Key"`
	ShortName string `json:"ShortName"`
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
type ArchiveStatus struct {
	Status string `json:"Status"`
}

type StintData struct {
	StintLines map[string]StintLine `json:"Lines"`
}

type PositionData struct {
	PositionLines map[string]PositionLine `json:"Lines"`
}

type PositionLine struct {
	Line int `json:"Line"`
}

type TrackStatus struct {
	Utc    time.Time `json:"Utc"`
	Status string    `json:"Status"`
}

type Position struct {
	SessionKey   int  `json:"SessionKey"`
	RacingNumber int  `json:"RacingNumber"`
	Position     *int `json:"Position"`
	Retired      bool `json:"Retired"`
	InPit        bool `json:"InPit"`
	PitOut       bool `json:"PitOut"`
	Stopped      bool `json:"Stopped"`
	Status       int  `json:"Status"`
	PositionOnly bool `json:"-"` // true when only positional data is available (DriverList)
}

type RawPosition struct {
	InPit        bool   `json:"InPit"`
	PitOut       bool   `json:"PitOut"`
	Stopped      bool   `json:"Stopped"`
	Retired      bool   `json:"Retired"`
	Status       int    `json:"Status"`
	Position     string `json:"Position"`
	Line         int    `json:"Line"`
	NumberOfLaps int    `json:"NumberOfLaps"`
}

type StintLine struct {
	Stints map[string]RawStint `json:"Stints"`
}

type RawStint struct {
	LapFlags        int    `json:"LapFlags"`
	Compound        string `json:"Compound"`
	New             string `json:"New"`
	TyresNotChanged string `json:"TyresNotChanged"`
	TotalLaps       int    `json:"TotalLaps"`
	StartLaps       int    `json:"StartLaps"`
	Timestamp       string `json:"Timestamp"`
}

type Stint struct {
	SessionKey      int    `json:"SessionKey"`
	RacingNumber    int    `json:"RacingNumber"`
	LapFlags        int    `json:"LapFlags"`
	Compound        string `json:"Compound"`
	New             bool   `json:"New"`
	TyresNotChanged int    `json:"TyresNotChanged"`
	TotalLaps       int    `json:"TotalLaps"`
	StartLaps       int    `json:"StartLaps"`
	Timestamp       string `json:"Timestamp"`
	StintNumber     int    `json:"StintNumber"`
}

type IntervalLine struct {
	GapToLeader             string `json:"GapToLeader,omitempty"`
	IntervalToPositionAhead Value  `json:"IntervalToPositionAhead"`
}

type Value struct {
	Value string `json:"Value"`
}

type UpdateData struct {
	C string    `json:"C"`
	M []Message `json:"M"`
}

type Message struct {
	H string        `json:"H"`
	M string        `json:"M"`
	A []interface{} `json:"A"`
}

type TimingData struct {
	Lines            map[string]interface{} `json:"Lines"`
	SessionPart      int                    `json:"SessionPart"`
	NoEntries        []int                  `json:"NoEntries"`
	CutOffTime       string                 `json:"CutOffTime"`
	CutOffPercentage string                 `json:"CutOffPercentage"`
}

type StatusSeries struct {
	Utc           time.Time `json:"Utc"`
	TrackStatus   string    `json:"TrackStatus,omitempty"`
	SessionStatus string    `json:"SessionStatus,omitempty"`
}

type InitialSessionData struct {
	R struct {
		SessionData struct {
			Series       []interface{}  `json:"Series"`
			StatusSeries []StatusSeries `json:"StatusSeries"`
			Kf           bool           `json:"_kf"`
		} `json:"SessionData"`
		SessionInfo struct {
			Meeting struct {
				Key          int    `json:"Key"`
				Name         string `json:"Name"`
				OfficialName string `json:"OfficialName"`
				Location     string `json:"Location"`
				Number       int    `json:"Number"`
				Country      struct {
					Key  int    `json:"Key"`
					Code string `json:"Code"`
					Name string `json:"Name"`
				} `json:"Country"`
				Circuit struct {
					Key       int    `json:"Key"`
					ShortName string `json:"ShortName"`
				} `json:"Circuit"`
			} `json:"Meeting"`
			ArchiveStatus struct {
				Status string `json:"Status"`
			} `json:"ArchiveStatus"`
			Key       int    `json:"Key"`
			Type      string `json:"Type"`
			Number    int    `json:"Number"`
			Name      string `json:"Name"`
			StartDate string `json:"StartDate"`
			EndDate   string `json:"EndDate"`
			GmtOffset string `json:"GmtOffset"`
			Path      string `json:"Path"`
			Kf        bool   `json:"_kf"`
		} `json:"SessionInfo"`
		DriverList map[string]DriverData `json:"DriverList"`
	} `json:"R"`
	I string `json:"I"`
}

func (d *DriverData) UnmarshalJSON(data []byte) error {
	var isBool bool
	if err := json.Unmarshal(data, &isBool); err == nil {
		return nil
	}

	type Alias DriverData
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*d = DriverData(alias)
	return nil
}

type SessionStatusStruct struct {
	Utc           string `json:"Utc"`
	SessionStatus string `json:"SessionStatus"`
}

type InitialData struct {
	R struct {
		TimingData struct {
			Lines map[string]struct {
				GapToLeader             string `json:"GapToLeader"`
				IntervalToPositionAhead struct {
					Value    string `json:"Value"`
					Catching bool   `json:"Catching"`
				} `json:"IntervalToPositionAhead"`
				Line         int    `json:"Line"`
				Position     string `json:"Position"`
				ShowPosition bool   `json:"ShowPosition"`
				RacingNumber string `json:"RacingNumber"`
				Retired      bool   `json:"Retired"`
				InPit        bool   `json:"InPit"`
				PitOut       bool   `json:"PitOut"`
				Stopped      bool   `json:"Stopped"`
				Status       int    `json:"Status"`
				NumberOfLaps int    `json:"NumberOfLaps"`
				Sectors      []struct {
					Stopped         bool   `json:"Stopped"`
					Value           string `json:"Value"`
					Status          int    `json:"Status"`
					OverallFastest  bool   `json:"OverallFastest"`
					PersonalFastest bool   `json:"PersonalFastest"`
					Segments        []struct {
						Status int `json:"Status"`
					} `json:"Segments"`
				} `json:"Sectors"`
				Speeds struct {
					I1 struct {
						Value           string `json:"Value"`
						Status          int    `json:"Status"`
						OverallFastest  bool   `json:"OverallFastest"`
						PersonalFastest bool   `json:"PersonalFastest"`
					} `json:"I1"`
					I2 struct {
						Value           string `json:"Value"`
						Status          int    `json:"Status"`
						OverallFastest  bool   `json:"OverallFastest"`
						PersonalFastest bool   `json:"PersonalFastest"`
					} `json:"I2"`
					FL struct {
						Value           string `json:"Value"`
						Status          int    `json:"Status"`
						OverallFastest  bool   `json:"OverallFastest"`
						PersonalFastest bool   `json:"PersonalFastest"`
					} `json:"FL"`
					ST struct {
						Value           string `json:"Value"`
						Status          int    `json:"Status"`
						OverallFastest  bool   `json:"OverallFastest"`
						PersonalFastest bool   `json:"PersonalFastest"`
					} `json:"ST"`
				} `json:"Speeds"`
				BestLapTime struct {
					Value string `json:"Value"`
					Lap   int    `json:"Lap"`
				} `json:"BestLapTime"`
				LastLapTime struct {
					Value           string `json:"Value"`
					Status          int    `json:"Status"`
					OverallFastest  bool   `json:"OverallFastest"`
					PersonalFastest bool   `json:"PersonalFastest"`
				} `json:"LastLapTime"`
			} `json:"Lines"`
			Withheld bool `json:"Withheld"`
			Kf       bool `json:"_kf"`
		} `json:"TimingData"`
		TimingStats struct {
			Withheld bool `json:"Withheld"`
			Lines    map[string]struct {
				Line                int    `json:"Line"`
				RacingNumber        string `json:"RacingNumber"`
				PersonalBestLapTime struct {
					Value string `json:"Value"`
				} `json:"PersonalBestLapTime"`
				BestSectors []struct {
					Value string `json:"Value"`
				} `json:"BestSectors"`
				BestSpeeds struct {
					I1 struct {
						Value string `json:"Value"`
					} `json:"I1"`
					I2 struct {
						Value string `json:"Value"`
					} `json:"I2"`
					FL struct {
						Value string `json:"Value"`
					} `json:"FL"`
					ST struct {
						Value string `json:"Value"`
					} `json:"ST"`
				} `json:"BestSpeeds"`
			} `json:"Lines"`
			SessionType string `json:"SessionType"`
			Kf          bool   `json:"_kf"`
		} `json:"TimingStats"`
		TimingAppData struct {
			Lines map[string]struct {
				RacingNumber string     `json:"RacingNumber"`
				Stints       []RawStint `json:"Stints"`
				Line         int        `json:"Line"`
				GridPos      string     `json:"GridPos"`
			} `json:"Lines"`
			Kf bool `json:"_kf"`
		} `json:"TimingAppData"`
		CarData   string `json:"CarData.Z"`
		PositionZ string `json:"Position.z"`
		LapCount  struct {
			CurrentLap int  `json:"CurrentLap"`
			TotalLaps  int  `json:"TotalLaps"`
			Kf         bool `json:"_kf"`
		} `json:"LapCount"`
		DriverList          map[string]DriverData `json:"DriverList"`
		RaceControlMessages struct {
			Messages []RaceControl `json:"Messages"`
		} `json:"RaceControlMessages"`
		WeatherData Weather `json:"WeatherData"`
	} `json:"R"`
	I string `json:"I"`
}

type DriverList struct {
	R struct {
		DriverList map[string]DriverData `json:"DriverList"`
	} `json:"R"`
	I string `json:"I"`
}

type Driver struct {
	RacingNumber  int    `json:"RacingNumber"`
	BroadcastName string `json:"BroadcastName"`
	FullName      string `json:"FullName"`
	Abbreviation  string `json:"Tla"`
	TeamName      string `json:"TeamName"`
	TeamColour    string `json:"TeamColour"`
	FirstName     string `json:"FirstName"`
	LastName      string `json:"LastName"`
	HeadshotUrl   string `json:"HeadshotUrl"`
	SessionKey    int    `json:"SessionKey"`
}

type DriverID struct {
	SessionKey   int `bson:"session_key"`
	RacingNumber int `bson:"racing_number"`
}

type DriverDocument struct {
	ID     DriverID `bson:"_id"`
	Driver `bson:",inline"`
}

type QualifyingState struct {
	SessionPart      int    `json:"SessionPart"`
	NoEntries        []int  `json:"NoEntries"`
	CutOffTime       string `json:"CutOffTime"`
	CutOffPercentage string `json:"CutOffPercentage"`
}

type QualifyingLapTime struct {
	Value string `json:"Value"`
	Lap   int    `json:"Lap"`
}

type QualifyingStats struct {
	TimeDiffToFastest       string `json:"TimeDiffToFastest"`
	TimeDifftoPositionAhead string `json:"TimeDifftoPositionAhead"`
}

type QualifyingPart struct {
	Utc            string `json:"Utc"`
	QualifyingPart int    `json:"QualifyingPart"`
}

type LapTimeMetric struct {
	SessionKey              int    `json:"SessionKey"`
	RacingNumber            string `json:"RacingNumber"`
	NumberOfLaps            int    `json:"NumberOfLaps"`
	GapToLeader             string `json:"GapToLeader"`
	InPit                   bool   `json:"InPit"`
	PitOut                  bool   `json:"PitOut"`
	Stopped                 bool   `json:"Stopped"`
	IntervalToPositionAhead struct {
		Value    string `json:"Value"`
		Catching bool   `json:"Catching"`
	}
	BestLapTime struct {
		Value string `json:"Value"`
		Lap   int    `json:"Lap"`
	} `json:"BestLapTimes"`
	LastLapTime struct {
		Value           string `json:"Value"`
		OverallFastest  bool   `json:"OverallFastest"`
		PersonalFastest bool   `json:"PersonalFastest"`
	} `json:"LastLapTime"`
	KnockedOut         bool                `json:"KnockedOut"`
	Cutoff             bool                `json:"Cutoff"`
	QualifyingBestLaps []QualifyingLapTime `json:"-" mapstructure:"-"`
	QualifyingStats    []QualifyingStats   `json:"-" mapstructure:"-"`
}

type SectorTime struct {
	SectorNumber    int    `json:"SectorNumber"`
	Value           string `json:"Value"`
	PersonalFastest bool   `json:"PersonalFastest"`
	OverallFastest  bool   `json:"OverallFastest"`
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
type CarData struct {
	Compressed string `json:"Compressed"`
}

type RaceControl struct {
	SessionKey int    `json:"SessionKey"`
	Utc        string `json:"Utc"`
	Category   string `json:"Category"`
	Flag       string `json:"Flag"`
	Scope      string `json:"Scope"`
	Message    string `json:"Message"`
	LapNumber  *int   `json:"LapNumber"`
}

type RawRaceControl struct {
	Messages []RaceControl `json:"Messages"`
}

type Root struct {
	Entries []Entry `json:"Entries"`
}

// Entry represents each entry in the Entries array
type Entry struct {
	Utc  time.Time      `json:"Utc"`
	Cars map[string]Car `json:"Cars"`
}

// Car represents a car in the Cars map
type Car struct {
	Channels map[string]int `json:"Channels"`
}

// PositionRoot represents decompressed Position.z data
type PositionRoot struct {
	Position []PositionEntry `json:"Position"`
}

// PositionEntry represents a timestamped set of car positions
type PositionEntry struct {
	Timestamp time.Time              `json:"Timestamp"`
	Entries   map[string]CarPosition `json:"Entries"`
}

// CarPosition represents a single car's GPS position
type CarPosition struct {
	Status string `json:"Status"`
	X      int    `json:"X"`
	Y      int    `json:"Y"`
	Z      int    `json:"Z"`
}

type UpdateSector struct {
	Sectors map[string]UpdateSectorEntry `json:"Sectors"`
}

type UpdateSectorEntry struct {
	Value           string                     `json:"Value"`
	PersonalFastest bool                       `json:"PersonalFastest"`
	OverallFastest  bool                       `json:"OverallFastest"`
	LapNumber       int                        `json:"LapNumber"`
	Segments        map[string]SegmentRawEntry `json:"Segments"`
}

type SegmentRawEntry struct {
	Status int `json:"Status"`
}

type SegmentStatus struct {
	SectorNumber  int
	SegmentNumber int
	Status        int
}

type AllSegments struct {
	RacingNumber int
	Segments     []SegmentStatus
	Utc          time.Time
}

type Sector struct {
	SectorNumber    int    `json:"SectorNumber"`
	Value           string `json:"Value"`
	PersonalFastest bool   `json:"PersonalFastest"`
	OverallFastest  bool   `json:"OverallFastest"`
	LapNumber       int    `json:"LapNumber"`
}

type AllSectors struct {
	RacingNumber int
	Sectors      []Sector
	Utc          time.Time
}

type Weather struct {
	AirTemp       string `json:"AirTemp"`
	TrackTemp     string `json:"TrackTemp"`
	Humidity      string `json:"Humidity"`
	Pressure      string `json:"Pressure"`
	Rainfall      string `json:"Rainfall"`
	WindSpeed     string `json:"WindSpeed"`
	WindDirection string `json:"WindDirection"`
}
