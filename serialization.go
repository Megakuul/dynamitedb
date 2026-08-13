package dynamitedb

import (
	"reflect"

	"github.com/fxamacker/cbor/v2"
)

var encoderOpts = cbor.EncOptions{
	Time: cbor.TimeRFC3339Nano,
}

var decoderOpts = cbor.DecOptions{
	MaxArrayElements: 10000,
	MaxMapPairs:      5000,
	MaxNestedLevels:  16,
}

var decoder, _ = decoderOpts.DecMode()

var encoder, _ = encoderOpts.EncMode()

// serializes the structure into raw database representation.
// reason for this wrapper is to run consistent tests and to be able to swap the underlying marshaller.
func serialize[T any](model T) ([]byte, error) {
	return encoder.Marshal(model)
}

// deserializes the raw database representation into a model structure.
// reason for this wrapper is to run consistent tests and to be able to swap the underlying marshaller.
// + this function also zero initializes dynamite models properly (filling default values into supported interface types).
func deserialize[T any](data []byte) (*T, error) {
	newVal := reflect.New(reflect.TypeFor[T]())
	initModel(newVal)
	new := newVal.Interface().(*T)
	if err := decoder.Unmarshal(data, new); err != nil {
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
		fieldValue := model.FieldByIndex(field.Index)

		method, ok := field.Type.MethodByName("zero")
		if ok {
			fieldValue.Set(reflect.New(method.Type.Out(0)))
			continue
		}
		switch field.Type {
		case reflect.TypeFor[KeyField]():
			fieldValue.Set(reflect.ValueOf(Key("")))
		case reflect.TypeFor[ETagField]():
			fieldValue.Set(reflect.Zero(reflect.TypeFor[ETagField]()))
		default:
			initModel(fieldValue)
		}
	}
}
