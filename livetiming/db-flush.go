package livetiming

import (
	"F1Telemetry-new-server/graph/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"time"
)

func SaveMeeting(dbClient *dynamodb.Client, ctx *context.Context, meeting MeetingDataDB) error {
	item := map[string]types.AttributeValue{
		"MeetingKey":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", meeting.Key)},
		"MeetingName":         &types.AttributeValueMemberS{Value: meeting.Name},
		"MeetingOfficialName": &types.AttributeValueMemberS{Value: meeting.OfficialName},
		"Location":            &types.AttributeValueMemberS{Value: meeting.Location},
		"CountryName":         &types.AttributeValueMemberS{Value: meeting.CountryName},
		"CountryCode":         &types.AttributeValueMemberS{Value: meeting.CountryCode},
		"CircuitShortName":    &types.AttributeValueMemberS{Value: meeting.Circuit},
	}
	_, err := dbClient.PutItem(*ctx, &dynamodb.PutItemInput{
		TableName: aws.String("meetings"),
		Item:      item,
	})
	return err
}

func SaveSession(dbClient *dynamodb.Client, ctx *context.Context, session SessionInfoDB) error {
	item := map[string]types.AttributeValue{
		"ArchiveStatus": &types.AttributeValueMemberS{Value: session.ArchiveStatus},
		"StartDate":     &types.AttributeValueMemberS{Value: session.StartDate},
		"EndDate":       &types.AttributeValueMemberS{Value: session.EndDate},
		"Type":          &types.AttributeValueMemberS{Value: session.Type},
		"GmtOffset":     &types.AttributeValueMemberS{Value: session.GmtOffset},
		"SessionKey":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", session.Key)},
		"MeetingKey":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", session.MeetingKey)},
		"Name":          &types.AttributeValueMemberS{Value: session.Name},
		"Path":          &types.AttributeValueMemberS{Value: session.Path},
	}
	_, err := dbClient.PutItem(*ctx, &dynamodb.PutItemInput{
		TableName: aws.String("sessions"),
		Item:      item,
	})
	return err
}

func SaveDrivers(dbClient *dynamodb.Client, ctx *context.Context, drivers []Driver, sessionInfoKey int) error {
	for _, driver := range drivers {
		item := map[string]types.AttributeValue{
			"RacingNumber":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", driver.RacingNumber)},
			"BroadcastName": &types.AttributeValueMemberS{Value: driver.BroadcastName},
			"FullName":      &types.AttributeValueMemberS{Value: driver.FullName},
			"Abbreviation":  &types.AttributeValueMemberS{Value: driver.Abbreviation},
			"TeamName":      &types.AttributeValueMemberS{Value: driver.TeamName},
			"TeamColour":    &types.AttributeValueMemberS{Value: driver.TeamColour},
			"FirstName":     &types.AttributeValueMemberS{Value: driver.FirstName},
			"LastName":      &types.AttributeValueMemberS{Value: driver.LastName},
			"HeadshotUrl":   &types.AttributeValueMemberS{Value: driver.HeadshotUrl},
			"SessionKey":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionInfoKey)},
		}
		_, err := dbClient.PutItem(*ctx, &dynamodb.PutItemInput{
			TableName: aws.String("drivers"),
			Item:      item,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func SavePositions(dbClient *dynamodb.Client, ctx *context.Context, positions []Position) error {

	for _, position := range positions {
		item := map[string]types.AttributeValue{
			"SessionKey":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", position.SessionKey)},
			"RacingNumber": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", position.RacingNumber)},
			"Position":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", position.Position)},
			"InPit":        &types.AttributeValueMemberBOOL{Value: position.InPit},
			"PitOut":       &types.AttributeValueMemberBOOL{Value: position.PitOut},
			"Retired":      &types.AttributeValueMemberBOOL{Value: position.Retired},
			"Stopped":      &types.AttributeValueMemberBOOL{Value: position.Stopped},
		}
		_, err := dbClient.PutItem(*ctx, &dynamodb.PutItemInput{
			TableName: aws.String("positions"),
			Item:      item,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveStints(dbClient *dynamodb.Client, ctx *context.Context, stints []Stint, sessionKey int) error {
	for _, stint := range stints {
		item := map[string]types.AttributeValue{
			"SessionKey":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
			"Compound":        &types.AttributeValueMemberS{Value: stint.Compound},
			"LapFlags":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", stint.LapFlags)},
			"New":             &types.AttributeValueMemberBOOL{Value: stint.New},
			"RacingNumber":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", stint.RacingNumber)},
			"StartLaps":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", stint.StartLaps)},
			"Timestamp":       &types.AttributeValueMemberS{Value: stint.Timestamp},
			"TyresNotChanged": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", stint.TyresNotChanged)},
			"TotalLaps":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", stint.TotalLaps)},
		}
		_, err := dbClient.PutItem(*ctx, &dynamodb.PutItemInput{
			TableName: aws.String("stints"),
			Item:      item,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveLapTime(dbClient *dynamodb.Client, ctx *context.Context, lapTime LapTimeMetric) error {
	lastLapTime, err := json.Marshal(lapTime.LastLapTime)
	if err != nil {
		return err
	}
	bestLapTime, err := json.Marshal(lapTime.BestLapTime)
	if err != nil {
		return err
	}
	intervalToPositionAhead, err := json.Marshal(lapTime.IntervalToPositionAhead)
	if err != nil {
		return err
	}
	item := map[string]types.AttributeValue{
		"SessionKey":              &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", lapTime.SessionKey)},
		"RacingNumber":            &types.AttributeValueMemberN{Value: lapTime.RacingNumber},
		"NumberOfLaps":            &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", lapTime.NumberOfLaps)},
		"GapToLeader":             &types.AttributeValueMemberS{Value: lapTime.GapToLeader},
		"IntervalToPositionAhead": &types.AttributeValueMemberS{Value: string(intervalToPositionAhead)},
		"BestLapTime":             &types.AttributeValueMemberS{Value: string(bestLapTime)},
		"LastLapTime":             &types.AttributeValueMemberS{Value: string(lastLapTime)},
	}
	_, err = dbClient.PutItem(*ctx, &dynamodb.PutItemInput{
		TableName: aws.String("timings"),
		Item:      item,
	})
	return err
}

func SaveSectors(dbClient *dynamodb.Client, ctx *context.Context, sessionKey int, sectors []*model.Sector) error {
	for _, sector := range sectors {
		var reference string
		if sector.LapNumber == 0 {
			reference = uuid.New().String()
		} else {
			reference = fmt.Sprintf("%d#%d#%d", sector.RacingNumber, sector.LapNumber, sector.SectorNumber)
		}
		item := map[string]types.AttributeValue{
			"SessionKey":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
			"RacingNumber":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sector.RacingNumber)},
			"Reference":       &types.AttributeValueMemberS{Value: reference},
			"SectorNumber":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sector.SectorNumber)},
			"Value":           &types.AttributeValueMemberS{Value: sector.Value},
			"OverallFastest":  &types.AttributeValueMemberBOOL{Value: sector.OverallFastest},
			"PersonalFastest": &types.AttributeValueMemberBOOL{Value: sector.PersonalFastest},
			"LapNumber":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sector.LapNumber)},
			"Utc":             &types.AttributeValueMemberS{Value: sector.Utc.Format("2006-01-02T15:04:05Z07:00")},
		}
		_, err := dbClient.PutItem(*ctx, &dynamodb.PutItemInput{
			TableName: aws.String("sectors"),
			Item:      item,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveRaceControlMessages(dbClient *dynamodb.Client, ctx *context.Context, sessionKey int, raceControlMessages []*model.RaceControl) error {
	for _, raceControlMessage := range raceControlMessages {
		category := raceControlMessage.Category
		flag := raceControlMessage.Flag
		scope := raceControlMessage.Scope
		timeStamp, err := parseTimestamp(raceControlMessage.Date)
		if err != nil {
			return err
		}
		item := map[string]types.AttributeValue{
			"SessionKey": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", sessionKey)},
			"Message":    &types.AttributeValueMemberS{Value: raceControlMessage.Message},
			"Category":   &types.AttributeValueMemberS{Value: dereferenceString(category)},
			"Utc":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", timeStamp)},
			"Flag":       &types.AttributeValueMemberS{Value: dereferenceString(flag)},
			"Scope":      &types.AttributeValueMemberS{Value: dereferenceString(scope)},
		}
		_, err = dbClient.PutItem(*ctx, &dynamodb.PutItemInput{
			TableName: aws.String("racecontrol"),
			Item:      item,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func dereferenceString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func parseTimestamp(timestamp string) (int64, error) {
	t, err := time.ParseInLocation("2006-01-02T15:04:05", timestamp, time.UTC)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return 0, err
	}
	return t.Unix(), nil
}
