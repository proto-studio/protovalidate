package time

import (
	"context"
	"time"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for minimum
type maxDiffRule struct {
	max time.Duration
}

// Evaluate takes a context and integer value and returns an error if the difference between the current server time and
// the time.Time value is less than than than the specified value.
func (rule *maxDiffRule) Evaluate(ctx context.Context, value time.Time) (time.Time, errors.ValidationErrorCollection) {
	if value.Sub(time.Now()) > rule.max {
		return value, errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "field must be on or before %s from now", rule.max),
		)
	}

	return value, nil
}

// WithMaxDiff returns a new child RuleSet that is constrained to the provided maximum time as a difference from the current
// time. If you want to test for absolute difference from now and the provided time then you may combine WithMinDiff and
// WithMaxDiff.
//
// Some examples:
// 0 will mean that the time just be after the current time.
// -15 * time.Minutes means that the time can be no more than 15 minutes in the past.
// 15 * time.Minutes means that the time can be no less than 15 minutes in the future.
func (v *TimeRuleSet) WithMaxDiff(max time.Duration) *TimeRuleSet {
	return v.WithRule(&maxDiffRule{
		max,
	})
}

// WithMaxDiff returns a new child RuleSet that is constrained to the provided maximum time as a difference from the current
// time. If you want to test for absolute difference from now and the provided time then you may combine WithMinDiff and
// WithMaxDiff.
//
// Some examples:
// 0 will mean that the time just be after the current time.
// -15 * time.Minutes means that the time can be no more than 15 minutes in the past.
// 15 * time.Minutes means that the time can be no less than 15 minutes in the future.
func (v *TimeStringRuleSet) WithMaxDiff(max time.Duration) *TimeStringRuleSet {
	return v.WithRule(&maxDiffRule{
		max,
	})
}
