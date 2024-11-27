package collections

type Laps struct {
	MeetingKey      int      `json:"meeting_key"`
	SessionKey      int      `json:"session_key"`
	DriverNumber    int      `json:"driver_number"`
	I1Speed         *int     `json:"i1_speed"`
	I2Speed         int      `json:"i2_speed"`
	StSpeed         int      `json:"st_speed"`
	DateStart       *string  `json:"date_start"`
	LapDuration     *float64 `json:"lap_duration"`
	IsPitOutLap     bool     `json:"is_pit_out_lap"`
	DurationSector1 *float64 `json:"duration_sector_1"`
	DurationSector2 float64  `json:"duration_sector_2"`
	DurationSector3 float64  `json:"duration_sector_3"`
	SegmentsSector1 []int    `json:"segments_sector_1"`
	SegmentsSector2 []int    `json:"segments_sector_2"`
	SegmentsSector3 []int    `json:"segments_sector_3"`
	LapNumber       int      `json:"lap_number"`
}
