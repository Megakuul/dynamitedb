package dynamitedb

import (
	"reflect"
)

// gettable defines the interface that must be implemented for fields that are used as value.
type gettable interface {
	// value just returns the underlying value.
	value() reflect.Value
}

// DataField defines a dynamite data field (initialize with dynamitedb.Data("blub")).
// Use Value() to retrieve the underlying value.
type DataField[T any] interface {
	gettable
	settable
	filterable
	// Value returns the raw data value.
	Value() T
}

var _ DataField[string] = dataField[string]{}

// dataField implements DataField to act as default embedding for model operations (e.g. that a filter also implements settable).
// This is required since dynamite uses a model struct for filter, update and insert.
type dataField[T any] struct{}

func (dataField[T]) Value() T {
	panic("incorrect DataField usage: cannot read value from filter or update operator")
}

func (dataField[T]) value() reflect.Value {
	panic("incorrect DataField usage: cannot read value from filter or update operator")
}

func (dataField[T]) set(reflect.Value) reflect.Value {
	panic("incorrect DataField usage: cannot set value from filter or value operator")
}

func (dataField[T]) filter(reflect.Value) bool {
	panic("incorrect DataField usage: cannot use filter from update or value operator")
}

func newData[T any](v T) *data[T] {
	return &data[T]{data: v}
}

var _ DataField[string] = data[string]{}

// internal data interface used only for returned values.
type data[T any] struct {
	dataField[T]
	data T
}

func (v data[T]) Value() T {
	return v.data
}

func (v data[T]) value() reflect.Value {
	return reflect.ValueOf(v.data)
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
