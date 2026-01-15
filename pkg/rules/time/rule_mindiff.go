package time

import (
	"context"
	"fmt"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for minimum
type minDiffRule struct {
	min time.Duration
}

// Evaluate takes a context and integer value and returns an error if the difference between the current server time and
// the time.Time value is less than than than the specified value.
func (rule *minDiffRule) Evaluate(ctx context.Context, value time.Time) errors.ValidationErrorCollection {
	if time.Until(value) < rule.min {
		return errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "field must be on or after %s from now", rule.min),
		)
	}

	return nil
}

// Replaces returns true for any minimum diff rule.
func (rule *minDiffRule) Replaces(x rules.Rule[time.Time]) bool {
	_, ok := x.(*minDiffRule)
	return ok
}

// String returns the string representation of the minimum diff rule.
// Example: WithMinDiff(1w2d)
func (rule *minDiffRule) String() string {
	return fmt.Sprintf("WithMinDiff(%s)", rule.min)
}

// WithMinDiff returns a new child RuleSet that is constrained to the provided minimum time as a difference from the current
// time. If you want to test for absolute difference from now and the provided time then you may combine WithMinDiff and
// WithMaxDiff.
//
// Some examples:
// 0 will mean that the time must be on or after the current time.
// -15 * time.Minutes means that the time can be no more than 15 minutes in the past.
// 15 * time.Minutes means that the time can be no less than 15 minutes in the future.
func (v *TimeRuleSet) WithMinDiff(min time.Duration) *TimeRuleSet {
	return v.WithRule(&minDiffRule{
		min,
	})
}
