package rules

import (
	"reflect"
)

// setter is responsible for assigning new values to the output.
// It abstracts the map vs struct logic making the RuleSet code cleaner and
// requiring less variables to be passed between validation function.
type setter[TK comparable] interface {
	Set(key TK, value any)
	SetBucket(bucketName, key TK, value any)
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

func (ms *mapSetter[TK]) SetBucket(bucketName, key TK, value any) {
	// Check if the bucket already exists
	bucketValue := ms.out.MapIndex(reflect.ValueOf(bucketName))
	if !bucketValue.IsValid() {
		// If no bucket exists, create a new map[TK]interface{}
		newMap := make(map[TK]interface{})
		ms.out.SetMapIndex(reflect.ValueOf(bucketName), reflect.ValueOf(newMap))
		bucketValue = reflect.ValueOf(newMap)
	} else {
		bucketValue = bucketValue.Elem()
	}

	// Set the key-value pair in the bucket
	if value == nil {
		elemType := bucketValue.Type().Elem()
		bucketValue.SetMapIndex(reflect.ValueOf(key), reflect.Zero(elemType))
	} else {
		bucketValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}
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

func (ss *structSetter[TK]) SetBucket(bucketName, key TK, value any) {
	// Get the field by bucket name
	field := ss.out.FieldByName(any(bucketName).(string))

	if !field.IsValid() || field.Kind() != reflect.Map {
		return
	}

	if field.IsNil() {
		// Initialize the map if it is nil
		mapType := field.Type()
		field.Set(reflect.MakeMap(mapType))
	}

	// Set the key-value pair in the map
	keyValue := reflect.ValueOf(key)
	valueValue := reflect.ValueOf(value)
	if value == nil {
		valueValue = reflect.Zero(field.Type().Elem())
	}
	field.SetMapIndex(keyValue, valueValue)
}

func (ss *structSetter[TK]) Map() bool {
	return false
}
