package livetiming

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
)

// DecompressZData decodes a base64+zlib compressed string into raw JSON bytes.
func DecompressZData(compressed string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(compressed)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}
	r := flate.NewReader(bytes.NewReader(decoded))
	defer r.Close()
	decompressed, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("deflate decompress failed: %w", err)
	}
	return decompressed, nil
}

// DecompressCarData decompresses a CarData.z compressed string into a Root struct.
func DecompressCarData(compressed string) (Root, error) {
	var root Root
	data, err := DecompressZData(compressed)
	if err != nil {
		return root, err
	}
	if err := json.Unmarshal(data, &root); err != nil {
		return root, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return root, nil
}

// DecompressPositionData decompresses a Position.z compressed string into a PositionRoot struct.
func DecompressPositionData(compressed string) (PositionRoot, error) {
	var posRoot PositionRoot
	data, err := DecompressZData(compressed)
	if err != nil {
		return posRoot, err
	}
	if err := json.Unmarshal(data, &posRoot); err != nil {
		return posRoot, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return posRoot, nil
}

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
	// Try TimingData first — it has full status fields (Retired, Stopped, etc.)
	positions, err := buildRacePositions(data)
	if err == nil && len(positions) > 0 {
		return positions, nil
	}

	// Fall back to DriverList (position-only data)
	var rawDriverList DriverList
	if err := json.Unmarshal(data, &rawDriverList); err != nil {
		return nil, nil
	}
	if rawDriverList.R.DriverList == nil {
		return nil, fmt.Errorf("no position data found")
	}
	var driverListPositions []Position
	for key, value := range rawDriverList.R.DriverList {
		if key == "_kf" {
			continue
		}
		var position Position
		position.RacingNumber, _ = strconv.Atoi(key)
		position.Position = &value.Line
		position.PositionOnly = true
		driverListPositions = append(driverListPositions, position)
	}
	return driverListPositions, nil
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
		positionNumber, _ := strconv.Atoi(value.Position)
		if positionNumber == 0 {
			positionNumber = value.Line
		}
		if positionNumber == 0 {
			continue
		}
		var position Position
		position.RacingNumber, _ = strconv.Atoi(key)
		position.Position = &positionNumber
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
	// Parse as raw map to preserve qualifying fields (BestLapTimes, Stats, KnockedOut, Cutoff)
	var rawData struct {
		R struct {
			TimingData struct {
				Lines map[string]interface{} `json:"Lines"`
			} `json:"TimingData"`
		} `json:"R"`
	}
	var lapTimeMetrics []LapTimeMetric
	if err := json.Unmarshal(data, &rawData); err != nil {
		return lapTimeMetrics, err
	}
	if rawData.R.TimingData.Lines == nil {
		return lapTimeMetrics, fmt.Errorf("No timing data found")
	}
	for key, value := range rawData.R.TimingData.Lines {
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
		parseQualifyingFields(value, &timing)
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
		if positionNumber == 0 {
			positionNumber = rawPosition.Line
		}
		if positionNumber == 0 {
			continue // No position data in this update
		}
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
			position.PositionOnly = true
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
			if timing.BestLapTime.Lap == 0 && timing.BestLapTime.Value == "" {
				continue
			}
			parseQualifyingFields(value, &timing)
			timing.RacingNumber = key
			lapTimeMetrics = append(lapTimeMetrics, timing)
		}

	}
	if len(lapTimeMetrics) == 0 {
		return nil, fmt.Errorf("No lap time metrics found")
	}

	return lapTimeMetrics, nil
}

// parseQualifyingFields extracts BestLapTimes and Stats from the raw timing data map.
// These fields can be either arrays (initial snapshot) or maps keyed by index (updates).
func parseQualifyingFields(raw interface{}, timing *LapTimeMetric) {
	rawMap, ok := raw.(map[string]interface{})
	if !ok {
		return
	}

	// Parse BestLapTimes (can be array or indexed map)
	if bestLapTimes, ok := rawMap["BestLapTimes"]; ok {
		timing.QualifyingBestLaps = decodeQualifyingLapTimes(bestLapTimes)
	}

	// Parse Stats (can be array or indexed map)
	if stats, ok := rawMap["Stats"]; ok {
		timing.QualifyingStats = decodeQualifyingStats(stats)
	}
}

func decodeQualifyingLapTimes(raw interface{}) []QualifyingLapTime {
	switch v := raw.(type) {
	case []interface{}:
		result := make([]QualifyingLapTime, len(v))
		for i, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if val, ok := m["Value"].(string); ok {
					result[i].Value = val
				}
				if lap, ok := m["Lap"].(float64); ok {
					result[i].Lap = int(lap)
				}
			}
		}
		return result
	case map[string]interface{}:
		// Indexed map format from updates (e.g. {"0": {...}, "1": {...}})
		maxIdx := -1
		for k := range v {
			idx, err := strconv.Atoi(k)
			if err == nil && idx > maxIdx {
				maxIdx = idx
			}
		}
		if maxIdx < 0 {
			return nil
		}
		result := make([]QualifyingLapTime, maxIdx+1)
		for k, item := range v {
			idx, err := strconv.Atoi(k)
			if err != nil {
				continue
			}
			if m, ok := item.(map[string]interface{}); ok {
				if val, ok := m["Value"].(string); ok {
					result[idx].Value = val
				}
				if lap, ok := m["Lap"].(float64); ok {
					result[idx].Lap = int(lap)
				}
			}
		}
		return result
	}
	return nil
}

func decodeQualifyingStats(raw interface{}) []QualifyingStats {
	switch v := raw.(type) {
	case []interface{}:
		result := make([]QualifyingStats, len(v))
		for i, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if val, ok := m["TimeDiffToFastest"].(string); ok {
					result[i].TimeDiffToFastest = val
				}
				if val, ok := m["TimeDifftoPositionAhead"].(string); ok {
					result[i].TimeDifftoPositionAhead = val
				}
			}
		}
		return result
	case map[string]interface{}:
		maxIdx := -1
		for k := range v {
			idx, err := strconv.Atoi(k)
			if err == nil && idx > maxIdx {
				maxIdx = idx
			}
		}
		if maxIdx < 0 {
			return nil
		}
		result := make([]QualifyingStats, maxIdx+1)
		for k, item := range v {
			idx, err := strconv.Atoi(k)
			if err != nil {
				continue
			}
			if m, ok := item.(map[string]interface{}); ok {
				if val, ok := m["TimeDiffToFastest"].(string); ok {
					result[idx].TimeDiffToFastest = val
				}
				if val, ok := m["TimeDifftoPositionAhead"].(string); ok {
					result[idx].TimeDifftoPositionAhead = val
				}
			}
		}
		return result
	}
	return nil
}

// BuildQualifyingState extracts session-level qualifying metadata from TimingData.
// Returns nil when the data does not contain qualifying fields (e.g. race sessions).
func BuildQualifyingState(data []byte) (*QualifyingState, error) {
	// Try initial snapshot format
	var initialData struct {
		R struct {
			TimingData TimingData `json:"TimingData"`
		} `json:"R"`
	}
	if err := json.Unmarshal(data, &initialData); err == nil && initialData.R.TimingData.SessionPart > 0 {
		return &QualifyingState{
			SessionPart:      initialData.R.TimingData.SessionPart,
			NoEntries:        initialData.R.TimingData.NoEntries,
			CutOffTime:       initialData.R.TimingData.CutOffTime,
			CutOffPercentage: initialData.R.TimingData.CutOffPercentage,
		}, nil
	}

	// Try update message format
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return nil, nil
	}
	for _, message := range updateData.M {
		elements := message.A
		if len(elements) < 2 {
			continue
		}
		if elements[0] != "TimingData" {
			continue
		}
		timingDataMap, ok := elements[1].(map[string]interface{})
		if !ok {
			continue
		}
		var td TimingData
		tdBytes, err := json.Marshal(timingDataMap)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(tdBytes, &td); err != nil {
			continue
		}
		if td.SessionPart > 0 {
			return &QualifyingState{
				SessionPart:      td.SessionPart,
				NoEntries:        td.NoEntries,
				CutOffTime:       td.CutOffTime,
				CutOffPercentage: td.CutOffPercentage,
			}, nil
		}
	}
	return nil, nil
}

// BuildQualifyingParts extracts qualifying phase transition timestamps from SessionData.Series.
func BuildQualifyingParts(data []byte) ([]QualifyingPart, error) {
	// Try initial snapshot format (miami.json style)
	var initial struct {
		R struct {
			SessionData struct {
				Series []QualifyingPart `json:"Series"`
			} `json:"SessionData"`
		} `json:"R"`
	}
	if err := json.Unmarshal(data, &initial); err == nil && len(initial.R.SessionData.Series) > 0 {
		var parts []QualifyingPart
		for _, s := range initial.R.SessionData.Series {
			if s.QualifyingPart > 0 {
				parts = append(parts, s)
			}
		}
		if len(parts) > 0 {
			return parts, nil
		}
	}

	// Try update message format
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return nil, nil
	}
	for _, message := range updateData.M {
		elements := message.A
		if len(elements) < 2 {
			continue
		}
		if elements[0] != "SessionData" {
			continue
		}
		sessionDataMap, ok := elements[1].(map[string]interface{})
		if !ok {
			continue
		}
		seriesRaw, ok := sessionDataMap["Series"]
		if !ok {
			continue
		}
		seriesBytes, err := json.Marshal(seriesRaw)
		if err != nil {
			continue
		}
		var seriesMap map[string]QualifyingPart
		if err := json.Unmarshal(seriesBytes, &seriesMap); err != nil {
			// Try array format
			var seriesArr []QualifyingPart
			if err := json.Unmarshal(seriesBytes, &seriesArr); err != nil {
				continue
			}
			var parts []QualifyingPart
			for _, s := range seriesArr {
				if s.QualifyingPart > 0 {
					parts = append(parts, s)
				}
			}
			if len(parts) > 0 {
				return parts, nil
			}
			continue
		}
		var parts []QualifyingPart
		for _, s := range seriesMap {
			if s.QualifyingPart > 0 {
				parts = append(parts, s)
			}
		}
		if len(parts) > 0 {
			return parts, nil
		}
	}
	return nil, nil
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
		for i, rawSector := range value.Sectors {
			var s Sector
			s.Value = rawSector.Value
			s.OverallFastest = rawSector.OverallFastest
			s.PersonalFastest = rawSector.PersonalFastest
			s.SectorNumber = i + 1
			s.LapNumber = value.NumberOfLaps
			if s.Value != "" {
				sector.Sectors = append(sector.Sectors, s)
			}
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

func BuildSegments(data []byte) ([]AllSegments, error) {
	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return nil, err
	}
	if initialData.R.TimingData.Lines == nil {
		return buildRaceSegments(data)
	}
	var allSegments []AllSegments
	for key, value := range initialData.R.TimingData.Lines {
		if value.Sectors == nil {
			continue
		}
		var driverSegments AllSegments
		driverSegments.RacingNumber, _ = strconv.Atoi(key)
		for sectorIdx, sector := range value.Sectors {
			for segIdx, seg := range sector.Segments {
				driverSegments.Segments = append(driverSegments.Segments, SegmentStatus{
					SectorNumber:  sectorIdx + 1,
					SegmentNumber: segIdx + 1,
					Status:        seg.Status,
				})
			}
		}
		if len(driverSegments.Segments) > 0 {
			allSegments = append(allSegments, driverSegments)
		}
	}
	if len(allSegments) == 0 {
		return nil, fmt.Errorf("no segments found")
	}
	return allSegments, nil
}

func buildRaceSegments(data []byte) ([]AllSegments, error) {
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return nil, err
	}
	if updateData.M == nil {
		return nil, fmt.Errorf("no segments found")
	}
	var allSegments []AllSegments
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
			var driverSegments AllSegments
			driverSegments.RacingNumber, _ = strconv.Atoi(key)
			driverSegments.Utc, _ = time.Parse(time.RFC3339, elements[2].(string))
			for sectorNumber, sectorEntry := range sectorTimes.Sectors {
				if sectorEntry.Segments == nil {
					continue
				}
				sectorNum, _ := strconv.Atoi(sectorNumber)
				for segNumber, segEntry := range sectorEntry.Segments {
					segNum, _ := strconv.Atoi(segNumber)
					driverSegments.Segments = append(driverSegments.Segments, SegmentStatus{
						SectorNumber:  sectorNum,
						SegmentNumber: segNum,
						Status:        segEntry.Status,
					})
				}
			}
			if len(driverSegments.Segments) > 0 {
				allSegments = append(allSegments, driverSegments)
			}
		}
	}
	if len(allSegments) == 0 {
		return nil, fmt.Errorf("no segments found")
	}
	return allSegments, nil
}

// ExtractPositionZCompressed extracts the raw compressed Position.z string without decompressing.
func ExtractPositionZCompressed(data []byte) (string, error) {
	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return "", err
	}
	compressed := initialData.R.PositionZ
	if compressed != "" {
		return compressed, nil
	}
	// Try extracting from update messages
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return "", err
	}
	for _, message := range updateData.M {
		elements := message.A
		if len(elements) < 2 {
			continue
		}
		if elements[0] != "Position.z" {
			continue
		}
		compressedStr, ok := elements[1].(string)
		if !ok {
			continue
		}
		return compressedStr, nil
	}
	return "", fmt.Errorf("no position data found")
}

// BuildPositionZ parses Position.z data from the initial snapshot.
func BuildPositionZ(data []byte) (PositionRoot, error) {
	var posRoot PositionRoot
	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return posRoot, err
	}
	compressed := initialData.R.PositionZ
	if compressed == "" {
		return buildUpdatePositionZ(data)
	}
	return DecompressPositionData(compressed)
}

func buildUpdatePositionZ(data []byte) (PositionRoot, error) {
	var posRoot PositionRoot
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return posRoot, err
	}
	for _, message := range updateData.M {
		elements := message.A
		if len(elements) < 2 {
			continue
		}
		if elements[0] != "Position.z" {
			continue
		}
		compressed, ok := elements[1].(string)
		if !ok {
			continue
		}
		return DecompressPositionData(compressed)
	}
	return posRoot, fmt.Errorf("no position data found")
}

// BuildLapCount parses LapCount from the initial snapshot.
func BuildLapCount(data []byte) (*LapCountData, error) {
	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return nil, err
	}
	lc := initialData.R.LapCount
	if lc.TotalLaps == 0 && lc.CurrentLap == 0 {
		return nil, fmt.Errorf("no lap count data found")
	}
	return &LapCountData{
		CurrentLap: lc.CurrentLap,
		TotalLaps:  lc.TotalLaps,
	}, nil
}

// BuildLapCountUpdate parses LapCount from update messages.
func BuildLapCountUpdate(data []byte) (*LapCountData, error) {
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return nil, err
	}
	for _, message := range updateData.M {
		elements := message.A
		if len(elements) < 2 {
			continue
		}
		if elements[0] != "LapCount" {
			continue
		}
		lcMap, ok := elements[1].(map[string]interface{})
		if !ok {
			continue
		}
		lcBytes, err := json.Marshal(lcMap)
		if err != nil {
			continue
		}
		var lc LapCountData
		if err := json.Unmarshal(lcBytes, &lc); err != nil {
			continue
		}
		return &lc, nil
	}
	return nil, fmt.Errorf("no lap count data found")
}

// BuildWeatherData parses WeatherData from the initial snapshot.
func BuildWeatherData(data []byte) (*Weather, error) {
	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return nil, err
	}
	w := initialData.R.WeatherData
	if w.AirTemp == "" && w.TrackTemp == "" {
		return nil, fmt.Errorf("no weather data found")
	}
	return &w, nil
}

// BuildWeatherDataUpdate parses WeatherData from update messages.
func BuildWeatherDataUpdate(data []byte) (*Weather, error) {
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return nil, err
	}
	for _, message := range updateData.M {
		elements := message.A
		if len(elements) < 2 {
			continue
		}
		if elements[0] != "WeatherData" {
			continue
		}
		weatherMap, ok := elements[1].(map[string]interface{})
		if !ok {
			continue
		}
		weatherBytes, err := json.Marshal(weatherMap)
		if err != nil {
			continue
		}
		var w Weather
		if err := json.Unmarshal(weatherBytes, &w); err != nil {
			continue
		}
		return &w, nil
	}
	return nil, fmt.Errorf("no weather data found")
}

// BuildTrackStatusUpdate parses TrackStatus topic update messages.
func BuildTrackStatusUpdate(data []byte) ([]TrackStatus, error) {
	var updateData UpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return nil, err
	}
	var statuses []TrackStatus
	for _, message := range updateData.M {
		elements := message.A
		if len(elements) < 2 {
			continue
		}
		if elements[0] != "TrackStatus" {
			continue
		}
		statusMap, ok := elements[1].(map[string]interface{})
		if !ok {
			continue
		}
		statusBytes, err := json.Marshal(statusMap)
		if err != nil {
			continue
		}
		var ts TrackStatus
		if err := json.Unmarshal(statusBytes, &ts); err != nil {
			continue
		}
		if ts.Status != "" {
			statuses = append(statuses, ts)
		}
	}
	if len(statuses) == 0 {
		return nil, fmt.Errorf("no track status found")
	}
	return statuses, nil
}
