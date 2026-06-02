package livetiming

import (
	appconfig "F1Telemetry-new-server/config"
	"F1Telemetry-new-server/graph"
	"F1Telemetry-new-server/mocksignalr"
	"F1Telemetry-new-server/ws"
	"context"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

const replayDataFile = "../australian-fp2-1.txt"

type tableSchema struct {
	name string
	pk   string
	pkT  ddbtypes.ScalarAttributeType
	sk   string // empty means no sort key
	skT  ddbtypes.ScalarAttributeType
}

// replaySchemas mirrors infra/stack.go:84-95 so the LocalStack tables match production.
var replaySchemas = []tableSchema{
	{name: "meetings", pk: "MeetingKey", pkT: ddbtypes.ScalarAttributeTypeN},
	{name: "sessions", pk: "MeetingKey", pkT: ddbtypes.ScalarAttributeTypeN, sk: "SessionKey", skT: ddbtypes.ScalarAttributeTypeN},
	{name: "drivers", pk: "SessionKey", pkT: ddbtypes.ScalarAttributeTypeN, sk: "RacingNumber", skT: ddbtypes.ScalarAttributeTypeN},
	{name: "positions", pk: "SessionKey", pkT: ddbtypes.ScalarAttributeTypeN, sk: "RacingNumber", skT: ddbtypes.ScalarAttributeTypeN},
	{name: "timings", pk: "SessionKey", pkT: ddbtypes.ScalarAttributeTypeN, sk: "RacingNumber", skT: ddbtypes.ScalarAttributeTypeN},
	{name: "sectors", pk: "SessionKey", pkT: ddbtypes.ScalarAttributeTypeN, sk: "Reference", skT: ddbtypes.ScalarAttributeTypeS},
	{name: "trackstatus", pk: "SessionKey", pkT: ddbtypes.ScalarAttributeTypeN, sk: "Utc", skT: ddbtypes.ScalarAttributeTypeN},
	{name: "racecontrol", pk: "SessionKey", pkT: ddbtypes.ScalarAttributeTypeN, sk: "Utc", skT: ddbtypes.ScalarAttributeTypeN},
	{name: "stints", pk: "SessionKey", pkT: ddbtypes.ScalarAttributeTypeN, sk: "RacingNumber", skT: ddbtypes.ScalarAttributeTypeN},
	{name: "weather", pk: "SessionKey", pkT: ddbtypes.ScalarAttributeTypeN, sk: "Utc", skT: ddbtypes.ScalarAttributeTypeN},
}

// TestReplaySessionPersistsToDynamoDB replays a recorded F1 session line-by-line through
// the real connection pipeline (Negotiate -> SetWebSocket -> ProcessSessionDataAndInfo)
// against a local mock SignalR server, with DynamoDB and S3 backed by LocalStack, then
// asserts the persisted data. This exercises the live path that unit tests skip.
func TestReplaySessionPersistsToDynamoDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping LocalStack replay integration test in -short mode")
	}

	ctx := context.Background()

	// 1. Start LocalStack. DynamoDB + S3 are community (free) features, so no auth
	// token is required. If LOCALSTACK_AUTH_TOKEN is set in the environment, forward it
	// into the container so Pro activation succeeds (the `localstack auth set-token` CLI
	// command only writes the CLI config file, which testcontainers does not read).
	opts := []testcontainers.ContainerCustomizer{}
	if token := os.Getenv("LOCALSTACK_AUTH_TOKEN"); token != "" {
		opts = append(opts, testcontainers.WithEnv(map[string]string{"LOCALSTACK_AUTH_TOKEN": token}))
	}
	lsContainer, err := localstack.Run(ctx, "localstack/localstack:latest", opts...)
	if err != nil {
		t.Fatalf("failed to start LocalStack: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(lsContainer); err != nil {
			t.Logf("failed to terminate LocalStack: %v", err)
		}
	})

	endpoint, err := lsContainer.PortEndpoint(ctx, "4566/tcp", "http")
	if err != nil {
		t.Fatalf("failed to get LocalStack endpoint: %v", err)
	}

	// 2. Build AWS clients pointed at LocalStack.
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		t.Fatalf("failed to load AWS config: %v", err)
	}
	dbClient := dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	// 3. Create tables + buckets.
	createTables(t, ctx, dbClient)
	createBucket(t, ctx, s3Client, "f1-livetiming-data")
	createBucket(t, ctx, s3Client, "f1-circuit-data")

	// 4. Start the mock SignalR server streaming the recorded session.
	mockSrv := httptest.NewServer(mocksignalr.Handler(replayDataFile))
	defer mockSrv.Close()
	cfg := &appconfig.Config{
		NegotiateBaseURL: mockSrv.URL,
		ConnectBaseURL:   strings.Replace(mockSrv.URL, "http://", "ws://", 1),
	}

	// 5. Construct pipeline dependencies.
	resolver := graph.NewResolver()
	hub := ws.NewHub(dbClient)
	go hub.Run()

	// 6. Drive the real connection flow.
	cookies, connObj, err := Negotiate(cfg)
	if err != nil {
		t.Fatalf("Negotiate failed: %v", err)
	}
	conn, _, err := SetWebSocket(cfg, connObj.ConnectionToken, cookies)
	if err != nil {
		t.Fatalf("SetWebSocket failed: %v", err)
	}
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		// Returns with an error when the mock closes the socket after replay — expected.
		_ = ProcessSessionDataAndInfo(conn, dbClient, s3Client, ctx, resolver, hub, connObj.KeepAliveTimeout)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Minute):
		t.Fatal("replay did not finish within timeout")
	}

	// 7. Verify the data landed in DynamoDB.
	const meetingKey = 1254
	const sessionKey = 9687

	if got := scanCount(t, ctx, dbClient, "meetings"); got != 1 {
		t.Errorf("meetings: expected 1 row, got %d", got)
	}
	if name := getStringAttr(t, ctx, dbClient, "meetings",
		map[string]ddbtypes.AttributeValue{"MeetingKey": numAttr(meetingKey)}, "MeetingName"); name != "Australian Grand Prix" {
		t.Errorf("meetings: expected MeetingName 'Australian Grand Prix', got %q", name)
	}

	sessionKeyAV := map[string]ddbtypes.AttributeValue{
		"MeetingKey": numAttr(meetingKey),
		"SessionKey": numAttr(sessionKey),
	}
	if name := getStringAttr(t, ctx, dbClient, "sessions", sessionKeyAV, "Name"); name != "Practice 2" {
		t.Errorf("sessions: expected Name 'Practice 2', got %q", name)
	}
	if typ := getStringAttr(t, ctx, dbClient, "sessions", sessionKeyAV, "Type"); typ != "Practice" {
		t.Errorf("sessions: expected Type 'Practice', got %q", typ)
	}
	// The archive-complete SessionInfo/SessionData update writes ArchiveStatus via UpdateItem
	// on the composite (MeetingKey, SessionKey). A regression to a partial key would fail
	// silently (logged only), leaving this attribute unset.
	if status := getStringAttr(t, ctx, dbClient, "sessions", sessionKeyAV, "ArchiveStatus"); status != "Complete" {
		t.Errorf("sessions: expected ArchiveStatus 'Complete', got %q", status)
	}

	if got := scanCount(t, ctx, dbClient, "drivers"); got != 20 {
		t.Errorf("drivers: expected 20 rows, got %d", got)
	}
	// positions is a current-state table keyed on (SessionKey, RacingNumber), so it
	// holds one row per driver. Every Position update upserts in place rather than
	// appending, so the row count stays at the driver count regardless of update volume.
	if got := scanCount(t, ctx, dbClient, "positions"); got != 20 {
		t.Errorf("positions: expected 20 rows, got %d", got)
	}
	if got := scanCount(t, ctx, dbClient, "trackstatus"); got == 0 {
		t.Errorf("trackstatus: expected at least 1 row, got 0")
	}
}

func createTables(t *testing.T, ctx context.Context, client *dynamodb.Client) {
	t.Helper()
	for _, s := range replaySchemas {
		attrs := []ddbtypes.AttributeDefinition{
			{AttributeName: aws.String(s.pk), AttributeType: s.pkT},
		}
		keys := []ddbtypes.KeySchemaElement{
			{AttributeName: aws.String(s.pk), KeyType: ddbtypes.KeyTypeHash},
		}
		if s.sk != "" {
			attrs = append(attrs, ddbtypes.AttributeDefinition{AttributeName: aws.String(s.sk), AttributeType: s.skT})
			keys = append(keys, ddbtypes.KeySchemaElement{AttributeName: aws.String(s.sk), KeyType: ddbtypes.KeyTypeRange})
		}
		_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName:            aws.String(s.name),
			AttributeDefinitions: attrs,
			KeySchema:            keys,
			BillingMode:          ddbtypes.BillingModePayPerRequest,
		})
		if err != nil {
			t.Fatalf("create table %s: %v", s.name, err)
		}
		if err := dynamodb.NewTableExistsWaiter(client).Wait(ctx,
			&dynamodb.DescribeTableInput{TableName: aws.String(s.name)}, 2*time.Minute); err != nil {
			t.Fatalf("waiting for table %s: %v", s.name, err)
		}
	}
}

func createBucket(t *testing.T, ctx context.Context, client *s3.Client, name string) {
	t.Helper()
	if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(name)}); err != nil {
		t.Fatalf("create bucket %s: %v", name, err)
	}
}

func scanCount(t *testing.T, ctx context.Context, client *dynamodb.Client, table string) int {
	t.Helper()
	out, err := client.Scan(ctx, &dynamodb.ScanInput{TableName: aws.String(table), Select: ddbtypes.SelectCount})
	if err != nil {
		t.Fatalf("scan %s: %v", table, err)
	}
	return int(out.Count)
}

func getStringAttr(t *testing.T, ctx context.Context, client *dynamodb.Client, table string, key map[string]ddbtypes.AttributeValue, attr string) string {
	t.Helper()
	out, err := client.GetItem(ctx, &dynamodb.GetItemInput{TableName: aws.String(table), Key: key})
	if err != nil {
		t.Fatalf("get item from %s: %v", table, err)
	}
	if out.Item == nil {
		t.Fatalf("get item from %s: no item for key %v", table, key)
	}
	v, ok := out.Item[attr].(*ddbtypes.AttributeValueMemberS)
	if !ok {
		t.Fatalf("get item from %s: attribute %q missing or not a string", table, attr)
	}
	return v.Value
}

func numAttr(n int) ddbtypes.AttributeValue {
	return &ddbtypes.AttributeValueMemberN{Value: strconv.Itoa(n)}
}
