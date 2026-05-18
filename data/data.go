// data provides the DataField implementation for raw data.
// This structure also dictates the underlying design of the json blobs (therefore it is an external interface that must be versioned!)
package data

import "reflect"

func New[T any](value T) *data[T] {
	return &data[T]{data: value}
}

type invalid struct{}

func (invalid) Filter(reflect.Value) bool {
	panic("invalid operation: datas are not supported in filter structs")
}

type data[T any] struct {
	invalid
	data T
}

func (v data[T]) Value() T {
	return v.data
}
