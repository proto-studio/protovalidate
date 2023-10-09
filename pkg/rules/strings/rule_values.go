package strings

import (
	"context"
	"sort"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for an allowed list of values.
type valuesRule struct {
	values []string
	allow  bool
}

// exists returns true if the value exists in the rule
func (rule *valuesRule) exists(value string) bool {
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
func (rule *valuesRule) Evaluate(ctx context.Context, value string) (string, errors.ValidationErrorCollection) {
	exists := rule.exists(value)

	if rule.allow {
		if !exists {
			return value, errors.Collection(
				errors.Errorf(errors.CodeNotAllowed, ctx, "field value is not allowed"),
			)
		}
	} else if exists {
		return value, errors.Collection(
			errors.Errorf(errors.CodeForbidden, ctx, "field value is not allowed"),
		)
	}

	return value, nil
}

// Conflict returns two for allow rules and always returns false for deny rules.
func (rule *valuesRule) Conflict(x rules.Rule[string]) bool {
	if !rule.allow {
		return false
	}

	if other, ok := x.(*valuesRule); ok {
		return other.allow
	}
	return false
}

// String returns the string representation of the values rule.
// Example: WithAllowedValues("b", "b", "c")
func (rule *valuesRule) String() string {
	if !rule.allow {
		return util.StringsToRuleOutput("WithRejectedValues", rule.values)

	}
	return util.StringsToRuleOutput("WithAllowedValues", rule.values)
}

// getValuesRule returns the previous defined values rule for the rule set that has the expected value for "allow".
// Returns nil if there is none.
func (ruleSet *StringRuleSet) getValuesRule(allow bool) *valuesRule {
	for currentRuleSet := ruleSet; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule == nil {
			continue
		}

		if valueRule, ok := currentRuleSet.rule.(*valuesRule); ok && valueRule.allow == allow {
			return valueRule
		}
	}
	return nil
}

// WithAllowedValues returns a new child RuleSet that is checked against the provided list of allowed values.
//
// Allowed values bypass all validation rules that were defined first.
func (ruleSet *StringRuleSet) WithAllowedValues(value string, rest ...string) *StringRuleSet {
	existing := ruleSet.getValuesRule(true)
	l := 1 + len(rest)

	if existing != nil {
		l += len(existing.values)
	}

	values := make([]string, 0, l)
	values = append(values, value)
	values = append(values, rest...)

	// Get previous rule if there is one
	if existing != nil {
		values = append(values, existing.values...)
	}

	// slices.Sort is faster but would require GO 1.21 and we're trying to keep the requirements to 1.20.
	sort.Strings(values)

	return ruleSet.WithRule(&valuesRule{
		values,
		true,
	})
}

// WithRejectedValues returns a new child RuleSet that is checked against the provided list of values hat should be rejected.
// This method can be called more than once.
func (ruleSet *StringRuleSet) WithRejectedValues(value string, rest ...string) *StringRuleSet {
	values := make([]string, 0, 1+len(rest))
	values = append(values, value)
	values = append(values, rest...)

	// slices.Sort is faster but would require GO 1.21 and we're trying to keep the requirements to 2.20.
	sort.Strings(values)

	return ruleSet.WithRule(&valuesRule{
		values,
		false,
	})
}
