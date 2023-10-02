package numbers

import (
	"context"
	"math"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

type floating interface {
	float64 | float32
}

// Implementation of RuleSet for floats.
type FloatRuleSet[T floating] struct {
	strict    bool
	base      int
	rule      rules.Rule[T]
	required  bool
	parent    *FloatRuleSet[T]
	rounding  Rounding
	precision int
}

// NewFloat32 creates a new float32 RuleSet.
func NewFloat32() *FloatRuleSet[float32] {
	return &FloatRuleSet[float32]{}
}

// NewFloat64 creates a new float64 RuleSet.
func NewFloat64() *FloatRuleSet[float64] {
	return &FloatRuleSet[float64]{}
}

// WithStrict returns a new child RuleSet with the strict flag applied.
// A strict rule will only validate if the value is already the correct type.
//
// With number types, any type will work in strict mode as long as it can be converted
// deterministically and without loss.
func (v *FloatRuleSet[T]) WithStrict() *FloatRuleSet[T] {
	return &FloatRuleSet[T]{
		strict:    true,
		parent:    v,
		base:      v.base,
		required:  v.required,
		rounding:  v.rounding,
		precision: v.precision,
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *FloatRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (v *FloatRuleSet[T]) WithRequired() *FloatRuleSet[T] {
	return &FloatRuleSet[T]{
		strict:    v.strict,
		parent:    v,
		base:      v.base,
		required:  true,
		rounding:  v.rounding,
		precision: v.precision,
	}
}

// Validate performs a validation of a RuleSet against a value and returns a value of the correct float type or
// a ValidationErrorCollection.
func (v *FloatRuleSet[T]) Validate(value any) (T, errors.ValidationErrorCollection) {
	return v.ValidateWithContext(value, context.Background())
}

// ValidateWithContext performs a validation of a RuleSet against a value and returns a value of the correct type or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by rules and error formatting.
func (v *FloatRuleSet[T]) ValidateWithContext(value any, ctx context.Context) (T, errors.ValidationErrorCollection) {
	allErrors := errors.Collection()

	floatval, validationErr := v.coerceFloat(value, ctx)

	if validationErr != nil {
		allErrors = append(allErrors, validationErr)
		return 0, allErrors
	}

	if v.rounding != RoundingNone {
		mul := math.Pow10(v.precision)
		tempFloatval := float64(floatval) * mul

		switch v.rounding {
		case RoundingDown:
			tempFloatval = math.Floor(tempFloatval)
		case RoundingUp:
			tempFloatval = math.Ceil(tempFloatval)
		case RoundingHalfUp:
			tempFloatval = math.Round(tempFloatval)
		case RoundingHalfEven:
			tempFloatval = math.RoundToEven(tempFloatval)
		}

		tempFloatval /= mul
		floatval = T(tempFloatval)
	}

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			newStr, err := currentRuleSet.rule.Evaluate(ctx, floatval)
			if err != nil {
				allErrors = append(allErrors, err...)
			} else {
				floatval = newStr
			}
		}
	}

	if len(allErrors) != 0 {
		return floatval, allErrors
	} else {
		return floatval, nil
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the given number type.
//
// Use this when implementing custom rules.
func (v *FloatRuleSet[T]) WithRule(rule rules.Rule[T]) *FloatRuleSet[T] {
	return &FloatRuleSet[T]{
		strict:    v.strict,
		parent:    v,
		base:      v.base,
		rule:      rule,
		required:  true,
		rounding:  v.rounding,
		precision: v.precision,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// for the given number type.
//
// Use this when implementing custom rules.
func (v *FloatRuleSet[T]) WithRuleFunc(rule rules.RuleFunc[T]) *FloatRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the number RuleSet in any Any rule set
// which can then be used in nested validation.
func (v *FloatRuleSet[T]) Any() rules.RuleSet[any] {
	return rules.WrapAny[T](v)
}

// typeName returns the name for the target integer type.
// Used for error message formatting.
func (v *FloatRuleSet[T]) typeName() string {
	return reflect.ValueOf(*new(T)).Kind().String()
}
