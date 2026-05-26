// Package dynamitedb provides an embedded singletable database engine running on s3.
package dynamitedb

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Bucket struct {
	name   string
	client *s3.Client
}

type BucketOption func(*s3.Options)

// New constructs a dynamitedb bucket pointing to the provided s3 url / bucket.
// Credentials are loaded with aws sdk (e.g. from env AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY).
func New(ctx context.Context, url, bucket string, opts ...BucketOption) (*Bucket, error) {
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
			o.UsePathStyle = true
			for _, opt := range opts {
				opt(o)
			}
		}),
	}, nil
}

// WithCredentials specifies a static access and secret key.
// This disables the default AWS SDK credential process.
func WithCredentials(accessKey, secretKey string) BucketOption {
	return func(o *s3.Options) {
		o.Credentials = credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	}
}

// WithRegion specifies a static bucket region.
// This disables the default AWS SDK region process.
func WithRegion(region string) BucketOption {
	return func(o *s3.Options) {
		o.Region = region
	}
}

// NewFromClient initializes a dynamitedb bucket from an existing aws s3 sdk client.
func NewFromClient(client *s3.Client, bucket string) *Bucket {
	return &Bucket{
		name:   bucket,
		client: client,
	}
}

var keySanitizer = regexp.MustCompile("^[A-Za-z0-9@._-]{1,100}$")

// constructBucketKey extracts, sanitizes and constructs an s3 bucket key string from the schema.
// Returns the raw s3 bucket key and whether it is an exact match or a prefix.
func constructBucketKey(filter reflect.Value) (string, bool, error) {
	partKey, partVal, partExact, err := retrievePartKey(filter)
	if err != nil {
		return "", false, err
	}
	if !keySanitizer.MatchString(partVal) {
		return "", false, fmt.Errorf("database partition key is malformed or contains unsafe characters")
	}

	sortKey, sortVal, sortExact, err := retrieveSortKey(filter)
	if err != nil {
		return "", false, err
	}
	if sortKey != "" && !keySanitizer.MatchString(sortVal) {
		return "", false, fmt.Errorf("database sort key is malformed or contains unsafe characters")
	}

	if sortKey == "" {
		return strings.Join([]string{partKey, partVal}, "/"), partExact, nil
	} else {
		// when specifying a sort key the part key is always an exact match
		return strings.Join([]string{partKey, partVal, sortKey, sortVal}, "/"), sortExact, nil
	}
}

// retrievePartKey extracts the partition key and partition key value from the schema.
// Returns partKey, partValue and if it requires an exactMatch.
func retrievePartKey(filter reflect.Value) (string, string, bool, error) {
	for field := range filter.Fields() {
		partKey := field.Tag.Get("pk")
		if partKey == "" {
			continue
		}
		partVal, ok := filter.FieldByIndex(field.Index).Interface().(KeyField)
		if !ok {
			return "", "", false, fmt.Errorf("partition key '%s' is unset", partKey)
		}
		segment, exact := partVal.query()
		return partKey, segment, exact, nil
	}
	return "", "", false, fmt.Errorf("no partition key found in schema")
}

// retrieveSortKey extracts the sort key and sort key value from the schema.
// If no sort key is present it returns an empty string.
func retrieveSortKey(filter reflect.Value) (string, string, bool, error) {
	for field := range filter.Fields() {
		if sortKey := field.Tag.Get("sk"); sortKey != "" {
			sortVal, ok := filter.FieldByIndex(field.Index).Interface().(KeyField)
			if !ok {
				return "", "", false, fmt.Errorf("sort key '%s' is unset", sortKey)
			}
			segment, exact := sortVal.query()
			return sortKey, segment, exact, nil
		}
	}
	return "", "", false, nil
}

// injectBucketKey takes the raw s3 key, parses it and inserts it into pk / sk fields of the target.
func injectBucketKey(target reflect.Value, key string) error {
	segments := strings.Split(key, "/")
	if len(segments) < 2 {
		return fmt.Errorf("invalid database key: expected '<part-type>/<part-id>/...' got '%s'", key)
	}
	partKey, partValue := segments[0], segments[1]
	if err := writePartKey(target, partKey, partValue); err != nil {
		return err
	}

	if len(segments) > 3 {
		sortKey, sortValue := segments[2], segments[3]
		if err := writeSortKey(target, sortKey, sortValue); err != nil {
			return err
		}
	}

	return nil
}

// writePartKey injects the provided value to the targets partition key with the matching key.
func writePartKey(target reflect.Value, key, value string) error {
	for field := range target.Fields() {
		if field.Tag.Get("pk") != key {
			continue
		}
		target.FieldByIndex(field.Index).Set(reflect.ValueOf(Key(value)))
		return nil
	}
	return fmt.Errorf("partition key '%s' not found in schema", key)
}

// writeSortKey injects the provided value to the targets sort key with the matching key.
func writeSortKey(target reflect.Value, key, value string) error {
	for field := range target.Fields() {
		if field.Tag.Get("sk") != key {
			continue
		}
		target.FieldByIndex(field.Index).Set(reflect.ValueOf(Key(value)))
		return nil
	}
	return fmt.Errorf("sort key '%s' not found in schema", key)
}
