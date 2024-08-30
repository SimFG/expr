package builtin

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/expr-lang/expr/internal/deref"
)

type DataView struct {
	Type   string              `json:"type,omitempty"`
	Value  string              `json:"value,omitempty"`
	Fields map[string]DataView `json:"fields,omitempty"`
}

var ZeroValue = DataView{
	Type:  "nil",
	Value: "nil",
}

func View(arg any) (result any) {
	defer func() {
		resultBytes, _ := json.Marshal(result)
		result = string(resultBytes)
	}()
	if arg == nil {
		return ZeroValue
	}

	argValue := deref.Value(reflect.ValueOf(arg))

	if !isComplicateKind(argValue.Kind()) {
		return getSimpleView(argValue)
	}

	if argValue.Kind() != reflect.Struct {
		return DataView{
			Type:  argValue.Type().String(),
			Value: fmt.Sprintf("%v", innerInterface(argValue)),
		}
	}

	argType := deref.Type(reflect.TypeOf(arg))
	fields := make(map[string]DataView)
	for i := 0; i < argValue.NumField(); i++ {
		fieldType := argType.Field(i)
		fieldValue := deref.Value(argValue.Field(i))
		if !isComplicateKind(fieldValue.Kind()) {
			fields[fieldType.Name] = getSimpleView(fieldValue)
		} else {
			fields[fieldType.Name] = DataView{
				Type: fieldType.Type.String(),
			}
		}
	}

	return DataView{
		Type:   argValue.Type().String(),
		Fields: fields,
	}
}

func isComplicateKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Struct, reflect.Chan, reflect.Func, reflect.Interface:
		return true
	}
	return false
}

func getSimpleView(arg reflect.Value) DataView {
	switch arg.Kind() {
	case reflect.Array, reflect.Slice:
		return DataView{
			Type:  arg.Type().String(),
			Value: fmt.Sprintf("%v", flatten(arg)),
		}
	case reflect.Map:
		keys := arg.MapKeys()
		out := make([][2]any, len(keys))
		// TODO how to show the detail when the value is a struct
		for i, key := range keys {
			out[i] = [2]any{innerInterface(key), innerInterface(arg.MapIndex(key))}
		}
		return DataView{
			Type:  arg.Type().String(),
			Value: fmt.Sprintf("%v", out),
		}
	}

	return DataView{
		Type:  arg.Type().String(),
		Value: fmt.Sprintf("%v", innerInterface(arg)),
	}
}

func innerInterface(v reflect.Value) any {
	if v.CanInterface() {
		return v.Interface()
	}
	return v
}
