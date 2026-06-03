package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func _main() {
	file, err := os.Open("australian-fp2-1.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the entire file into memory to avoid the single-line minified JSON trap
	bytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	content := string(bytes)

	// We count the exact raw F1 Stream JSON keys, not your DB struct tags.
	// strings.Count accurately finds every occurrence, even within nested arrays on a single line.

	positionsCount := strings.Count(content, `"Position":`)

	// Sectors are broadcast in a "Sectors" array or object depending on the message type
	sectorsCount := strings.Count(content, `"Sectors":`)

	// Stints are tracked using the "Compound" key in the raw stream
	stintsCount := strings.Count(content, `"Compound":`)

	// Intervals and gaps are tracked via TimeDiffToFastest and TimeDiffToPositionAhead
	// Counting "TimeDiffToFastest" gives us the exact number of interval blocks broadcast
	intervalsCount := strings.Count(content, `"TimeDiffToFastest":`)

	fmt.Println("=== Add These Strong Assertions to Your Test ===")
	fmt.Printf("assert.Equal(t, %d, len(actualPositions), \"Should save the exact number of PositionsDB records\")\n", positionsCount)
	fmt.Printf("assert.Equal(t, %d, len(actualSectors), \"Should save the exact number of SectorDB records\")\n", sectorsCount)
	fmt.Printf("assert.Equal(t, %d, len(actualStints), \"Should save the exact number of StintDB records\")\n", stintsCount)
	fmt.Printf("assert.Equal(t, %d, len(actualIntervals), \"Should save the exact number of IntervalsDB records\")\n", intervalsCount)
}
