// update DataFields are used to update structures.
// Update fields should only be used in the context of update operations (update). Calling anything but Update() on them panics.
package update

import (
	"maps"
	"reflect"

	"github.com/megakuul/dynamitdb/types"
)

// Set simply overwrites the previous value.
func Set[T types.DataConstraint](operand T) *setUpdate[T] {
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
func Emplace[T []string](operand T) *appendUpdate[T] {
	return &appendUpdate[T]{new: operand}
}

type invalid[T types.DataConstraint] struct{}

func (invalid[T]) Value() T {
	panic("invalid operation: values are not supported in update structs")
}

func (invalid[T]) Filter(reflect.Value) bool {
	panic("invalid operation: filters are not supported in update structs")
}

type setUpdate[T types.DataConstraint] struct {
	invalid[T]
	new T
}

func (u setUpdate[T]) Update(original T) T {
	return u.new
}

type incUpdate[T int | float64] struct {
	invalid[T]
	new T
}

func (u incUpdate[T]) Update(original T) T {
	return u.new + original
}

type mulUpdate[T int | float64] struct {
	invalid[T]
	new T
}

func (u mulUpdate[T]) Update(original T) T {
	return u.new * original
}

type toggleUpdate[T bool] struct {
	invalid[T]
}

func (u toggleUpdate[T]) Update(original T) T {
	return !original
}

type appendUpdate[T []string] struct {
	invalid[T]
	new T
}

func (u appendUpdate[T]) Update(original T) T {
	return append(original, u.new...)
}

type emplaceUpdate[T map[string]string] struct {
	invalid[T]
	new T
}

func (u emplaceUpdate[T]) Update(original T) T {
	maps.Copy(original, u.new)
	return original
}
