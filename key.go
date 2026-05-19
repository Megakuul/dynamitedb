package dynamitdb

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// TODO if this bottlenecks just write a small whitelist char checker
var keySanitizer = regexp.MustCompile("^[A-Za-z0-9._-]{1,100}$")

// constructBucketKey extracts, sanitizes and constructs an s3 bucket key string from the schema.
func constructBucketKey(filter reflect.Value) (string, error) {
	bucketKey := []string{}
	partKey, partVal, err := retrievePartKey(filter)
	if err != nil {
		return "", err
	}
	if !keySanitizer.MatchString(partKey) {
		return "", fmt.Errorf("database partition key contains unsafe characters")
	}
	bucketKey = append(bucketKey, partKey, partVal)

	sortKey, sortVal, err := retrieveSortKey(filter)
	if err != nil {
		return "", err
	}
	if sortKey != "" {
		if !keySanitizer.MatchString(sortKey) {
			return "", fmt.Errorf("database sort key contains unsafe characters")
		}
		bucketKey = append(bucketKey, sortKey, sortVal)
	}

	return strings.Join(bucketKey, "/"), nil
}

// retrievePartKey extracts the partition key and partition key value from the schema.
func retrievePartKey(filter reflect.Value) (string, string, error) {
	for field := range filter.Fields() {
		if partKey := field.Tag.Get("pk"); partKey != "" {
			partVal, ok := filter.FieldByIndex(field.Index).Interface().(KeyField)
			if !ok {
				return "", "", fmt.Errorf("partition key '%s' is unset", partKey)
			}
			return partKey, partVal.Value(), nil
		}
	}
	return "", "", fmt.Errorf("no partition key found in schema")
}

// retrieveSortKey extracts the sort key and sort key value from the schema.
// If no sort key is present it returns an empty string.
func retrieveSortKey(filter reflect.Value) (string, string, error) {
	for field := range filter.Fields() {
		if sortKey := field.Tag.Get("sk"); sortKey != "" {
			sortVal, ok := filter.FieldByIndex(field.Index).Interface().(KeyField)
			if !ok {
				return "", "", fmt.Errorf("sort key '%s' is unset", sortKey)
			}
			return sortKey, sortVal.Value(), nil
		}
	}
	return "", "", nil
}
