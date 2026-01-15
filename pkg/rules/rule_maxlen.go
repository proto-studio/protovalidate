package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for maximum length
type maxLenRule[TV any, T lengthy[TV]] struct {
	max int
}

// Evaluate takes a context and array/slice value and returns an error if it is not equal or lower in length than the specified value.
func (rule *maxLenRule[TV, T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	if len(value) > rule.max {
		return errors.Collection(
			errors.Error(errors.CodeMaxLen, ctx, rule.max),
		)
	}
	return nil
}

// Replaces returns true for any maximum length rule.
func (rule *maxLenRule[TV, T]) Replaces(x Rule[T]) bool {
	_, ok := x.(*maxLenRule[TV, T])
	return ok
}

// String returns the string representation of the maximum length rule.
// Example: WithMaxLen(2)
func (rule *maxLenRule[TV, T]) String() string {
	return fmt.Sprintf("WithMaxLen(%d)", rule.max)
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided maximum array/slice length.
// The maxLen is applied proactively during item processing, stopping validation of items after the maximum
// length is reached. This allows maxLen to be enforced without requiring all items to be buffered in memory.
func (v *SliceRuleSet[T]) WithMaxLen(max int) *SliceRuleSet[T] {
	newRuleSet := v.clone(
		sliceWithLabel[T](fmt.Sprintf("WithMaxLen(%d)", max)),
		sliceWithConflictType[T](conflictTypeMaxLen),
	)
	newRuleSet.maxLen = max
	return newRuleSet
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided maximum string length.
func (v *StringRuleSet) WithMaxLen(max int) *StringRuleSet {
	return v.WithRule(&maxLenRule[any, string]{
		max: max,
	})
}
