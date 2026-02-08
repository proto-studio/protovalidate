package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Replaces is an interface for types that can check if they replace other rules.
// Types implementing this interface have a Replaces method that takes a Rule[T] and returns whether they replace it.
type Replaces[T any] interface {
	// Replaces returns true if this rule should replace the given rule.
	// It may be used to remove duplicate rules when the new rule replaces the existing rule.
	// The new rule should be kept and the old rule should be disabled.
	//
	// For example, if minimum is called a second time with a higher value, the new value should be taken.
	// If both rules are kept then the effective minimum is the smaller of the two.
	//
	// This method may return false if the previous rule should never be discarded.
	//
	// This method should be called even if the two rule types are not the same.
	Replaces(r Rule[T]) bool
}

// Rule defines the interface for validation rules.
// Rule implementations validate a value of type T and return any validation errors.
//
// Rule extends Replaces[T], which provides the Replaces method for checking if one rule
// should replace another rule during conflict resolution.
//
// Rule extends fmt.Stringer, which provides the String method for obtaining a string representation
// of the rule for debugging purposes.
type Rule[T any] interface {
	Replaces[T]
	fmt.Stringer

	// Evaluate takes in a context and value and returns any validation errors.
	Evaluate(ctx context.Context, value T) errors.ValidationError
}

// RuleFunc implements the Rule interface for functions.
type RuleFunc[T any] func(ctx context.Context, value T) errors.ValidationError

// Evaluate calls the rule function and returns the results.
func (rule RuleFunc[T]) Evaluate(ctx context.Context, value T) errors.ValidationError {
	return rule(ctx, value)
}

// Replaces always returns false for rule functions.
// Replaces cannot perform deduplication; implement the interface instead.
func (rule RuleFunc[T]) Replaces(_ Rule[T]) bool {
	return false
}

// String returns "WithRuleFunc(<function>)" for function rules.
// String cannot be customized; implement the interface instead.
func (rule RuleFunc[T]) String() string {
	return "WithRuleFunc(<function>)"
}
