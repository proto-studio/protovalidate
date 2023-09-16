package arrays

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for maximum length
type maxLenRule[T any] struct {
	max int
}

// Evaluate takes a context and array/slice value and returns an error if it is not equal or higher in length than the specified value.
func (rule *maxLenRule[T]) Evaluate(ctx context.Context, value []T) ([]T, errors.ValidationErrorCollection) {
	if len(value) > rule.max {
		return value, errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "list cannot be more than %d items long", rule.max),
		)
	}

	return value, nil
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided maximum array/slice length.
func (v *ArrayRuleSet[T]) WithMaxLen(min int) *ArrayRuleSet[T] {
	return v.WithRule(&maxLenRule[T]{
		min,
	})
}
