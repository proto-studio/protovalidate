package arrays

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for minimum length
type minLenRule[T any] struct {
	min int
}

// Evaluate takes a context and array/slice value and returns an error if it is not equal or lower in length than the specified value.
func (rule *minLenRule[T]) Evaluate(ctx context.Context, value []T) ([]T, errors.ValidationErrorCollection) {
	if len(value) < rule.min {
		return value, errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "list must be at least %d items long", rule.min),
		)
	}

	return value, nil
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided minimum array/slice length.
func (v *ArrayRuleSet[T]) WithMinLen(min int) *ArrayRuleSet[T] {
	return v.WithRule(&minLenRule[T]{
		min,
	})
}
