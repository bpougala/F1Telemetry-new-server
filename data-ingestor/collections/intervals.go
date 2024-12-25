package collections

type Interval struct {
	Date         string `json:"date"`
	DriverNumber int    `json:"driver_number"`
	MeetingKey   int    `json:"meeting_key"`
	SessionKey   int    `json:"session_key"`
	GapToLeader  *any   `json:"gap_to_leader,omitempty"`
	Interval     *any   `json:"interval,omitempty"`
}

type IntervalsDB struct {
	SessionKey   int    `json:"sessionkey"`
	DriverNumber int    `json:"drivernumber"`
	Interval     string `json:"interval"`
	GapToLeader  string `json:"gaptoleader"`
	Timestamp    string `json:"timestamp"`
}

type RawInterval struct {
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
}

type Timing struct {
	SessionKey              int    `json:"session_key"`
	RacingNumber            int    `json:"racing_number"`
	GapToLeader             string `json:"gap_to_leader"`
	IntervalToPositionAhead struct {
		Value    string `json:"value"`
		Catching bool   `json:"catching"`
	} `json:"interval_to_position_ahead"`
	Position     int  `json:"position"`
	Retired      bool `json:"retired"`
	InPit        bool `json:"in_pit"`
	PitOut       bool `json:"pit_out"`
	Stopped      bool `json:"stopped"`
	Status       int  `json:"status"`
	NumberOfLaps int  `json:"number_of_laps"`
	Sectors      []struct {
		Stopped         bool   `json:"stopped"`
		Value           string `json:"value"`
		Status          int    `json:"status"`
		OverallFastest  bool   `json:"overall_fastest"`
		PersonalFastest bool   `json:"personal_fastest"`
		Segments        []struct {
			Status int `json:"status"`
		} `json:"segments"`
	} `json:"sectors"`
	BestLapTime struct {
		Value string `json:"value"`
		Lap   int    `json:"lap"`
	} `json:"best_lap_time"`
	LastLapTime struct {
		Value           string `json:"value"`
		Status          int    `json:"status"`
		OverallFastest  bool   `json:"overall_fastest"`
		PersonalFastest bool   `json:"personal_fastest"`
	} `json:"last_lap_time"`
}
