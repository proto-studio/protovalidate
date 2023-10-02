package time

import (
	"context"
	"time"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for maximum
type maxTimeRule struct {
	max time.Time
}

// Evaluate takes a context and integer value and returns an error if it is not equal or lower than the specified value.
func (rule *maxTimeRule) Evaluate(ctx context.Context, value time.Time) (time.Time, errors.ValidationErrorCollection) {
	if value.After(rule.max) {
		return value, errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "field must be on or before %s", rule.max),
		)
	}

	return value, nil
}

// WithMin returns a new child RuleSet that is constrained to the provided minimum time value.
func (v *TimeRuleSet) WithMax(max time.Time) *TimeRuleSet {
	return v.WithRule(&maxTimeRule{
		max,
	})
}

// WithMin returns a new child RuleSet that is constrained to the provided minimum time value.
func (v *TimeStringRuleSet) WithMax(max time.Time) *TimeStringRuleSet {
	return v.WithRule(&maxTimeRule{
		max,
	})
}
