package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for maximum
type maxRule[T integer | floating] struct {
	max T
	fmt string
}

// Evaluate takes a context and integer value and returns an error if it is not equal or higher than the specified value.
func (rule *maxRule[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	if value > rule.max {
		return errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "field cannot be greater than %d", rule.max),
		)
	}

	return nil
}

// Conflict returns true for any maximum rule.
func (rule *maxRule[T]) Conflict(x Rule[T]) bool {
	_, ok := x.(*maxRule[T])
	return ok
}

// String returns the string representation of the maximum rule.
// Example: WithMax(2)
func (rule *maxRule[T]) String() string {
	return fmt.Sprintf("WithMax(%"+rule.fmt+")", rule.max)
}

// WithMax returns a new child RuleSet that is constrained to the provided maximum value.
func (v *IntRuleSet[T]) WithMax(max T) *IntRuleSet[T] {
	return v.WithRule(&maxRule[T]{
		max,
		"d",
	})
}

// WithMax returns a new child RuleSet that is constrained to the provided maximum value.
func (v *FloatRuleSet[T]) WithMax(max T) *FloatRuleSet[T] {
	return v.WithRule(&maxRule[T]{
		max,
		"f",
	})
}
