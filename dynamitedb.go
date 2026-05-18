package dynamitdb

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	bucket string
	client *s3.Client
}

// New constructs a dynamitdb client to the provided s3 url / bucket.
// Credentials are loaded with aws sdk (e.g. from env AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY).
func New(ctx context.Context, url, bucket string) (*Client, error) {
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

	return &Client{
		bucket: bucket,
		client: s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(url)
		}),
	}, nil
}

// Get fetches a single entry from database based on the keypair in v (also writing to v).
// Requires PK (and if defined by schema SK) to be set in v.
func (c *Client) Get(ctx context.Context, filter any) error {
	input := reflect.ValueOf(filter)

	key, err := constructBucketKey(input)
	if err != nil {
		return err
	}
	resp, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, filter)
}

// Create inserts the provided structure to the database if not exists.
func (c *Client) Create(ctx context.Context, value any) error {
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
func (c *Client) Put(ctx context.Context, value any) error {
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
