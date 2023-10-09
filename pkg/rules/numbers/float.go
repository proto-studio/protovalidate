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
	rules.NoConflict[T]
	strict    bool
	rule      rules.Rule[T]
	required  bool
	parent    *FloatRuleSet[T]
	rounding  Rounding
	precision int
	label     string
}

// NewFloat32 creates a new float32 RuleSet.
func NewFloat32() *FloatRuleSet[float32] {
	return &FloatRuleSet[float32]{
		label: "FloatRuleSet[float32]",
	}
}

// NewFloat64 creates a new float64 RuleSet.
func NewFloat64() *FloatRuleSet[float64] {
	return &FloatRuleSet[float64]{
		label: "FloatRuleSet[float64]",
	}
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
		required:  v.required,
		rounding:  v.rounding,
		precision: v.precision,
		label:     "WithStrict()",
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
		required:  true,
		rounding:  v.rounding,
		precision: v.precision,
		label:     "WithRequired()",
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
	floatval, validationErr := v.coerceFloat(value, ctx)

	if validationErr != nil {
		return 0, errors.Collection(validationErr)
	}

	return v.Evaluate(ctx, floatval)
}

// Evaluate performs a validation of a RuleSet against a float value and returns a float value of the
// same type or a ValidationErrorCollection.
func (v *FloatRuleSet[T]) Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection) {
	if v.rounding != RoundingNone {
		mul := math.Pow10(v.precision)
		tempFloatval := float64(value) * mul

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
		value = T(tempFloatval)
	}

	allErrors := errors.Collection()

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			newFloat, err := currentRuleSet.rule.Evaluate(ctx, value)
			if err != nil {
				allErrors = append(allErrors, err...)
			} else {
				value = newFloat
			}
		}
	}

	if len(allErrors) != 0 {
		return value, allErrors
	} else {
		return value, nil
	}
}

// noConflict returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *FloatRuleSet[T]) noConflict(rule rules.Rule[T]) *FloatRuleSet[T] {
	if ruleSet.rule != nil {

		// Conflicting rules, skip this and return the parent
		if rule.Conflict(ruleSet.rule) {
			return ruleSet.parent.noConflict(rule)
		}

	}

	if ruleSet.parent == nil {
		return ruleSet
	}

	newParent := ruleSet.parent.noConflict(rule)

	if newParent == ruleSet.parent {
		return ruleSet
	}

	return &FloatRuleSet[T]{
		strict:    ruleSet.strict,
		rule:      ruleSet.rule,
		required:  ruleSet.required,
		parent:    newParent,
		rounding:  ruleSet.rounding,
		precision: ruleSet.precision,
		label:     ruleSet.label,
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the given number type.
//
// Use this when implementing custom rules.
func (ruleSet *FloatRuleSet[T]) WithRule(rule rules.Rule[T]) *FloatRuleSet[T] {
	return &FloatRuleSet[T]{
		strict:    ruleSet.strict,
		parent:    ruleSet.noConflict(rule),
		rule:      rule,
		required:  true,
		rounding:  ruleSet.rounding,
		precision: ruleSet.precision,
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

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *FloatRuleSet[T]) String() string {
	label := ruleSet.label

	if label == "" && ruleSet.rule != nil {
		label = ruleSet.rule.String()
	}

	if ruleSet.parent != nil {
		return ruleSet.parent.String() + "." + label
	}
	return label
}
