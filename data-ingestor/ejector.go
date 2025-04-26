package data_ingestor

import (
	"F1Telemetry-new-server/data-ingestor/collections"
	"F1Telemetry-new-server/graph/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
	"time"
)

func FetchMeetings(ddbClient *dynamodb.Client, ctx context.Context) ([]model.Meeting, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String("meetings"),
	}

	result, err := ddbClient.Scan(ctx, input)
	if err != nil {
		return nil, err
	}

	var meetings []model.Meeting
	for _, item := range result.Items {
		meetingKeyAttr, ok := item["MeetingKey"].(*types.AttributeValueMemberN)
		if !ok {
			continue
		}
		var meetingKey int
		if _, err := fmt.Sscanf(meetingKeyAttr.Value, "%d", &meetingKey); err != nil {
			continue
		}
		meeting := model.Meeting{
			MeetingKey:          meetingKey,
			MeetingName:         getStringAttr(item, "MeetingName"),
			MeetingOfficialName: getStringAttr(item, "MeetingOfficialName"),
			Location:            getStringAttr(item, "Location"),
			CountryName:         getStringAttr(item, "CountryName"),
			CountryCode:         getStringAttr(item, "CountryCode"),
			CircuitShortName:    getStringAttr(item, "CircuitShortName"),
		}
		meetings = append(meetings, meeting)
	}
	return meetings, nil
}

func FetchSessions(dbClient *dynamodb.Client, ctx context.Context, meetingKey int) ([]model.Session, error) {
	input := &dynamodb.QueryInput{
		TableName: aws.String("sessions"),
		KeyConditions: map[string]types.Condition{
			"MeetingKey": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: fmt.Sprintf("%d", meetingKey)},
				},
			},
		},
	}

	result, err := dbClient.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	var sessions []model.Session
	for _, item := range result.Items {
		session := model.Session{
			SessionKey:  getIntAttr(item, "SessionKey"),
			MeetingKey:  getIntAttr(item, "MeetingKey"),
			SessionName: getStringAttr(item, "Name"),
			DateStart:   getStringAttr(item, "StartDate"),
			DateEnd:     getStringAttr(item, "EndDate"),
			SessionType: getStringAttr(item, "Type"),
			GmtOffset:   getStringAttr(item, "GmtOffset"),
			Status:      getStringAttr(item, "ArchiveStatus"),
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func FetchDrivers(dbClient *dynamodb.Client, ctx context.Context, sessionKey int) ([]model.Driver, error) {
	input := &dynamodb.QueryInput{
		TableName: aws.String("drivers"),
		KeyConditions: map[string]types.Condition{
			"SessionKey": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
				},
			},
		},
	}

	result, err := dbClient.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	var drivers []model.Driver
	for _, item := range result.Items {
		teamColour := getStringAttr(item, "TeamColour")
		teamName := getStringAttr(item, "TeamName")
		abbreviation := getStringAttr(item, "Abbreviation")
		driver := model.Driver{
			SessionKey:    getIntAttr(item, "SessionKey"),
			RacingNumber:  getIntAttr(item, "RacingNumber"),
			FirstName:     getStringAttr(item, "FirstName"),
			LastName:      getStringAttr(item, "LastName"),
			TeamColour:    &teamColour,
			TeamName:      &teamName,
			Abbreviation:  &abbreviation,
			FullName:      getStringAttr(item, "FullName"),
			BroadcastName: getStringAttr(item, "BroadcastName"),
			HeadshotURL:   getStringAttr(item, "HeadshotUrl"),
		}
		drivers = append(drivers, driver)
	}
	return drivers, nil
}

func FetchStints(dbClient *dynamodb.Client, ctx context.Context, sessionKey int) ([]model.Stint, error) {
	input := &dynamodb.QueryInput{
		TableName: aws.String("stints"),
		KeyConditions: map[string]types.Condition{
			"SessionKey": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
				},
			},
		},
	}

	result, err := dbClient.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	var stints []model.Stint
	for _, item := range result.Items {
		stint := model.Stint{
			RacingNumber:    getIntAttr(item, "RacingNumber"),
			LapFlags:        getIntAttr(item, "LapFlags"),
			Compound:        getStringAttr(item, "Compound"),
			New:             getBoolAttr(item, "New"),
			TyresNotChanged: getIntAttr(item, "TyresNotChanged"),
			TotalLaps:       getIntAttr(item, "TotalLaps"),
			StartLaps:       getIntAttr(item, "StartLaps"),
			Timestamp:       getStringAttr(item, "Timestamp"),
		}
		stints = append(stints, stint)
	}
	return stints, nil
}

func FetchSectors(dbClient *dynamodb.Client, ctx context.Context, sessionKey int) ([]model.Sector, error) {
	input := &dynamodb.QueryInput{
		TableName: aws.String("sectors"),
		KeyConditions: map[string]types.Condition{
			"SessionKey": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
				},
			},
		},
	}

	result, err := dbClient.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	var sectors []model.Sector
	for _, item := range result.Items {
		time, _ := time.Parse(time.RFC3339, getStringAttr(item, "Utc"))
		sector := model.Sector{
			RacingNumber:    getIntAttr(item, "RacingNumber"),
			SectorNumber:    getIntAttr(item, "SectorNumber"),
			Value:           getStringAttr(item, "Value"),
			LapNumber:       getIntAttr(item, "LapNumber"),
			Utc:             &time,
			OverallFastest:  getBoolAttr(item, "OverallFastest"),
			PersonalFastest: getBoolAttr(item, "PersonalFastest"),
		}
		sectors = append(sectors, sector)
	}
	return sectors, nil
}

func FetchPositions(dbClient *dynamodb.Client, ctx context.Context, sessionKey int) ([]model.Position, error) {
	input := &dynamodb.QueryInput{
		TableName: aws.String("positions"),
		KeyConditions: map[string]types.Condition{
			"SessionKey": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
				},
			},
		},
	}

	result, err := dbClient.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	var positions []model.Position
	for _, item := range result.Items {
		position := model.Position{
			SessionKey:   getIntAttr(item, "SessionKey"),
			RacingNumber: getIntAttr(item, "RacingNumber"),
			InPit:        getBoolAttr(item, "InPit"),
			PitOut:       getBoolAttr(item, "PitOut"),
			Retired:      getBoolAttr(item, "Retired"),
			Stopped:      getBoolAttr(item, "Stopped"),
			Position:     getIntAttr(item, "Position"),
		}
		positions = append(positions, position)
	}
	return positions, nil
}

func FetchTimings(dbClient *dynamodb.Client, ctx context.Context, sessionKey int) ([]model.Timing, error) {
	input := &dynamodb.QueryInput{
		TableName: aws.String("timings"),
		KeyConditions: map[string]types.Condition{
			"SessionKey": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
				},
			},
		},
	}

	result, err := dbClient.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	var timings []model.Timing
	for _, item := range result.Items {
		interval := getStringAttr(item, "IntervalToPositionAhead")
		var intervalToPositionAhead collections.IntervalToPositionAhead
		if err := json.Unmarshal([]byte(interval), &intervalToPositionAhead); err != nil {
			return nil, err
		}
		bestLapTime := getStringAttr(item, "BestLapTime")
		var bestLapTimeStruct collections.BestLapTime
		if err := json.Unmarshal([]byte(bestLapTime), &bestLapTimeStruct); err != nil {
			return nil, err
		}
		lastLapTime := getStringAttr(item, "LastLapTime")
		var lastLapTimeStruct collections.LastLapTime
		if err := json.Unmarshal([]byte(lastLapTime), &lastLapTimeStruct); err != nil {
			return nil, err
		}
		timing := model.Timing{
			SessionKey:              getIntAttr(item, "SessionKey"),
			RacingNumber:            getIntAttr(item, "RacingNumber"),
			Status:                  getIntAttr(item, "Status"),
			InPit:                   getBoolAttr(item, "InPit"),
			PitOut:                  getBoolAttr(item, "PitOut"),
			Stopped:                 getBoolAttr(item, "Stopped"),
			Retired:                 getBoolAttr(item, "Retired"),
			LastLapTime:             lastLapTimeStruct,
			IntervalToPositionAhead: intervalToPositionAhead,
			GapToLeader:             getStringAttr(item, "GapToLeader"),
			BestLapTime:             bestLapTimeStruct,
		}
		timings = append(timings, timing)
	}
	return timings, nil
}

func FetchRaceControl(dbClient *dynamodb.Client, ctx context.Context, sessionKey int) ([]model.RaceControl, error) {
	input := &dynamodb.QueryInput{
		TableName: aws.String("racecontrol"),
		KeyConditions: map[string]types.Condition{
			"SessionKey": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
				},
			},
		},
	}
	result, err := dbClient.Query(ctx, input)
	if err != nil {
		return nil, err
	}
	var raceControlMessages []model.RaceControl
	for _, item := range result.Items {
		date, err := getTimeStringFromTimestamp(getIntAttr(item, "Utc"))
		if err != nil {
			return nil, err
		}
		category := getStringAttr(item, "Category")
		flag := getStringAttr(item, "Flag")
		lapNumber := getIntAttr(item, "LapNumber")
		scope := getStringAttr(item, "Scope")
		raceControl := model.RaceControl{
			Date:      date,
			Category:  &category,
			Flag:      &flag,
			LapNumber: &lapNumber,
			Message:   getStringAttr(item, "Message"),
			Scope:     &scope,
		}
		raceControlMessages = append(raceControlMessages, raceControl)
	}
	return raceControlMessages, nil
}

func getStringAttr(item map[string]types.AttributeValue, key string) string {
	if val, ok := item[key].(*types.AttributeValueMemberS); ok {
		return val.Value
	}
	return ""
}

func getIntAttr(item map[string]types.AttributeValue, key string) int {
	if val, ok := item[key].(*types.AttributeValueMemberN); ok {
		intVal, err := strconv.Atoi(val.Value)
		if err != nil {
			return 0
		}
		return intVal
	}
	return 0
}

func getBoolAttr(item map[string]types.AttributeValue, key string) bool {
	if val, ok := item[key].(*types.AttributeValueMemberBOOL); ok {
		return val.Value
	}
	return false
}

func getTimeStringFromTimestamp(timestamp int) (string, error) {
	intTimeStamp := int64(timestamp)
	t := time.Unix(intTimeStamp, 0).UTC()
	return t.Format(time.RFC3339), nil
}
