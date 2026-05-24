package dynamitedb

import (
	"reflect"
	"time"
)

// dataConstraint simply wraps all supported dynamite types to avoid boilerplate.
type dataConstraint interface {
	string | int | float64 | bool | time.Time | time.Duration | []string | map[string]string
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
	panic("incorrect DataField usage: cannot read value from filter or update operator")
}

func (dataFallback[T]) update(input T) T {
	panic("incorrect DataField usage: cannot use value or filter as update operator")
}

func (dataFallback[T]) filter(reflect.Value) bool {
	panic("incorrect DataField usage: cannot use value or update as filter operator")
}

func newData[T dataConstraint](v T) *data[T] {
	return &data[T]{data: v}
}

// internal data interface used only for returned values.
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
