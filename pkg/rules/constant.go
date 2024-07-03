package rules

import (
	"context"
	"fmt"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
)

// ConstantRuleSet implements RuleSet that returns an error for
// any value that does not match the constant.
//
// This is primary used for conditional rules. To test a constant of a specific
// type it is usually best to use that type.
type ConstantRuleSet[T comparable] struct {
	required bool
	value    T
	empty    T // Leave this empty
}

// Constant creates a new Constant rule set for the specified value.
func Constant[T comparable](value T) *ConstantRuleSet[T] {
	return &ConstantRuleSet[T]{
		value: value,
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *ConstantRuleSet[T]) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (ruleSet *ConstantRuleSet[T]) WithRequired() *ConstantRuleSet[T] {
	if ruleSet.required {
		return ruleSet
	}

	return &ConstantRuleSet[T]{
		value:    ruleSet.value,
		required: true,
	}
}

// Validate performs a validation of a RuleSet against a value and returns the unaltered supplied value
// or a ValidationErrorCollection.
func (ruleSet *ConstantRuleSet[T]) Validate(value any) (T, errors.ValidationErrorCollection) {
	return ruleSet.ValidateWithContext(value, context.Background())
}

// ValidateWithContext performs a validation of a RuleSet against a value and returns the unaltered supplied value
// or a ValidationErrorCollection.
//
// Also, takes a Context which can be used by rules and error formatting.
func (ruleSet *ConstantRuleSet[T]) ValidateWithContext(value any, ctx context.Context) (T, errors.ValidationErrorCollection) {
	v, ok := value.(T)
	if !ok {
		return ruleSet.empty, errors.Collection(errors.NewCoercionError(ctx, reflect.TypeOf(ruleSet.empty).String(), reflect.TypeOf(value).String()))
	}
	return ruleSet.Evaluate(ctx, v)
}

// Evaluate performs a validation of a RuleSet against a value and returns a value of the same type
// as the wrapped RuleSet or a ValidationErrorCollection. The wrapped rules are called before any rules
// added directly to the WrapConstantRuleSet.
//
// For WrapAny, Evaluate is identical to ValidateWithContext except for the argument order.
func (ruleSet *ConstantRuleSet[T]) Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection) {
	if value != ruleSet.value {
		return ruleSet.empty, errors.Collection(errors.Errorf(errors.CodePattern, ctx, "value does not match"))
	}
	return value, nil
}

// Conflict returns true for all rules since by definition no rule can be a superset of it.
func (ruleSet *ConstantRuleSet[T]) Conflict(other Rule[T]) bool {
	return true
}

// Any is an identity function for this implementation and returns the current rule set.
func (ruleSet *ConstantRuleSet[T]) Any() RuleSet[any] {
	return WrapAny[T](ruleSet)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *ConstantRuleSet[T]) String() string {
	str := fmt.Sprintf(`ConstantRuleSet(%v)`, ruleSet.value)
	if ruleSet.required {
		return str + ".WithRequired()"
	}
	return str
}
