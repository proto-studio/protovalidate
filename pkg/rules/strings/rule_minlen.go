package strings

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
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

// WithMaxLen returns a new child RuleSet that is constrained to the provided minimum string length.
func (v *StringRuleSet) WithMinLen(min int) *StringRuleSet {
	return v.WithRule(&minLenRule{
		min,
	})
}
