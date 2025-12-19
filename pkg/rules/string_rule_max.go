package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for maximum string value (lexicographical comparison)
type stringMaxRule struct {
	max string
}

// Evaluate takes a context and string value and returns an error if it is lexicographically greater than the specified maximum value.
func (rule *stringMaxRule) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	if value > rule.max {
		return errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "value must be less than or equal to %q", truncateString(rule.max)),
		)
	}

	return nil
}

// Conflict returns true for any maximum or exclusive maximum string value rule.
func (rule *stringMaxRule) Conflict(x Rule[string]) bool {
	_, ok1 := x.(*stringMaxRule)
	_, ok2 := x.(*stringMaxExclusiveRule)
	return ok1 || ok2
}

// String returns the string representation of the maximum string value rule.
// Example: WithMax("xyz")
func (rule *stringMaxRule) String() string {
	return fmt.Sprintf("WithMax(%q)", rule.max)
}

// WithMax returns a new child RuleSet that is constrained to the provided maximum string value (inclusive).
// Strings are compared using lexicographical comparison.
func (v *StringRuleSet) WithMax(max string) *StringRuleSet {
	return v.WithRule(&stringMaxRule{
		max: max,
	})
}
