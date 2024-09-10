package rules

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for minimum length.
type stringMinLenRule struct {
	min int
}

// Evaluate takes a context and string value and returns an error if it is not equal or greater in length than the specified value.
func (rule *stringMinLenRule) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	if len(value) < rule.min {
		return errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "field must be at least %d characters long", rule.min),
		)
	}

	return nil
}

// Conflict returns true for any minimum length rule.
func (rule *stringMinLenRule) Conflict(x Rule[string]) bool {
	_, ok := x.(*stringMinLenRule)
	return ok
}

// String returns the string representation of the minimum length rule.
// Example: WithMinLen(2)
func (rule *stringMinLenRule) String() string {
	return fmt.Sprintf("WithMinLen(%d)", rule.min)
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided minimum string length.
func (v *StringRuleSet) WithMinLen(min int) *StringRuleSet {
	return v.WithRule(&stringMinLenRule{
		min,
	})
}
