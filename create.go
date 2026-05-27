package dynamitedb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

// Create inserts the provided structure to the database if not exists.
func Create[T any](ctx context.Context, bucket *Bucket, model *T, opts ...Option) error {
	modelVal := reflect.ValueOf(model)
	key, exact, err := constructBucketKey(modelVal.Elem())
	if err != nil {
		return err
	} else if !exact {
		return fmt.Errorf("create database call requires exact key match")
	}
	// the model is actually just an update applied to an empty insert object.
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
		Bucket:      aws.String(bucket.name),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		IfNoneMatch: aws.String("*"),
		Expires:     options.expires,
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		var sErr smithy.APIError
		if errors.As(err, &sErr) && sErr.ErrorCode() == "PreconditionFailed" {
			return ErrAlreadyExists
		}
		return errors.New(err.Error())
	}
	return nil
}
