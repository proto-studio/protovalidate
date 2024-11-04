package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for minimum length
type minLenRule[TV any, T lengthy[TV]] struct {
	min int
	msg string
}

// Evaluate takes a context and array/slice value and returns an error if it is not equal or lower in length than the specified value.
func (rule *minLenRule[TV, T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	if len(value) < rule.min {
		return errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, rule.msg, rule.min),
		)
	}
	return nil
}

// Conflict returns true for any minimum length rule.
func (rule *minLenRule[TV, T]) Conflict(x Rule[T]) bool {
	_, ok := x.(*minLenRule[TV, T])
	return ok
}

// String returns the string representation of the minimum length rule.
// Example: WithMinLen(2)
func (rule *minLenRule[TV, T]) String() string {
	return fmt.Sprintf("WithMinLen(%d)", rule.min)
}

// WithMinLen returns a new child RuleSet that is constrained to the provided minimum array/slice length.
func (v *SliceRuleSet[T]) WithMinLen(min int) *SliceRuleSet[T] {
	return v.WithRule(&minLenRule[T, []T]{
		min,
		"list must be at least %d items long",
	})
}

// WithMinLen returns a new child RuleSet that is constrained to the provided minimum string length.
func (v *StringRuleSet) WithMinLen(min int) *StringRuleSet {
	return v.WithRule(&minLenRule[any, string]{
		min,
		"value must be at least %d characters long",
	})
}
