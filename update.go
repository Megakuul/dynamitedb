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
)

// Update changes the entry on the database (identified by PK / SK).
// For update operations see "Field" API on the schema.
// Update will modify v to represent the final state of the updated entry.
func Update[T any](ctx context.Context, bucket *Bucket, update *T, opts ...Option) error {
	key, exact, err := constructBucketKey(reflect.ValueOf(update).Elem())
	if err != nil {
		return err
	} else if !exact {
		return fmt.Errorf("update database call requires exact key match")
	}
	options := &options{
		expires: nil,
	}
	for _, opt := range opts {
		opt(options)
	}
	originalResp, err := bucket.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket.name),
		Key:    aws.String(key),
	})
	if err != nil {
		if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
			return ErrNotFound
		}
		return err
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
		Bucket:  aws.String(bucket.name),
		Key:     aws.String(key),
		Body:    bytes.NewReader(updatedBody),
		IfMatch: originalResp.ETag,
		Expires: options.expires,
	})
	if err != nil {
		return err
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
	if original.Kind() == reflect.Pointer {
		applyUpdate(original.Elem(), update.Elem())
		return
	}
	for field := range update.Fields() {
		if !field.IsExported() {
			continue
		}
		switch field.Type {
		case reflect.TypeFor[KeyField]():
			continue
		case reflect.TypeFor[DataField[string]]():
			applyFieldUpdate[string](original, update, field.Index)
		case reflect.TypeFor[DataField[int]]():
			applyFieldUpdate[int](original, update, field.Index)
		case reflect.TypeFor[DataField[float64]]():
			applyFieldUpdate[float64](original, update, field.Index)
		case reflect.TypeFor[DataField[bool]]():
			applyFieldUpdate[bool](original, update, field.Index)
		case reflect.TypeFor[DataField[[]string]]():
			applyFieldUpdate[[]string](original, update, field.Index)
		case reflect.TypeFor[DataField[map[string]string]]():
			applyFieldUpdate[map[string]string](original, update, field.Index)
		default:
			if field.Type.Kind() == reflect.Pointer || field.Type.Kind() == reflect.Struct {
				applyUpdate(
					original.FieldByIndex(field.Index),
					update.FieldByIndex(field.Index),
				)
			}
			continue
		}
	}
}

// applyFieldUpdate applies the defined update operation from "update" to "original".
func applyFieldUpdate[T dataConstraint](original, update reflect.Value, index []int) {
	updateField, ok := update.FieldByIndex(index).Interface().(DataField[T])
	if !ok {
		return
	}
	originalField, ok := original.FieldByIndex(index).Interface().(DataField[T])
	if !ok {
		var new T
		original.FieldByIndex(index).Set(reflect.ValueOf(
			Data(updateField.update(new))),
		)
		return
	}
	original.FieldByIndex(index).Set(reflect.ValueOf(
		Data(updateField.update(originalField.Value()))),
	)
}
