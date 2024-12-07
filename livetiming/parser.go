package livetiming

import (
	"encoding/json"
	"fmt"
	"strconv"
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
		return positions, fmt.Errorf("Incorrect input data")
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
