package collections

type StintDB struct {
	Compound        string `json:"compound"`
	LapFlags        int    `json:"lapflags"`
	New             bool   `json:"new"`
	RacingNumber    int    `json:"racingnumber"`
	SessionKey      int    `json:"sessionkey"`
	StartLaps       int    `json:"startlaps"`
	Timestamp       string `json:"timestamp"`
	TyresNotChanged int    `json:"tyresnotchanged"`
	TotalLaps       int    `json:"totallaps"`
}
