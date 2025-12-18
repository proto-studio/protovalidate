package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for exclusive minimum string value (lexicographical comparison)
type stringMoreRule struct {
	more string
}

// Evaluate takes a context and string value and returns an error if it is lexicographically less than or equal to the specified value.
func (rule *stringMoreRule) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	if value <= rule.more {
		return errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "value must be greater than %q", truncateString(rule.more)),
		)
	}

	return nil
}

// Conflict returns true for any minimum or exclusive minimum string value rule.
func (rule *stringMoreRule) Conflict(x Rule[string]) bool {
	_, ok1 := x.(*stringMinRule)
	_, ok2 := x.(*stringMoreRule)
	return ok1 || ok2
}

// String returns the string representation of the exclusive minimum string value rule.
// Example: WithMore("abc")
func (rule *stringMoreRule) String() string {
	return fmt.Sprintf("WithMore(%q)", rule.more)
}

// WithMore returns a new child RuleSet that is constrained to values greater than the provided string value (exclusive).
// Strings are compared using lexicographical comparison.
func (v *StringRuleSet) WithMore(more string) *StringRuleSet {
	return v.WithRule(&stringMoreRule{
		more: more,
	})
}
