package rules

import (
	"context"
	"fmt"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
)

var baseInt IntRuleSet[int] = IntRuleSet[int]{
	base:  10,
	label: "IntRuleSet[int]",
}

var baseUint IntRuleSet[uint] = IntRuleSet[uint]{
	base:  10,
	label: "IntRuleSet[uint]",
}

var baseInt8 IntRuleSet[int8] = IntRuleSet[int8]{
	base:  10,
	label: "IntRuleSet[int8]",
}

var baseUint8 IntRuleSet[uint8] = IntRuleSet[uint8]{
	base:  10,
	label: "IntRuleSet[uint8]",
}

var baseInt16 IntRuleSet[int16] = IntRuleSet[int16]{
	base:  10,
	label: "IntRuleSet[int16]",
}

var baseUint16 IntRuleSet[uint16] = IntRuleSet[uint16]{
	base:  10,
	label: "IntRuleSet[uint16]",
}

var baseInt32 IntRuleSet[int32] = IntRuleSet[int32]{
	base:  10,
	label: "IntRuleSet[int32]",
}

var baseUint32 IntRuleSet[uint32] = IntRuleSet[uint32]{
	base:  10,
	label: "IntRuleSet[uint32]",
}

var baseInt64 IntRuleSet[int64] = IntRuleSet[int64]{
	base:  10,
	label: "IntRuleSet[int64]",
}

var baseUint64 IntRuleSet[uint64] = IntRuleSet[uint64]{
	base:  10,
	label: "IntRuleSet[uint64]",
}

type integer interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Implementation of RuleSet for integers.
type IntRuleSet[T integer] struct {
	NoConflict[T]
	strict   bool
	base     int
	rule     Rule[T]
	required bool
	withNil  bool
	parent   *IntRuleSet[T]
	rounding Rounding
	label    string
}

// Int creates a new integer RuleSet.
func Int() *IntRuleSet[int] {
	return &baseInt
}

// Uint creates a new unsigned integer RuleSet.
func Uint() *IntRuleSet[uint] {
	return &baseUint
}

// Int8 creates a new 8 bit integer RuleSet.
func Int8() *IntRuleSet[int8] {
	return &baseInt8
}

// Uint8 creates a new unsigned 8 bit integer RuleSet.
func Uint8() *IntRuleSet[uint8] {
	return &baseUint8
}

// Int16 creates a new 16 bit integer RuleSet.
func Int16() *IntRuleSet[int16] {
	return &baseInt16
}

// Uint16 creates a new unsigned 16 bit integer RuleSet.
func Uint16() *IntRuleSet[uint16] {
	return &baseUint16
}

// Int32 creates a new 32 bit integer RuleSet.
func Int32() *IntRuleSet[int32] {
	return &baseInt32
}

// Uint32 creates a new unsigned 32 bit integer RuleSet.
func Uint32() *IntRuleSet[uint32] {
	return &baseUint32
}

// Int64 creates a new int64 RuleSet.
func Int64() *IntRuleSet[int64] {
	return &baseInt64
}

// Uint64 creates a new unsigned 64 bit integer RuleSet.
func Uint64() *IntRuleSet[uint64] {
	return &baseUint64
}

// WithStrict returns a new child RuleSet with the strict flag applied.
// A strict rule will only validate if the value is already the correct type.
//
// With number types, any type will work in strict mode as long as it can be converted
// deterministically and without loss.
func (v *IntRuleSet[T]) WithStrict() *IntRuleSet[T] {
	return &IntRuleSet[T]{
		strict:   true,
		parent:   v,
		base:     v.base,
		required: v.required,
		withNil:  v.withNil,
		rounding: v.rounding,
		label:    "WithStrict()",
	}
}

// WithBase returns a new child rule set with the number base set.
// The base will be used to convert strings to numbers.
// The base has no effect if the RuleSet is strict since strict sets will not convert types.
//
// The default is base 10.
func (v *IntRuleSet[T]) WithBase(base int) *IntRuleSet[T] {
	return &IntRuleSet[T]{
		strict:   v.strict,
		parent:   v,
		base:     base,
		required: v.required,
		withNil:  v.withNil,
		rounding: v.rounding,
		label:    fmt.Sprintf("WithBase(%d)", base),
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *IntRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (v *IntRuleSet[T]) WithRequired() *IntRuleSet[T] {
	return &IntRuleSet[T]{
		strict:   v.strict,
		parent:   v,
		base:     v.base,
		required: true,
		withNil:  v.withNil,
		rounding: v.rounding,
		label:    "WithRequired()",
	}
}

// WithNil returns a new child rule set with the withNil flag set.
// Use WithNil when you want to allow values to be explicitly set to nil if the output parameter supports nil values.
// By default, WithNil is false.
func (v *IntRuleSet[T]) WithNil() *IntRuleSet[T] {
	return &IntRuleSet[T]{
		strict:   v.strict,
		parent:   v,
		base:     v.base,
		required: v.required,
		withNil:  true,
		rounding: v.rounding,
		label:    "WithNil()",
	}
}

// Apply performs a validation of a RuleSet against a value and assigns the result to the output parameter.
// It returns a ValidationErrorCollection if any validation errors occur.
func (ruleSet *IntRuleSet[T]) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, ruleSet.withNil, input, output); handled {
		return err
	}

	// Ensure output is a non-nil pointer
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Output must be a non-nil pointer",
		))
	}

	// Attempt to coerce the input value to an integer
	intval, validationErr := ruleSet.coerceInt(input, ctx)
	if validationErr != nil {
		return errors.Collection(validationErr)
	}

	// Handle setting the value in output
	outputElem := outputVal.Elem()

	var assignable bool

	// If output is a nil interface, or an assignable type, set it directly to the new integer value
	if (outputElem.Kind() == reflect.Interface && outputElem.IsNil()) ||
		(outputElem.Kind() == reflect.Int || outputElem.Kind() == reflect.Int8 ||
			outputElem.Kind() == reflect.Int16 || outputElem.Kind() == reflect.Int32 ||
			outputElem.Kind() == reflect.Int64 || outputElem.Type().AssignableTo(reflect.TypeOf(intval))) {

		outputElem.Set(reflect.ValueOf(intval))
		assignable = true
	}

	// If the types are incompatible, return an error
	if !assignable {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Cannot assign %T to %T", intval, outputElem.Interface(),
		))
	}

	allErrors := errors.Collection()

	for currentRuleSet := ruleSet; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			if err := currentRuleSet.rule.Evaluate(ctx, intval); err != nil {
				allErrors = append(allErrors, err...)
			}
		}
	}

	if len(allErrors) != 0 {
		return allErrors
	}
	return nil
}

// Evaluate performs a validation of a RuleSet against an integer value and returns an integer value of the
// same type or a ValidationErrorCollection.
func (v *IntRuleSet[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	allErrors := errors.Collection()

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			if err := currentRuleSet.rule.Evaluate(ctx, value); err != nil {
				allErrors = append(allErrors, err...)
			}
		}
	}

	if len(allErrors) != 0 {
		return allErrors
	} else {
		return nil
	}
}

// withoutConflicts returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *IntRuleSet[T]) withoutConflicts(rule Rule[T]) *IntRuleSet[T] {
	if ruleSet.rule != nil {

		// Conflicting rules, skip this and return the parent
		if rule.Conflict(ruleSet.rule) {
			return ruleSet.parent.withoutConflicts(rule)
		}

	}

	if ruleSet.parent == nil {
		return ruleSet
	}

	newParent := ruleSet.parent.withoutConflicts(rule)

	if newParent == ruleSet.parent {
		return ruleSet
	}

	return &IntRuleSet[T]{
		strict:   ruleSet.strict,
		base:     ruleSet.base,
		rule:     ruleSet.rule,
		required: ruleSet.required,
		withNil:  ruleSet.withNil,
		parent:   newParent,
		rounding: ruleSet.rounding,
		label:    ruleSet.label,
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the given number type.
//
// Use this when implementing custom rules.
func (ruleSet *IntRuleSet[T]) WithRule(rule Rule[T]) *IntRuleSet[T] {
	return &IntRuleSet[T]{
		strict:   ruleSet.strict,
		rule:     rule,
		parent:   ruleSet.withoutConflicts(rule),
		base:     ruleSet.base,
		required: ruleSet.required,
		withNil:  ruleSet.withNil,
		rounding: ruleSet.rounding,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// for the given number type.
//
// Use this when implementing custom rules.
func (v *IntRuleSet[T]) WithRuleFunc(rule RuleFunc[T]) *IntRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the number RuleSet in any Any rule set
// which can then be used in nested validation.
func (v *IntRuleSet[T]) Any() RuleSet[any] {
	return WrapAny[T](v)
}

// typeName returns the name for the target integer type.
// Used for error message formatting.
func (v *IntRuleSet[T]) typeName() string {
	return reflect.ValueOf(*new(T)).Kind().String()
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *IntRuleSet[T]) String() string {
	label := ruleSet.label

	if label == "" && ruleSet.rule != nil {
		label = ruleSet.rule.String()
	}

	if ruleSet.parent != nil {
		return ruleSet.parent.String() + "." + label
	}
	return label
}
