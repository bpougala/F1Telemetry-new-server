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

type Position struct {
	SessionKey   int  `json:"SessionKey"`
	RacingNumber int  `json:"DriverNumber"`
	Position     int  `json:"Position"`
	Retired      bool `json:"Retired"`
	InPit        bool `json:"InPit"`
	PitOut       bool `json:"PitOut"`
	Stopped      bool `json:"Stopped"`
	Status       int  `json:"Status"`
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
	Lines map[string]interface{} `json:"Lines"`
}

type InitialSessionData struct {
	R struct {
		SessionData struct {
			Series       []interface{} `json:"Series"`
			StatusSeries []struct {
				Utc           time.Time `json:"Utc"`
				TrackStatus   string    `json:"TrackStatus,omitempty"`
				SessionStatus string    `json:"SessionStatus,omitempty"`
			} `json:"StatusSeries"`
			Kf bool `json:"_kf"`
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
		CarData  string `json:"CarData.Z"`
		LapCount struct {
			CurrentLap int  `json:"CurrentLap"`
			TotalLaps  int  `json:"TotalLaps"`
			Kf         bool `json:"_kf"`
		} `json:"LapCount"`
		DriverList          map[string]DriverData `json:"DriverList"`
		RaceControlMessages struct {
			Messages []RaceControl `json:"Messages"`
		} `json:"RaceControlMessages"`
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

type LapTimeMetric struct {
	SessionKey              int    `json:"SessionKey"`
	RacingNumber            string `json:"RacingNumber"`
	NumberOfLaps            int    `json:"NumberOfLaps"`
	GapToLeader             string `json:"GapToLeader"`
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
	SessionKey int    `json:"sessionkey"`
	Utc        string `json:"utc"`
	Category   string `json:"category"`
	Flag       string `json:"flag"`
	Scope      string `json:"scope"`
	Message    string `json:"message"`
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

type UpdateSector struct {
	Sectors map[string]Sector `json:"Sectors"`
}

type Sector struct {
	SectorNumber    int    `json:"SectorNumber"`
	Value           string `json:"Value"`
	PersonalFastest bool   `json:"PersonalFastest"`
	OverallFastest  bool   `json:"OverallFastest"`
}

type AllSectors struct {
	RacingNumber int
	Sectors      []Sector
	Utc          time.Time
}
