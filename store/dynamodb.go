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
// TODO: stream results instead - use dynamodbstreams service
func (store DynamoDb) ListAll() []map[string]types.AttributeValue {
	result, err := store.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: &store.tableName,
	})
	if err != nil {
		log.Fatalf("failed to scan table, %v", err)
	}

	return result.Items
}

// Get Fetch a single item from table
func (store DynamoDb) Get(id string) (map[string]types.AttributeValue, error) {
	result, err := store.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: &store.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}

	return result.Item, nil
}
