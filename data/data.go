package data

import "reflect"

func String(value string) *stringData {
	return &stringData{data: value}
}

func Int(value int) *intData {
	return &intData{data: value}
}

func Float(value float64) *floatData {
	return &floatData{data: value}
}

func Slice(value []string) *sliceData {
	return &sliceData{data: value}
}

func Map(value map[string]string) *mapData {
	return &mapData{data: value}
}

type invalid struct{}

func (invalid) Filter(reflect.Value) bool {
	panic("invalid operation: datas are not supported in filter structs")
}

type stringData struct {
	invalid
	data string
}

func (v stringData) Value() string {
	return v.data
}

type intData struct {
	invalid
	data int
}

func (v intData) Value() int {
	return v.data
}

type floatData struct {
	invalid
	data float64
}

func (v floatData) Value() float64 {
	return v.data
}

type sliceData struct {
	invalid
	data []string
}

func (v sliceData) Value() []string {
	return v.data
}

type mapData struct {
	invalid
	data map[string]string
}

func (v mapData) Value() map[string]string {
	return v.data
}
