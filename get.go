package dynamitdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"

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
		if _, ok := errors.AsType[*types.NotFound](err); ok {
			return nil, ErrNotFound
		}
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	output, err := deserialize[T](body)
	if err != nil {
		return nil, err
	}
	if !checkFilter(reflect.ValueOf(output), filterVal) {
		return nil, ErrNotFound
	}
	return output, nil
}

// QueryOptions defines an additional configuration for queries.
type QueryOptions[T any] struct {
	// specifies the limit of elements to return.
	// caution: even with a limit of 10 Query() will scan million of entries if it finds no filter matches.
	Limit int
	// StartAfter specifies an item from where to start querying (useful for pagination).
	// This parameter requires a filter object with an exact key match (other filter props are not checked).
	StartAfter *T
}

// Query scans (by filter) and returns all database entries indexed by key prefix matches.
// Caution: if key prefixes are not properly organized a query can easily rawdog scan millions of entries.
func Query[T any](ctx context.Context, bucket *Bucket, filter *T, opts QueryOptions[T]) ([]*T, error) {
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

	var startAfter *string
	if opts.StartAfter != nil {
		startAfterVal := reflect.ValueOf(opts.StartAfter)
		startAfterKey, exact, err := constructBucketKey(startAfterVal.Elem())
		if err != nil {
			return nil, err
		} else if !exact {
			return nil, fmt.Errorf("startAfter filter requires an exact key match")
		}
		startAfter = aws.String(startAfterKey)
	}

	resultModels := []*T{}

	paginator := s3.NewListObjectsV2Paginator(bucket.client, &s3.ListObjectsV2Input{
		Bucket:     aws.String(bucket.name),
		Prefix:     aws.String(key),
		StartAfter: startAfter,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for {
			if len(page.Contents) < 1 || len(resultModels) >= opts.Limit {
				break
			}
			nextBatch := int(math.Min(
				float64(opts.Limit-len(resultModels)),
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
						return err
					}
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return err
					}
					result, err := deserialize[T](body)
					if err != nil {
						return err
					}
					if !checkFilter(reflect.ValueOf(result), filterVal) {
						return nil
					}
					batchResults[i] = result
					return nil
				})
			}
			if err = scanGroup.Wait(); err != nil {
				return nil, err
			}
			for _, result := range batchResults {
				if result != nil {
					resultModels = append(resultModels, result)
				}
			}
			page.Contents = page.Contents[nextBatch:]
		}
	}

	return resultModels, nil
}

// checkFilter traverses the filter structure and performs all non-nil checks (if any of them fails it returns false).
// Expects abi compatible objects.
func checkFilter(original, filter reflect.Value) bool {
	if original.Kind() == reflect.Pointer {
		return checkFilter(original.Elem(), filter.Elem())
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
		case reflect.TypeFor[DataField[[]string]]():
			if !checkFieldFilter[[]string](original, filter, field.Index) {
				return false
			}
		case reflect.TypeFor[DataField[map[string]string]]():
			if !checkFieldFilter[map[string]string](original, filter, field.Index) {
				return false
			}
		default:
			if field.Type.Kind() == reflect.Pointer || field.Type.Kind() == reflect.Struct {
				return checkFilter(
					original.FieldByIndex(field.Index),
					filter.FieldByIndex(field.Index),
				)
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
	return filterField.filter(original)
}
