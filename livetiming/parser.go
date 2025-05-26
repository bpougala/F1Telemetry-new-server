package livetiming

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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
		return buildUpdatedSession(data)
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
		drivers = append(drivers, driver)
	}
	return drivers, nil
}
func BuildTrackStatus(data []byte) ([]TrackStatus, error) {
	var trackStatus []TrackStatus
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return trackStatus, err
	}
	var initialData InitialSessionData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return trackStatus, fmt.Errorf("failed to unmarshal data: %v", err)
	}
	if initialData.R.SessionData.StatusSeries == nil {
		return trackStatus, fmt.Errorf("no track status found")
	}
	statusSeries := initialData.R.SessionData.StatusSeries
	if len(statusSeries) == 0 {
		return trackStatus, fmt.Errorf("no track status found")
	}
	for _, status := range statusSeries {
		var trackStatusItem TrackStatus
		trackStatusItem.Status = status.TrackStatus
		trackStatusItem.Utc = status.Utc
		if trackStatusItem.Status != "" {
			trackStatus = append(trackStatus, trackStatusItem)
		}
	}
	return trackStatus, nil
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
		position.Position = &value.Line
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
		position.Position = &value.Line
		position.InPit = value.InPit
		position.PitOut = value.PitOut
		position.Stopped = value.Stopped
		position.Status = value.Status
		position.Retired = value.Retired
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

func buildFinalSession(elements []interface{}) (SessionInfoDB, error) {
	raw, ok := elements[1].(map[string]interface{})
	if !ok {
		var sessionInfo SessionInfoDB
		return sessionInfo, fmt.Errorf("unexpected type for elements[1], expected map[string]interface{}")
	}
	statusSeries := raw
	value, ok := statusSeries["StatusSeries"].(map[string]interface{})
	if !ok {
		var sessionInfo SessionInfoDB
		return sessionInfo, fmt.Errorf("unexpected type for StatusSeries, expected map[string]interface{}")
	}
	var sessionInfo SessionInfoDB
	for _, v := range value {
		var statusInfoStruct SessionStatusStruct
		err := mapstructure.Decode(v, &statusInfoStruct)
		if err != nil {
			return sessionInfo, err
		}
		if statusInfoStruct.Utc == "" || statusInfoStruct.SessionStatus == "" {
			return sessionInfo, fmt.Errorf("no session info data available")
		}
		sessionInfo.ArchiveStatus = statusInfoStruct.SessionStatus
	}
	return sessionInfo, nil
}

func buildUpdatedSession(data []byte) (SessionInfoDB, error) {
	var updatedData UpdateData
	var updatedSession SessionInfoDB
	if err := json.Unmarshal(data, &updatedData); err != nil {
		return updatedSession, err
	}
	if len(updatedData.M) == 0 {
		return updatedSession, fmt.Errorf("no data available")
	}
	for _, message := range updatedData.M {
		elements := message.A
		if len(elements) != 3 {
			continue
		}
		if elements[0] != "SessionInfo" {
			return buildFinalSession(elements)
		}
		sessionDataMap := elements[1].(map[string]interface{})
		sessionDataMapBytes, err := json.Marshal(sessionDataMap)
		if err != nil {
			return updatedSession, fmt.Errorf("could not parse session data map as bytes")
		}
		var sessionInfo SessionInfo
		if err := json.Unmarshal(sessionDataMapBytes, &sessionInfo); err != nil {
			return updatedSession, fmt.Errorf("could not unmarshal session data map bytes to json")
		}
		updatedSession.ArchiveStatus = sessionInfo.ArchiveStatus.Status
		updatedSession.Key = sessionInfo.Key
		updatedSession.StartDate = sessionInfo.StartDate
		updatedSession.EndDate = sessionInfo.EndDate
		updatedSession.Type = sessionInfo.Type
		updatedSession.GmtOffset = sessionInfo.GmtOffset
		updatedSession.MeetingKey = sessionInfo.Meeting.Key
		updatedSession.Name = sessionInfo.Name
		updatedSession.Path = sessionInfo.Path
	}
	return updatedSession, nil
}

func buildTimingDataPositionsUpdate(message Message) ([]Position, error) {
	var positions []Position
	elements := message.A
	if elements[0] != "TimingData" {
		return positions, fmt.Errorf("incorrect input data")
	}
	var timingData TimingData
	timingDataMap, err := json.Marshal(elements[1])
	if err != nil {
		return positions, fmt.Errorf("could not parse timing data map as bytes")
	}
	err = json.Unmarshal(timingDataMap, &timingData)
	if err != nil {
		return positions, err
	}
	for key, value := range timingData.Lines {
		var rawPosition RawPosition
		err = mapstructure.Decode(value, &rawPosition)
		if err != nil {
			continue
		}
		positionNumber, _ := strconv.Atoi(rawPosition.Position)
		racingNumber, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		position := Position{
			RacingNumber: racingNumber,
			Position:     &positionNumber,
			InPit:        rawPosition.InPit,
			PitOut:       rawPosition.PitOut,
			Stopped:      rawPosition.Stopped,
			Status:       rawPosition.Status,
			Retired:      rawPosition.Retired,
		}
		positions = append(positions, position)
	}
	if len(positions) == 0 {
		return nil, fmt.Errorf("no positions found")
	}
	return positions, nil
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
			return buildTimingDataPositionsUpdate(message)
		}
		var positionData map[string]PositionLine
		err := mapstructure.Decode(elements[1], &positionData)
		if err != nil {
			continue
		}
		for key, value := range positionData {
			var position Position
			position.RacingNumber, _ = strconv.Atoi(key)
			position.Position = &value.Line
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
	Messages map[string]interface{}
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
		rawMessagesMap, ok := elements[1].(map[string]interface{})
		if !ok {
			continue
		}
		for _, value := range rawMessagesMap["Messages"].(map[string]interface{}) {
			var raceControl RaceControl
			err := mapstructure.Decode(value, &raceControl)
			if err != nil {
				continue
			}
			raceControlMessages = append(raceControlMessages, raceControl)
		}
	}
	if len(raceControlMessages) == 0 {
		return raceControlMessages, fmt.Errorf("no race control messages")
	}
	return raceControlMessages, nil
}

func buildRaceStints(data []byte) ([]Stint, error) {
	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return nil, err
	}
	var stints []Stint
	if initialData.R.TimingAppData.Lines == nil {
		return nil, fmt.Errorf("no stints found")
	}
	for _, value := range initialData.R.TimingAppData.Lines {
		var stint Stint
		stint.RacingNumber, _ = strconv.Atoi(value.RacingNumber)
		for count, rawStint := range value.Stints {
			stint.StintNumber = count
			stint.Compound = rawStint.Compound
			stint.LapFlags = rawStint.LapFlags
			stint.New, _ = strconv.ParseBool(rawStint.New)
			stint.StartLaps = rawStint.StartLaps
			stint.TotalLaps = rawStint.TotalLaps
			stint.TyresNotChanged, _ = strconv.Atoi(rawStint.TyresNotChanged)
			stint.Timestamp = rawStint.Timestamp
			stints = append(stints, stint)
		}
	}
	if len(stints) == 0 {
		return nil, fmt.Errorf("no stints found")
	}
	return stints, nil
}

func BuildStints(data []byte) ([]Stint, error) {
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return nil, err
	}
	var stints []Stint
	if updateData.M == nil {
		return buildRaceStints(data)
	}
	for _, message := range updateData.M {
		elements := message.A
		if len(elements) != 3 {
			continue
		}
		if elements[0] != "TimingAppData" {
			continue
		}
		var stintData StintData
		stintDataMap := elements[1].(map[string]interface{})
		stintDataMapBytes, err := json.Marshal(stintDataMap)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(stintDataMapBytes, &stintData); err != nil {
			continue
		}
		for racingNumber, value := range stintData.StintLines {
			for stintNumber, rawStint := range value.Stints {
				var stint Stint
				stint.RacingNumber, _ = strconv.Atoi(racingNumber)
				stint.StintNumber, _ = strconv.Atoi(stintNumber)
				stint.Compound = rawStint.Compound
				stint.LapFlags = rawStint.LapFlags
				stint.New, _ = strconv.ParseBool(rawStint.New)
				stint.StartLaps = rawStint.StartLaps
				stint.TotalLaps = rawStint.TotalLaps
				stint.TyresNotChanged, _ = strconv.Atoi(rawStint.TyresNotChanged)
				stint.Timestamp = rawStint.Timestamp
				stints = append(stints, stint)
			}
		}
	}
	if len(stints) == 0 {
		return stints, fmt.Errorf("no stints found")
	}
	return stints, nil
}

func BuildSectors(data []byte) ([]AllSectors, error) {
	var sectors []AllSectors
	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return sectors, err
	}
	if initialData.R.TimingData.Lines == nil {
		return buildRaceSectors(data)
	}
	for key, value := range initialData.R.TimingData.Lines {
		if value.Sectors == nil {
			continue
		}
		var sector AllSectors
		sector.RacingNumber, _ = strconv.Atoi(key)
		var firstSector Sector
		rawFirstSector := value.Sectors[0]

		firstSector.Value = rawFirstSector.Value
		firstSector.OverallFastest = rawFirstSector.OverallFastest
		firstSector.PersonalFastest = rawFirstSector.PersonalFastest
		firstSector.SectorNumber = 1
		firstSector.LapNumber = value.NumberOfLaps
		if firstSector.Value != "" {
			sector.Sectors = append(sector.Sectors, firstSector)
		}

		var secondSector Sector
		rawSecondSector := value.Sectors[1]
		secondSector.Value = rawSecondSector.Value
		secondSector.OverallFastest = rawSecondSector.OverallFastest
		secondSector.PersonalFastest = rawSecondSector.PersonalFastest
		secondSector.SectorNumber = 2
		secondSector.LapNumber = value.NumberOfLaps
		if secondSector.Value != "" {
			sector.Sectors = append(sector.Sectors, secondSector)
		}

		var thirdSector Sector
		rawThirdSector := value.Sectors[2]
		thirdSector.Value = rawThirdSector.Value
		thirdSector.OverallFastest = rawThirdSector.OverallFastest
		thirdSector.PersonalFastest = rawThirdSector.PersonalFastest
		thirdSector.SectorNumber = 3
		thirdSector.LapNumber = value.NumberOfLaps
		if thirdSector.Value != "" {
			sector.Sectors = append(sector.Sectors, thirdSector)
		}
		sectors = append(sectors, sector)
	}

	return sectors, nil
}

func buildRaceSectors(data []byte) ([]AllSectors, error) {
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return nil, err
	}
	if updateData.M == nil {
		return nil, fmt.Errorf("no sectors found")
	}
	var sectors []AllSectors
	for _, message := range updateData.M {
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
			var sectorTimes UpdateSector
			err = mapstructure.Decode(value, &sectorTimes)
			if err != nil {
				continue
			}
			if sectorTimes.Sectors == nil {
				continue
			}
			var driverSectors AllSectors
			driverSectors.RacingNumber, _ = strconv.Atoi(key)
			driverSectors.Utc, _ = time.Parse(time.RFC3339, elements[2].(string))
			for sectorNumber, sectorTime := range sectorTimes.Sectors {
				if sectorTime.Value == "" {
					continue
				}
				var sector Sector
				sector.SectorNumber, _ = strconv.Atoi(sectorNumber)
				sector.Value = sectorTime.Value
				sector.OverallFastest = sectorTime.OverallFastest
				sector.PersonalFastest = sectorTime.PersonalFastest
				sector.LapNumber = sectorTime.LapNumber
				driverSectors.Sectors = append(driverSectors.Sectors, sector)
			}
			if driverSectors.Sectors != nil {
				sectors = append(sectors, driverSectors)
			}
		}
	}
	if len(sectors) == 0 {
		return nil, fmt.Errorf("no sectors found")
	}
	return sectors, nil
}
