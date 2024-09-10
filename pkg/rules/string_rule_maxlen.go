package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for maximum length
type stringMaxLenRule struct {
	max int
}

// Evaluate takes a context and string value and returns an error if it is not equal or higher in length than the specified value.
func (rule *stringMaxLenRule) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	if len(value) > rule.max {
		return errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "field must be at most %d characters long", rule.max),
		)
	}

	return nil
}

// Conflict returns true for any maximum length rule.
func (rule *stringMaxLenRule) Conflict(x Rule[string]) bool {
	_, ok := x.(*stringMaxLenRule)
	return ok
}

// String returns the string representation of the maximum length rule.
// Example: WithMaxLen(2)
func (rule *stringMaxLenRule) String() string {
	return fmt.Sprintf("WithMaxLen(%d)", rule.max)
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided maximum string length.
func (v *StringRuleSet) WithMaxLen(max int) *StringRuleSet {
	return v.WithRule(&stringMaxLenRule{
		max,
	})
}
