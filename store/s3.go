package store

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	bucket string
	prefix string
	client *s3.Client
}

func New(client *s3.Client, bucket, prefix string) S3 {
	return S3{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}
}

func (store S3) Get(key string) ([]byte, error) {
	fullKey := store.prefix + "/" + key

	fmt.Println("*** Full key:", fullKey)

	res, err := store.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &store.bucket,
		Key:    &fullKey,
	})
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return data, nil
}

func (store S3) ListKeys() ([]string, error) {
	prefix := store.prefix

	res, err := store.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: &store.bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, err
	}

	fmt.Println("prefix:", prefix)
	a := fmt.Sprintf("%s/", prefix)
	fmt.Println("a:", a)

	var keys []string
	for _, item := range res.Contents {
		key := strings.TrimPrefix(strings.TrimPrefix(*item.Key, prefix), "/")
		if key != "." {
			keys = append(keys, key)
		}
	}

	// TODO: paginate (currently does 1000 max)

	fmt.Println("keys:", keys)

	return keys, nil
}
