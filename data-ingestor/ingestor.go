package data_ingestor

import (
	"F1Telemetry-new-server/data-ingestor/collections"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var BASE_URL = "https://api.openf1.org/v1"

func IngestMeetings(year int) ([]collections.Meeting, error) {
	uri := fmt.Sprintf("%s/meetings?year=%d", BASE_URL, year)
	resp, err := http.Get(uri)
	var meetings []collections.Meeting
	if err != nil {
		return meetings, err
	}
	if resp.StatusCode != 200 {
		return meetings, fmt.Errorf("error fetching data: %s", resp.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &meetings)
	if err != nil {
		return meetings, err
	}
	return meetings, nil
}

func IngestSession(meetingKey int) ([]collections.Sessions, error) {
	uri := fmt.Sprintf("%s/sessions?meeting_key=%d", BASE_URL, meetingKey)
	resp, err := http.Get(uri)
	var sessions []collections.Sessions
	if err != nil {
		return sessions, err
	}
	if resp.StatusCode != 200 {
		return sessions, fmt.Errorf("error fetching data: %s", resp.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &sessions)
	if err != nil {
		return sessions, err
	}
	return sessions, nil
}

func IngestDrivers(sessionKey int) ([]collections.Driver, error) {
	uri := fmt.Sprintf("%s/drivers?session_key=%d", BASE_URL, sessionKey)
	resp, err := http.Get(uri)
	var drivers []collections.Driver
	if err != nil {
		return drivers, err
	}
	if resp.StatusCode != 200 {
		return drivers, fmt.Errorf("error fetching data: %s", resp.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &drivers)
	if err != nil {
		return drivers, err
	}
	return drivers, nil
}

func IngestPositions(sessionKey int) ([]collections.Position, error) {
	uri := fmt.Sprintf("%s/position?session_key=%d", BASE_URL, sessionKey)
	resp, err := http.Get(uri)
	var positions []collections.Position
	if err != nil {
		return positions, err
	}
	if resp.StatusCode != 200 {
		return positions, fmt.Errorf("error fetching data: %s", resp.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &positions)
	if err != nil {
		return positions, err
	}
	return positions, nil
}

/*var BASE_URL = "https://livetiming.formula1.com/static"

type Meetings struct {
	Year     int       `json:"Year"`
	Meetings []Meeting `json:"Meetings"`
}

type Meeting struct {
	Sessions     []Session `json:"Sessions"`
	Key          int       `json:"Key"`
	Code         string    `json:"Code"`
	Number       int       `json:"Number"`
	Location     string    `json:"Location"`
	OfficialName string    `json:"OfficialName"`
	Name         string    `json:"Name"`
	Country      struct {
		Key  int    `json:"Key"`
		Code string `json:"Code"`
		Name string `json:"Name"`
	} `json:"Country"`
	Circuit struct {
		Key       int    `json:"Key"`
		ShortName string `json:"ShortName"`
	} `json:"Circuit"`
}

type Session struct {
	Key       int    `json:"Key"`
	Type      string `json:"Type"`
	Number    int    `json:"Number"`
	Name      string `json:"Name"`
	StartDate string `json:"StartDate"`
	EndDate   string `json:"EndDate"`
	GmtOffset string `json:"GmtOffset"`
	Path      string `json:"Path"`
}

func IngestMeetings(

func IngestMeetings(year int) (Meetings, error) {
	var meetings Meetings
	url := fmt.Sprintf("%s/%d/Index.json", BASE_URL, year)
	resp, err := http.Get(url)

	if err != nil {
		return meetings, err
	}
	if resp.StatusCode != 200 {
		return meetings, fmt.Errorf("error fetching data: %s", resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return meetings, err
	}
	body = removeBOM(body)
	err = json.Unmarshal(body, &meetings)
	if err != nil {
		return meetings, err
	}
	return meetings, nil
}

func removeBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

func GetMeetingKeys(meetings Meetings) []int {
	var keys []int
	for _, meeting := range meetings.Meetings {
		keys = append(keys, meeting.Key)
	}
	return keys
}

func GetSessionKeys(meetings Meetings, meetingKey int) []int {
	var keys []int
	for _, meeting := range meetings.Meetings {
		if meeting.Key == meetingKey {
			for _, session := range meeting.Sessions {
				keys = append(keys, session.Key)
			}
		}
	}
	return keys
}

func GetLatestMeetingKey(meetings Meetings) int {
	return GetMeetingKeys(meetings)[0]
}

func GetLatestSessionKey(meetings Meetings, meetingKey int) int {
	return GetSessionKeys(meetings, meetingKey)[0]
}*/
