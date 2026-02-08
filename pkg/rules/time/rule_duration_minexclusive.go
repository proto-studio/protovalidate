package time

import (
	"context"
	"fmt"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for exclusive minimum duration
type minExclusiveDurationRule struct {
	min time.Duration
}

// Evaluate takes a context and duration value and returns an error if it is not greater than the specified value (exclusive).
func (rule *minExclusiveDurationRule) Evaluate(ctx context.Context, value time.Duration) errors.ValidationError {
	// Exclusive: value must be > min, so reject if value <= min
	if value <= rule.min {
		return errors.Errorf(errors.CodeMinExclusive, ctx, "below minimum", "must be greater than %s", rule.min)
	}

	return nil
}

// Replaces returns true for any minimum or exclusive minimum rule.
func (rule *minExclusiveDurationRule) Replaces(x rules.Rule[time.Duration]) bool {
	_, ok1 := x.(*minDurationRule)
	_, ok2 := x.(*minExclusiveDurationRule)
	return ok1 || ok2
}

// String returns the string representation of the exclusive minimum rule.
// Example: WithMinExclusive(1h)
func (rule *minExclusiveDurationRule) String() string {
	return fmt.Sprintf("WithMinExclusive(%s)", rule.min)
}

// WithMinExclusive returns a new child RuleSet that is constrained to values greater than the provided duration value (exclusive).
func (v *DurationRuleSet) WithMinExclusive(min time.Duration) *DurationRuleSet {
	return v.WithRule(&minExclusiveDurationRule{
		min,
	})
}
