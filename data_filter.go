package dynamitdb

import (
	"reflect"
	"slices"
	"strings"
)

// Eq checks for an exact match with the operand.
// For slices and map this enforces a deep equal state.
func Eq[T any](operand T) *eqFilter[T] {
	return &eqFilter[T]{rhs: reflect.ValueOf(operand)}
}

// NotEq is just !Eq.
func NotEq[T any](operand T) *notEqFilter[T] {
	return &notEqFilter[T]{rhs: reflect.ValueOf(operand)}
}

// In checks if the database value is inside the provided list.
// Functionally equivalent to Eq but just with an operand for loop.
func In[T any](operands []T) inFilter[T] {
	rhsSlice := make([]reflect.Value, len(operands))
	for _, operand := range operands {
		rhsSlice = append(rhsSlice, reflect.ValueOf(operand))
	}
	return inFilter[T]{rhsSlice: rhsSlice}
}

// NotIn is just !In.
func NotIn[T any](operands []T) notInFilter[T] {
	rhsSlice := make([]reflect.Value, len(operands))
	for _, operand := range operands {
		rhsSlice = append(rhsSlice, reflect.ValueOf(operand))
	}
	return notInFilter[T]{rhsSlice: rhsSlice}
}

// Contains checks if the string or slice contains the operand.
func Contains(operand string) *containsFilter {
	return &containsFilter{search: operand}
}

// Has checks if the map contains the specified key value pair.
func Has(key, value string) *hasFilter {
	return &hasFilter{key: key, value: value}
}

// GreaterThan compares exactly what it says.
func GreaterThan[T int | float64](operand T) *greaterThanFilter[T] {
	return &greaterThanFilter[T]{operand: operand}
}

// GreaterOrEqThan compares exactly what it says.
func GreaterOrEqThan[T int | float64](operand T) *greaterOrEqThanFilter[T] {
	return &greaterOrEqThanFilter[T]{operand: operand}
}

// LessThan compares exactly what it says.
func LessThan[T int | float64](operand T) *lessThanFilter[T] {
	return &lessThanFilter[T]{operand: operand}
}

// LessThanOrEq compares exactly what it says.
func LessOrEqThan[T int | float64](operand T) *lessOrEqThanFilter[T] {
	return &lessOrEqThanFilter[T]{operand: operand}
}

type eqFilter[T any] struct {
	dataFallback[T]
	rhs reflect.Value
}

func (q eqFilter[T]) filter(lhs reflect.Value) bool {
	if lhs.Type() != q.rhs.Type() {
		return false
	}
	if lhs.Comparable() {
		return q.rhs.Equal(lhs)
	}
	return reflect.DeepEqual(q.rhs.Interface(), lhs.Interface())
}

type notEqFilter[T any] struct {
	dataFallback[T]
	rhs reflect.Value
}

func (q notEqFilter[T]) filter(lhs reflect.Value) bool {
	return !(eqFilter[T]{rhs: q.rhs}).filter(lhs)
}

type inFilter[T any] struct {
	dataFallback[T]
	rhsSlice []reflect.Value
}

func (q inFilter[T]) filter(lhs reflect.Value) bool {
	for _, rhs := range q.rhsSlice {
		if lhs.Type() != rhs.Type() {
			continue
		}
		if lhs.Comparable() {
			if rhs.Equal(lhs) {
				return true
			}
			continue
		} else if reflect.DeepEqual(rhs.Interface(), lhs.Interface()) {
			return true
		}
	}
	return false
}

type notInFilter[T any] struct {
	dataFallback[T]
	rhsSlice []reflect.Value
}

func (q notInFilter[T]) filter(lhs reflect.Value) bool {
	return !(inFilter[T]{rhsSlice: q.rhsSlice}).filter(lhs)
}

type containsFilter struct {
	dataFallback[string]
	search string
}

func (q containsFilter) filter(lhs reflect.Value) bool {
	switch lhs.Kind() {
	case reflect.String:
		return strings.Contains(lhs.String(), q.search)
	case reflect.Slice:
		if slice, ok := lhs.Interface().([]string); ok {
			return slices.Contains(slice, q.search)
		}
	}
	return false
}

type hasFilter struct {
	dataFallback[string]
	key, value string
}

func (q hasFilter) filter(lhs reflect.Value) bool {
	hashmap, _ := lhs.Interface().(map[string]string)
	return hashmap != nil && hashmap[q.key] == q.value
}

type greaterThanFilter[T float64 | int] struct {
	dataFallback[T]
	operand T
}

func (q greaterThanFilter[T]) filter(lhs reflect.Value) bool {
	switch lhs.Kind() {
	case reflect.Int:
		return lhs.Int() > int64(q.operand)
	case reflect.Float64:
		return lhs.Float() > float64(q.operand)
	default:
		return false
	}
}

type greaterOrEqThanFilter[T float64 | int] struct {
	dataFallback[T]
	operand T
}

func (q greaterOrEqThanFilter[T]) filter(lhs reflect.Value) bool {
	switch lhs.Kind() {
	case reflect.Int:
		return lhs.Int() >= int64(q.operand)
	case reflect.Float64:
		return lhs.Float() >= float64(q.operand)
	default:
		return false
	}
}

type lessThanFilter[T float64 | int] struct {
	dataFallback[T]
	operand T
}

func (q lessThanFilter[T]) filter(lhs reflect.Value) bool {
	switch lhs.Kind() {
	case reflect.Int:
		return lhs.Int() < int64(q.operand)
	case reflect.Float64:
		return lhs.Float() < float64(q.operand)
	default:
		return false
	}
}

type lessOrEqThanFilter[T float64 | int] struct {
	dataFallback[T]
	operand T
}

func (q lessOrEqThanFilter[T]) filter(lhs reflect.Value) bool {
	switch lhs.Kind() {
	case reflect.Int:
		return lhs.Int() <= int64(q.operand)
	case reflect.Float64:
		return lhs.Float() <= float64(q.operand)
	default:
		return false
	}
}
