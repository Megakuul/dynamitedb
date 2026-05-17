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
func constructBucketKey(schema reflect.Value) (string, error) {
	bucketKey := []string{}
	partKey, partVal, err := retrievePartKey(schema)
	if err != nil {
		return "", err
	}
	if !keySanitizer.MatchString(partKey) {
		return "", fmt.Errorf("database partition key contains unsafe characters")
	}
	bucketKey = append(bucketKey, partKey, partVal)

	sortKey, sortVal, err := retrieveSortKey(schema)
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
func retrievePartKey(schema reflect.Value) (string, string, error) {
	for field := range schema.Fields() {
		if partKey := field.Tag.Get("pk"); partKey != "" {
			partVal := schema.FieldByName(field.Name).String()
			if partVal == "" {
				return "", "", fmt.Errorf("partition key '%s' is unset", partKey)
			}
			return partKey, partVal, nil
		}
	}
	return "", "", fmt.Errorf("no partition key found in schema")
}

// retrievePartKey extracts the sort key and sort key value from the schema.
// If no sort key is present it returns an empty string.
func retrieveSortKey(schema reflect.Value) (string, string, error) {
	for field := range schema.Fields() {
		if partKey := field.Tag.Get("sk"); partKey != "" {
			partVal := schema.FieldByName(field.Name).String()
			if partVal == "" {
				return "", "", fmt.Errorf("sort key '%s' is unset", partKey)
			}
			return partKey, partVal, nil
		}
	}
	return "", "", nil
}
