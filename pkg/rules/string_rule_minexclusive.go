package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for exclusive minimum string value (lexicographical comparison)
type stringMinExclusiveRule struct {
	min string
}

// Evaluate takes a context and string value and returns an error if it is lexicographically less than or equal to the specified value.
func (rule *stringMinExclusiveRule) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	if value <= rule.min {
		return errors.Collection(
			errors.Error(errors.CodeMinExclusive, ctx, util.TruncateString(rule.min)),
		)
	}

	return nil
}

// Replaces returns true for any minimum or exclusive minimum string value rule.
func (rule *stringMinExclusiveRule) Replaces(x Rule[string]) bool {
	_, ok1 := x.(*stringMinRule)
	_, ok2 := x.(*stringMinExclusiveRule)
	return ok1 || ok2
}

// String returns the string representation of the exclusive minimum string value rule.
// Example: WithMinExclusive("abc")
func (rule *stringMinExclusiveRule) String() string {
	truncated := util.TruncateString(rule.min)
	return fmt.Sprintf("WithMinExclusive(%q)", truncated)
}

// WithMinExclusive returns a new child RuleSet that is constrained to values greater than the provided string value (exclusive).
// Strings are compared using lexicographical comparison.
func (v *StringRuleSet) WithMinExclusive(min string) *StringRuleSet {
	return v.WithRule(&stringMinExclusiveRule{
		min: min,
	})
}
