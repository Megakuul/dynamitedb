package dynamitedb

import (
	"encoding/json"
	"reflect"
	"time"
)

// serializes the structure into raw database representation.
// reason for this wrapper is to run consistent tests and to be able to swap the underlying marshaller.
func serialize[T any](model T) ([]byte, error) {
	return json.Marshal(model)
}

// deserializes the raw database representation into a model structure.
// reason for this wrapper is to run consistent tests and to be able to swap the underlying marshaller.
// + this function also zero initializes dynamite models properly (filling default values into supported interface types).
func deserialize[T any](data []byte) (*T, error) {
	newVal := reflect.New(reflect.TypeFor[T]())
	initModel(newVal)
	new := newVal.Interface().(*T)
	if err := json.Unmarshal(data, new); err != nil {
		return nil, err
	}
	return new, nil
}

// initModel traverses the model and applies default values to all dynamite fields.
// This function will not touch anything that is not a KeyField, DataField or Struct / *Struct.
func initModel(model reflect.Value) {
	if model.Kind() == reflect.Pointer {
		// initialize nil struct fields
		if model.IsNil() && model.Type().Elem().Kind() == reflect.Struct {
			model.Set(reflect.New(model.Type().Elem()))
		}
		initModel(model.Elem())
		return
	} else if model.Kind() != reflect.Struct {
		return
	}
	for field := range model.Fields() {
		if !field.IsExported() {
			continue
		}
		fieldVal := model.FieldByIndex(field.Index)
		switch field.Type {
		case reflect.TypeFor[KeyField]():
			fieldVal.Set(reflect.ValueOf(Key("")))
		case reflect.TypeFor[DataField[string]]():
			fieldVal.Set(reflect.ValueOf(newData("")))
		case reflect.TypeFor[DataField[int]]():
			fieldVal.Set(reflect.ValueOf(newData(0)))
		case reflect.TypeFor[DataField[float64]]():
			fieldVal.Set(reflect.ValueOf(newData(0.0)))
		case reflect.TypeFor[DataField[bool]]():
			fieldVal.Set(reflect.ValueOf(newData(false)))
		case reflect.TypeFor[DataField[time.Time]]():
			fieldVal.Set(reflect.ValueOf(newData(time.Time{})))
		case reflect.TypeFor[DataField[time.Duration]]():
			fieldVal.Set(reflect.ValueOf(newData(time.Duration{})))
		case reflect.TypeFor[DataField[[]string]]():
			fieldVal.Set(reflect.ValueOf(newData([]string{})))
		case reflect.TypeFor[DataField[map[string]string]]():
			fieldVal.Set(reflect.ValueOf(newData(map[string]string{})))
		default:
			initModel(model.FieldByIndex(field.Index))
		}
	}
}
