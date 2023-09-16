package numbers

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for maximum
type maxRule[T integer | floating] struct {
	max T
}

// Evaluate takes a context and integer value and returns an error if it is not equal or higher than the specified value.
func (rule *maxRule[T]) Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection) {
	if value > rule.max {
		return value, errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "field cannot be greater than %d", rule.max),
		)
	}

	return value, nil
}

// WithMax returns a new child RuleSet that is constrained to the provided maximum value.
func (v *IntRuleSet[T]) WithMax(max T) *IntRuleSet[T] {
	return v.WithRule(&maxRule[T]{
		max,
	})
}

// WithMax returns a new child RuleSet that is constrained to the provided maximum value.
func (v *FloatRuleSet[T]) WithMax(max T) *FloatRuleSet[T] {
	return v.WithRule(&maxRule[T]{
		max,
	})
}
