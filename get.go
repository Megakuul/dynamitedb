package dynamitdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/megakuul/dynamitdb/data"
)

// Get fetches a single entry from database based on the filter.
// If the query finds multiple entries (when using BeginsWith) it returns the first match.
func Get[T any](ctx context.Context, bucket *Bucket, filter *T) (*T, error) {
	key, exact, err := constructBucketKey(reflect.ValueOf(filter).Elem())
	if err != nil {
		return nil, err
	} else if !exact {
		return nil, fmt.Errorf("get database call requires exact key match")
	}
	resp, err := bucket.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket.name),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var output T
	initModel(reflect.ValueOf(output))
	err = json.Unmarshal(body, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func Query[T any](ctx context.Context, bucket *Bucket, filter *T) (*T, error) {
	key, exact, err := constructBucketKey(reflect.ValueOf(filter).Elem())
	if err != nil {
		return nil, err
	}
	resp, err := bucket.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket.name),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var output T
	initModel(reflect.ValueOf(output))
	err = json.Unmarshal(body, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

// initModel traverses the model and applies default values to all nil fields.
func initModel(model reflect.Value) {
	if model.Kind() == reflect.Pointer {
		initModel(model.Elem())
		return
	}
	for field := range model.Fields() {
		if !field.IsExported() {
			continue
		}
		fieldVal := model.FieldByIndex(field.Index)
		if !fieldVal.IsNil() {
			continue
		}
		switch field.Type {
		case reflect.TypeFor[KeyField]():
			continue
		case reflect.TypeFor[DataField[string]]():
			fieldVal.Set(reflect.ValueOf(data.New("")))
		case reflect.TypeFor[DataField[int]]():
			fieldVal.Set(reflect.ValueOf(data.New(0)))
		case reflect.TypeFor[DataField[float64]]():
			fieldVal.Set(reflect.ValueOf(data.New(0.0)))
		case reflect.TypeFor[DataField[bool]]():
			fieldVal.Set(reflect.ValueOf(data.New(false)))
		case reflect.TypeFor[DataField[[]string]]():
			fieldVal.Set(reflect.ValueOf(data.New([]string{})))
		case reflect.TypeFor[DataField[map[string]string]]():
			fieldVal.Set(reflect.ValueOf(data.New(map[string]string{})))
		default:
			if field.Type.Kind() == reflect.Pointer || field.Type.Kind() == reflect.Struct {
				initModel(model.FieldByIndex(field.Index))
			}
			continue
		}
	}
}
