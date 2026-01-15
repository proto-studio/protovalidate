package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for exclusive minimum
type minExclusiveRule[T integer | floating] struct {
	min T
	fmt string
}

// Evaluate takes a context and value and returns an error if it is not greater than the specified value (exclusive).
func (rule *minExclusiveRule[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	if value <= rule.min {
		return errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "field must be greater than %d", rule.min),
		)
	}

	return nil
}

// Replaces returns true for any minimum or exclusive minimum rule.
func (rule *minExclusiveRule[T]) Replaces(x Rule[T]) bool {
	_, ok1 := x.(*minRule[T])
	_, ok2 := x.(*minExclusiveRule[T])
	return ok1 || ok2
}

// String returns the string representation of the exclusive minimum rule.
// Example: WithMinExclusive(2)
func (rule *minExclusiveRule[T]) String() string {
	return fmt.Sprintf("WithMinExclusive(%"+rule.fmt+")", rule.min)
}

// WithMinExclusive returns a new child RuleSet that is constrained to values greater than the provided value (exclusive).
func (v *IntRuleSet[T]) WithMinExclusive(min T) *IntRuleSet[T] {
	return v.WithRule(&minExclusiveRule[T]{
		min,
		"d",
	})
}

// WithMinExclusive returns a new child RuleSet that is constrained to values greater than the provided value (exclusive).
func (v *FloatRuleSet[T]) WithMinExclusive(min T) *FloatRuleSet[T] {
	return v.WithRule(&minExclusiveRule[T]{
		min,
		"g",
	})
}
