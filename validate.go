// Package validate is used to create new rule sets that can be used to
// validate data and return usable errors. It implements the most common
// data types.
//
// This main package is used for convenience.
// You can also import packages independently from the subdirectories and
// implement your own rules and rule sets.
package validate

import (
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/arrays"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/rules/objects"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/rules/time"
)

// Array returns a new rule set that can be used to validate arrays of a
// specific type.
func Array[T any]() *arrays.ArrayRuleSet[T] {
	return arrays.New[T]()
}

// ArrayAny returns a new rule set that can be used to validate arrays of
// any type.
//
// These are useful for array that come from untyped formats such as Json.
func ArrayAny() *arrays.ArrayRuleSet[any] {
	return arrays.New[any]()
}

// Constant returns a new rule set that can be used to validate a constant value.
// Constant rule set only return without error when the value is exactly equal
// including type and pointer address (if relevant).
func Constant[T comparable](value T) *rules.ConstantRuleSet[T] {
	return rules.Constant[T](value)
}

// Interface returns a new rule set that can be used to validate a value that
// implements the specific interface.
func Interface[T comparable](value T) *rules.InterfaceRuleSet[T] {
	return rules.Interface[T]()
}

// Int returns a new rule set that can be used to validate integers with type int.
func Int() *numbers.IntRuleSet[int] {
	return numbers.NewInt()
}

// Uint returns a new rule set that can be used to validate integers with type uint32.
func Uint() *numbers.IntRuleSet[uint] {
	return numbers.NewUint()
}

// Int8 returns a new rule set that can be used to validate integers with type int8.
func Int8() *numbers.IntRuleSet[int8] {
	return numbers.NewInt8()
}

// Uint8 returns a new rule set that can be used to validate integers with type uint32.
func Uint8() *numbers.IntRuleSet[uint8] {
	return numbers.NewUint8()
}

// Int16 returns a new rule set that can be used to validate integers with type int16.
func Int16() *numbers.IntRuleSet[int16] {
	return numbers.NewInt16()
}

// Uint16 returns a new rule set that can be used to validate integers with type uint32.
func Uint16() *numbers.IntRuleSet[uint16] {
	return numbers.NewUint16()
}

// Int32 returns a new rule set that can be used to validate integers with type int32.
func Int32() *numbers.IntRuleSet[int32] {
	return numbers.NewInt32()
}

// Uint32 returns a new rule set that can be used to validate integers with type uint32.
func Uint32() *numbers.IntRuleSet[uint32] {
	return numbers.NewUint32()
}

// Int64 returns a new rule set that can be used to validate integers with type int64.
func Int64() *numbers.IntRuleSet[int64] {
	return numbers.NewInt64()
}

// Uint64 returns a new rule set that can be used to validate integers with type uint64.
func Uint64() *numbers.IntRuleSet[uint64] {
	return numbers.NewUint64()
}

// Float64 returns a new rule set that can be used to validate floating point numbers with type float64.
func Float32() *numbers.FloatRuleSet[float32] {
	return numbers.NewFloat32()
}

// Float64 returns a new rule set that can be used to validate floating point numbers with type float64.
func Float64() *numbers.FloatRuleSet[float64] {
	return numbers.NewFloat64()
}

// Map returns a new rule set that can be used to validate a map containing
// a string as a key and a single data type as the value.
func Map[T any]() *objects.ObjectRuleSet[map[string]T, string, T] {
	return objects.NewObjectMap[T]()
}

// Map returns a new rule set that can be used to validate a map containing
// a string as a key and values of any type.
//
// These are useful for maps that come from untyped formats such as Json.
func MapAny() *objects.ObjectRuleSet[map[string]any, string, any] {
	return objects.NewObjectMap[any]()
}

// Object returns a validator that can be used to validate an object of an
// arbitrary data type.
//
// Using the "validate" annotation you can may input values to different
// properties of the object. This is useful for converting unstructured maps
// created from Json and converting to an object.
func Object[T any]() *objects.ObjectRuleSet[T, string, any] {
	return objects.New[T]()
}

// String returns a new rule set that can be used to validate strings.
func String() *strings.StringRuleSet {
	return strings.New()
}

// Domain returns a new rule set that can be used to validate domain names.
func Domain() *net.DomainRuleSet {
	return net.NewDomain()
}

// Email returns a new rule set that can be used to validate domain names.
func Email() *net.EmailRuleSet {
	return net.NewEmail()
}

// Time returns a new rule set that can be used to validate time.Time objects.
// Input can be either a time.Time instance or a string.
//
// When accepting string as an input make sure to call WithLayouts to specify the
// desired input formats.
func Time() *time.TimeRuleSet {
	return time.NewTime()
}

// Time returns a new rule set that can be used to validate time objects.
// This behaves identical to Time except that the value returned from validation
// is a string representation of the time.
//
// It takes the desired output layout as an argument.
// See: time.Format
//
// Input can be either a time.Time instance or a string.
//
// When accepting string it will default the same same format as the output. You can
// call WithLayouts to override the default behavior.
func TimeString(layout string) *time.TimeStringRuleSet {
	return time.NewTimeString(layout)
}
