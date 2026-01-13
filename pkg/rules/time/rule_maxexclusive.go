package time

import (
	"context"
	"fmt"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for exclusive maximum
type maxExclusiveTimeRule struct {
	max time.Time
}

// Evaluate takes a context and time value and returns an error if it is not before the specified value (exclusive).
func (rule *maxExclusiveTimeRule) Evaluate(ctx context.Context, value time.Time) errors.ValidationErrorCollection {
	// Exclusive: value must be < max, so reject if value >= max
	if !value.Before(rule.max) {
		return errors.Collection(
			errors.Errorf(errors.CodeMaxExclusive, ctx, "above maximum", "must be before %s", rule.max),
		)
	}

	return nil
}

// Conflict returns true for any maximum or exclusive maximum rule.
func (rule *maxExclusiveTimeRule) Conflict(x rules.Rule[time.Time]) bool {
	_, ok1 := x.(*maxTimeRule)
	_, ok2 := x.(*maxExclusiveTimeRule)
	return ok1 || ok2
}

// String returns the string representation of the exclusive maximum rule.
// Example: WithMaxExclusive(2023...)
func (rule *maxExclusiveTimeRule) String() string {
	return fmt.Sprintf("WithMaxExclusive(%s)", rule.max)
}

// WithMaxExclusive returns a new child RuleSet that is constrained to values before the provided time value (exclusive).
func (v *TimeRuleSet) WithMaxExclusive(max time.Time) *TimeRuleSet {
	return v.WithRule(&maxExclusiveTimeRule{
		max,
	})
}
