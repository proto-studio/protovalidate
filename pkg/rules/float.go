package rules

import (
	"context"
	"math"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
)

var baseFloat32 FloatRuleSet[float32] = FloatRuleSet[float32]{
	label: "FloatRuleSet[float32]",
}

var baseFloat64 FloatRuleSet[float64] = FloatRuleSet[float64]{
	label: "FloatRuleSet[float64]",
}

type floating interface {
	float64 | float32
}

// Implementation of RuleSet for floats.
type FloatRuleSet[T floating] struct {
	NoConflict[T]
	strict    bool
	rule      Rule[T]
	required  bool
	withNil   bool
	parent    *FloatRuleSet[T]
	rounding  Rounding
	precision int
	label     string
}

// Float32 creates a new float32 RuleSet.
func Float32() *FloatRuleSet[float32] {
	return &baseFloat32
}

// Float64 creates a new float64 RuleSet.
func Float64() *FloatRuleSet[float64] {
	return &baseFloat64
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
		withNil:   v.withNil,
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
// WithRequired is used when nesting a RuleSet and a value is not allowed to be omitted.
func (v *FloatRuleSet[T]) WithRequired() *FloatRuleSet[T] {
	return &FloatRuleSet[T]{
		strict:    v.strict,
		parent:    v,
		required:  true,
		withNil:   v.withNil,
		rounding:  v.rounding,
		precision: v.precision,
		label:     "WithRequired()",
	}
}

// WithNil returns a new child rule set with the withNil flag set.
// WithNil allows values to be explicitly set to nil if the output parameter supports nil values.
// By default, WithNil is false.
func (v *FloatRuleSet[T]) WithNil() *FloatRuleSet[T] {
	return &FloatRuleSet[T]{
		strict:    v.strict,
		parent:    v,
		required:  v.required,
		withNil:   true,
		rounding:  v.rounding,
		precision: v.precision,
		label:     "WithNil()",
	}
}

// Apply performs validation of a RuleSet against a value and assigns the result to the output parameter.
// Apply returns a ValidationErrorCollection if any validation errors occur.
func (v *FloatRuleSet[T]) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, v.withNil, input, output); handled {
		return err
	}

	// Ensure output is a non-nil pointer
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Output must be a non-nil pointer",
		))
	}

	// Attempt to coerce the input value to the correct float type
	floatval, validationErr := v.coerceFloat(input, ctx)
	if validationErr != nil {
		return errors.Collection(validationErr)
	}

	// Apply rounding if specified
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

	// Handle setting the value in output
	outputElem := outputVal.Elem()

	var assignable bool

	// If output is a nil interface, or an assignable type, set it directly to the new float value
	if (outputElem.Kind() == reflect.Interface && outputElem.IsNil()) ||
		(outputElem.Kind() == reflect.Float32 || outputElem.Kind() == reflect.Float64 ||
			outputElem.Type().AssignableTo(reflect.TypeOf(floatval))) {

		outputElem.Set(reflect.ValueOf(floatval))
		assignable = true
	}

	// If the types are incompatible, return an error
	if !assignable {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Cannot assign %T to %T", floatval, outputElem.Interface(),
		))
	}

	allErrors := errors.Collection()

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			if err := currentRuleSet.rule.Evaluate(ctx, floatval); err != nil {
				allErrors = append(allErrors, err...)
			}
		}
	}

	if len(allErrors) != 0 {
		return allErrors
	}
	return nil
}

// Evaluate performs validation of a RuleSet against a float value and returns a ValidationErrorCollection.
func (v *FloatRuleSet[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	var out T
	return v.Apply(ctx, value, &out)
}

// noConflict returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *FloatRuleSet[T]) noConflict(rule Rule[T]) *FloatRuleSet[T] {
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
		withNil:   ruleSet.withNil,
		parent:    newParent,
		rounding:  ruleSet.rounding,
		precision: ruleSet.precision,
		label:     ruleSet.label,
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the given number type.
func (ruleSet *FloatRuleSet[T]) WithRule(rule Rule[T]) *FloatRuleSet[T] {
	return &FloatRuleSet[T]{
		strict:    ruleSet.strict,
		parent:    ruleSet.noConflict(rule),
		rule:      rule,
		required:  ruleSet.required,
		withNil:   ruleSet.withNil,
		rounding:  ruleSet.rounding,
		precision: ruleSet.precision,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// for the given number type.
func (v *FloatRuleSet[T]) WithRuleFunc(rule RuleFunc[T]) *FloatRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the number RuleSet in an Any rule set
// which can then be used in nested validation.
func (v *FloatRuleSet[T]) Any() RuleSet[any] {
	return WrapAny[T](v)
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
