package time

import (
	"context"
	"fmt"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
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

// Conflict returns true for any minimum rule.
func (rule *minTimeRule) Conflict(x rules.Rule[time.Time]) bool {
	_, ok := x.(*minTimeRule)
	return ok
}

// String returns the string representation of the minimum rule.
// Example: WithMin(2023...)
func (rule *minTimeRule) String() string {
	return fmt.Sprintf("WithMin(%s)", rule.min)
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
