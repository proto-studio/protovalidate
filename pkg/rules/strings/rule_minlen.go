package strings

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for minimum length.
type minLenRule struct {
	min int
}

// Evaluate takes a context and string value and returns an error if it is not equal or greater in length than the specified value.
func (rule *minLenRule) Evaluate(ctx context.Context, value string) (string, errors.ValidationErrorCollection) {
	if len(value) < rule.min {
		return value, errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "field must be at least %d characters long", rule.min),
		)
	}

	return value, nil
}

// Conflict returns true for any minimum length rule.
func (rule *minLenRule) Conflict(x rules.Rule[string]) bool {
	_, ok := x.(*minLenRule)
	return ok
}

// String returns the string representation of the minimum length rule.
// Example: WithMinLen(2)
func (rule *minLenRule) String() string {
	return fmt.Sprintf("WithMinLen(%d)", rule.min)
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided minimum string length.
func (v *StringRuleSet) WithMinLen(min int) *StringRuleSet {
	return v.WithRule(&minLenRule{
		min,
	})
}
