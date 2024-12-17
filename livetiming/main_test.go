package livetiming

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func echo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	sessionData, err := os.ReadFile("test/session-data.json")
	if err != nil {
		fmt.Printf("Error reading session data: %v\n", err)
		return
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, sessionData); err != nil {
			return
		}
	}
}

func serverSendingDriverList(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	sessionData, err := os.ReadFile("test/driver-list.json")
	if err != nil {
		fmt.Printf("Error reading driver list data: %v\n", err)
		return
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, sessionData); err != nil {
			return
		}
	}
}

func serverSendingTimingAppRaceUpdates(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	sessionData, err := os.ReadFile("test/timing-data-race.json")
	if err != nil {
		fmt.Printf("Error reading timing data: %v\n", err)
		return
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, sessionData); err != nil {
			return
		}
	}
}

func serverSendingTimingAppQualifyingUpdates(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	sessionData, err := os.ReadFile("test/timing-data-qualifying.json")
	if err != nil {
		fmt.Printf("Error reading timing data: %v\n", err)
		return
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, sessionData); err != nil {
			return
		}
	}
}

func serverSendingTimingAppStints(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	sessionData, err := os.ReadFile("test/timing-app-data-race.json")
	if err != nil {
		fmt.Printf("Error reading timing data: %v\n", err)
		return
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, sessionData); err != nil {
			return
		}
	}
}

func serverSendingTimingDataUpdate(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	sessionData, err := os.ReadFile("test/update-timing-race.json")
	if err != nil {
		fmt.Printf("Error reading session data: %v\n", err)
		return
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, sessionData); err != nil {
			return
		}
	}
}

func serverSendingPositionUpdate(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	sessionData, err := os.ReadFile("test/update-driver-list.json")
	if err != nil {
		fmt.Printf("Error reading session data: %v\n", err)
		return
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, sessionData); err != nil {
			return
		}
	}
}

func serverSendingCarDataInfo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	carData, err := os.ReadFile("test/cardata.json")
	if err != nil {
		fmt.Printf("Error reading car data: %v\n", err)
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, carData); err != nil {
			return
		}
	}
}

func serverSendingSessionUpdate(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	sessionData, err := os.ReadFile("test/session-update.json")
	if err != nil {
		fmt.Printf("Error reading session data: %v\n", err)
		return
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, sessionData); err != nil {
			return
		}
	}
}

func serverSendingTimingAppDataUpdate(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	sessionData, err := os.ReadFile("test/timing-app-data-update.json")
	if err != nil {
		fmt.Printf("Error reading timing app data: %v\n", err)
		return
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, sessionData); err != nil {
			return
		}
	}
}

func serverSendingDummyMessage(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	if err != nil {
		fmt.Printf("Error reading session data: %v\n", err)
		return
	}
	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, []byte("{}")); err != nil {
			return
		}
	}
}

func getWebSocketClient() (*websocket.Conn, error) {
	server := httptest.NewServer(http.HandlerFunc(serverSendingDriverList))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func sendDummyMessage(client *websocket.Conn) error {
	message := []byte("Hello")
	if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
		return err
	}
	return nil
}

func prepareClientReturningDummyMessage() (*websocket.Conn, error) {
	server := httptest.NewServer(http.HandlerFunc(serverSendingDummyMessage))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func prepareClientreturningTimingAppFirstResponse() ([]byte, error) {
	server := httptest.NewServer(http.HandlerFunc(serverSendingTimingAppRaceUpdates))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		return nil, fmt.Errorf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("Could not read message from websocket: %v", err)
	}
	return response, nil
}

func TestShouldBuildMeetingDataStructWhenOpeningWebSocket(t *testing.T) {
	expectedMeetingData := MeetingData{
		Key:          1252,
		Name:         "Abu Dhabi Grand Prix",
		OfficialName: "FORMULA 1 ETIHAD AIRWAYS ABU DHABI GRAND PRIX 2024",
		Location:     "Yas Island",
		Number:       24,
		Country: CountryData{
			Key:  21,
			Code: "UAE",
			Name: "United Arab Emirates",
		},
		Circuit: CircuitData{
			Key:       70,
			ShortName: "Yas Marina Circuit",
		},
	}
	client, err := getWebSocketClient()
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	realMeetingData, err := BuildMeetingData(response)
	if err != nil {
		t.Fatalf("Could not build session data and session info structs: %v", err)
	}
	if realMeetingData != expectedMeetingData {
		t.Fatalf("Expected meeting data: %v, got: %v", expectedMeetingData, realMeetingData)
	}
}

func TestShouldReturnErrorWhenOtherMessagesAreSent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(serverSendingDummyMessage))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	meetingData, err := BuildMeetingData(response)
	if err == nil {
		t.Fatalf("Expected error, got meeting data: %v", meetingData)
	}
	var realMeetingData MeetingData
	if meetingData != realMeetingData {
		t.Fatalf("Expected meeting data: %v, got: %v", realMeetingData, meetingData)
	}
}

func TestShouldBuildSessionInfoStructWhenOpeningWebSocket(t *testing.T) {
	expectedSessionInfo := SessionInfoDB{
		ArchiveStatus: "Complete",
		StartDate:     "2024-12-07T14:30:00",
		EndDate:       "2024-12-07T15:30:00",
		Type:          "Practice",
		GmtOffset:     "04:00:00",
		Key:           9657,
		MeetingKey:    1252,
		Name:          "Practice 3",
		Path:          "2024/2024-12-08_Abu_Dhabi_Grand_Prix/2024-12-07_Practice_3/",
	}
	client, err := getWebSocketClient()
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	realSessionInfo, err := BuildSessionInfo(response)
	if err != nil {
		t.Fatalf("Could not build session data and session info structs: %v", err)
	}
	if realSessionInfo != expectedSessionInfo {
		t.Fatalf("Expected session info: %v, got: %v", expectedSessionInfo, realSessionInfo)
	}
}

func TestShouldNotCreateSessionInfoWhenOtherMessagesAreSent(t *testing.T) {
	client, err := prepareClientReturningDummyMessage()
	defer client.Close()
	if err != nil {
		t.Fatalf("Could not prepare client: %v", err)
	}
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	sessionInfo, err := BuildSessionInfo(response)
	if err == nil {
		t.Fatalf("Expected error, got session info: %v", sessionInfo)
	}
	if sessionInfo != (SessionInfoDB{}) {
		t.Fatalf("Expected empty session info, got: %v", sessionInfo)
	}
}

func TestShouldCreateDriverList(t *testing.T) {
	maxVerstappen := Driver{
		RacingNumber:  1,
		BroadcastName: "M VERSTAPPEN",
		FullName:      "Max VERSTAPPEN",
		Abbreviation:  "VER",
		TeamName:      "Red Bull Racing",
		TeamColour:    "3671C6",
		FirstName:     "Max",
		LastName:      "Verstappen",
		HeadshotUrl:   "https://media.formula1.com/d_driver_fallback_image.png/content/dam/fom-website/drivers/M/MAXVER01_Max_Verstappen/maxver01.png.transform/1col/image.png",
		CountryCode:   "NED",
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingDriverList))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	driverList, err := BuildDriverList(response)
	if err != nil {
		t.Fatalf("Could not build driver list: %v", err)
	}
	sort.Slice(driverList, func(i, j int) bool {
		return driverList[i].RacingNumber < driverList[j].RacingNumber
	})
	if len(driverList) != 20 {
		t.Fatalf("Expected 20 drivers, got: %d", len(driverList))
	}
	if driverList[0] != maxVerstappen {
		t.Fatalf("Expected driver list: %v, got: %v", maxVerstappen, driverList[1])
	}
}

func TestShouldNotCreateDriverListIfOtherMessageIsSent(t *testing.T) {
	client, err := prepareClientReturningDummyMessage()
	defer client.Close()
	if err != nil {
		t.Fatalf("Could not prepare client: %v", err)
	}
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	driverList, err := BuildDriverList(response)
	if err == nil {
		t.Fatalf("Expected error, got driver list: %v", driverList)
	}
	if driverList != nil {
		t.Fatalf("Expected nil driver list, got: %v", driverList)
	}
}

func TestShouldCreatePositionsWhenDriverListFirstSubscribes(t *testing.T) {
	maxVerstappenPosition := Position{
		SessionKey:   0,
		RacingNumber: 1,
		Position:     4,
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingDriverList))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	positionsList, err := BuildPositions(response)
	if err != nil {
		t.Fatalf("Could not build positions list: %v", err)
	}
	sort.Slice(positionsList, func(i, j int) bool {
		return positionsList[i].RacingNumber < positionsList[j].RacingNumber
	})
	if len(positionsList) != 20 {
		t.Fatalf("Expected 20 positions, got: %d", len(positionsList))
	}
	if positionsList[0] != maxVerstappenPosition {
		t.Fatalf("Expected position: %v, got: %v", maxVerstappenPosition, positionsList[0])
	}
}

func TestShouldCreatePositionsWhenTimingAppSendsUpdatesDuringRace(t *testing.T) {
	maxVerstappenPosition := Position{
		SessionKey:   0,
		RacingNumber: 1,
		Position:     6,
		InPit:        false,
		PitOut:       false,
		Stopped:      false,
		Status:       1088,
	}
	response, err := prepareClientreturningTimingAppFirstResponse()
	if err != nil {
		t.Fatalf("%v", err)
	}
	positionsList, err := BuildPositions(response)
	if err != nil {
		t.Fatalf("Could not build positions list: %v", err)
	}
	sort.Slice(positionsList, func(i, j int) bool {
		return positionsList[i].RacingNumber < positionsList[j].RacingNumber
	})
	if len(positionsList) != 20 {
		t.Fatalf("Expected 20 positions, got: %d", len(positionsList))
	}
	if positionsList[0] != maxVerstappenPosition {
		t.Fatalf("Expected position: %v, got: %v", maxVerstappenPosition, positionsList[0])
	}
}

func TestShouldCreatePositionsWhenTimingAppSendsUpdatesDuringQualifying(t *testing.T) {
	maxVerstappenPosition := Position{
		SessionKey:   0,
		RacingNumber: 1,
		Position:     11,
		InPit:        true,
		PitOut:       false,
		Stopped:      false,
		Status:       272,
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingTimingAppQualifyingUpdates))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	positionsList, err := BuildPositions(response)
	if err != nil {
		t.Fatalf("Could not build positions list: %v", err)
	}
	sort.Slice(positionsList, func(i, j int) bool {
		return positionsList[i].RacingNumber < positionsList[j].RacingNumber
	})
	if len(positionsList) != 20 {
		t.Fatalf("Expected 20 positions, got: %d", len(positionsList))
	}
	if positionsList[0] != maxVerstappenPosition {
		t.Fatalf("Expected position: %v, got: %v", maxVerstappenPosition, positionsList[0])
	}
}

func TestShouldNotCreatePositionsWhenDummyMessageIsSent(t *testing.T) {
	client, err := prepareClientReturningDummyMessage()
	defer client.Close()
	if err != nil {
		t.Fatalf("Could not prepare client: %v", err)
	}
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	positionsList, err := BuildPositions(response)
	if err == nil {
		t.Fatalf("Expected error, got positions list: %v", positionsList)
	}
	if positionsList != nil {
		t.Fatalf("Expected nil positions list, got: %v", positionsList)
	}
}

func TestShouldCreateLapTimeWhenFirstSubscribing(t *testing.T) {
	expectedTimingData := LapTimeMetric{
		SessionKey:   0,
		RacingNumber: "1",
		NumberOfLaps: 58,
		GapToLeader:  "+49.847",
		IntervalToPositionAhead: struct {
			Value    string `json:"Value"`
			Catching bool   `json:"Catching"`
		}{Value: "+12.309", Catching: true},
		LastLapTime: struct {
			Value           string `json:"Value"`
			OverallFastest  bool   `json:"OverallFastest"`
			PersonalFastest bool   `json:"PersonalFastest"`
		}{Value: "1:28.780", OverallFastest: false, PersonalFastest: false},
		BestLapTime: struct {
			Value string `json:"Value"`
			Lap   int    `json:"Lap"`
		}{Value: "1:27.765", Lap: 56},
	}
	response, err := prepareClientreturningTimingAppFirstResponse()
	if err != nil {
		t.Fatalf("%v", err)
	}
	timingData, err := BuildTimingData(response)
	if err != nil {
		t.Fatalf("Could not build timing data: %v", err)
	}
	if len(timingData) != 19 {
		t.Fatalf("Expected 19 timing data, got: %d", len(timingData))
	}
	sort.Slice(timingData, func(i, j int) bool {
		racingNumberI, _ := strconv.Atoi(timingData[i].RacingNumber)
		racingNumberJ, _ := strconv.Atoi(timingData[j].RacingNumber)
		return racingNumberI < racingNumberJ
	})
	if timingData[0] != expectedTimingData {
		t.Fatalf("Expected timing data: %v, got: %v", expectedTimingData, timingData[0])
	}
}

func TestShouldCreateLapTimeWhenReceivingUpdate(t *testing.T) {
	expectedTimingData := LapTimeMetric{
		SessionKey:   0,
		RacingNumber: "63",
		NumberOfLaps: 10,
		BestLapTime: struct {
			Value string `json:"Value"`
			Lap   int    `json:"Lap"`
		}{
			Value: "1:23.805",
			Lap:   6,
		},
		LastLapTime: struct {
			Value           string `json:"Value"`
			OverallFastest  bool   `json:"OverallFastest"`
			PersonalFastest bool   `json:"PersonalFastest"`
		}{
			Value:           "1:23.805",
			OverallFastest:  false,
			PersonalFastest: true,
		},
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingTimingDataUpdate))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	timingData, err := BuildTimingData(response)
	if err != nil {
		t.Fatalf("Could not build timing data: %v", err)
	}
	if len(timingData) != 1 {
		t.Fatalf("Expected 1 timing data, got: %d", len(timingData))
	}
	if timingData[0] != expectedTimingData {
		t.Fatalf("Expected timing data: %v, got: %v", expectedTimingData, timingData)
	}
}

func TestShouldCreatePositionWhenReceivingUpdate(t *testing.T) {
	lanceStrollPosition := Position{
		SessionKey:   0,
		RacingNumber: 18,
		Position:     8,
		InPit:        false,
		PitOut:       false,
		Stopped:      false,
		Status:       0,
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingPositionUpdate))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	positionsList, err := BuildPositions(response)
	if err != nil {
		t.Fatalf("Could not build positions list: %v", err)
	}
	sort.Slice(positionsList, func(i, j int) bool {
		return positionsList[i].Position < positionsList[j].Position
	})
	if len(positionsList) != 6 {
		t.Fatalf("Expected 6 positions, got: %d", len(positionsList))
	}
	if positionsList[0] != lanceStrollPosition {
		t.Fatalf("Expected position: %v, got: %v", lanceStrollPosition, positionsList[0])
	}
}

func TestShouldNotCreateLapTimeWhenDummyMessageIsSent(t *testing.T) {
	client, err := prepareClientReturningDummyMessage()
	defer client.Close()
	if err != nil {
		t.Fatalf("Could not prepare client: %v", err)
	}
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	timingData, err := BuildTimingData(response)
	if err == nil {
		t.Fatalf("Expected error, got timing data: %v", timingData)
	}
	if timingData != nil {
		t.Fatalf("Expected nil timing data, got: %v", timingData)
	}
}

func TestShouldCreateCarDataWhenFirstSubscribing(t *testing.T) {
	carDataElem := CarData{
		Compressed: "7ZQ9DsIwDIXv4rkgO3Z+mhVxA1hADAghgYQ6QLeqd6cENgJEZkBCXdyo7RfbL87rYN605+P+AnHdwbLdQQSDRiZEE8YF+Wg52nrK4p2wrKCC2fY8/N0B3cLssG2a/Sm9QIhYgUmRUxSIhFKBfTxlWIS+Tx8KWEwk3jm8cYRa8JtqSV2uy4Bch/CevScNyqQmpxE5osJujdEmZi2o1df4XKtoS1tl7Thx2Tg9g1Km0Ytbo9XJ2gwo6D6wKanTnqrPHU4JGHLiDgNsi2QadnhnZD546wOPRva9kQlSPRrZaGS/NjKumf/SyDb9FQ==",
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingCarDataInfo))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	compressedCarData, err := BuildCarData(response)
	if err != nil {
		t.Fatalf("Could not build car data: %v", err)
	}
	if compressedCarData != carDataElem {
		t.Fatalf("Expected car data: %v, got: %v", carDataElem, compressedCarData)
	}
}

func TestShouldCreateCarDataWhenReceivingUpdate(t *testing.T) {
	carDataElem := CarData{
		Compressed: "7ZQ9DsIwDIXv4rkgO3Z+mhVxA1hADAghgYQ6QLeqd6cENgJEZkBCXdyo7RfbL87rYN605+P+AnHdwbLdQQSDRiZEE8YF+Wg52nrK4p2wrKCC2fY8/N0B3cLssG2a/Sm9QIhYgUmRUxSIhFKBfTxlWIS+Tx8KWEwk3jm8cYRa8JtqSV2uy4Bch/CevScNyqQmpxE5osJujdEmZi2o1df4XKtoS1tl7Thx2Tg9g1Km0Ytbo9XJ2gwo6D6wKanTnqrPHU4JGHLiDgNsi2QadnhnZD546wOPRva9kQlSPRrZaGS/NjKumf/SyDb9FQ==",
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingTimingDataUpdate))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	compressedCarData, err := BuildCarData(response)
	if err != nil {
		t.Fatalf("Could not build car data: %v", err)
	}
	if compressedCarData != carDataElem {
		t.Fatalf("Expected car data: %v, got: %v", carDataElem, compressedCarData)
	}
}

func TestShouldCreateRaceControlMessageWhenFirstSubscribing(t *testing.T) {
	raceControlMessage := RaceControl{
		Utc:      "2024-11-30T17:46:16",
		Category: "Other",
		Message:  "PINK HEAD PADDING MATERIAL MUST BE USED",
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingCarDataInfo))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	raceControls, err := BuildRaceControl(response)
	if err != nil {
		t.Fatalf("Could not build race control message")
	}
	raceControl := raceControls[0]
	if raceControl != raceControlMessage {
		t.Fatalf("Expected message: %v but got: %v\n", raceControlMessage, raceControl)
	}
}

func TestShouldCreateRaceControlMessageWhenReceivingUpdate(t *testing.T) {
	raceControlMessage := RaceControl{
		Utc:      "2024-11-30T17:46:16",
		Category: "Other",
		Message:  "PINK HEAD PADDING MATERIAL MUST BE USED",
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingSessionUpdate))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	raceControl, err := BuildRaceControl(response)
	if err != nil {
		t.Fatalf("Could not build race control message: %v\n", err)
	}
	if raceControl[0] != raceControlMessage {
		t.Fatalf("Expected message: %v but got: %v\n", raceControlMessage, raceControl)
	}
}

func TestShouldCreateStintWhenFirstSubscribing(t *testing.T) {
	firstStint := Stint{
		RacingNumber:    1,
		LapFlags:        1,
		Compound:        "HARD",
		New:             true,
		TyresNotChanged: 0,
		TotalLaps:       29,
		StartLaps:       0,
		StintNumber:     1,
	}
	secondStint := Stint{
		RacingNumber:    1,
		LapFlags:        0,
		Compound:        "MEDIUM",
		New:             true,
		TyresNotChanged: 0,
		TotalLaps:       29,
		StartLaps:       0,
		StintNumber:     0,
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingTimingAppStints))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	stints, err := BuildStints(response)
	if err != nil {
		t.Fatalf("Could not build stints: %v", err)
	}
	if len(stints) != 48 {
		t.Fatalf("Expected 48 stints, got: %d", len(stints))
	}
	sort.Slice(stints, func(i, j int) bool {
		return stints[i].RacingNumber < stints[j].RacingNumber
	})
	if stints[0] != firstStint {
		t.Fatalf("Expected first stint: %v, got: %v", firstStint, stints[0])
	}
	if stints[1] != secondStint {
		t.Fatalf("Expected second stint: %v, got: %v", secondStint, stints[1])
	}
}

func TestShouldCreateStintWhenReceivingUpdate(t *testing.T) {
	expectedStint := Stint{
		RacingNumber:    1,
		LapFlags:        0,
		Compound:        "MEDIUM",
		New:             true,
		TyresNotChanged: 0,
		TotalLaps:       0,
		StartLaps:       0,
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingTimingAppDataUpdate))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	stints, err := BuildStints(response)
	if err != nil {
		t.Fatalf("Could not build stints: %v", err)
	}
	if len(stints) != 17 {
		t.Fatalf("Expected 17 stints, got: %d", len(stints))
	}
	sort.Slice(stints, func(i, j int) bool {
		return stints[i].RacingNumber < stints[j].RacingNumber
	})
	if stints[0] != expectedStint {
		t.Fatalf("Expected stint: %v, got: %v", expectedStint, stints[0])
	}
}

func TestShouldCreateSectorsWhenFirstMessageIsSent(t *testing.T) {
	driverSectors := AllSectors{
		RacingNumber: 4,
		Sectors: []Sector{
			{
				SectorNumber:    1,
				Value:           "17.513",
				PersonalFastest: true,
				OverallFastest:  false,
			},
			{
				SectorNumber:    2,
				Value:           "39.702",
				PersonalFastest: false,
				OverallFastest:  false,
			},
			{
				SectorNumber:    3,
				Value:           "32.643",
				PersonalFastest: false,
				OverallFastest:  false,
			},
		},
	}
	response, err := prepareClientreturningTimingAppFirstResponse()
	if err != nil {
		t.Fatalf("%v", err)
	}
	sectors, err := BuildSectors(response)
	if err != nil {
		t.Fatalf("Could not build sectors: %v", err)
	}
	sort.Slice(sectors, func(i, j int) bool {
		return sectors[i].RacingNumber < sectors[j].RacingNumber
	})
	if len(sectors) != 20 {
		t.Fatalf("Expected 20 sectors, got: %d", len(sectors))
	}
	if sectors[1].RacingNumber != driverSectors.RacingNumber || !compareSectors(sectors[1].Sectors, driverSectors.Sectors) {
		t.Fatalf("Expected sectors: %v, got: %v", driverSectors, sectors[1])
	}
}

func compareSectors(a, b []Sector) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestShouldCreateSectorsWhenUpdateIsSent(t *testing.T) {
	driverSectors := AllSectors{
		RacingNumber: 63,
		Sectors: []Sector{
			{
				SectorNumber:    2,
				Value:           "30.403",
				OverallFastest:  false,
				PersonalFastest: false,
			},
		},
		Utc: func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2024-12-07T14:33:13.195Z")
			return t
		}(),
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingTimingDataUpdate))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not open a websocket connection: %v", err)
	}
	defer client.Close()
	err = sendDummyMessage(client)
	if err != nil {
		t.Fatalf("Could not write message to websocket: %v", err)
	}
	_, response, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message from websocket: %v", err)
	}
	sectors, err := BuildSectors(response)
	if err != nil {
		t.Fatalf("Could not build sectors: %v", err)
	}
	fmt.Println(sectors)
	if len(sectors) != 1 {
		t.Fatalf("Expected 1 sectors, got: %d", len(sectors))
	}
	if sectors[0].RacingNumber != driverSectors.RacingNumber || sectors[0].Utc != driverSectors.Utc || !compareSectors(sectors[0].Sectors, driverSectors.Sectors) {
		t.Fatalf("Expected sectors: %v, got: %v", driverSectors, sectors[0])
	}

}
