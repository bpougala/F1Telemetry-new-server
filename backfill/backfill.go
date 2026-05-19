package backfill

import (
	dataingestor "F1Telemetry-new-server/data-ingestor"
	"F1Telemetry-new-server/graph/model"
	"F1Telemetry-new-server/livetiming"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// --- Index.json types ---

type IndexResponse struct {
	Year     int            `json:"Year"`
	Meetings []IndexMeeting `json:"Meetings"`
}

type IndexMeeting struct {
	Sessions     []IndexSession `json:"Sessions"`
	Key          int            `json:"Key"`
	Code         string         `json:"Code"`
	Number       int            `json:"Number"`
	Location     string         `json:"Location"`
	OfficialName string         `json:"OfficialName"`
	Name         string         `json:"Name"`
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

type IndexSession struct {
	Key       int    `json:"Key"`
	Type      string `json:"Type"`
	Number    int    `json:"Number"`
	Name      string `json:"Name"`
	StartDate string `json:"StartDate"`
	EndDate   string `json:"EndDate"`
	GmtOffset string `json:"GmtOffset"`
	Path      string `json:"Path"`
}

// --- Helpers ---

const staticBaseURL = "https://livetiming.formula1.com/static"

func fetchStaticJSON(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}
	// Strip UTF-8 BOM
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	return body, nil
}

// wrapInR wraps raw JSON data in {"R":{<key>:<data>},"I":"1"} so existing
// Build* parsers (which expect the SignalR initial snapshot envelope) can
// parse it directly.
func wrapInR(key string, data json.RawMessage) []byte {
	inner := fmt.Sprintf(`{"R":{%q:%s},"I":"1"}`, key, string(data))
	return []byte(inner)
}

// --- Entry points ---

// Run parses the year from os.Args and runs the backfill.
func Run() {
	year := 2026
	if len(os.Args) > 2 {
		if y, err := strconv.Atoi(os.Args[2]); err == nil {
			year = y
		}
	}
	ctx := context.Background()
	dbClient, err := dataingestor.GetDynamoClient(&ctx)
	if err != nil {
		log.Fatalf("failed to create DynamoDB client: %v", err)
	}
	fmt.Printf("Starting backfill for year %d\n", year)
	if err := backfillYear(dbClient, ctx, year); err != nil {
		log.Fatalf("backfill failed: %v", err)
	}
	fmt.Println("Backfill complete")
}

// --- Orchestration ---

func backfillYear(dbClient *dynamodb.Client, ctx context.Context, year int) error {
	indexURL := fmt.Sprintf("%s/%d/Index.json", staticBaseURL, year)
	body, err := fetchStaticJSON(indexURL)
	if err != nil {
		return fmt.Errorf("fetch index: %w", err)
	}
	var index IndexResponse
	if err := json.Unmarshal(body, &index); err != nil {
		return fmt.Errorf("parse index: %w", err)
	}
	fmt.Printf("Found %d meetings for %d\n", len(index.Meetings), year)

	// Build set of existing session keys
	existingSessionKeys := make(map[int]bool)
	meetings, err := dataingestor.FetchMeetings(dbClient, ctx)
	if err != nil {
		fmt.Printf("Warning: could not fetch existing meetings: %v\n", err)
	}
	for _, m := range meetings {
		sessions, err := dataingestor.FetchSessions(dbClient, ctx, m.MeetingKey)
		if err != nil {
			continue
		}
		for _, s := range sessions {
			existingSessionKeys[s.SessionKey] = true
		}
	}
	fmt.Printf("Found %d existing sessions in DB\n", len(existingSessionKeys))

	// Build set of meeting keys that need circuit_key backfill
	meetingsNeedCircuitKey := make(map[int]bool)
	for _, m := range meetings {
		if m.CircuitKey == 0 {
			meetingsNeedCircuitKey[m.MeetingKey] = true
		}
	}

	// Build set of sessions that already have weather data
	existingWeatherKeys := make(map[int]bool)
	for sessionKey := range existingSessionKeys {
		weather, err := dataingestor.FetchWeather(dbClient, ctx, sessionKey)
		if err == nil && len(weather) > 0 {
			existingWeatherKeys[sessionKey] = true
		}
	}

	for _, meeting := range index.Meetings {
		// Re-save meeting if circuit_key is missing
		if meetingsNeedCircuitKey[meeting.Key] {
			meetingDB := livetiming.MeetingDataDB{
				Key:          meeting.Key,
				Name:         meeting.Name,
				OfficialName: meeting.OfficialName,
				Location:     meeting.Location,
				Number:       meeting.Number,
				CountryName:  meeting.Country.Name,
				CountryCode:  meeting.Country.Code,
				Circuit:      meeting.Circuit.ShortName,
				CircuitKey:   meeting.Circuit.Key,
			}
			if err := livetiming.SaveMeeting(dbClient, &ctx, meetingDB); err != nil {
				fmt.Printf("  [Meeting] save circuit_key error: %v\n", err)
			} else {
				fmt.Printf("Backfilled circuit_key for meeting %d (%s)\n", meeting.Key, meeting.Name)
			}
		}

		for _, session := range meeting.Sessions {
			if existingSessionKeys[session.Key] {
				if existingWeatherKeys[session.Key] {
					fmt.Printf("Skipping session %d (%s - %s) — already in DB with weather\n", session.Key, meeting.Name, session.Name)
					continue
				}
				fmt.Printf("Backfilling weather for existing session %d (%s - %s)...\n", session.Key, meeting.Name, session.Name)
				basePath := fmt.Sprintf("%s/%s", staticBaseURL, session.Path)
				backfillWeatherData(dbClient, ctx, basePath, session.Key)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			fmt.Printf("Backfilling session %d (%s - %s)...\n", session.Key, meeting.Name, session.Name)
			backfillSession(dbClient, ctx, meeting, session)
			time.Sleep(500 * time.Millisecond)
		}
	}
	return nil
}

func backfillSession(dbClient *dynamodb.Client, ctx context.Context, meeting IndexMeeting, session IndexSession) {
	basePath := fmt.Sprintf("%s/%s", staticBaseURL, session.Path)

	// 1. SessionInfo → meeting + session
	backfillSessionInfo(dbClient, ctx, basePath, session.Key)

	// 2. DriverList → drivers
	backfillDriverList(dbClient, ctx, basePath, session.Key)

	// 3. TimingData → positions, timings, sectors
	backfillTimingData(dbClient, ctx, basePath, session.Key)

	// 4. TimingAppData → stints
	backfillTimingAppData(dbClient, ctx, basePath, session.Key)

	// 5. RaceControlMessages → race control
	backfillRaceControl(dbClient, ctx, basePath, session.Key)

	// 6. SessionData → track status
	backfillSessionData(dbClient, ctx, basePath, session.Key)

	// 7. WeatherData → weather
	backfillWeatherData(dbClient, ctx, basePath, session.Key)

	fmt.Printf("  Done session %d\n", session.Key)
}

func backfillSessionInfo(dbClient *dynamodb.Client, ctx context.Context, basePath string, sessionKey int) {
	data, err := fetchStaticJSON(basePath + "SessionInfo.json")
	if err != nil {
		fmt.Printf("  [SessionInfo] fetch error: %v\n", err)
		return
	}
	wrapped := wrapInR("SessionInfo", data)

	meetingData, err := livetiming.BuildMeetingData(wrapped)
	if err == nil {
		meetingDB := livetiming.ConvertMeetingToDB(meetingData)
		if err := livetiming.SaveMeeting(dbClient, &ctx, meetingDB); err != nil {
			fmt.Printf("  [SessionInfo] save meeting error: %v\n", err)
		}
	} else {
		fmt.Printf("  [SessionInfo] parse meeting error: %v\n", err)
	}

	sessionInfo, err := livetiming.BuildSessionInfo(wrapped)
	if err == nil && sessionInfo.Key != 0 {
		if err := livetiming.SaveSession(dbClient, &ctx, sessionInfo); err != nil {
			fmt.Printf("  [SessionInfo] save session error: %v\n", err)
		}
	} else if err != nil {
		fmt.Printf("  [SessionInfo] parse session error: %v\n", err)
	}
}

func backfillDriverList(dbClient *dynamodb.Client, ctx context.Context, basePath string, sessionKey int) {
	data, err := fetchStaticJSON(basePath + "DriverList.json")
	if err != nil {
		fmt.Printf("  [DriverList] fetch error: %v\n", err)
		return
	}
	wrapped := wrapInR("DriverList", data)

	drivers, err := livetiming.BuildDriverList(wrapped)
	if err != nil {
		fmt.Printf("  [DriverList] parse error: %v\n", err)
		return
	}
	for i := range drivers {
		drivers[i].SessionKey = sessionKey
	}
	if err := livetiming.SaveDrivers(dbClient, &ctx, drivers, sessionKey); err != nil {
		fmt.Printf("  [DriverList] save error: %v\n", err)
	}
}

func backfillTimingData(dbClient *dynamodb.Client, ctx context.Context, basePath string, sessionKey int) {
	data, err := fetchStaticJSON(basePath + "TimingData.json")
	if err != nil {
		fmt.Printf("  [TimingData] fetch error: %v\n", err)
		return
	}
	wrapped := wrapInR("TimingData", data)

	// Positions
	positions, err := livetiming.BuildPositions(wrapped)
	if err == nil {
		for i := range positions {
			positions[i].SessionKey = sessionKey
		}
		if err := livetiming.SavePositions(dbClient, &ctx, positions); err != nil {
			fmt.Printf("  [TimingData] save positions error: %v\n", err)
		}
	}

	// Timing data (lap times)
	timingData, err := livetiming.BuildTimingData(wrapped)
	if err == nil {
		for _, timing := range timingData {
			timing.SessionKey = sessionKey
			if err := livetiming.SaveLapTime(dbClient, &ctx, timing); err != nil {
				fmt.Printf("  [TimingData] save lap time error: %v\n", err)
			}
		}
	}

	// Sectors
	sectors, err := livetiming.BuildSectors(wrapped)
	if err == nil {
		var sectorsModel []*model.Sector
		for _, sectorTime := range sectors {
			for _, sector := range sectorTime.Sectors {
				sectorsModel = append(sectorsModel, &model.Sector{
					LapNumber:       sector.LapNumber,
					RacingNumber:    sectorTime.RacingNumber,
					SectorNumber:    sector.SectorNumber,
					Value:           sector.Value,
					OverallFastest:  sector.OverallFastest,
					PersonalFastest: sector.PersonalFastest,
					Utc:             &sectorTime.Utc,
				})
			}
		}
		if err := livetiming.SaveSectors(dbClient, &ctx, sessionKey, sectorsModel); err != nil {
			fmt.Printf("  [TimingData] save sectors error: %v\n", err)
		}
	}
}

func backfillTimingAppData(dbClient *dynamodb.Client, ctx context.Context, basePath string, sessionKey int) {
	data, err := fetchStaticJSON(basePath + "TimingAppData.json")
	if err != nil {
		fmt.Printf("  [TimingAppData] fetch error: %v\n", err)
		return
	}
	wrapped := wrapInR("TimingAppData", data)

	stints, err := livetiming.BuildStints(wrapped)
	if err != nil {
		fmt.Printf("  [TimingAppData] parse error: %v\n", err)
		return
	}
	if err := livetiming.SaveStints(dbClient, &ctx, stints, sessionKey); err != nil {
		fmt.Printf("  [TimingAppData] save error: %v\n", err)
	}
}

func backfillRaceControl(dbClient *dynamodb.Client, ctx context.Context, basePath string, sessionKey int) {
	data, err := fetchStaticJSON(basePath + "RaceControlMessages.json")
	if err != nil {
		fmt.Printf("  [RaceControl] fetch error: %v\n", err)
		return
	}
	wrapped := wrapInR("RaceControlMessages", data)

	raceControlMessages, err := livetiming.BuildRaceControl(wrapped)
	if err != nil {
		fmt.Printf("  [RaceControl] parse error: %v\n", err)
		return
	}
	var raceControlModel []*model.RaceControl
	for _, message := range raceControlMessages {
		raceControlModel = append(raceControlModel, &model.RaceControl{
			Message:  message.Message,
			Category: &message.Category,
			Date:     message.Utc,
			Flag:     &message.Flag,
			Scope:    &message.Scope,
		})
	}
	if err := livetiming.SaveRaceControlMessages(dbClient, &ctx, sessionKey, raceControlModel); err != nil {
		fmt.Printf("  [RaceControl] save error: %v\n", err)
	}
}

func backfillSessionData(dbClient *dynamodb.Client, ctx context.Context, basePath string, sessionKey int) {
	data, err := fetchStaticJSON(basePath + "SessionData.json")
	if err != nil {
		fmt.Printf("  [SessionData] fetch error: %v\n", err)
		return
	}
	wrapped := wrapInR("SessionData", data)

	trackStatus, err := livetiming.BuildTrackStatus(wrapped)
	if err != nil {
		fmt.Printf("  [SessionData] parse error: %v\n", err)
		return
	}
	var trackStatusModel []*model.TrackStatus
	for _, status := range trackStatus {
		trackStatusModel = append(trackStatusModel, &model.TrackStatus{
			Status:     status.Status,
			Timestamp:  status.Utc.Format(time.RFC3339),
			SessionKey: sessionKey,
		})
	}
	if err := livetiming.SaveTrackStatus(dbClient, &ctx, trackStatusModel); err != nil {
		fmt.Printf("  [SessionData] save error: %v\n", err)
	}
}

func backfillWeatherData(dbClient *dynamodb.Client, ctx context.Context, basePath string, sessionKey int) {
	data, err := fetchStaticJSON(basePath + "WeatherData.json")
	if err != nil {
		fmt.Printf("  [WeatherData] fetch error: %v\n", err)
		return
	}
	wrapped := wrapInR("WeatherData", data)

	weather, err := livetiming.BuildWeatherData(wrapped)
	if err != nil {
		fmt.Printf("  [WeatherData] parse error: %v\n", err)
		return
	}

	airTemp, _ := strconv.ParseFloat(weather.AirTemp, 64)
	trackTemp, _ := strconv.ParseFloat(weather.TrackTemp, 64)
	humidity, _ := strconv.ParseFloat(weather.Humidity, 64)
	pressure, _ := strconv.ParseFloat(weather.Pressure, 64)
	windSpeed, _ := strconv.ParseFloat(weather.WindSpeed, 64)
	windDirection, _ := strconv.Atoi(weather.WindDirection)
	rainfall := weather.Rainfall == "1"

	weatherModel := &model.Weather{
		SessionKey:       sessionKey,
		AirTemperature:   airTemp,
		TrackTemperature: trackTemp,
		Humidity:         humidity,
		Pressure:         pressure,
		Rainfall:         rainfall,
		WindSpeed:        windSpeed,
		WindDirection:    windDirection,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
	}
	if err := livetiming.SaveWeather(dbClient, &ctx, weatherModel); err != nil {
		fmt.Printf("  [WeatherData] save error: %v\n", err)
	}
}
