package client

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/guardian/cq-source-image-packages/store"
	"github.com/rs/zerolog"
)

type Client struct {
	logger          zerolog.Logger
	Spec            Spec
	S3Store         store.S3
	BakesTable      store.DynamoDb
	RecipesTable    store.DynamoDb
	BaseImagesTable store.DynamoDb
}

func (c *Client) ID() string {
	return "guardian/image-packages"
}

func (c *Client) Logger() *zerolog.Logger {
	return &c.logger
}

func New(ctx context.Context, logger zerolog.Logger, s *Spec) (Client, error) {

	// Loads credentials from the default credential chain.
	// Locally, set the AWS_PROFILE environment variable, or run `make serve`.
	// See https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials.
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("eu-west-1"),
	)
	if err != nil {
		log.Fatalf("unable to load AWS config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	s3Store := store.NewS3(s3Client, s.AmigoBucketName, "packagelists")

	dynamoDbClient := dynamodb.NewFromConfig(cfg)
	bakesTable := store.NewDynamoDb(dynamoDbClient, s.AmigoBakesTableName)
	recipesTable := store.NewDynamoDb(dynamoDbClient, s.AmigoRecipesTableName)
	baseImagesTable := store.NewDynamoDb(dynamoDbClient, s.AmigoBaseImagesTableName)

	return Client{
		logger:          logger,
		Spec:            *s,
		S3Store:         s3Store,
		BakesTable:      bakesTable,
		RecipesTable:    recipesTable,
		BaseImagesTable: baseImagesTable,
	}, nil
}
