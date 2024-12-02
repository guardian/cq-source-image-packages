package store

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDb struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDb(client *dynamodb.Client, tableName string) DynamoDb {
	return DynamoDb{
		client:    client,
		tableName: tableName,
	}
}

// ListAll List all items in table
func (store DynamoDb) ListAll() []map[string]types.AttributeValue {
	result, err := store.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: &store.tableName,
	})
	if err != nil {
		log.Fatalf("failed to scan table, %v", err)
	}

	// If the scan limit is hit, the result will contain a LastEvaluatedKey
	// In this case we'll need to modify the code to paginate through the results.
	if result.LastEvaluatedKey != nil {
		log.Fatalf("Table %s has more items than the scan limit", store.tableName)
	}

	log.Printf("Found %d items in table %s", result.ScannedCount, store.tableName)

	return result.Items
}
