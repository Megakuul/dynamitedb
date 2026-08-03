package dynamitedb

import (
	"maps"
	"time"
)

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
func Append[T []string](items ...string) *appendUpdate[T] {
	return &appendUpdate[T]{new: items}
}

// Remove removes the provided items from the slice.
// This is an update operator.
func Remove[T []string](items ...string) *removeUpdate[T] {
	remove := map[string]bool{}
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

type customUpdate[T any] struct {
	dataFallback[T]
	change func(T) T
}

func (u customUpdate[T]) update(original T) T {
	return u.change(original)
}

type setUpdate[T any] struct {
	dataFallback[T]
	new T
}

func (u setUpdate[T]) update(original T) T {
	return u.new
}

type incrementUpdate[T int | float64 | time.Duration] struct {
	dataFallback[T]
	new T
}

func (u incrementUpdate[T]) update(original T) T {
	return u.new + original
}

type multiplyUpdate[T int | float64 | time.Duration] struct {
	dataFallback[T]
	new T
}

func (u multiplyUpdate[T]) update(original T) T {
	return u.new * original
}

type toggleUpdate[T bool] struct {
	dataFallback[T]
}

func (u toggleUpdate[T]) update(original T) T {
	return !original
}

type addUpdate[T time.Time] struct {
	dataFallback[T]
	add time.Duration
}

func (u addUpdate[T]) update(original T) T {
	if originalTime, ok := any(original).(time.Time); ok {
		return T(originalTime.Add(u.add))
	}
	return original
}

type appendUpdate[T []string] struct {
	dataFallback[T]
	new T
}

func (u appendUpdate[T]) update(original T) T {
	return append(original, u.new...)
}

type removeUpdate[T []string] struct {
	dataFallback[T]
	remove map[string]bool
}

func (u removeUpdate[T]) update(original T) T {
	newSlice := make(T, 0, len(original))
	for _, item := range original {
		if u.remove[item] {
			continue
		}
		newSlice = append(newSlice, item)
	}
	return newSlice
}

type emplaceUpdate[K comparable, V any] struct {
	dataFallback[map[K]V]
	new map[K]V
}

func (u emplaceUpdate[K, V]) update(original map[K]V) map[K]V {
	if original == nil {
		original = make(map[K]V)
	}
	maps.Copy(original, u.new)
	return original
}
