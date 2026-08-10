package dynamitedb

import (
	"bytes"
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

// Update changes the entry on the database (identified by PK / SK).
// For update operations see "Field" API on the schema.
// Update will modify v to represent the final state of the updated entry.
func Update[T any](ctx context.Context, bucket *Bucket, update *T, opts ...Option) error {
	modelVal := reflect.ValueOf(update)
	key, exact, err := constructBucketKey(modelVal.Elem())
	if err != nil {
		return err
	} else if !exact {
		return fmt.Errorf("update database call requires exact key match")
	}
	etag := retrieveETag(modelVal.Elem())
	options := &options{
		expires: nil,
	}
	for _, opt := range opts {
		opt(options)
	}
	originalResp, err := bucket.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket:  aws.String(bucket.name),
		Key:     aws.String(key),
		IfMatch: etag,
	})
	if err != nil {
		var sErr smithy.APIError
		if errors.As(err, &sErr) && sErr.ErrorCode() == "PreconditionFailed" {
			return ErrConcurrencyConflict
		}
		if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
			return ErrNotFound
		}
		return errors.New(err.Error())
	}
	// certain providers do not implement IfMatch on GetObject. Therefore this extra check.
	if etag != nil && *originalResp.ETag != *etag {
		return ErrConcurrencyConflict
	}
	originalBody, err := io.ReadAll(originalResp.Body)
	if err != nil {
		return err
	}

	updatedBody, err := updateObject(originalBody, update)
	if err != nil {
		return err
	}

	_, err = bucket.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket.name),
		Key:         aws.String(key),
		Body:        bytes.NewReader(updatedBody),
		IfMatch:     originalResp.ETag,
		Expires:     options.expires,
		ContentType: aws.String("application/cbor"),
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

// updateObject parses a raw database object blob applies the update and returns the serialized final blob.
func updateObject[T any](originalBody []byte, update *T) ([]byte, error) {
	original, err := deserialize[T](originalBody)
	if err != nil {
		return nil, err
	}
	applyUpdate(reflect.ValueOf(original), reflect.ValueOf(update))

	return serialize(original)
}

// applyUpdate traverses the update structure and applies supported non-nil
// Field types to the original object. Expects abi compatible objects.
func applyUpdate(original, update reflect.Value) {
	if update.Kind() == reflect.Pointer {
		if !original.Elem().IsValid() {
			original.Set(reflect.New(original.Type().Elem()))
		}
		applyUpdate(original.Elem(), update.Elem())
		return
	}
	if update.Kind() != reflect.Struct {
		return
	}
	for field := range update.Fields() {
		if !field.IsExported() {
			continue
		}
		switch field.Type {
		case reflect.TypeFor[KeyField](), reflect.TypeFor[ETagField]():
			continue
		}
		if fieldUpdate, ok := update.FieldByIndex(field.Index).Interface().(settable); ok {
			originalField := original.FieldByIndex(field.Index)
			originalValue, ok := originalField.Interface().(gettable)
			if !ok { // this means the field is not initialized / nil
				originalField.Set(fieldUpdate.set(reflect.Zero(field.Type)))
			} else {
				originalField.Set(fieldUpdate.set(originalValue.value()))
			}
			continue
		}
		if field.Type.Kind() == reflect.Pointer || field.Type.Kind() == reflect.Struct {
			applyUpdate(
				original.FieldByIndex(field.Index),
				update.FieldByIndex(field.Index),
			)
		}
	}
}
