package dynamitedb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/sync/errgroup"
)

// Get fetches a single entry from database based on the filter.
// This operation performs no ListObject call and is more efficient but only works with exact key matches.
func Get[T any](ctx context.Context, bucket *Bucket, filter *T) (*T, error) {
	filterVal := reflect.ValueOf(filter)
	key, exact, err := constructBucketKey(filterVal.Elem())
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
		if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
			return nil, ErrNotFound
		}
		return nil, errors.New(err.Error())
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	output, err := deserialize[T](body)
	if err != nil {
		return nil, err
	}
	outputVal := reflect.ValueOf(output)
	if err := injectBucketKey(outputVal.Elem(), key); err != nil {
		return nil, fmt.Errorf("key writeback failed: %v", err)
	}
	if !checkFilter(outputVal, filterVal) {
		return nil, ErrFilterMismatch
	}
	return outputVal.Interface().(*T), nil
}

// Query scans (by filter) and returns all database entries indexed by key prefix matches.
// Caution: if key prefixes are not properly organized a query can easily rawdog scan millions of entries.
func Query[T any](ctx context.Context, bucket *Bucket, filter *T, opts ...Option) ([]*T, error) {
	filterVal := reflect.ValueOf(filter)
	key, exact, err := constructBucketKey(filterVal.Elem())
	if err != nil {
		return nil, err
	}
	if exact {
		// for exact matches just optimize it and use Get().
		output, err := Get(ctx, bucket, filter)
		if err != nil && errors.Is(err, ErrNotFound) {
			return []*T{}, nil
		}
		return []*T{output}, err
	}

	options := &options{
		limit:      100000,
		startAfter: nil,
	}
	for _, opt := range opts {
		opt(options)
	}

	outputModels := []*T{}

	paginator := s3.NewListObjectsV2Paginator(bucket.client, &s3.ListObjectsV2Input{
		Bucket:     aws.String(bucket.name),
		Prefix:     aws.String(key),
		StartAfter: options.startAfter,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		for {
			if len(page.Contents) < 1 || len(outputModels) >= options.limit {
				break
			}
			nextBatch := int(math.Min(
				float64(options.limit-len(outputModels)),
				float64(len(page.Contents)),
			))
			batchResults := make([]*T, nextBatch)
			scanGroup, scanCtx := errgroup.WithContext(ctx)
			for i, object := range page.Contents[:nextBatch] {
				scanGroup.Go(func() error {
					resp, err := bucket.client.GetObject(scanCtx, &s3.GetObjectInput{
						Bucket: aws.String(bucket.name),
						Key:    object.Key,
					})
					if err != nil {
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
					outputVal := reflect.ValueOf(output)
					if err := injectBucketKey(outputVal.Elem(), *object.Key); err != nil {
						return fmt.Errorf("key writeback failed: %v", err)
					}
					if !checkFilter(outputVal, filterVal) {
						return nil
					}
					batchResults[i] = outputVal.Interface().(*T)
					return nil
				})
			}
			if err = scanGroup.Wait(); err != nil {
				return nil, err
			}
			for _, result := range batchResults {
				if result != nil {
					outputModels = append(outputModels, result)
				}
			}
			page.Contents = page.Contents[nextBatch:]
		}
	}

	return outputModels, nil
}

// checkFilter traverses the filter structure and performs all non-nil checks (if any of them fails it returns false).
// Expects abi compatible objects.
func checkFilter(original, filter reflect.Value) bool {
	if filter.Kind() == reflect.Pointer {
		if filter.IsNil() {
			return true
		}
		if original.IsNil() {
			return false
		}
		return checkFilter(original.Elem(), filter.Elem())
	}
	if filter.Kind() != reflect.Struct {
		return true
	}
	for field := range filter.Fields() {
		if !field.IsExported() {
			continue
		}
		switch field.Type {
		case reflect.TypeFor[KeyField]():
			continue
		case reflect.TypeFor[DataField[string]]():
			if !checkFieldFilter[string](original, filter, field.Index) {
				return false
			}
		case reflect.TypeFor[DataField[int]]():
			if !checkFieldFilter[int](original, filter, field.Index) {
				return false
			}
		case reflect.TypeFor[DataField[float64]]():
			if !checkFieldFilter[float64](original, filter, field.Index) {
				return false
			}
		case reflect.TypeFor[DataField[bool]]():
			if !checkFieldFilter[bool](original, filter, field.Index) {
				return false
			}
		case reflect.TypeFor[DataField[time.Time]]():
			if !checkFieldFilter[time.Time](original, filter, field.Index) {
				return false
			}
		case reflect.TypeFor[DataField[time.Duration]]():
			if !checkFieldFilter[time.Duration](original, filter, field.Index) {
				return false
			}
		case reflect.TypeFor[DataField[[]string]]():
			if !checkFieldFilter[[]string](original, filter, field.Index) {
				return false
			}
		case reflect.TypeFor[DataField[map[string]string]]():
			if !checkFieldFilter[map[string]string](original, filter, field.Index) {
				return false
			}
		default:
			if !checkFilter(
				original.FieldByIndex(field.Index),
				filter.FieldByIndex(field.Index),
			) {
				return false
			}
		}
	}
	return true
}

// checkFieldFilter applies the filter to original and returns whether it matches.
func checkFieldFilter[T dataConstraint](original, filter reflect.Value, index []int) bool {
	filterField, ok := filter.FieldByIndex(index).Interface().(DataField[T])
	if !ok {
		return true
	}
	originalField, ok := original.FieldByIndex(index).Interface().(DataField[T])
	if !ok {
		return false
	}
	return filterField.filter(reflect.ValueOf(originalField.Value()))
}
