// data provides the DataField implementation for raw data.
// This structure also dictates the underlying design of the json blobs (therefore it is an external interface that must be versioned!)
package data

import (
	"reflect"

	"github.com/megakuul/dynamitdb/types"
)

func New[T types.DataConstraint](value T) *data[T] {
	return &data[T]{data: value}
}

type invalid[T types.DataConstraint] struct{}

func (invalid[T]) Update(T) T {
	panic("invalid operation: data fields are not supported in update structs")
}

func (invalid[T]) Filter(reflect.Value) bool {
	panic("invalid operation: data fields are not supported in filter structs")
}

type data[T types.DataConstraint] struct {
	invalid[T]
	data T
}

func (v data[T]) Value() T {
	return v.data
}
