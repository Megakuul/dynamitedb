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
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Bucket struct {
	name   string
	client *s3.Client
}

// New constructs a dynamitedb bucket pointing to the provided s3 url / bucket.
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
			o.UsePathStyle = true
		}),
	}, nil
}

// NewFromClient initializes a dynamitedb bucket from an existing aws s3 sdk client.
func NewFromClient(client *s3.Client, bucket string) *Bucket {
	return &Bucket{
		name:   bucket,
		client: client,
	}
}

// TODO if this bottlenecks just write a small whitelist char checker
var keySanitizer = regexp.MustCompile("^[A-Za-z0-9._-]{1,100}$")

// constructBucketKey extracts, sanitizes and constructs an s3 bucket key string from the schema.
// Returns the raw s3 bucket key and whether it is an exact match or a prefix.
func constructBucketKey(filter reflect.Value) (string, bool, error) {
	partKey, partVal, partExact, err := retrievePartKey(filter)
	if err != nil {
		return "", false, err
	}
	if !keySanitizer.MatchString(partKey) {
		return "", false, fmt.Errorf("database partition key contains unsafe characters")
	}

	sortKey, sortVal, sortExact, err := retrieveSortKey(filter)
	if err != nil {
		return "", false, err
	}
	if sortKey != "" && !keySanitizer.MatchString(sortKey) {
		return "", false, fmt.Errorf("database sort key contains unsafe characters")
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
