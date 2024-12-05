package livetiming

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func (stint *Stint) UnmarshalJSON(data []byte) error {
	var rawStintData map[string]StintLine
	if err := json.Unmarshal(data, &rawStintData); err != nil {
		return err
	}
	for _, value := range rawStintData {
		for _, rawStint := range value.Stints {
			if rawStint.Compound != "" {
				stint.LapFlags = rawStint.LapFlags
				stint.Compound = rawStint.Compound
				stint.New = rawStint.New
				stint.TyresNotChanged = rawStint.TyresNotChanged
				stint.TotalLaps = rawStint.TotalLaps
				stint.StartLaps = rawStint.StartLaps
			}
		}
	}
	return nil
}

func ParseInterfaceAsStint(rawStintData interface{}) (Stint, error) {
	var stint Stint
	if rawStintDataMap, ok := rawStintData.(map[string]interface{}); ok {
		for _, value := range rawStintDataMap {
			if stintLine, ok := value.(map[string]interface{}); ok {
				for key, value := range stintLine {
					stint.DriverNumber, _ = strconv.Atoi(key)
					for _, v := range value.(map[string]interface{}) {
						if _, ok := v.(float64); !ok {
							for _, rawStint := range v.(map[string]interface{}) {
								if stintMap, ok := rawStint.(map[string]interface{}); ok {
									if lapFlags, ok := stintMap["LapFlags"]; ok {
										stint.LapFlags = int(lapFlags.(float64))
									}
									if compound, ok := stintMap["Compound"]; ok {
										if compound.(string) == "" {
											continue
										}
										stint.Compound = compound.(string)
									} else {
										return stint, fmt.Errorf("no compound")
									}
									if isNew, ok := stintMap["New"]; ok {
										stint.New = isNew.(string)
									}
									if tyresNotChanged, ok := stintMap["TyresNotChanged"]; ok {
										stint.TyresNotChanged = tyresNotChanged.(string)
									}
									if totalLaps, ok := stintMap["TotalLaps"]; ok {
										stint.TotalLaps = int(totalLaps.(float64))
									}
									if startLaps, ok := stintMap["StartLaps"]; ok {
										stint.StartLaps = int(startLaps.(float64))
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return stint, nil
}

func ParseInterfaceAsPosition(rawPositionData interface{}) (Position, error) {
	var position Position
	if rawPositionDataMap, ok := rawPositionData.(map[string]interface{}); ok {
		for key, value := range rawPositionDataMap {
			position.DriverNumber, _ = strconv.Atoi(key)
			if positionLine, ok := value.(map[string]interface{}); ok {
				if positionLine["Line"] != nil {
					position.Position = int(positionLine["Line"].(float64))
				} else {
					return position, fmt.Errorf("no position")
				}
			}
		}
	}
	return position, nil
}

func (position *Position) UnmarshalJSON(data []byte) error {
	var rawPositionData map[string]PositionLine
	if err := json.Unmarshal(data, &rawPositionData); err != nil {
		return err
	}
	for key, value := range rawPositionData {
		position.Position = value.Line
		position.DriverNumber, _ = strconv.Atoi(key)
	}
	return nil
}

func isPositionData(data interface{}) bool {
	if obj, ok := data.(map[string]interface{}); ok {
		if lines, ok := obj["Lines"].(map[string]interface{}); ok {
			for _, value := range lines {
				if entry, ok := value.(map[string]interface{}); ok {
					if _, ok := entry["Line"]; ok {
						return true
					}
				}
			}
		}
	}
	return false
}

func isStintData(data interface{}) bool {
	if obj, ok := data.(map[string]interface{}); ok {
		if lines, ok := obj["Lines"].(map[string]interface{}); ok {
			for _, value := range lines {
				if entry, ok := value.(map[string]interface{}); ok {
					if _, ok := entry["Stints"]; ok {
						return true
					}
				}
			}
		}
	}
	return false
}

func ParseInitialData(data InitialData) {
	var positions []Position
	lines := data.R.TimingData.Lines
	for _, value := range lines {
		var position Position
		position.DriverNumber, _ = strconv.Atoi(value.RacingNumber)
		position.Position, _ = strconv.Atoi(value.Position)
		positions = append(positions, position)
	}
}

func Scanner(data []byte) (Stint, error) {
	var obj map[string]interface{}
	var stint Stint
	if err := json.Unmarshal(data, &obj); err != nil {
		return stint, err
	}
	if len(obj) == 0 {
		return stint, fmt.Errorf("empty JSON")
	}
	if _, ok := obj["C"]; !ok {
		return stint, fmt.Errorf("no control")
	}
	if _, ok := obj["M"]; !ok {
		return stint, fmt.Errorf("no data block")
	}
	var webSocketMessage WebSocketMessage
	if err := json.Unmarshal(data, &webSocketMessage); err != nil {
		return stint, fmt.Errorf("failed to unmarshal data: %v", err)
	}
	if webSocketMessage.Control == "" {
		return stint, fmt.Errorf("no control")
	}
	if len(webSocketMessage.DataBlock) == 0 {
		return stint, fmt.Errorf("no data block")
	}
	for _, dataBlock := range webSocketMessage.DataBlock {
		if len(dataBlock.DataFeed) < 3 {
			return stint, fmt.Errorf("data feed too short")
		}
		if dataBlock.DataFeed[0] == "TimingAppData" {
			if isStintData(dataBlock.DataFeed[1]) {
				stint, err := ParseInterfaceAsStint(dataBlock.DataFeed[1])
				if err != nil {
					return stint, err
				}
				stint.Timestamp = dataBlock.DataFeed[2].(string)
				return stint, nil
			} else if isPositionData(dataBlock.DataFeed[1]) {
				position, err := ParseInterfaceAsPosition(dataBlock.DataFeed[1])
				if err != nil {
					return stint, err
				}
			} else {
				var position Position
				if str, ok := dataBlock.DataFeed[1].(string); ok {
					if err := json.Unmarshal([]byte(str), &position); err != nil {
						return stint, fmt.Errorf("failed to unmarshal position: %v", err)
					}
					fmt.Printf("Position: %d\n", position.Position)
				}
			}
		}
	}
	return stint, fmt.Errorf("no stint data")
}
