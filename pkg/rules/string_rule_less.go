package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for exclusive maximum string value (lexicographical comparison)
type stringLessRule struct {
	less string
}

// Evaluate takes a context and string value and returns an error if it is lexicographically greater than or equal to the specified value.
func (rule *stringLessRule) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	if value >= rule.less {
		return errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "value must be less than %q", truncateString(rule.less)),
		)
	}

	return nil
}

// Conflict returns true for any maximum or exclusive maximum string value rule.
func (rule *stringLessRule) Conflict(x Rule[string]) bool {
	_, ok1 := x.(*stringMaxRule)
	_, ok2 := x.(*stringLessRule)
	return ok1 || ok2
}

// String returns the string representation of the exclusive maximum string value rule.
// Example: WithLess("xyz")
func (rule *stringLessRule) String() string {
	return fmt.Sprintf("WithLess(%q)", rule.less)
}

// WithLess returns a new child RuleSet that is constrained to values less than the provided string value (exclusive).
// Strings are compared using lexicographical comparison.
func (v *StringRuleSet) WithLess(less string) *StringRuleSet {
	return v.WithRule(&stringLessRule{
		less: less,
	})
}
