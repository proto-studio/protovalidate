package numbers

import (
	"context"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

type integer interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Implementation of RuleSet for integers.
type IntRuleSet[T integer] struct {
	strict   bool
	base     int
	rule     rules.Rule[T]
	required bool
	parent   *IntRuleSet[T]
	rounding Rounding
}

// NewInt creates a new integer RuleSet.
func NewInt() *IntRuleSet[int] {
	return &IntRuleSet[int]{
		base: 10,
	}
}

// NewUint creates a new unsigned integer RuleSet.
func NewUint() *IntRuleSet[uint] {
	return &IntRuleSet[uint]{
		base: 10,
	}
}

// NewInt8 creates a new 8 bit integer RuleSet.
func NewInt8() *IntRuleSet[int8] {
	return &IntRuleSet[int8]{
		base: 10,
	}
}

// NewUint8 creates a new unsigned 8 bit integer RuleSet.
func NewUint8() *IntRuleSet[uint8] {
	return &IntRuleSet[uint8]{
		base: 10,
	}
}

// NewInt16 creates a new 16 bit integer RuleSet.
func NewInt16() *IntRuleSet[int16] {
	return &IntRuleSet[int16]{
		base: 10,
	}
}

// NewUint16 creates a new unsigned 16 bit integer RuleSet.
func NewUint16() *IntRuleSet[uint16] {
	return &IntRuleSet[uint16]{
		base: 10,
	}
}

// NewInt32 creates a new 32 bit integer RuleSet.
func NewInt32() *IntRuleSet[int32] {
	return &IntRuleSet[int32]{
		base: 10,
	}
}

// NewUint32 creates a new unsigned 32 bit integer RuleSet.
func NewUint32() *IntRuleSet[uint32] {
	return &IntRuleSet[uint32]{
		base: 10,
	}
}

// NewInt64 creates a new int64 RuleSet.
func NewInt64() *IntRuleSet[int64] {
	return &IntRuleSet[int64]{
		base: 10,
	}
}

// NewUint64 creates a new unsigned 64 bit integer RuleSet.
func NewUint64() *IntRuleSet[uint64] {
	return &IntRuleSet[uint64]{
		base: 10,
	}
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
		rounding: v.rounding,
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
		rounding: v.rounding,
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
		rounding: v.rounding,
	}
}

// Validate performs a validation of a RuleSet against a value and returns a value of the correct integer type or
// a ValidationErrorCollection.
func (v *IntRuleSet[T]) Validate(value any) (T, errors.ValidationErrorCollection) {
	return v.ValidateWithContext(value, context.Background())
}

// ValidateWithContext performs a validation of a RuleSet against a value and returns a value of the correct type or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by validaton rules and error formatting.
func (v *IntRuleSet[T]) ValidateWithContext(value any, ctx context.Context) (T, errors.ValidationErrorCollection) {
	allErrors := errors.Collection()

	intval, validationErr := v.coerceInt(value, ctx)

	if validationErr != nil {
		allErrors.Add(validationErr)
		return 0, allErrors
	}

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			newStr, err := currentRuleSet.rule.Evaluate(ctx, intval)
			if err != nil {
				allErrors.Add(err.All()...)
			} else {
				intval = newStr
			}
		}
	}

	if allErrors.Size() != 0 {
		return intval, allErrors
	} else {
		return intval, nil
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the given number type.
//
// Use this when implementing custom rules.
func (v *IntRuleSet[T]) WithRule(rule rules.Rule[T]) *IntRuleSet[T] {
	return &IntRuleSet[T]{
		strict:   v.strict,
		rule:     rule,
		parent:   v,
		base:     v.base,
		required: v.required,
		rounding: v.rounding,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// for the given number type.
//
// Use this when implementing custom rules.
func (v *IntRuleSet[T]) WithRuleFunc(rule rules.RuleFunc[T]) *IntRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the number RuleSet in any Any rule set
// which can then be used in nested validation.
func (v *IntRuleSet[T]) Any() rules.RuleSet[any] {
	return rules.WrapAny[T](v)
}

// typeName returns the name for the target integer type.
// Used for error message formatting.
func (v *IntRuleSet[T]) typeName() string {
	return reflect.ValueOf(*new(T)).Kind().String()
}
