package livetiming

import (
	"F1Telemetry-new-server/graph"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const replayArchiveURL = "https://livetiming.formula1.com/static/2026/2026-05-24_Canadian_Grand_Prix/2026-05-24_Race/Position.z.jsonStream"

// ReplayPositionZ fetches archived Position.z data and replays the last
// portion of the race into the resolver at real-time pace (~250ms between frames).
// It loops forever so subscribers always see data.
func ReplayPositionZ(resolver *graph.Resolver) {
	lines, err := fetchArchiveLines()
	if err != nil {
		fmt.Printf("replay: failed to fetch archive: %v\n", err)
		return
	}

	// Take the last 600 lines (~2.5 minutes of data at ~4 lines/sec)
	start := len(lines) - 600
	if start < 0 {
		start = 0
	}
	lines = lines[start:]
	fmt.Printf("replay: loaded %d Position.z frames, looping\n", len(lines))

	for {
		for _, compressed := range lines {
			posRoot, err := DecompressPositionData(compressed)
			if err != nil {
				fmt.Printf("replay: decompress error: %v\n", err)
				continue
			}
			resolver.NotifyDriverLocationSubscribers(positionRootToDriverPositions(posRoot))
			time.Sleep(250 * time.Millisecond)
		}
		fmt.Println("replay: loop complete, restarting")
	}
}

func fetchArchiveLines() ([]string, error) {
	req, err := http.NewRequest("GET", replayArchiveURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "BestHTTP")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	raw := string(bytes.TrimPrefix(body, []byte("\xef\xbb\xbf")))
	allLines := strings.Split(strings.TrimSpace(raw), "\n")

	var result []string
	for _, line := range allLines {
		line = strings.TrimRight(line, "\r")
		quoteIdx := strings.Index(line, "\"")
		if quoteIdx == -1 {
			continue
		}
		compressed := strings.Trim(line[quoteIdx:], "\"")
		if compressed != "" {
			result = append(result, compressed)
		}
	}
	return result, nil
}
