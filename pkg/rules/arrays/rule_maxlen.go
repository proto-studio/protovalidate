package arrays

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
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

// Conflict returns true for any maximum length rule.
func (rule *maxLenRule[T]) Conflict(x rules.Rule[[]T]) bool {
	_, ok := x.(*maxLenRule[T])
	return ok
}

// String returns the string representation of the maximum length rule.
// Example: WithMaxLen(2)
func (rule *maxLenRule[T]) String() string {
	return fmt.Sprintf("WithMaxLen(%d)", rule.max)
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided maximum array/slice length.
func (v *ArrayRuleSet[T]) WithMaxLen(min int) *ArrayRuleSet[T] {
	return v.WithRule(&maxLenRule[T]{
		min,
	})
}
