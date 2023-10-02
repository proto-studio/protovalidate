package time

import (
	"context"
	"time"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for minimum
type minTimeRule struct {
	min time.Time
}

// Evaluate takes a context and integer value and returns an error if it is not equal or later than than the specified value.
func (rule *minTimeRule) Evaluate(ctx context.Context, value time.Time) (time.Time, errors.ValidationErrorCollection) {
	if value.Before(rule.min) {
		return value, errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "field must be on or after %s", rule.min),
		)
	}

	return value, nil
}

// WithMin returns a new child RuleSet that is constrained to the provided minimum time value.
func (v *TimeRuleSet) WithMin(min time.Time) *TimeRuleSet {
	return v.WithRule(&minTimeRule{
		min,
	})
}

// WithMin returns a new child RuleSet that is constrained to the provided minimum time value.
func (v *TimeStringRuleSet) WithMin(min time.Time) *TimeStringRuleSet {
	return v.WithRule(&minTimeRule{
		min,
	})
}
