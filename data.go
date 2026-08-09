package dynamitedb

import (
	"reflect"
)

// DataField defines a dynamite data field (initialize with dynamitedb.Data("blub")).
// Use Value() to retrieve the underlying value.
// The reason for this abstraction is that it allows you to use the model struct
// as filter, update and insert structure aswell.
type DataField[T any] interface {
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

func newData[T any](v T) *data[T] {
	return &data[T]{data: v}
}

// internal data interface used only for returned values.
type data[T any] struct {
	dataFallback[T]
	data T
}

func (v data[T]) Value() T {
	return v.data
}

func (v *data[T]) UnmarshalCBOR(data []byte) error {
	newKey, err := deserialize[T](data)
	if err != nil {
		return err
	}
	v.data = *newKey
	return nil
}

func (v data[T]) MarshalCBOR() ([]byte, error) {
	return serialize(v.data)
}
