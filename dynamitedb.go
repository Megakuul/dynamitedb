package dynamitdb

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Bucket struct {
	name   string
	client *s3.Client
}

// New constructs a dynamitedb bucket to the provided s3 url / bucket.
// Credentials are loaded with aws sdk (e.g. from env AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY).
func New(ctx context.Context, url, bucket string) (*Bucket, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithHTTPClient(&http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}))
	if err != nil {
		return nil, err
	}

	return &Bucket{
		name: bucket,
		client: s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(url)
		}),
	}, nil
}

// Create inserts the provided structure to the database if not exists.
func (c *Bucket) Create(ctx context.Context, value any) error {
	input := reflect.ValueOf(value)

	key, err := constructBucketKey(input)
	if err != nil {
		return err
	}
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		IfNoneMatch: aws.String("*"),
	})
	if err != nil {
		return err
	}
	return nil
}

// Put inserts the provided structure to the database (replaces previous data).
func (c *Bucket) Put(ctx context.Context, value any) error {
	input := reflect.ValueOf(value)

	key, err := constructBucketKey(input)
	if err != nil {
		return err
	}
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(body),
	})
	if err != nil {
		return err
	}
	return nil
}
