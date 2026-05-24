package graph

import (
	"context"

	dataingestor "F1Telemetry-new-server/data-ingestor"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var resolverCtx = context.Background()
var dbClient, _ = dataingestor.GetDynamoClient(&resolverCtx)

// GetDBClient returns the shared DynamoDB client for query resolvers.
func GetDBClient() *dynamodb.Client {
	return dbClient
}
