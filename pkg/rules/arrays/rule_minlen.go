package arrays

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
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

// Conflict returns true for any minimum length rule.
func (rule *minLenRule[T]) Conflict(x rules.Rule[[]T]) bool {
	_, ok := x.(*minLenRule[T])
	return ok
}

// String returns the string representation of the minimum length rule.
// Example: WithMinLen(2)
func (rule *minLenRule[T]) String() string {
	return fmt.Sprintf("WithMinLen(%d)", rule.min)
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided minimum array/slice length.
func (v *ArrayRuleSet[T]) WithMinLen(min int) *ArrayRuleSet[T] {
	return v.WithRule(&minLenRule[T]{
		min,
	})
}
