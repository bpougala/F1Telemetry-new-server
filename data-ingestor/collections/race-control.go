package collections

type RaceControl struct {
	SessionKey int    `json:"session_key"`
	Utc        string `json:"date"`
	Category   string `json:"category"`
	Flag       string `json:"flag"`
	Message    string `json:"message"`
	Scope      string `json:"scope"`
}
