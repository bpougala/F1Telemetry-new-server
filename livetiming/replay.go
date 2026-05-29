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
// portion of the race into the resolver using actual timestamp deltas for pacing.
// It loops forever so subscribers always see data.
func ReplayPositionZ(resolver *graph.Resolver) {
	frames, err := fetchArchiveFrames()
	if err != nil {
		fmt.Printf("replay: failed to fetch archive: %v\n", err)
		return
	}

	// Take the last 600 frames (~2.5 minutes of data at ~4 frames/sec)
	start := len(frames) - 600
	if start < 0 {
		start = 0
	}
	frames = frames[start:]
	fmt.Printf("replay: loaded %d Position.z frames, looping\n", len(frames))

	for {
		for i, frame := range frames {
			posRoot, err := DecompressPositionData(frame.compressed)
			if err != nil {
				fmt.Printf("replay: decompress error: %v\n", err)
				continue
			}
			resolver.NotifyDriverLocationSubscribers(positionRootToDriverPositions(posRoot))

			if i+1 < len(frames) {
				delta := frames[i+1].timestamp.Sub(frame.timestamp)
				if delta <= 0 || delta > 2*time.Second {
					delta = 250 * time.Millisecond
				}
				time.Sleep(delta)
			}
		}
		fmt.Println("replay: loop complete, restarting")
	}
}

type replayFrame struct {
	timestamp  time.Time
	compressed string
}

func fetchArchiveFrames() ([]replayFrame, error) {
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

	var result []replayFrame
	for _, line := range allLines {
		line = strings.TrimRight(line, "\r")
		quoteIdx := strings.Index(line, "\"")
		if quoteIdx == -1 {
			continue
		}
		compressed := strings.Trim(line[quoteIdx:], "\"")
		if compressed == "" {
			continue
		}
		ts, err := time.Parse("15:04:05.000", strings.TrimSpace(line[:quoteIdx]))
		if err != nil {
			continue
		}
		result = append(result, replayFrame{timestamp: ts, compressed: compressed})
	}
	return result, nil
}
