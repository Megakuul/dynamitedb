package dynamitdb

import "reflect"

type KeyField interface {
	Value() string
	Query() (string, bool)
}

type DataField[T string | int | float64 | []string | map[string]string] interface {
	Value() T
	// Update(T) T // TODO implement this for $inc
	Filter(reflect.Value) bool
}
