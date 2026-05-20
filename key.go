package dynamitdb

// KeyField defines a dynamite index key (initialize with dynamitedb.Key("123")).
// Tag it with `pk:"user"` to mark partition key or `sk:"order"` for sort key.
// Every model requires at least the partition key (sort key is optional).
// Multiple partition or sort keys are not allowed (dynamite will simply use the first in this situation).
type KeyField interface {
	// Value returns the raw key value.
	Value() string
	// query returns the key segment and whether it should be an exact match or a prefix.
	query() (string, bool)
}

// keyFallback implements keyfield to act as default embedding for model operations.
// This is required since dynamite uses a model struct for filter, update and insert.
type keyFallback struct{}

func (keyFallback) Value() string {
	return ""
}

func (keyFallback) query() (string, bool) {
	return "", true
}

// Key initializes a new partition or sort key.
func Key(id string) *key {
	return &key{key: id}
}

type key struct {
	keyFallback
	key string
}

func (v key) Value() string {
	return v.key
}
