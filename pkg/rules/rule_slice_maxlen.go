package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for maximum length
type maxLenRule[T any] struct {
	max int
}

// Evaluate takes a context and array/slice value and returns an error if it is not equal or higher in length than the specified value.
func (rule *maxLenRule[T]) Evaluate(ctx context.Context, value []T) errors.ValidationErrorCollection {
	if len(value) > rule.max {
		return errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "list cannot be more than %d items long", rule.max),
		)
	}
	return nil
}

// Conflict returns true for any maximum length rule.
func (rule *maxLenRule[T]) Conflict(x Rule[[]T]) bool {
	_, ok := x.(*maxLenRule[T])
	return ok
}

// String returns the string representation of the maximum length rule.
// Example: WithMaxLen(2)
func (rule *maxLenRule[T]) String() string {
	return fmt.Sprintf("WithMaxLen(%d)", rule.max)
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided maximum array/slice length.
func (v *SliceRuleSet[T]) WithMaxLen(min int) *SliceRuleSet[T] {
	return v.WithRule(&maxLenRule[T]{
		min,
	})
}
