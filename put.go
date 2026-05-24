package dynamitedb

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Put inserts the provided structure to the database (replaces previous data).
func Put[T any](ctx context.Context, bucket *Bucket, model *T, opts ...Option) error {
	modelVal := reflect.ValueOf(model)
	key, exact, err := constructBucketKey(modelVal.Elem())
	if err != nil {
		return err
	} else if !exact {
		return fmt.Errorf("put database call requires exact key match")
	}
	insert := reflect.New(reflect.TypeFor[T]())
	applyUpdate(insert, modelVal)
	body, err := serialize(insert.Interface().(*T))
	if err != nil {
		return err
	}
	options := &options{
		expires: nil,
	}
	for _, opt := range opts {
		opt(options)
	}
	_, err = bucket.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:  aws.String(bucket.name),
		Key:     aws.String(key),
		Body:    bytes.NewReader(body),
		Expires: options.expires,
	})
	if err != nil {
		return err
	}
	return nil
}
