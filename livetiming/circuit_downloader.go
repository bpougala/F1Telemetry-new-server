package livetiming

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const circuitBucketName = "f1-circuit-data"

// DownloadAndStoreCircuitSync downloads circuit geometry from the multiviewer API
// and uploads it to S3 if not already present.
func DownloadAndStoreCircuitSync(s3Client *s3.Client, ctx context.Context, circuitKey int, year int) error {
	s3Key := fmt.Sprintf("circuits/%d/%d.json", circuitKey, year)

	// Check if already in S3
	_, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(circuitBucketName),
		Key:    aws.String(s3Key),
	})
	if err == nil {
		fmt.Printf("circuit %d/%d already in S3, skipping\n", circuitKey, year)
		return nil
	}

	url := fmt.Sprintf("https://api.multiviewer.app/api/v1/circuits/%d/%d", circuitKey, year)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "F1Telemetry-Server/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading circuit %d/%d: %w", circuitKey, year, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("circuit API returned status %d for %d/%d", resp.StatusCode, circuitKey, year)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading circuit response: %w", err)
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(circuitBucketName),
		Key:         aws.String(s3Key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("uploading circuit to S3 (%s): %w", s3Key, err)
	}

	fmt.Printf("circuit %d/%d uploaded to S3\n", circuitKey, year)
	return nil
}

// DownloadAndStoreCircuit is the async wrapper for live streaming use.
func DownloadAndStoreCircuit(s3Client *s3.Client, ctx context.Context, circuitKey int, year int) {
	go func() {
		if err := DownloadAndStoreCircuitSync(s3Client, ctx, circuitKey, year); err != nil {
			fmt.Printf("error downloading circuit: %v\n", err)
		}
	}()
}
