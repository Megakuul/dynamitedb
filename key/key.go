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
