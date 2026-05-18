// key provides the KeyField implementation for raw index values.
// This structure also dictates the underlying design of the json blobs (therefore it is an external interface that must be versioned!)
package key

func New(id string) *key {
	return &key{key: id}
}

type invalid struct{}

func (invalid) Query() (string, bool) {
	panic("invalid operation: keys are not supported in query structs")
}

type key struct {
	invalid
	key string
}

func (v key) Value() string {
	return v.key
}
