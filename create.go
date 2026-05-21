package dynamitdb

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Create inserts the provided structure to the database if not exists.
func Create[T any](ctx context.Context, bucket *Bucket, model *T, opts ...Option) error {
	key, exact, err := constructBucketKey(reflect.ValueOf(model).Elem())
	if err != nil {
		return err
	} else if !exact {
		return fmt.Errorf("create database call requires exact key match")
	}
	body, err := serialize(model)
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
		Bucket:      aws.String(bucket.name),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		IfNoneMatch: aws.String("*"),
		Expires:     options.expires,
	})
	if err != nil {
		return err
	}
	return nil
}
