package rules

import (
	"context"
	"slices"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for an allowed list of values.
type valuesRule[T integer | floating] struct {
	values []T
	allow  bool
}

// exists returns true if the value exists in the rule
func (rule *valuesRule[T]) exists(value T) bool {
	low, high := 0, len(rule.values)-1

	for low <= high {
		mid := (low + high) / 2

		if rule.values[mid] == value {
			return true
		}

		if rule.values[mid] < value {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return false
}

// Evaluate takes a context and string value and returns an error depending on whether the value is in a list
// of allowed or denied values.
func (rule *valuesRule[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	exists := rule.exists(value)

	if rule.allow {
		if !exists {
			return errors.Collection(
				errors.Error(errors.CodeNotAllowed, ctx),
			)
		}
	} else if exists {
		return errors.Collection(
			errors.Error(errors.CodeForbidden, ctx),
		)
	}

	return nil
}

// Conflict returns two for allow rules and always returns false for deny rules.
func (rule *valuesRule[T]) Conflict(x Rule[T]) bool {
	if !rule.allow {
		return false
	}

	if other, ok := x.(*valuesRule[T]); ok {
		return other.allow
	}
	return false
}

// String returns the string representation of the values rule.
// Example: WithAllowedValues("b", "b", "c")
func (rule *valuesRule[T]) String() string {
	if !rule.allow {
		return util.StringsToRuleOutput("WithRejectedValues", rule.values)

	}
	return util.StringsToRuleOutput("WithAllowedValues", rule.values)
}

// getValuesRule returns the previous defined values rule for the rule set that has the expected value for "allow".
// Returns nil if there is none.
func (ruleSet *IntRuleSet[T]) getValuesRule(allow bool) *valuesRule[T] {
	for currentRuleSet := ruleSet; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule == nil {
			continue
		}

		if valueRule, ok := currentRuleSet.rule.(*valuesRule[T]); ok && valueRule.allow == allow {
			return valueRule
		}
	}
	return nil
}

// WithAllowedValues returns a new child RuleSet that is checked against the provided list of allowed values.
//
// This method can be called more than once and the allowed values are cumulative.
// Allowed values must still pass all other rules.
func (ruleSet *IntRuleSet[T]) WithAllowedValues(value T, rest ...T) *IntRuleSet[T] {
	existing := ruleSet.getValuesRule(true)
	l := 1 + len(rest)

	if existing != nil {
		l += len(existing.values)
	}

	values := make([]T, 0, l)
	values = append(values, value)
	values = append(values, rest...)

	// Get previous rule if there is one
	if existing != nil {
		values = append(values, existing.values...)
	}

	slices.Sort(values)

	return ruleSet.WithRule(&valuesRule[T]{
		values,
		true,
	})
}

// WithRejectedValues returns a new child RuleSet that is checked against the provided list of values hat should be rejected.
// This method can be called more than once.
//
// Rejected values will always be rejected even if they are in the allowed values list.
func (ruleSet *IntRuleSet[T]) WithRejectedValues(value T, rest ...T) *IntRuleSet[T] {
	values := make([]T, 0, 1+len(rest))
	values = append(values, value)
	values = append(values, rest...)

	slices.Sort(values)

	return ruleSet.WithRule(&valuesRule[T]{
		values,
		false,
	})
}
