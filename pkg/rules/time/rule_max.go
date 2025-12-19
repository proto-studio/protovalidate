package time

import (
	"context"
	"fmt"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for maximum
type maxTimeRule struct {
	max time.Time
}

// Evaluate takes a context and integer value and returns an error if it is not equal or lower than the specified value.
func (rule *maxTimeRule) Evaluate(ctx context.Context, value time.Time) errors.ValidationErrorCollection {
	if value.After(rule.max) {
		return errors.Collection(
			errors.Errorf(errors.CodeMax, ctx, "field must be on or before %s", rule.max),
		)
	}

	return nil
}

// Conflict returns true for any maximum or exclusive maximum rule.
func (rule *maxTimeRule) Conflict(x rules.Rule[time.Time]) bool {
	_, ok1 := x.(*maxTimeRule)
	_, ok2 := x.(*maxExclusiveTimeRule)
	return ok1 || ok2
}

// String returns the string representation of the maximum rule.
// Example: WithMax(2023...)
func (rule *maxTimeRule) String() string {
	return fmt.Sprintf("WithMax(%s)", rule.max)
}

// WithMin returns a new child RuleSet that is constrained to the provided minimum time value.
func (v *TimeRuleSet) WithMax(max time.Time) *TimeRuleSet {
	return v.WithRule(&maxTimeRule{
		max,
	})
}
