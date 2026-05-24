package dynamitedb

import (
	"maps"
	"time"
)

// CustomUpdate allows you to perform a custom update action on the field.
// This is an update operator.
func CustomUpdate[T dataConstraint](update func(databaseValue T) (updatedValue T)) *customUpdate[T] {
	return &customUpdate[T]{change: update}
}

// Set simply overwrites the previous value.
// This is an update operator.
func Set[T dataConstraint](operand T) *setUpdate[T] {
	return &setUpdate[T]{new: operand}
}

// Increment increments the value with operand (can be positive or negative).
// This is an update operator.
func Increment[T int | float64](operand T) *incrementUpdate[T] {
	return &incrementUpdate[T]{new: operand}
}

// Multiply multiplies the value with operand.
// This is an update operator.
func Multiply[T int | float64](operand T) *multiplyUpdate[T] {
	return &multiplyUpdate[T]{new: operand}
}

// Toggle changes the value to !value.
// This is an update operator.
func Toggle() *toggleUpdate[bool] {
	return &toggleUpdate[bool]{}
}

// Append appends the slice to the previous slice.
// This is an update operator.
func Append[T []string](operand T) *appendUpdate[T] {
	return &appendUpdate[T]{new: operand}
}

// Emplace updates the previous map with the provided new kv pairs.
// This is an update operator.
func Emplace[T map[string]string](operand T) *emplaceUpdate[T] {
	return &emplaceUpdate[T]{new: operand}
}

type customUpdate[T dataConstraint] struct {
	dataFallback[T]
	change func(T) T
}

func (u customUpdate[T]) update(original T) T {
	return u.change(original)
}

type setUpdate[T dataConstraint] struct {
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

type appendUpdate[T []string] struct {
	dataFallback[T]
	new T
}

func (u appendUpdate[T]) update(original T) T {
	return append(original, u.new...)
}

type emplaceUpdate[T map[string]string] struct {
	dataFallback[T]
	new T
}

func (u emplaceUpdate[T]) update(original T) T {
	if original == nil {
		original = make(T)
	}
	maps.Copy(original, u.new)
	return original
}
