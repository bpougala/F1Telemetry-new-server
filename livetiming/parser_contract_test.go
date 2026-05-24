package livetiming

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

// wrapInUpdateEnvelope takes a raw mock payload [TopicName, Data, Timestamp]
// and wraps it into a full UpdateData envelope that Build* functions expect.
func wrapInUpdateEnvelope(mockPayload []byte) ([]byte, error) {
	var a []interface{}
	if err := json.Unmarshal(mockPayload, &a); err != nil {
		return nil, err
	}
	envelope := UpdateData{
		C: "test",
		M: []Message{{
			H: "Streaming",
			M: "feed",
			A: a,
		}},
	}
	return json.Marshal(envelope)
}

func intPtr(v int) *int { return &v }

func TestContractDriverList(t *testing.T) {
	raw, err := os.ReadFile("../mocks/DriverList_0001.json")
	if err != nil {
		t.Fatalf("Failed to read mock file: %v", err)
	}
	data, err := wrapInUpdateEnvelope(raw)
	if err != nil {
		t.Fatalf("Failed to wrap in envelope: %v", err)
	}

	positions, err := BuildPositions(data)
	if err != nil {
		t.Fatalf("BuildPositions failed: %v", err)
	}

	expected := []Position{
		{RacingNumber: 1, Position: intPtr(7)},
		{RacingNumber: 3, Position: intPtr(6)},
		{RacingNumber: 5, Position: intPtr(12)},
		{RacingNumber: 6, Position: intPtr(13)},
		{RacingNumber: 12, Position: intPtr(2)},
		{RacingNumber: 14, Position: intPtr(11)},
		{RacingNumber: 16, Position: intPtr(5)},
		{RacingNumber: 23, Position: intPtr(1)},
		{RacingNumber: 27, Position: intPtr(10)},
		{RacingNumber: 31, Position: intPtr(14)},
		{RacingNumber: 41, Position: intPtr(9)},
		{RacingNumber: 44, Position: intPtr(4)},
		{RacingNumber: 63, Position: intPtr(3)},
		{RacingNumber: 81, Position: intPtr(8)},
	}

	sort.Slice(positions, func(i, j int) bool {
		return positions[i].RacingNumber < positions[j].RacingNumber
	})

	if len(positions) != len(expected) {
		t.Fatalf("Expected %d positions, got %d", len(expected), len(positions))
	}
	for i, exp := range expected {
		got := positions[i]
		if got.RacingNumber != exp.RacingNumber {
			t.Fatalf("Position[%d]: expected RacingNumber %d, got %d", i, exp.RacingNumber, got.RacingNumber)
		}
		if got.Position == nil || *got.Position != *exp.Position {
			t.Fatalf("Position[%d] (driver %d): expected Position %d, got %v", i, exp.RacingNumber, *exp.Position, got.Position)
		}
	}
}

func TestContractRaceControlMessages(t *testing.T) {
	raw, err := os.ReadFile("../mocks/RaceControlMessages_0038.json")
	if err != nil {
		t.Fatalf("Failed to read mock file: %v", err)
	}
	data, err := wrapInUpdateEnvelope(raw)
	if err != nil {
		t.Fatalf("Failed to wrap in envelope: %v", err)
	}

	messages, err := BuildRaceControl(data)
	if err != nil {
		t.Fatalf("BuildRaceControl failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 race control message, got %d", len(messages))
	}

	expected := RaceControl{
		Category: "Other",
		Message:  "CAR 27 (HUL) TIME 1:34.852 DELETED - TRACK LIMITS AT TURN 14 (NEXT LAP)",
		Utc:      "2026-05-23T20:38:44",
	}
	if messages[0] != expected {
		t.Fatalf("Expected %+v, got %+v", expected, messages[0])
	}
}

func TestContractTimingAppDataTotalLaps(t *testing.T) {
	raw, err := os.ReadFile("../mocks/TimingAppData_0027.json")
	if err != nil {
		t.Fatalf("Failed to read mock file: %v", err)
	}
	data, err := wrapInUpdateEnvelope(raw)
	if err != nil {
		t.Fatalf("Failed to wrap in envelope: %v", err)
	}

	stints, err := BuildStints(data)
	if err != nil {
		t.Fatalf("BuildStints failed: %v", err)
	}

	sort.Slice(stints, func(i, j int) bool {
		return stints[i].RacingNumber < stints[j].RacingNumber
	})

	expected := []Stint{
		{RacingNumber: 5, StintNumber: 0, TotalLaps: 1},
		{RacingNumber: 23, StintNumber: 0, TotalLaps: 2},
		{RacingNumber: 27, StintNumber: 0, TotalLaps: 1},
	}

	if len(stints) != len(expected) {
		t.Fatalf("Expected %d stints, got %d", len(expected), len(stints))
	}
	for i, exp := range expected {
		got := stints[i]
		if got.RacingNumber != exp.RacingNumber {
			t.Fatalf("Stint[%d]: expected RacingNumber %d, got %d", i, exp.RacingNumber, got.RacingNumber)
		}
		if got.StintNumber != exp.StintNumber {
			t.Fatalf("Stint[%d]: expected StintNumber %d, got %d", i, exp.StintNumber, got.StintNumber)
		}
		if got.TotalLaps != exp.TotalLaps {
			t.Fatalf("Stint[%d]: expected TotalLaps %d, got %d", i, exp.TotalLaps, got.TotalLaps)
		}
	}
}

func TestContractTimingAppDataNewCompound(t *testing.T) {
	raw, err := os.ReadFile("../mocks/TimingAppData_0173.json")
	if err != nil {
		t.Fatalf("Failed to read mock file: %v", err)
	}
	data, err := wrapInUpdateEnvelope(raw)
	if err != nil {
		t.Fatalf("Failed to wrap in envelope: %v", err)
	}

	stints, err := BuildStints(data)
	if err != nil {
		t.Fatalf("BuildStints failed: %v", err)
	}

	if len(stints) != 1 {
		t.Fatalf("Expected 1 stint, got %d", len(stints))
	}

	expected := Stint{
		RacingNumber: 18,
		StintNumber:  1,
		Compound:     "SOFT",
		New:          true,
	}
	got := stints[0]
	if got.RacingNumber != expected.RacingNumber {
		t.Fatalf("Expected RacingNumber %d, got %d", expected.RacingNumber, got.RacingNumber)
	}
	if got.StintNumber != expected.StintNumber {
		t.Fatalf("Expected StintNumber %d, got %d", expected.StintNumber, got.StintNumber)
	}
	if got.Compound != expected.Compound {
		t.Fatalf("Expected Compound %s, got %s", expected.Compound, got.Compound)
	}
	if got.New != expected.New {
		t.Fatalf("Expected New %v, got %v", expected.New, got.New)
	}
}

func TestContractTimingAppDataMultipleStints(t *testing.T) {
	raw, err := os.ReadFile("../mocks/TimingAppData_0178.json")
	if err != nil {
		t.Fatalf("Failed to read mock file: %v", err)
	}
	data, err := wrapInUpdateEnvelope(raw)
	if err != nil {
		t.Fatalf("Failed to wrap in envelope: %v", err)
	}

	stints, err := BuildStints(data)
	if err != nil {
		t.Fatalf("BuildStints failed: %v", err)
	}

	sort.Slice(stints, func(i, j int) bool {
		return stints[i].StintNumber < stints[j].StintNumber
	})

	if len(stints) != 2 {
		t.Fatalf("Expected 2 stints, got %d", len(stints))
	}

	// Stint 0: only TotalLaps update
	if stints[0].RacingNumber != 11 {
		t.Fatalf("Stint[0]: expected RacingNumber 11, got %d", stints[0].RacingNumber)
	}
	if stints[0].StintNumber != 0 {
		t.Fatalf("Stint[0]: expected StintNumber 0, got %d", stints[0].StintNumber)
	}
	if stints[0].TotalLaps != 6 {
		t.Fatalf("Stint[0]: expected TotalLaps 6, got %d", stints[0].TotalLaps)
	}

	// Stint 1: new compound with TyresNotChanged
	if stints[1].RacingNumber != 11 {
		t.Fatalf("Stint[1]: expected RacingNumber 11, got %d", stints[1].RacingNumber)
	}
	if stints[1].StintNumber != 1 {
		t.Fatalf("Stint[1]: expected StintNumber 1, got %d", stints[1].StintNumber)
	}
	if stints[1].Compound != "UNKNOWN" {
		t.Fatalf("Stint[1]: expected Compound UNKNOWN, got %s", stints[1].Compound)
	}
	if stints[1].New != false {
		t.Fatalf("Stint[1]: expected New false, got %v", stints[1].New)
	}
	if stints[1].TyresNotChanged != 1 {
		t.Fatalf("Stint[1]: expected TyresNotChanged 1, got %d", stints[1].TyresNotChanged)
	}
}

func TestContractTimingDataSparseUpdate(t *testing.T) {
	raw, err := os.ReadFile("../mocks/TimingData_0282.json")
	if err != nil {
		t.Fatalf("Failed to read mock file: %v", err)
	}
	data, err := wrapInUpdateEnvelope(raw)
	if err != nil {
		t.Fatalf("Failed to wrap in envelope: %v", err)
	}

	// Sparse update with only Segments data — no sector Values, so BuildSectors should return error
	_, err = BuildSectors(data)
	if err == nil {
		t.Fatalf("Expected error from BuildSectors on sparse TimingData update, got nil")
	}

	// No position data in this update either
	_, err = BuildPositions(data)
	if err == nil {
		t.Fatalf("Expected error from BuildPositions on sparse TimingData update, got nil")
	}

	// No timing data (no BestLapTime)
	_, err = BuildTimingData(data)
	if err == nil {
		t.Fatalf("Expected error from BuildTimingData on sparse TimingData update, got nil")
	}

	// BuildSegments should succeed on this sparse update
	segments, segErr := BuildSegments(data)
	if segErr != nil {
		t.Fatalf("Expected BuildSegments to succeed, got error: %v", segErr)
	}
	if len(segments) != 1 {
		t.Fatalf("Expected 1 driver segment group, got %d", len(segments))
	}
	if segments[0].RacingNumber != 31 {
		t.Fatalf("Expected RacingNumber 31, got %d", segments[0].RacingNumber)
	}
	if len(segments[0].Segments) != 1 {
		t.Fatalf("Expected 1 segment, got %d", len(segments[0].Segments))
	}
	seg := segments[0].Segments[0]
	if seg.SectorNumber != 2 {
		t.Fatalf("Expected SectorNumber 2, got %d", seg.SectorNumber)
	}
	if seg.SegmentNumber != 5 {
		t.Fatalf("Expected SegmentNumber 5, got %d", seg.SegmentNumber)
	}
	if seg.Status != 2049 {
		t.Fatalf("Expected Status 2049, got %d", seg.Status)
	}
}
