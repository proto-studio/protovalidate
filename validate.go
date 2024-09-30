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
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/rules/time"
)

// Array returns a new rule set that can be used to validate arrays of a
// specific type.
func Array[T any]() *rules.SliceRuleSet[T] {
	return rules.Slice[T]()
}

// ArrayAny returns a new rule set that can be used to validate arrays of
// any type.
//
// These are useful for array that come from untyped formats such as Json.
func ArrayAny() *rules.SliceRuleSet[any] {
	return rules.Slice[any]()
}

// Constant returns a new rule set that can be used to validate a constant value.
// Constant rule set only return without error when the value is exactly equal
// including type and pointer address (if relevant).
func Constant[T comparable](value T) *rules.ConstantRuleSet[T] {
	return rules.Constant[T](value)
}

// Interface returns a new rule set that can be used to validate a value that
// implements the specific interface.
func Interface[T any]() *rules.InterfaceRuleSet[T] {
	return rules.Interface[T]()
}

// Int returns a new rule set that can be used to validate integers with type int.
func Int() *rules.IntRuleSet[int] {
	return rules.Int()
}

// Uint returns a new rule set that can be used to validate integers with type uint32.
func Uint() *rules.IntRuleSet[uint] {
	return rules.Uint()
}

// Int8 returns a new rule set that can be used to validate integers with type int8.
func Int8() *rules.IntRuleSet[int8] {
	return rules.Int8()
}

// Uint8 returns a new rule set that can be used to validate integers with type uint32.
func Uint8() *rules.IntRuleSet[uint8] {
	return rules.Uint8()
}

// Int16 returns a new rule set that can be used to validate integers with type int16.
func Int16() *rules.IntRuleSet[int16] {
	return rules.Int16()
}

// Uint16 returns a new rule set that can be used to validate integers with type uint32.
func Uint16() *rules.IntRuleSet[uint16] {
	return rules.Uint16()
}

// Int32 returns a new rule set that can be used to validate integers with type int32.
func Int32() *rules.IntRuleSet[int32] {
	return rules.Int32()
}

// Uint32 returns a new rule set that can be used to validate integers with type uint32.
func Uint32() *rules.IntRuleSet[uint32] {
	return rules.Uint32()
}

// Int64 returns a new rule set that can be used to validate integers with type int64.
func Int64() *rules.IntRuleSet[int64] {
	return rules.Int64()
}

// Uint64 returns a new rule set that can be used to validate integers with type uint64.
func Uint64() *rules.IntRuleSet[uint64] {
	return rules.Uint64()
}

// Float64 returns a new rule set that can be used to validate floating point numbers with type float64.
func Float32() *rules.FloatRuleSet[float32] {
	return rules.Float32()
}

// Float64 returns a new rule set that can be used to validate floating point numbers with type float64.
func Float64() *rules.FloatRuleSet[float64] {
	return rules.Float64()
}

// Map returns a new rule set that can be used to validate a map containing
// a string as a key and a single data type as the value.
func Map[T any]() *rules.ObjectRuleSet[map[string]T, string, T] {
	return rules.StringMap[T]()
}

// Map returns a new rule set that can be used to validate a map containing
// a string as a key and values of any type.
//
// These are useful for maps that come from untyped formats such as Json.
func MapAny() *rules.ObjectRuleSet[map[string]any, string, any] {
	return rules.StringMap[any]()
}

// Object returns a validator that can be used to validate an object of an
// arbitrary data type.
//
// Using the "validate" annotation you can may input values to different
// properties of the object. This is useful for converting unstructured maps
// created from Json and converting to an object.
func Object[T any]() *rules.ObjectRuleSet[T, string, any] {
	return rules.Struct[T]()
}

// String returns a new rule set that can be used to validate strings.
func String() *rules.StringRuleSet {
	return rules.String()
}

// Domain returns a new rule set that can be used to validate domain names.
func Domain() *net.DomainRuleSet {
	return net.Domain()
}

// URI returns a new rule set that can be used to validate URIs / URLs.
func URI() *net.URIRuleSet {
	return net.URI()
}

// Email returns a new rule set that can be used to validate domain names.
func Email() *net.EmailRuleSet {
	return net.Email()
}

// Time returns a new rule set that can be used to validate time.Time objects.
// Input can be either a time.Time instance or a string.
//
// When accepting string as an input make sure to call WithLayouts to specify the
// desired input formats.
func Time() *time.TimeRuleSet {
	return time.Time()
}
