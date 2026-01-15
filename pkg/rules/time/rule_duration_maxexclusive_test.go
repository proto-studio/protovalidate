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

// TestDurationRuleSet_WithMaxExclusive tests:
// - Durations before the maximum pass validation
// - Durations equal to the maximum fail validation (exclusive)
// - Durations after the maximum fail validation
func TestDurationRuleSet_WithMaxExclusive(t *testing.T) {
	max := 1 * internalTime.Hour
	before := 30 * internalTime.Minute
	after := 2 * internalTime.Hour

	ruleSet := time.Duration().WithMaxExclusive(max).Any()

	// before is before max, should pass
	testhelpers.MustApply(t, ruleSet, before)

	// max is equal to max, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, max, errors.CodeMaxExclusive)

	// after is after max, should fail
	testhelpers.MustNotApply(t, ruleSet, after, errors.CodeMaxExclusive)
}

// TestDurationRuleSet_WithMaxExclusive_Conflict tests:
// - Only one WithMaxExclusive can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMaxExclusive is used.
func TestDurationRuleSet_WithMaxExclusive_Conflict(t *testing.T) {
	max := 1 * internalTime.Hour
	before := 30 * internalTime.Minute
	after := 2 * internalTime.Hour

	// Create an initial rule set with before and after values
	ruleSet := time.Duration().WithMaxExclusive(max).WithMinExclusive(before)

	// Prepare an output variable for Apply
	var output internalTime.Duration

	// Apply with a duration equal to the threshold, expecting an error (exclusive)
	err := ruleSet.Apply(context.TODO(), max, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a duration after the threshold, expecting an error
	err = ruleSet.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a duration before the threshold, expecting no error
	// Use a value strictly between minExclusive (before) and maxExclusive (max)
	middle := before + 15*internalTime.Minute // 45 minutes, which is > 30 min and < 1 hour
	err = ruleSet.Apply(context.TODO(), middle, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different before value and validate
	ruleSet2 := ruleSet.WithMaxExclusive(after)

	// Apply with a duration exactly at the new threshold, expecting an error (exclusive)
	err = ruleSet2.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a duration before the new threshold, expecting no error
	err = ruleSet2.Apply(context.TODO(), max, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := fmt.Sprintf("DurationRuleSet.WithMaxExclusive(%s).WithMinExclusive(%s)", max, before)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = fmt.Sprintf("DurationRuleSet.WithMinExclusive(%s).WithMaxExclusive(%s)", before, after)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
