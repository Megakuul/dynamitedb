package dynamitdb

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Update changes the entry on the database (identified by PK / SK).
// For update operations see "Field" API on the schema.
// Update will modify v to represent the final state of the updated entry.
func (c *Client) Update(ctx context.Context, v any) error {
	update := reflect.ValueOf(v)
	key, err := constructBucketKey(update)
	if err != nil {
		return err
	}
	originalResp, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
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

	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:  aws.String(c.bucket),
		Key:     aws.String(key),
		Body:    bytes.NewReader(updatedBody),
		IfMatch: originalResp.ETag,
	})
	if err != nil {
		return err
	}
	return nil
}

// updateObject parses a raw database object blob applies the update and returns the serialized final blob.
func updateObject(originalBody []byte, update reflect.Value) ([]byte, error) {
	var original reflect.Value
	if update.Type().Kind() == reflect.Pointer {
		original = reflect.New(update.Type().Elem())
	} else {
		original = reflect.New(update.Type())
	}
	err := json.Unmarshal(originalBody, original.Interface())
	if err != nil {
		return nil, err
	}
	applyUpdate(original, update)

	return json.Marshal(original.Interface())
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
			fallthrough
		case reflect.TypeFor[DataField[int]]():
			fallthrough
		case reflect.TypeFor[DataField[float64]]():
			fallthrough
		case reflect.TypeFor[DataField[[]string]]():
			fallthrough
		case reflect.TypeFor[DataField[map[string]string]]():
			updateField := update.FieldByIndex(field.Index)
			if updateField.IsNil() {
				continue
			}
			original.FieldByIndex(field.Index).Set(updateField)
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
