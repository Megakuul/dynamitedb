package dynamitdb

import (
	"maps"
)

// Set simply overwrites the previous value.
func Set[T dataConstraint](operand T) *setUpdate[T] {
	return &setUpdate[T]{new: operand}
}

// Inc increments the value with operand (can be positive or negative).
func Inc[T int | float64](operand T) *incUpdate[T] {
	return &incUpdate[T]{new: operand}
}

// Mul multiplies the value with operand.
func Mul[T int | float64](operand T) *mulUpdate[T] {
	return &mulUpdate[T]{new: operand}
}

// Toggle changes the value to !value.
func Toggle() *toggleUpdate[bool] {
	return &toggleUpdate[bool]{}
}

// Append appends the slice to the previous slice.
func Append[T []string](operand T) *appendUpdate[T] {
	return &appendUpdate[T]{new: operand}
}

// Emplace updates the previous map with the provided new kv pairs.
func Emplace[T map[string]string](operand T) *emplaceUpdate[T] {
	return &emplaceUpdate[T]{new: operand}
}

type setUpdate[T dataConstraint] struct {
	dataFallback[T]
	new T
}

func (u setUpdate[T]) update(original T) T {
	return u.new
}

type incUpdate[T int | float64] struct {
	dataFallback[T]
	new T
}

func (u incUpdate[T]) update(original T) T {
	return u.new + original
}

type mulUpdate[T int | float64] struct {
	dataFallback[T]
	new T
}

func (u mulUpdate[T]) update(original T) T {
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
