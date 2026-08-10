package dynamitedb

import (
	"reflect"
	"strconv"
)

// ETagField defines a dynamite OCC etag.
// Tag it with `etag:"true"` to mark for automatic insertion.
// Multiple etags are not allowed (dynamite will simply use the first in this situation).
type ETagField *string

// retrieveETag extracts the etag from the filter.
// Returns ETag (might be nil if no etag provided).
func retrieveETag(filter reflect.Value) ETagField {
	for field := range filter.Fields() {
		if ok, _ := strconv.ParseBool(field.Tag.Get("etag")); ok {
			etag, _ := filter.FieldByIndex(field.Index).Interface().(ETagField)
			return etag
		}
	}
	return nil
}

// injectETag takes the etag and inserts it into the etag field of the target.
func injectETag(target reflect.Value, etag string) {
	for field := range target.Fields() {
		if ok, _ := strconv.ParseBool(field.Tag.Get("etag")); ok {
			target.FieldByIndex(field.Index).Set(reflect.ValueOf(ETagField(&etag)))
		}
	}
}
