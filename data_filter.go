package dynamitedb

import (
	"reflect"
	"slices"
	"strings"
	"time"
)

// CustomFilter allows you to perform a custom filter check on the field.
// This is a filter operator.
func CustomFilter[T dataConstraint](check func(databaseValue T) bool) *customFilter[T] {
	return &customFilter[T]{check: check}
}

// Eq checks for an exact match with the operand.
// For slices and map this enforces a deep equal state.
// This is a filter operator.
func Eq[T dataConstraint](operand T) *eqFilter[T] {
	return &eqFilter[T]{rhs: reflect.ValueOf(operand)}
}

// NotEq is just !Eq.
// This is a filter operator.
func NotEq[T dataConstraint](operand T) *notEqFilter[T] {
	return &notEqFilter[T]{rhs: reflect.ValueOf(operand)}
}

// In checks if the database value is inside the provided list.
// Functionally equivalent to Eq but just with an operand for loop.
// This is a filter operator.
func In[T dataConstraint](operands ...T) inFilter[T] {
	rhsSlice := make([]reflect.Value, len(operands))
	for i, operand := range operands {
		rhsSlice[i] = reflect.ValueOf(operand)
	}
	return inFilter[T]{rhsSlice: rhsSlice}
}

// NotIn is just !In.
// This is a filter operator.
func NotIn[T dataConstraint](operands ...T) notInFilter[T] {
	rhsSlice := make([]reflect.Value, len(operands))
	for i, operand := range operands {
		rhsSlice[i] = reflect.ValueOf(operand)
	}
	return notInFilter[T]{rhsSlice: rhsSlice}
}

// Includes checks if the string contains the operand.
// This is a filter operator.
func Includes[T string](operand string) *includesFilter[T] {
	return &includesFilter[T]{search: operand}
}

// Contains checks if the slice contains all of the operands.
// This is a filter operator.
func Contains[T []string](operands ...string) *containsFilter[T] {
	return &containsFilter[T]{searches: operands}
}

// Has checks if the map contains the specified key value pair.
// This is a filter operator.
func Has[T map[string]string](key, value string) *hasFilter[T] {
	return &hasFilter[T]{key: key, value: value}
}

// Before checks if the database value is before the specified time.
// This is a filter operator.
func Before[T time.Time](operand T) *beforeFilter[T] {
	return &beforeFilter[T]{operand: operand}
}

// After checks if the database value is after the specified time.
// This is a filter operator.
func After[T time.Time](operand T) *afterFilter[T] {
	return &afterFilter[T]{operand: operand}
}

// GreaterThan compares exactly what it says.
// This is a filter operator.
func GreaterThan[T int | float64 | time.Duration](operand T) *greaterThanFilter[T] {
	return &greaterThanFilter[T]{operand: operand}
}

// GreaterOrEqThan compares exactly what it says.
// This is a filter operator.
func GreaterOrEqThan[T int | float64 | time.Duration](operand T) *greaterOrEqThanFilter[T] {
	return &greaterOrEqThanFilter[T]{operand: operand}
}

// LessThan compares exactly what it says.
// This is a filter operator.
func LessThan[T int | float64 | time.Duration](operand T) *lessThanFilter[T] {
	return &lessThanFilter[T]{operand: operand}
}

// LessThanOrEq compares exactly what it says.
// This is a filter operator.
func LessOrEqThan[T int | float64 | time.Duration](operand T) *lessOrEqThanFilter[T] {
	return &lessOrEqThanFilter[T]{operand: operand}
}

type customFilter[T any] struct {
	dataFallback[T]
	check func(T) bool
}

func (q customFilter[T]) filter(lhs reflect.Value) bool {
	if lhsValue, ok := lhs.Interface().(T); ok {
		return q.check(lhsValue)
	}
	return false
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

type includesFilter[T string] struct {
	dataFallback[T]
	search string
}

func (q includesFilter[T]) filter(lhs reflect.Value) bool {
	return lhs.Kind() == reflect.String && strings.Contains(lhs.String(), q.search)
}

type containsFilter[T []string] struct {
	dataFallback[T]
	searches []string
}

func (q containsFilter[T]) filter(lhs reflect.Value) bool {
	if slice, _ := lhs.Interface().([]string); slice != nil {
		for _, search := range q.searches {
			if !slices.Contains(slice, search) {
				return false
			}
		}
		return true
	}
	return false
}

type hasFilter[T map[string]string] struct {
	dataFallback[T]
	key, value string
}

func (q hasFilter[T]) filter(lhs reflect.Value) bool {
	hashmap, _ := lhs.Interface().(map[string]string)
	return hashmap != nil && hashmap[q.key] == q.value
}

type afterFilter[T time.Time] struct {
	dataFallback[T]
	operand T
}

func (q afterFilter[T]) filter(lhs reflect.Value) bool {
	if lhsValue, ok := lhs.Interface().(time.Time); ok {
		return lhsValue.After(any(q.operand).(time.Time))
	}
	return false
}

type beforeFilter[T time.Time] struct {
	dataFallback[T]
	operand T
}

func (q beforeFilter[T]) filter(lhs reflect.Value) bool {
	if lhsValue, ok := lhs.Interface().(time.Time); ok {
		return lhsValue.Before(any(q.operand).(time.Time))
	}
	return false
}

type greaterThanFilter[T float64 | int | time.Duration] struct {
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

type greaterOrEqThanFilter[T float64 | int | time.Duration] struct {
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

type lessThanFilter[T float64 | int | time.Duration] struct {
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

type lessOrEqThanFilter[T float64 | int | time.Duration] struct {
	dataFallback[T]
	operand T
}

func (q lessOrEqThanFilter[T]) filter(lhs reflect.Value) bool {
	switch lhs.Kind() {
	case reflect.Int | reflect.Int64: // 64 for duration
		return lhs.Int() <= int64(q.operand)
	case reflect.Float64:
		return lhs.Float() <= float64(q.operand)
	default:
		return false
	}
}
