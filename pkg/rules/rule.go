package rules

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

type Rule[T any] interface {
	// Evaluate takes in a context and value and returns the new value or a collection or errors.
	Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection)

	// Conflict returns true if two rules should not co-exist.
	// It may be used to remove duplicate rules when the new rule conflicts the existing rule.
	// The new rule should be kept and the old rule should be disabled.
	//
	// For example, if minimum is called a second time with a higher value, the new value should be taken.
	// If both rules are kept then the effective minimum is the smaller of the two.
	//
	// This method may return false if the previous rule should never be discarded.
	//
	// This method should be called even if the two rule types are not the same.
	Conflict(Rule[T]) bool

	// Returns the string representation of the rule for debugging.
	String() string
}

// RuleFunc implements the Rule interface for functions.
type RuleFunc[T any] func(ctx context.Context, value T) (T, errors.ValidationErrorCollection)

// Evaluate calls the rule function and returns the results.
func (rule RuleFunc[T]) Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection) {
	return rule(ctx, value)
}

// Conflict always returns false for rule functions.
// To perform deduplication, implement the interface instead.
func (rule RuleFunc[T]) Conflict(_ Rule[T]) bool {
	return false
}

// String always returns WithRule(func()) for function rules.
// To use a different string, implement the interface instead.
func (rule RuleFunc[T]) String() string {
	return "WithRuleFunc(...)"
}
