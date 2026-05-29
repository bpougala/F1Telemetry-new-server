package backfill

import (
	dataingestor "F1Telemetry-new-server/data-ingestor"
	"F1Telemetry-new-server/livetiming"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func RunCircuits() {
	ctx := context.Background()
	dbClient, err := dataingestor.GetDynamoClient(&ctx)
	if err != nil {
		log.Fatalf("failed to create DynamoDB client: %v", err)
	}
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	s3Client := s3.NewFromConfig(awsCfg)

	year := time.Now().Year()
	fmt.Printf("Starting circuit backfill for year %d\n", year)

	meetings, err := dataingestor.FetchMeetings(dbClient, ctx)
	if err != nil {
		log.Fatalf("failed to fetch meetings: %v", err)
	}

	seen := make(map[int]bool)
	var circuitKeys []int
	for _, m := range meetings {
		if m.CircuitKey != 0 && !seen[m.CircuitKey] {
			seen[m.CircuitKey] = true
			circuitKeys = append(circuitKeys, m.CircuitKey)
		}
	}

	fmt.Printf("Found %d unique circuit keys\n", len(circuitKeys))

	for _, ck := range circuitKeys {
		fmt.Printf("  Downloading circuit %d...\n", ck)
		if err := livetiming.DownloadAndStoreCircuitSync(s3Client, ctx, ck, year); err != nil {
			fmt.Printf("  Error: %v\n", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("Circuit backfill complete")
}
