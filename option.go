package dynamitedb

import (
	"reflect"
	"time"
)

// options defines a global dynamite setting struct.
// certain settings are only important to specific operations.
// operations that use options must first default initialize the fields they use.
type options struct {
	// limit defines the maximum number of elements that should be fetched (not processed though).
	limit int
	// startAfter defines an item from where the bucket call will start querying (passed to s3 StartAfter param).
	startAfter *string
	// expires defines if the specified item should expire (if set it will be expired by then via the s3 lifecycle system).
	expires *time.Time
}

type Option func(*options)

// WithLimit specifies the limit of elements to return.
// caution: even with a limit of 10 Query() will scan million of entries if it finds no filter matches.
func WithLimit(limit int) func(*options) {
	return func(q *options) {
		q.limit = limit
	}
}

// WithStartAfter specifies an item from where to start querying (useful for pagination).
// This parameter requires a filter object with an exact key match (other filter props are not checked).
func WithStartAfter[T any](filter *T) func(*options) {
	return func(q *options) {
		key, exact, err := constructBucketKey(reflect.ValueOf(filter).Elem())
		if err != nil || !exact {
			return
		}
		q.startAfter = &key
	}
}

// WithExpires can be specified to expire the updated or inserted item at a certain point in time.
// Notice that this relies on the S3 lifecycle and may not be immediately removed depending on implementation.
func WithExpires(expires time.Time) func(*options) {
	return func(q *options) {
		q.expires = &expires
	}
}
