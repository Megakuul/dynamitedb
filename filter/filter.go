package filter

import "reflect"

func Eq[T any](operand T) *eqFilter[T] {
	return &eqFilter[T]{rhs: reflect.ValueOf(operand)}
}

func NotEq[T any](operand T) *notEqFilter[T] {
	return &notEqFilter[T]{rhs: reflect.ValueOf(operand)}
}

func In[T any](operands []T) inFilter[T] {
	rhsSlice := []reflect.Value{}
	for _, operand := range operands {
		rhsSlice = append(rhsSlice, reflect.ValueOf(operand))
	}
	return inFilter[T]{rhsSlice: rhsSlice}
}

func NotIn[T any](operands []T) notInFilter[T] {
	rhsSlice := []reflect.Value{}
	for _, operand := range operands {
		rhsSlice = append(rhsSlice, reflect.ValueOf(operand))
	}
	return notInFilter[T]{rhsSlice: rhsSlice}
}

type invalid[T any] struct{}

func (invalid[T]) Value() T {
	panic("invalid operation: values are not supported in filter structs")
}

type eqFilter[T any] struct {
	invalid[T]
	rhs reflect.Value
}

func (q eqFilter[T]) Filter(lhs reflect.Value) bool {
	if lhs.Type() != q.rhs.Type() {
		return false
	}
	if lhs.Comparable() {
		return q.rhs.Equal(lhs)
	}
	return reflect.DeepEqual(q.rhs.Interface(), lhs.Interface())
}

type notEqFilter[T any] struct {
	invalid[T]
	rhs reflect.Value
}

func (q notEqFilter[T]) Filter(lhs reflect.Value) bool {
	return !(eqFilter[T]{rhs: q.rhs}).Filter(lhs)
}

type inFilter[T any] struct {
	invalid[T]
	rhsSlice []reflect.Value
}

func (q inFilter[T]) Filter(lhs reflect.Value) bool {
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
	invalid[T]
	rhsSlice []reflect.Value
}

func (q notInFilter[T]) Filter(lhs reflect.Value) bool {
	return !(inFilter[T]{rhsSlice: q.rhsSlice}).Filter(lhs)
}
