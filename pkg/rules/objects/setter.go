package objects

import "reflect"

// setter is responsible for assigning new values to the output.
// It abstracts the map vs struct logic making the RuleSet code cleaner and
// requiring less variables to be passed between validation function.
type setter interface {
	Set(key string, value any)
}

// mapSetter is an implementation of the setter for
type mapSetter struct {
	out reflect.Value
}

func (ms *mapSetter) Set(key string, value any) {
	ms.out.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))

}

type structSetter struct {
	out     reflect.Value
	mapping map[string]string
}

func (ss *structSetter) Set(key string, value any) {
	field := ss.out.FieldByName(ss.mapping[key])
	if field.Kind() == reflect.Ptr && reflect.ValueOf(value).Kind() != reflect.Ptr {
		valPtr := reflect.New(reflect.TypeOf(value))
		valPtr.Elem().Set(reflect.ValueOf(value))
		field.Set(valPtr)
	} else {
		field.Set(reflect.ValueOf(value))
	}
}
