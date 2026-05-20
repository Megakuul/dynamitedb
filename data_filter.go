package dynamitdb

import "reflect"

// Eq checks for an exact match with the operand.
func Eq[T any](operand T) *eqFilter[T] {
	return &eqFilter[T]{rhs: reflect.ValueOf(operand)}
}

// NotEq is just !Eq.
func NotEq[T any](operand T) *notEqFilter[T] {
	return &notEqFilter[T]{rhs: reflect.ValueOf(operand)}
}

// In checks if the database value is inside the provided list.
func In[T any](operands []T) inFilter[T] {
	rhsSlice := []reflect.Value{}
	for _, operand := range operands {
		rhsSlice = append(rhsSlice, reflect.ValueOf(operand))
	}
	return inFilter[T]{rhsSlice: rhsSlice}
}

// NotIn is just !In.
func NotIn[T any](operands []T) notInFilter[T] {
	rhsSlice := []reflect.Value{}
	for _, operand := range operands {
		rhsSlice = append(rhsSlice, reflect.ValueOf(operand))
	}
	return notInFilter[T]{rhsSlice: rhsSlice}
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
		if lhs.Comparable() && !rhs.Equal(lhs) {
			continue
		}
		if !reflect.DeepEqual(rhs.Interface(), lhs.Interface()) {
			continue
		}
		return true
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
