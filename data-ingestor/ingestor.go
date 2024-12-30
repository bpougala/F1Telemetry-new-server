package data_ingestor

import (
	"F1Telemetry-new-server/data-ingestor/collections"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
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

func IngestPositionsForDB(sessionKey int) ([]interface{}, error) {
	uri := "https://livetiming.formula1.com/static/2024/2024-11-23_Las_Vegas_Grand_Prix/2024-11-23_Race/TimingData.json"
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching data: %s", resp.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	var rawTimings collections.RawInterval
	err = json.Unmarshal(body, &rawTimings)
	if err != nil {
		panic(err)
	}
	var positions []interface{}
	for _, timingLine := range rawTimings.Lines {
		var position collections.PositionsDB
		position.SessionKey = sessionKey
		position.RacingNumber, _ = strconv.Atoi(timingLine.RacingNumber)
		position.Position, _ = strconv.Atoi(timingLine.Position)
		positions = append(positions, position)
	}

	return positions, nil
}

func IngestTimings(sessionKey int) ([]interface{}, error) {
	uri := "https://livetiming.formula1.com/static/2024/2024-11-23_Las_Vegas_Grand_Prix/2024-11-23_Race/TimingData.json"
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching data: %s", resp.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	var rawTimings collections.RawInterval
	err = json.Unmarshal(body, &rawTimings)
	if err != nil {
		panic(err)
	}
	var timingsInterface []interface{}
	for _, timingLine := range rawTimings.Lines {
		var timing collections.Timing
		timing.SessionKey = sessionKey
		timing.RacingNumber, _ = strconv.Atoi(timingLine.RacingNumber)
		timing.Position, _ = strconv.Atoi(timingLine.Position)
		timing.Stopped = timingLine.Stopped
		timing.InPit = timingLine.InPit
		timing.PitOut = timingLine.PitOut
		timing.Status = timingLine.Status
		timing.NumberOfLaps = timingLine.NumberOfLaps
		timing.Sectors = []struct {
			Stopped         bool   `json:"stopped"`
			Value           string `json:"value"`
			Status          int    `json:"status"`
			OverallFastest  bool   `json:"overall_fastest"`
			PersonalFastest bool   `json:"personal_fastest"`
			Segments        []struct {
				Status int `json:"status"`
			} `json:"segments"`
		}(timingLine.Sectors)
		timing.BestLapTime = struct {
			Value string `json:"value"`
			Lap   int    `json:"lap"`
		}(timingLine.BestLapTime)
		timing.LastLapTime = struct {
			Value           string `json:"value"`
			Status          int    `json:"status"`
			OverallFastest  bool   `json:"overall_fastest"`
			PersonalFastest bool   `json:"personal_fastest"`
		}(timingLine.LastLapTime)
		timing.GapToLeader = timingLine.GapToLeader
		timing.IntervalToPositionAhead = struct {
			Value    string `json:"value"`
			Catching bool   `json:"catching"`
		}(timingLine.IntervalToPositionAhead)
		timing.Retired = timingLine.Retired
		timingsInterface = append(timingsInterface, timing)
	}
	return timingsInterface, nil
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
	if err != nil {
		return positions, err
	}
	err = json.Unmarshal(body, &positions)
	if err != nil {
		return positions, err
	}
	return positions, nil
}

type RaceControl struct {
	SessionKey int    `json:"sessionkey"`
	Utc        string `json:"utc"`
	Category   string `json:"category"`
	Flag       string `json:"flag"`
	Scope      string `json:"scope"`
	Message    string `json:"message"`
}

type RaceControlMessages struct {
	Messages []RaceControl `json:"Messages"`
}

func IngestRaceControl(sessionKey int) ([]interface{}, error) {
	uri := "https://livetiming.formula1.com/static/2024/2024-11-23_Las_Vegas_Grand_Prix/2024-11-23_Race/RaceControlMessages.json"
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching data: %s", resp.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	var raceControlInt []interface{}
	var data RaceControlMessages
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	for _, raceControlMessage := range data.Messages {
		var raceControl collections.RaceControl
		raceControl.SessionKey = sessionKey
		raceControl.Flag = raceControlMessage.Flag
		raceControl.Category = raceControlMessage.Category
		raceControl.Scope = raceControlMessage.Scope
		raceControl.Utc = raceControlMessage.Utc
		raceControl.Message = raceControlMessage.Message
		raceControlInt = append(raceControlInt, raceControl)
	}
	return raceControlInt, nil
}

func IngestLaps(sessionKey int) ([]collections.Laps, error) {
	url := fmt.Sprintf("%s/laps?session_key=%d", BASE_URL, sessionKey)
	var laps []collections.Laps
	resp, err := http.Get(url)
	retries := 3
	for retries > 0 {
		if err != nil {
			retries -= 1
		} else if resp.StatusCode != 200 {
			retries -= 1
		} else {
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					panic(err)
				}
			}(resp.Body)
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				retries -= 1
			} else {
				err = json.Unmarshal(body, &laps)
				return laps, err
			}
		}
	}
	return laps, err
}

func IngestIntervals(sessionKey int) ([]collections.Interval, error) {
	uri := fmt.Sprintf("%s/intervals?session_key=%d", BASE_URL, sessionKey)
	resp, err := http.Get(uri)
	var intervals []collections.Interval
	if err != nil {
		return intervals, err
	}
	if resp.StatusCode != 200 {
		return intervals, fmt.Errorf("error fetching data: %s", resp.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return intervals, err
	}
	err = json.Unmarshal(body, &intervals)
	if err != nil {
		return intervals, err
	}
	return intervals, nil
}

func IngestCarData(sessionKey int, driverNumber int) ([]collections.CarData, error) {
	uri := fmt.Sprintf("%s/car_data?session_key=%d&driver_number=%d", BASE_URL, sessionKey, driverNumber)
	client := http.Client{Timeout: 25 * time.Second}
	retries := 3
	var err error
	var carData []collections.CarData

	for retries > 0 {
		fmt.Printf("number of retries: %d\n", retries)
		resp, err := client.Get(uri)

		if err != nil {
			fmt.Printf("error fetching new data: %s\n", err.Error())
			retries -= 1
		} else if resp.StatusCode != 200 {
			fmt.Printf("status code is not 200: %s\n", resp.Status)
			retries -= 1
		} else {
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					panic(err)
				}
			}(resp.Body)
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("error reading body: %s\n", err.Error())
				retries -= 1
			} else {
				err = json.Unmarshal(body, &carData)
				if err != nil {
					fmt.Printf("error unmarshalling data: %s\n", err.Error())
					retries -= 1
				}
				return carData, nil
			}
		}
	}
	fmt.Println("what is going on")
	return carData, err
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
