package time

import (
	"context"
	"fmt"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for exclusive minimum
type minExclusiveTimeRule struct {
	min time.Time
}

// Evaluate takes a context and time value and returns an error if it is not after the specified value (exclusive).
func (rule *minExclusiveTimeRule) Evaluate(ctx context.Context, value time.Time) errors.ValidationError {
	// Exclusive: value must be > min, so reject if value <= min
	if !value.After(rule.min) {
		return errors.Errorf(errors.CodeMinExclusive, ctx, "below minimum", "must be after %s", rule.min)
	}

	return nil
}

// Replaces returns true for any minimum or exclusive minimum rule.
func (rule *minExclusiveTimeRule) Replaces(x rules.Rule[time.Time]) bool {
	_, ok1 := x.(*minTimeRule)
	_, ok2 := x.(*minExclusiveTimeRule)
	return ok1 || ok2
}

// String returns the string representation of the exclusive minimum rule.
// Example: WithMinExclusive(2023...)
func (rule *minExclusiveTimeRule) String() string {
	return fmt.Sprintf("WithMinExclusive(%s)", rule.min)
}

// WithMinExclusive returns a new child RuleSet that is constrained to values after the provided time value (exclusive).
func (v *TimeRuleSet) WithMinExclusive(min time.Time) *TimeRuleSet {
	return v.WithRule(&minExclusiveTimeRule{
		min,
	})
}
