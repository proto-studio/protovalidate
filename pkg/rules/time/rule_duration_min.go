package time

import (
	"context"
	"fmt"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for minimum duration
type minDurationRule struct {
	min time.Duration
}

// Evaluate takes a context and duration value and returns an error if it is less than the specified value.
func (rule *minDurationRule) Evaluate(ctx context.Context, value time.Duration) errors.ValidationErrorCollection {
	if value < rule.min {
		return errors.Collection(
			errors.Errorf(errors.CodeMin, ctx, "below minimum", "must be at least %s", rule.min),
		)
	}

	return nil
}

// Replaces returns true for any minimum or exclusive minimum rule.
func (rule *minDurationRule) Replaces(x rules.Rule[time.Duration]) bool {
	_, ok1 := x.(*minDurationRule)
	_, ok2 := x.(*minExclusiveDurationRule)
	return ok1 || ok2
}

// String returns the string representation of the minimum rule.
// Example: WithMin(1h)
func (rule *minDurationRule) String() string {
	return fmt.Sprintf("WithMin(%s)", rule.min)
}

// WithMin returns a new child RuleSet that is constrained to the provided minimum duration value.
func (v *DurationRuleSet) WithMin(min time.Duration) *DurationRuleSet {
	return v.WithRule(&minDurationRule{
		min,
	})
}
