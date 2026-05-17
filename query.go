package dynamitdb

import "reflect"

func Eq[T any](operand T) *eqQuery {
	return &eqQuery{rhs: reflect.ValueOf(operand)}
}

func NotEq[T any](operand T) *notEqQuery {
	return &notEqQuery{rhs: reflect.ValueOf(operand)}
}

func In[T any](operands []T) inQuery {
	rhsSlice := []reflect.Value{}
	for _, operand := range operands {
		rhsSlice = append(rhsSlice, reflect.ValueOf(operand))
	}
	return inQuery{rhsSlice: rhsSlice}
}

func NotIn[T any](operands []T) notInQuery {
	rhsSlice := []reflect.Value{}
	for _, operand := range operands {
		rhsSlice = append(rhsSlice, reflect.ValueOf(operand))
	}
	return notInQuery{rhsSlice: rhsSlice}
}

type query interface {
	match(reflect.Value) bool
}

type eqQuery struct {
	rhs reflect.Value
}

func (q eqQuery) match(lhs reflect.Value) bool {
	if lhs.Type() != q.rhs.Type() {
		return false
	}
	if lhs.Comparable() {
		return q.rhs.Equal(lhs)
	}
	return reflect.DeepEqual(q.rhs.Interface(), lhs.Interface())
}

type notEqQuery struct {
	rhs reflect.Value
}

func (q notEqQuery) match(lhs reflect.Value) bool {
	return !(eqQuery{rhs: q.rhs}).match(lhs)
}

type inQuery struct {
	rhsSlice []reflect.Value
}

func (q inQuery) match(lhs reflect.Value) bool {
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

type notInQuery struct {
	rhsSlice []reflect.Value
}

func (q notInQuery) match(lhs reflect.Value) bool {
	return !(inQuery{rhsSlice: q.rhsSlice}).match(lhs)
}
