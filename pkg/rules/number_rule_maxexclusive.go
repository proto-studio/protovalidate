package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for exclusive maximum
type maxExclusiveRule[T integer | floating] struct {
	max T
	fmt string
}

// Evaluate takes a context and value and returns an error if it is not less than the specified value (exclusive).
func (rule *maxExclusiveRule[T]) Evaluate(ctx context.Context, value T) errors.ValidationError {
	if value >= rule.max {
		return errors.Error(errors.CodeMaxExclusive, ctx, rule.max)
	}

	return nil
}

// Replaces returns true for any maximum or exclusive maximum rule.
func (rule *maxExclusiveRule[T]) Replaces(x Rule[T]) bool {
	_, ok1 := x.(*maxRule[T])
	_, ok2 := x.(*maxExclusiveRule[T])
	return ok1 || ok2
}

// String returns the string representation of the exclusive maximum rule.
// Example: WithMaxExclusive(2)
func (rule *maxExclusiveRule[T]) String() string {
	return fmt.Sprintf("WithMaxExclusive(%"+rule.fmt+")", rule.max)
}

// WithMaxExclusive returns a new child RuleSet that is constrained to values less than the provided value (exclusive).
func (v *IntRuleSet[T]) WithMaxExclusive(max T) *IntRuleSet[T] {
	return v.WithRule(&maxExclusiveRule[T]{
		max,
		"d",
	})
}

// WithMaxExclusive returns a new child RuleSet that is constrained to values less than the provided value (exclusive).
func (v *FloatRuleSet[T]) WithMaxExclusive(max T) *FloatRuleSet[T] {
	return v.WithRule(&maxExclusiveRule[T]{
		max,
		"g",
	})
}
