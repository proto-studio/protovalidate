package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

const maxStringDisplayLength = 50

// truncateString truncates a string to a maximum length, adding ellipsis if truncated.
func truncateString(s string) string {
	if len(s) <= maxStringDisplayLength {
		return s
	}
	return s[:maxStringDisplayLength] + "..."
}

// Implements the Rule interface for minimum string value (lexicographical comparison)
type stringMinRule struct {
	min string
}

// Evaluate takes a context and string value and returns an error if it is lexicographically less than the specified minimum value.
func (rule *stringMinRule) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	if value < rule.min {
		return errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "value must be greater than or equal to %q", truncateString(rule.min)),
		)
	}

	return nil
}

// Conflict returns true for any minimum or exclusive minimum string value rule.
func (rule *stringMinRule) Conflict(x Rule[string]) bool {
	_, ok1 := x.(*stringMinRule)
	_, ok2 := x.(*stringMinExclusiveRule)
	return ok1 || ok2
}

// String returns the string representation of the minimum string value rule.
// Example: WithMin("abc")
func (rule *stringMinRule) String() string {
	return fmt.Sprintf("WithMin(%q)", rule.min)
}

// WithMin returns a new child RuleSet that is constrained to the provided minimum string value (inclusive).
// Strings are compared using lexicographical comparison.
func (v *StringRuleSet) WithMin(min string) *StringRuleSet {
	return v.WithRule(&stringMinRule{
		min: min,
	})
}
