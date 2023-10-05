package numbers

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for minimum
type minRule[T integer | floating] struct {
	min T
	fmt string
}

// Evaluate takes a context and integer value and returns an error if it is not equal or greater than the specified value.
func (rule *minRule[T]) Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection) {
	if value < rule.min {
		return value, errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "field must be greater than %d", rule.min),
		)
	}

	return value, nil
}

// Conflict returns true for any minimum rule.
func (rule *minRule[T]) Conflict(x rules.Rule[T]) bool {
	_, ok := x.(*minRule[T])
	return ok
}

// String returns the string representation of the minimum rule.
// Example: WithMin(2)
func (rule *minRule[T]) String() string {
	return fmt.Sprintf("WithMin(%"+rule.fmt+")", rule.min)
}

// WithMin returns a new child RuleSet that is constrained to the provided minimum value.
func (v *IntRuleSet[T]) WithMin(min T) *IntRuleSet[T] {
	return v.WithRule(&minRule[T]{
		min,
		"d",
	})
}

// WithMin returns a new child RuleSet that is constrained to the provided minimum value.
func (v *FloatRuleSet[T]) WithMin(min T) *FloatRuleSet[T] {
	return v.WithRule(&minRule[T]{
		min,
		"f",
	})
}
