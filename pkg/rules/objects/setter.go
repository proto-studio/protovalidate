package objects

import (
	"reflect"
)

// setter is responsible for assigning new values to the output.
// It abstracts the map vs struct logic making the RuleSet code cleaner and
// requiring less variables to be passed between validation function.
type setter[TK comparable] interface {
	Set(key TK, value any)
	Map() bool
}

// mapSetter is an implementation of the setter for
type mapSetter[TK comparable] struct {
	out reflect.Value
}

func (ms *mapSetter[TK]) Set(key TK, value any) {
	if value == nil {
		elemType := ms.out.Type().Elem()
		ms.out.SetMapIndex(reflect.ValueOf(key), reflect.Zero(elemType))
		return
	}
	ms.out.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
}

func (ms *mapSetter[TK]) Map() bool {
	return true
}

type structSetter[TK comparable] struct {
	out     reflect.Value
	mapping map[TK]TK
}

func (ss *structSetter[TK]) Set(key TK, value any) {
	field := ss.out.FieldByName(any(ss.mapping[key]).(string))

	valueReflect := reflect.ValueOf(value)

	if value == nil {
		field.Set(reflect.Zero(field.Type()))
		return
	}

	if field.Kind() == reflect.Ptr {
		if valueReflect.Kind() == reflect.Ptr {
			field.Set(valueReflect)
		} else {
			valPtr := reflect.New(field.Type().Elem())
			valPtr.Elem().Set(valueReflect)
			field.Set(valPtr)
		}
	} else {
		field.Set(valueReflect)
	}
}

func (ss *structSetter[TK]) Map() bool {
	return false
}
