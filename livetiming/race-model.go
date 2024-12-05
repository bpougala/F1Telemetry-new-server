package livetiming

type WebSocketMessage struct {
	Control   string `json:"C"`
	DataBlock []struct {
		SubscribeType string        `json:"H"`
		ResultType    string        `json:"M"`
		DataFeed      []interface{} `json:"A"`
	} `json:"M"`
}

type IntervalData struct {
	Lines map[string]IntervalLine `json:"Lines"`
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
	DriverNumber int `json:"DriverNumber"`
	Position     int `json:"Position"`
}
type StintLine struct {
	Stints map[string]Stint `json:"Stints"`
}

type Stint struct {
	DriverNumber    int    `json:"DriverNumber"`
	LapFlags        int    `json:"LapFlags"`
	Compound        string `json:"Compound"`
	New             string `json:"New"`
	TyresNotChanged string `json:"TyresNotChanged"`
	TotalLaps       int    `json:"TotalLaps"`
	StartLaps       int    `json:"StartLaps"`
	Timestamp       string `json:"Timestamp"`
}

type IntervalLine struct {
	GapToLeader             string `json:"GapToLeader,omitempty"`
	IntervalToPositionAhead Value  `json:"IntervalToPositionAhead"`
}

type Value struct {
	Value string `json:"Value"`
}

type LapTimeData struct {
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
				RacingNumber string        `json:"RacingNumber"`
				Stints       []interface{} `json:"Stints"`
				Line         int           `json:"Line"`
				GridPos      string        `json:"GridPos"`
			} `json:"Lines"`
			Kf bool `json:"_kf"`
		} `json:"TimingAppData"`
		LapCount struct {
			CurrentLap int  `json:"CurrentLap"`
			TotalLaps  int  `json:"TotalLaps"`
			Kf         bool `json:"_kf"`
		} `json:"LapCount"`
	} `json:"R"`
	I string `json:"I"`
}
