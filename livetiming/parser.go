package livetiming

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

func BuildMeetingData(data []byte) (MeetingData, error) {
	var meetingData MeetingData
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return meetingData, err
	}
	var initialSessionData InitialSessionData
	if err := json.Unmarshal(data, &initialSessionData); err != nil {
		return meetingData, fmt.Errorf("failed to unmarshal data: %v", err)
	}
	if initialSessionData.R.SessionInfo.Meeting.Key == 0 {
		return meetingData, fmt.Errorf("Incorrect input data")
	}
	meeting := initialSessionData.R.SessionInfo.Meeting
	meetingData.Key = meeting.Key
	meetingData.Name = meeting.Name
	meetingData.OfficialName = meeting.OfficialName
	meetingData.Location = meeting.Location
	meetingData.Number = meeting.Number
	meetingData.Country = meeting.Country
	meetingData.Circuit = meeting.Circuit
	return meetingData, nil
}

func BuildSessionInfo(data []byte) (SessionInfoDB, error) {
	var sessionInfo SessionInfoDB
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return sessionInfo, err
	}
	var initialSessionData InitialSessionData
	if err := json.Unmarshal(data, &initialSessionData); err != nil {
		return sessionInfo, fmt.Errorf("failed to unmarshal data: %v", err)
	}
	if initialSessionData.R.SessionInfo.Key == 0 {
		return sessionInfo, fmt.Errorf("Incorrect input data")
	}
	session := initialSessionData.R.SessionInfo
	sessionInfo.ArchiveStatus = session.ArchiveStatus.Status
	sessionInfo.StartDate = session.StartDate
	sessionInfo.EndDate = session.EndDate
	sessionInfo.Type = session.Type
	sessionInfo.GmtOffset = session.GmtOffset
	sessionInfo.Key = session.Key
	sessionInfo.MeetingKey = session.Meeting.Key
	sessionInfo.Name = session.Name
	sessionInfo.Path = session.Path
	return sessionInfo, nil
}

func BuildDriverList(data []byte) ([]Driver, error) {
	var drivers []Driver
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return drivers, err
	}
	var rawDriverList InitialSessionData
	if err := json.Unmarshal(data, &rawDriverList); err != nil {
		return drivers, err
	}
	if rawDriverList.R.DriverList == nil {
		return drivers, fmt.Errorf("Incorrect input data")
	}
	for key, value := range rawDriverList.R.DriverList {
		if key == "_kf" {
			continue
		}
		var driver Driver
		driver.RacingNumber, _ = strconv.Atoi(key)
		driver.BroadcastName = value.BroadcastName
		driver.FullName = value.FullName
		driver.Abbreviation = value.Tla
		driver.TeamName = value.TeamName
		driver.TeamColour = value.TeamColour
		driver.FirstName = value.FirstName
		driver.LastName = value.LastName
		driver.HeadshotUrl = value.HeadshotUrl
		driver.CountryCode = value.CountryCode
		drivers = append(drivers, driver)
	}
	return drivers, nil
}

func BuildPositions(data []byte) ([]Position, error) {
	var positions []Position
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return positions, err
	}
	var rawDriverList DriverList
	if err := json.Unmarshal(data, &rawDriverList); err != nil {
		return positions, nil
	}
	if rawDriverList.R.DriverList == nil {
		return buildRacePositions(data)
	}
	for key, value := range rawDriverList.R.DriverList {
		if key == "_kf" {
			continue
		}
		var position Position
		position.RacingNumber, _ = strconv.Atoi(key)
		position.Position = value.Line
		positions = append(positions, position)
	}
	return positions, nil
}

func buildRacePositions(data []byte) ([]Position, error) {
	var initialData InitialData
	var positions []Position
	if err := json.Unmarshal(data, &initialData); err != nil {
		return positions, err
	}
	if initialData.R.TimingData.Lines == nil {
		return buildUpdatedPositions(data)
	}

	for key, value := range initialData.R.TimingData.Lines {
		if key == "_kf" {
			continue
		}
		var position Position
		position.RacingNumber, _ = strconv.Atoi(key)
		position.Position = value.Line
		position.InPit = value.InPit
		position.PitOut = value.PitOut
		position.Stopped = value.Stopped
		position.Status = value.Status
		positions = append(positions, position)
	}
	return positions, nil
}

func buildRaceTimingData(data []byte) ([]LapTimeMetric, error) {
	var initialData InitialData
	var lapTimeMetrics []LapTimeMetric
	if err := json.Unmarshal(data, &initialData); err != nil {
		return lapTimeMetrics, err
	}
	if initialData.R.TimingData.Lines == nil {
		return lapTimeMetrics, fmt.Errorf("No timing data found")
	}
	for key, value := range initialData.R.TimingData.Lines {
		if key == "_kf" {
			continue
		}
		var timing LapTimeMetric
		err := mapstructure.Decode(value, &timing)
		if err != nil {
			continue
		}
		if timing.BestLapTime.Lap == 0 && timing.BestLapTime.Value == "" {
			continue
		}
		timing.RacingNumber = key
		lapTimeMetrics = append(lapTimeMetrics, timing)
	}
	if len(lapTimeMetrics) == 0 {
		return lapTimeMetrics, fmt.Errorf("No lap time metrics found")
	}
	return lapTimeMetrics, nil
}

func buildUpdatedPositions(data []byte) ([]Position, error) {
	var updateTimingData UpdateData
	if err := json.Unmarshal(data, &updateTimingData); err != nil {
		return nil, err
	}
	var positions []Position
	for _, message := range updateTimingData.M {
		elements := message.A
		if len(elements) != 3 {
			continue
		}
		if elements[0] != "DriverList" {
			continue
		}
		var positionData map[string]PositionLine
		err := mapstructure.Decode(elements[1], &positionData)
		if err != nil {
			continue
		}
		for key, value := range positionData {
			var position Position
			position.RacingNumber, _ = strconv.Atoi(key)
			position.Position = value.Line
			positions = append(positions, position)
		}
	}
	if len(positions) == 0 {
		return nil, fmt.Errorf("No positions found")
	}
	return positions, nil
}

func BuildTimingData(data []byte) ([]LapTimeMetric, error) {
	var updateTimingData UpdateData
	if err := json.Unmarshal(data, &updateTimingData); err != nil {
		return nil, err
	}
	if updateTimingData.M == nil {
		return buildRaceTimingData(data)
	}
	var lapTimeMetrics []LapTimeMetric
	for _, message := range updateTimingData.M {
		elements := message.A
		if len(elements) != 3 {
			continue
		}
		if elements[0] != "TimingData" {
			continue
		}
		timingDataMap := elements[1].(map[string]interface{})
		timingDataMapBytes, err := json.Marshal(timingDataMap)
		if err != nil {
			continue
		}
		var timingData TimingData
		if err := json.Unmarshal(timingDataMapBytes, &timingData); err != nil {
			continue
		}
		for key, value := range timingData.Lines {
			var timing LapTimeMetric
			err := mapstructure.Decode(value, &timing)
			if err != nil {
				continue
			}
			if timing.BestLapTime.Lap == 0 || timing.BestLapTime.Value == "" {
				continue
			}
			timing.RacingNumber = key
			lapTimeMetrics = append(lapTimeMetrics, timing)
		}

	}
	if len(lapTimeMetrics) == 0 {
		return nil, fmt.Errorf("No lap time metrics found")
	}

	return lapTimeMetrics, nil
}

func BuildCarData(data []byte) (CarData, error) {
	var carData CarData
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return carData, err
	}
	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return carData, err
	}
	compressedCarData := initialData.R.CarData
	if compressedCarData == "" {
		return buildUpdateCarData(data)
	}
	carData.Compressed = compressedCarData
	return carData, nil
}

func buildUpdateCarData(data []byte) (CarData, error) {
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return CarData{}, err
	}
	var carData CarData
	for _, message := range updateData.M {
		elements := message.A
		if len(elements) != 3 {
			continue
		}
		if elements[0] != "CarData.z" {
			continue
		}
		carData.Compressed = elements[1].(string)
	}
	if carData.Compressed == "" {
		return CarData{}, fmt.Errorf("no car data found")
	}
	return carData, nil
}

func BuildRaceControl(data []byte) ([]RaceControl, error) {
	var raceControlMessages []RaceControl
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return raceControlMessages, err
	}
	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return raceControlMessages, err
	}
	if initialData.R.RaceControlMessages.Messages == nil {
		return buildUpdateRaceControl(data)
	}
	for _, message := range initialData.R.RaceControlMessages.Messages {
		var raceControl RaceControl
		err := mapstructure.Decode(message, &raceControl)
		if err != nil {
			continue
		}
		raceControlMessages = append(raceControlMessages, raceControl)
	}
	return raceControlMessages, nil
}

type RaceControlMessages struct {
	Messages []RaceControl `json:"Messages"`
}

func buildUpdateRaceControl(data []byte) ([]RaceControl, error) {
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return nil, err
	}
	var raceControlMessages []RaceControl
	for _, message := range updateData.M {
		elements := message.A
		if len(elements) != 3 {
			continue
		}
		if elements[0] != "RaceControlMessages" {
			continue
		}
		var messages RaceControlMessages
		err := mapstructure.Decode(elements[1], &messages)
		if err != nil {
			fmt.Println(err)
			continue
		}
		raceControlMessages = append(raceControlMessages, messages.Messages...)
	}
	if len(raceControlMessages) == 0 {
		return raceControlMessages, fmt.Errorf("no race control messages")
	}
	return raceControlMessages, nil
}
