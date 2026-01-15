package time

import (
	"context"
	"fmt"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for exclusive maximum duration
type maxExclusiveDurationRule struct {
	max time.Duration
}

// Evaluate takes a context and duration value and returns an error if it is not less than the specified value (exclusive).
func (rule *maxExclusiveDurationRule) Evaluate(ctx context.Context, value time.Duration) errors.ValidationErrorCollection {
	// Exclusive: value must be < max, so reject if value >= max
	if value >= rule.max {
		return errors.Collection(
			errors.Errorf(errors.CodeMaxExclusive, ctx, "above maximum", "must be less than %s", rule.max),
		)
	}

	return nil
}

// Replaces returns true for any maximum or exclusive maximum rule.
func (rule *maxExclusiveDurationRule) Replaces(x rules.Rule[time.Duration]) bool {
	_, ok1 := x.(*maxDurationRule)
	_, ok2 := x.(*maxExclusiveDurationRule)
	return ok1 || ok2
}

// String returns the string representation of the exclusive maximum rule.
// Example: WithMaxExclusive(24h)
func (rule *maxExclusiveDurationRule) String() string {
	return fmt.Sprintf("WithMaxExclusive(%s)", rule.max)
}

// WithMaxExclusive returns a new child RuleSet that is constrained to values less than the provided duration value (exclusive).
func (v *DurationRuleSet) WithMaxExclusive(max time.Duration) *DurationRuleSet {
	return v.WithRule(&maxExclusiveDurationRule{
		max,
	})
}
