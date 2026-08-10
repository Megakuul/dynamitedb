package dynamitedb

import (
	"maps"
	"reflect"
	"time"
)

// settable defines the interface that must be implemented for fields that can be updated.
type settable interface {
	// set changes the field data by taking the old raw value in and returning an updated value wrapped as data implementation.
	set(reflect.Value) reflect.Value
}

// CustomUpdate allows you to perform a custom update action on the field.
// This is an update operator.
func CustomUpdate[T any](update func(databaseValue T) (updatedValue T)) *customUpdate[T] {
	return &customUpdate[T]{change: update}
}

// Set simply overwrites the previous value.
// This is an update operator.
func Set[T any](operand T) *setUpdate[T] {
	return &setUpdate[T]{new: operand}
}

// Increment increments the value with operand (can be positive or negative).
// This is an update operator.
func Increment[T int | float64 | time.Duration](operand T) *incrementUpdate[T] {
	return &incrementUpdate[T]{new: operand}
}

// Multiply multiplies the value with operand.
// This is an update operator.
func Multiply[T int | float64 | time.Duration](operand T) *multiplyUpdate[T] {
	return &multiplyUpdate[T]{new: operand}
}

// Toggle changes the value to !value.
// This is an update operator.
func Toggle() *toggleUpdate[bool] {
	return &toggleUpdate[bool]{}
}

// Add adds the provided duration to the database time.
// This is an update operator.
func Add[T time.Time](operand time.Duration) *addUpdate[T] {
	return &addUpdate[T]{add: operand}
}

// Append appends the items to the previous slice.
// This is an update operator.
func Append[T any](items ...T) *appendUpdate[T] {
	return &appendUpdate[T]{new: items}
}

// Remove removes the provided items from the slice.
// This is an update operator.
func Remove[T comparable](items ...T) *removeUpdate[T] {
	remove := map[T]bool{}
	for _, item := range items {
		remove[item] = true
	}
	return &removeUpdate[T]{remove: remove}
}

// Emplace updates the previous map with the provided new kv pairs.
// This is an update operator.
func Emplace[K comparable, V any](operand map[K]V) *emplaceUpdate[K, V] {
	return &emplaceUpdate[K, V]{new: operand}
}

var _ DataField[string] = customUpdate[string]{}

type customUpdate[T any] struct {
	dataField[T]
	change func(T) T
}

func (u customUpdate[T]) set(original reflect.Value) reflect.Value {
	value, _ := original.Interface().(T)
	return reflect.ValueOf(newData(u.change(value)))
}

var _ DataField[string] = setUpdate[string]{}

type setUpdate[T any] struct {
	dataField[T]
	new T
}

func (u setUpdate[T]) set(_ reflect.Value) reflect.Value {
	return reflect.ValueOf(newData(u.new))
}

var _ DataField[float64] = incrementUpdate[float64]{}

type incrementUpdate[T int | float64 | time.Duration] struct {
	dataField[T]
	new T
}

func (u incrementUpdate[T]) set(original reflect.Value) reflect.Value {
	value, _ := original.Interface().(T)
	return reflect.ValueOf(newData(u.new + value))
}

var _ DataField[float64] = multiplyUpdate[float64]{}

type multiplyUpdate[T int | float64 | time.Duration] struct {
	dataField[T]
	new T
}

func (u multiplyUpdate[T]) set(original reflect.Value) reflect.Value {
	value, _ := original.Interface().(T)
	return reflect.ValueOf(newData(u.new * value))
}

var _ DataField[bool] = toggleUpdate[bool]{}

type toggleUpdate[T bool] struct {
	dataField[T]
}

func (u toggleUpdate[T]) set(original reflect.Value) reflect.Value {
	value, _ := original.Interface().(T)
	return reflect.ValueOf(newData(!value))
}

var _ DataField[time.Time] = addUpdate[time.Time]{}

type addUpdate[T time.Time] struct {
	dataField[T]
	add time.Duration
}

func (u addUpdate[T]) set(original reflect.Value) reflect.Value {
	value, _ := original.Interface().(time.Time)
	return reflect.ValueOf(newData(value.Add(u.add)))
}

var _ DataField[[]string] = appendUpdate[string]{}

type appendUpdate[T any] struct {
	dataField[[]T]
	new []T
}

func (u appendUpdate[T]) set(original reflect.Value) reflect.Value {
	value, _ := original.Interface().([]T)
	return reflect.ValueOf(newData(append(value, u.new...)))
}

var _ DataField[[]string] = removeUpdate[string]{}

type removeUpdate[T comparable] struct {
	dataField[[]T]
	remove map[T]bool
}

func (u removeUpdate[T]) set(original reflect.Value) reflect.Value {
	value, _ := original.Interface().([]T)
	newSlice := make([]T, 0, len(value))
	for _, item := range value {
		if u.remove[item] {
			continue
		}
		newSlice = append(newSlice, item)
	}
	return reflect.ValueOf(newData(newSlice))
}

var _ DataField[map[string]string] = emplaceUpdate[string, string]{}

type emplaceUpdate[K comparable, V any] struct {
	dataField[map[K]V]
	new map[K]V
}

func (u emplaceUpdate[K, V]) set(original reflect.Value) reflect.Value {
	value, _ := original.Interface().(map[K]V)
	if value == nil {
		value = make(map[K]V)
	}
	maps.Copy(value, u.new)
	return reflect.ValueOf(newData(value))
}
