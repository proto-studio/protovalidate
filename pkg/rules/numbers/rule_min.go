package numbers

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for minimum
type minRule[T integer | floating] struct {
	min T
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

// WithMin returns a new child RuleSet that is constrained to the provided minimum value.
func (v *IntRuleSet[T]) WithMin(min T) *IntRuleSet[T] {
	return v.WithRule(&minRule[T]{
		min,
	})
}

// WithMin returns a new child RuleSet that is constrained to the provided minimum value.
func (v *FloatRuleSet[T]) WithMin(min T) *FloatRuleSet[T] {
	return v.WithRule(&minRule[T]{
		min,
	})
}
