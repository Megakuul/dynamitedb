package dynamitedb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// Delete removes the specified object from database.
// If the filter does not match an object this becomes a noop returning nil.
func Delete[T any](ctx context.Context, bucket *Bucket, filter *T) error {
	filterVal := reflect.ValueOf(filter)
	key, exact, err := constructBucketKey(filterVal.Elem())
	if err != nil {
		return err
	} else if !exact {
		return fmt.Errorf("delete database call requires exact key match")
	}

	resp, err := bucket.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket.name),
		Key:    aws.String(key),
	})
	if err != nil {
		if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
			return nil
		}
		return errors.New(err.Error())
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	output, err := deserialize[T](body)
	if err != nil {
		return err
	}
	if !checkFilter(reflect.ValueOf(output), filterVal) {
		return nil
	}
	_, err = bucket.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket:  aws.String(bucket.name),
		Key:     aws.String(key),
		IfMatch: resp.ETag,
	})
	if err != nil {
		var sErr smithy.APIError
		if errors.As(err, &sErr) && sErr.ErrorCode() == "PreconditionFailed" {
			return ErrConcurrencyConflict
		}
		return errors.New(err.Error())
	}
	return nil
}
