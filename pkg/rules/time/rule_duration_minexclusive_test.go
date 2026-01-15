package time_test

import (
	"context"
	"fmt"
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/time"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestDurationRuleSet_WithMinExclusive tests:
// - Durations before the minimum fail validation
// - Durations equal to the minimum fail validation (exclusive)
// - Durations after the minimum pass validation
func TestDurationRuleSet_WithMinExclusive(t *testing.T) {
	min := 1 * internalTime.Hour
	before := 30 * internalTime.Minute
	after := 2 * internalTime.Hour

	ruleSet := time.Duration().WithMinExclusive(min).Any()

	// before is before min, should fail
	testhelpers.MustNotApply(t, ruleSet, before, errors.CodeMinExclusive)

	// min is equal to min, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, min, errors.CodeMinExclusive)

	// after is after min, should pass
	testhelpers.MustApply(t, ruleSet, after)
}

// TestDurationRuleSet_WithMinExclusive_Conflict tests:
// - Only one WithMinExclusive can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMinExclusive is used.
func TestDurationRuleSet_WithMinExclusive_Conflict(t *testing.T) {
	min := 1 * internalTime.Hour
	before := 30 * internalTime.Minute
	after := 2 * internalTime.Hour

	// Create an initial rule set with after and before values
	ruleSet := time.Duration().WithMinExclusive(min).WithMaxExclusive(after)

	// Prepare an output variable for Apply
	var output internalTime.Duration

	// Apply with a duration equal to the threshold, expecting an error (exclusive)
	err := ruleSet.Apply(context.TODO(), min, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a duration before the threshold, expecting an error
	err = ruleSet.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a duration after the threshold, expecting no error
	err = ruleSet.Apply(context.TODO(), after-30*internalTime.Minute, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different before value and validate
	ruleSet2 := ruleSet.WithMinExclusive(before)

	// Apply with a duration exactly at the new threshold, expecting an error (exclusive)
	err = ruleSet2.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a duration after the new threshold, expecting no error
	err = ruleSet2.Apply(context.TODO(), min, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := fmt.Sprintf("DurationRuleSet.WithMinExclusive(%s).WithMaxExclusive(%s)", min, after)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = fmt.Sprintf("DurationRuleSet.WithMaxExclusive(%s).WithMinExclusive(%s)", after, before)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
