package livetiming

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const s3BucketName = "f1-livetiming-data"

type S3Logger struct {
	client       *s3.Client
	ctx          context.Context
	raceName     string
	sessionName  string
	partCounters map[string]int
	disabled     bool
	mu           sync.Mutex
}

func NewS3Logger(client *s3.Client, ctx context.Context) *S3Logger {
	return &S3Logger{
		client:       client,
		ctx:          ctx,
		partCounters: make(map[string]int),
		// Per-message archival is irrelevant to load tests and floods S3; allow
		// disabling it so it doesn't compete with the live broadcast path.
		disabled: os.Getenv("DISABLE_S3_LOGGER") != "",
	}
}

func (l *S3Logger) SetSession(raceName, sessionName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.raceName = strings.ReplaceAll(raceName, " ", "_")
	l.sessionName = strings.ReplaceAll(sessionName, " ", "_")
	l.partCounters = make(map[string]int)
}

func (l *S3Logger) LogMessage(dataType string, data []byte) {
	if l.disabled {
		return
	}
	l.mu.Lock()
	if l.raceName == "" || l.sessionName == "" {
		l.mu.Unlock()
		return
	}
	l.partCounters[dataType]++
	part := l.partCounters[dataType]
	key := fmt.Sprintf("%s/%s/%s_%04d.json", l.raceName, l.sessionName, dataType, part)
	l.mu.Unlock()

	go func() {
		_, err := l.client.PutObject(l.ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s3BucketName),
			Key:         aws.String(key),
			Body:        bytes.NewReader(data),
			ContentType: aws.String("application/json"),
		})
		if err != nil {
			fmt.Printf("error uploading to S3 (%s): %v\n", key, err)
		}
	}()
}
