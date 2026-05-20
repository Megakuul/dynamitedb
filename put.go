package dynamitdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Put inserts the provided structure to the database (replaces previous data).
func Put[T any](ctx context.Context, bucket *Bucket, model *T) error {
	modelVal := reflect.ValueOf(model)

	key, exact, err := constructBucketKey(modelVal)
	if err != nil {
		return err
	} else if !exact {
		return fmt.Errorf("put database call requires exact key match")
	}
	body, err := json.Marshal(model)
	if err != nil {
		return err
	}
	_, err = bucket.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket.name),
		Key:    aws.String(key),
		Body:   bytes.NewReader(body),
	})
	if err != nil {
		return err
	}
	return nil
}
