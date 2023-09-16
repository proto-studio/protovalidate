package strings

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for maximum length
type maxLenRule struct {
	max int
}

// Evaluate takes a context and string value and returns an error if it is not equal or higher in length than the specified value.
func (rule *maxLenRule) Evaluate(ctx context.Context, value string) (string, errors.ValidationErrorCollection) {
	if len(value) > rule.max {
		return value, errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "field must be at most %d characters long", rule.max),
		)
	}

	return value, nil
}

// WithMaxLen returns a new child RuleSet that is constrained to the provided maximum string length.
func (v *StringRuleSet) WithMaxLen(max int) *StringRuleSet {
	return v.WithRule(&maxLenRule{
		max,
	})
}
