package store

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	client *s3.Client
	bucket string
	prefix string
}

func NewS3(client *s3.Client, bucket, prefix string) S3 {
	return S3{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}
}

// Get retrieves a single object from S3
func (store S3) Get(key string) ([]byte, error) {
	fullKey := store.prefix + "/" + key

	result, err := store.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &store.bucket,
		Key:    &fullKey,
	})
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := result.Body.Close(); cerr != nil {
			log.Printf("failed to close result body: %v", cerr)
		}
	}()

	return data, nil
}