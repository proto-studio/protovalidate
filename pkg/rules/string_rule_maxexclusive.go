package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for exclusive maximum string value (lexicographical comparison)
type stringMaxExclusiveRule struct {
	max string
}

// Evaluate takes a context and string value and returns an error if it is lexicographically greater than or equal to the specified value.
func (rule *stringMaxExclusiveRule) Evaluate(ctx context.Context, value string) errors.ValidationError {
	if value >= rule.max {
		return errors.Error(errors.CodeMaxExclusive, ctx, util.TruncateString(rule.max))
	}

	return nil
}

// Replaces returns true for any maximum or exclusive maximum string value rule.
func (rule *stringMaxExclusiveRule) Replaces(x Rule[string]) bool {
	_, ok1 := x.(*stringMaxRule)
	_, ok2 := x.(*stringMaxExclusiveRule)
	return ok1 || ok2
}

// String returns the string representation of the exclusive maximum string value rule.
// Example: WithMaxExclusive("xyz")
func (rule *stringMaxExclusiveRule) String() string {
	truncated := util.TruncateString(rule.max)
	return fmt.Sprintf("WithMaxExclusive(%q)", truncated)
}

// WithMaxExclusive returns a new child RuleSet that is constrained to values less than the provided string value (exclusive).
// Strings are compared using lexicographical comparison.
func (v *StringRuleSet) WithMaxExclusive(max string) *StringRuleSet {
	return v.WithRule(&stringMaxExclusiveRule{
		max: max,
	})
}
