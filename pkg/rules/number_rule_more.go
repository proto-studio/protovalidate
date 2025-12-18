package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for exclusive minimum
type moreRule[T integer | floating] struct {
	more T
	fmt  string
}

// Evaluate takes a context and value and returns an error if it is not greater than the specified value (exclusive).
func (rule *moreRule[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	if value <= rule.more {
		return errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "field must be greater than %d", rule.more),
		)
	}

	return nil
}

// Conflict returns true for any minimum or exclusive minimum rule.
func (rule *moreRule[T]) Conflict(x Rule[T]) bool {
	_, ok1 := x.(*minRule[T])
	_, ok2 := x.(*moreRule[T])
	return ok1 || ok2
}

// String returns the string representation of the exclusive minimum rule.
// Example: WithMore(2)
func (rule *moreRule[T]) String() string {
	return fmt.Sprintf("WithMore(%"+rule.fmt+")", rule.more)
}

// WithMore returns a new child RuleSet that is constrained to values greater than the provided value (exclusive).
func (v *IntRuleSet[T]) WithMore(more T) *IntRuleSet[T] {
	return v.WithRule(&moreRule[T]{
		more,
		"d",
	})
}

// WithMore returns a new child RuleSet that is constrained to values greater than the provided value (exclusive).
func (v *FloatRuleSet[T]) WithMore(more T) *FloatRuleSet[T] {
	return v.WithRule(&moreRule[T]{
		more,
		"f",
	})
}
