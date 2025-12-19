package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for exclusive maximum string value (lexicographical comparison)
type stringMaxExclusiveRule struct {
	max string
}

// Evaluate takes a context and string value and returns an error if it is lexicographically greater than or equal to the specified value.
func (rule *stringMaxExclusiveRule) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	if value >= rule.max {
		return errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "value must be less than %q", truncateString(rule.max)),
		)
	}

	return nil
}

// Conflict returns true for any maximum or exclusive maximum string value rule.
func (rule *stringMaxExclusiveRule) Conflict(x Rule[string]) bool {
	_, ok1 := x.(*stringMaxRule)
	_, ok2 := x.(*stringMaxExclusiveRule)
	return ok1 || ok2
}

// String returns the string representation of the exclusive maximum string value rule.
// Example: WithMaxExclusive("xyz")
func (rule *stringMaxExclusiveRule) String() string {
	return fmt.Sprintf("WithMaxExclusive(%q)", rule.max)
}

// WithMaxExclusive returns a new child RuleSet that is constrained to values less than the provided string value (exclusive).
// Strings are compared using lexicographical comparison.
func (v *StringRuleSet) WithMaxExclusive(max string) *StringRuleSet {
	return v.WithRule(&stringMaxExclusiveRule{
		max: max,
	})
}
