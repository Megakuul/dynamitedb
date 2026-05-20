package dynamitdb

import (
	"reflect"
)

// dataConstraint simply wraps all supported dynamite types to avoid boilerplate.
type dataConstraint interface {
	string | int | float64 | bool | []string | map[string]string
}

// DataField defines a dynamite data field (initialize with dynamitedb.Data("blub")).
// Use Value() to retrieve the underlying value.
// The reason for this abstraction is that it allows you to use the model struct
// as filter, update and insert structure aswell.
type DataField[T dataConstraint] interface {
	// Value returns the raw data value.
	Value() T
	// update executes an update expr on the value and returns the new value.
	update(T) T
	// filter checks if the provided model matches on this field.
	filter(reflect.Value) bool
}

// dataFallback implements datafield to act as default embedding for model operations.
// This is required since dynamite uses a model struct for filter, update and insert.
type dataFallback[T any] struct{}

func (dataFallback[T]) Value() T {
	var def T
	return def
}

func (dataFallback[T]) update(input T) T {
	return input
}

func (dataFallback[T]) filter(reflect.Value) bool {
	return false
}

// Data initializes a new data field.
func Data[T dataConstraint](value T) *data[T] {
	return &data[T]{data: value}
}

type data[T dataConstraint] struct {
	dataFallback[T]
	data T
}

func (v data[T]) Value() T {
	return v.data
}

func (v *data[T]) UnmarshalJSON(data []byte) error {
	newKey, err := deserialize[T](data)
	if err != nil {
		return err
	}
	v.data = *newKey
	return nil
}

func (v data[T]) MarshalJSON() ([]byte, error) {
	return serialize(v.data)
}
