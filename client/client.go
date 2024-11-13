package client

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/guardian/cq-source-images-instances/store"
	"github.com/rs/zerolog"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	logger zerolog.Logger
	Spec   Spec
	Store  store.S3
}

func (c *Client) ID() string {
	return "guardian/images-instances"
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
		fmt.Fprintf(os.Stderr, "unable to load AWS config, %v", err)
		os.Exit(1)
	}

	client := s3.NewFromConfig(cfg)

	s3store := store.New(client, s.AmigoBucketName, "packagelists")

	return Client{
		logger: logger,
		Spec:   *s,
		Store:  s3store,
	}, nil
}
