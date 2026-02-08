package time

import (
	"context"
	"fmt"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for maximum duration
type maxDurationRule struct {
	max time.Duration
}

// Evaluate takes a context and duration value and returns an error if it is greater than the specified value.
func (rule *maxDurationRule) Evaluate(ctx context.Context, value time.Duration) errors.ValidationError {
	if value > rule.max {
		return errors.Errorf(errors.CodeMax, ctx, "above maximum", "must be at most %s", rule.max)
	}

	return nil
}

// Replaces returns true for any maximum or exclusive maximum rule.
func (rule *maxDurationRule) Replaces(x rules.Rule[time.Duration]) bool {
	_, ok1 := x.(*maxDurationRule)
	_, ok2 := x.(*maxExclusiveDurationRule)
	return ok1 || ok2
}

// String returns the string representation of the maximum rule.
// Example: WithMax(24h)
func (rule *maxDurationRule) String() string {
	return fmt.Sprintf("WithMax(%s)", rule.max)
}

// WithMax returns a new child RuleSet that is constrained to the provided maximum duration value.
func (v *DurationRuleSet) WithMax(max time.Duration) *DurationRuleSet {
	return v.WithRule(&maxDurationRule{
		max,
	})
}
