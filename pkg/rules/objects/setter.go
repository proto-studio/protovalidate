package objects

import "reflect"

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
	if field.Kind() == reflect.Ptr && reflect.ValueOf(value).Kind() != reflect.Ptr {
		valPtr := reflect.New(reflect.TypeOf(value))
		valPtr.Elem().Set(reflect.ValueOf(value))
		field.Set(valPtr)
	} else {
		field.Set(reflect.ValueOf(value))
	}
}

func (ss *structSetter[TK]) Map() bool {
	return false
}
