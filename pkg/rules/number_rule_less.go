package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for exclusive maximum
type lessRule[T integer | floating] struct {
	less T
	fmt  string
}

// Evaluate takes a context and value and returns an error if it is not less than the specified value (exclusive).
func (rule *lessRule[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	if value >= rule.less {
		return errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "field must be less than %d", rule.less),
		)
	}

	return nil
}

// Conflict returns true for any maximum or exclusive maximum rule.
func (rule *lessRule[T]) Conflict(x Rule[T]) bool {
	_, ok1 := x.(*maxRule[T])
	_, ok2 := x.(*lessRule[T])
	return ok1 || ok2
}

// String returns the string representation of the exclusive maximum rule.
// Example: WithLess(2)
func (rule *lessRule[T]) String() string {
	return fmt.Sprintf("WithLess(%"+rule.fmt+")", rule.less)
}

// WithLess returns a new child RuleSet that is constrained to values less than the provided value (exclusive).
func (v *IntRuleSet[T]) WithLess(less T) *IntRuleSet[T] {
	return v.WithRule(&lessRule[T]{
		less,
		"d",
	})
}

// WithLess returns a new child RuleSet that is constrained to values less than the provided value (exclusive).
func (v *FloatRuleSet[T]) WithLess(less T) *FloatRuleSet[T] {
	return v.WithRule(&lessRule[T]{
		less,
		"f",
	})
}
