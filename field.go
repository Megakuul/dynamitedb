package dynamitdb

import (
	"reflect"

	"github.com/megakuul/dynamitdb/types"
)

type KeyField interface {
	Value() string
	Query() (string, bool)
}

type DataField[T types.DataConstraint] interface {
	Value() T
	Update(T) T
	Filter(reflect.Value) bool
}
