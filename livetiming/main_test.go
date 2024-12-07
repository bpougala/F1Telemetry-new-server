package livetiming

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
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
		Position:     2,
		InPit:        false,
		PitOut:       false,
		Stopped:      false,
		Status:       64,
	}
	server := httptest.NewServer(http.HandlerFunc(serverSendingTimingAppRaceUpdates))
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
