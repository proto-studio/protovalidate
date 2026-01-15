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

// TestDurationRuleSet_WithMax tests:
// - Durations before the maximum pass validation
// - Durations equal to the maximum pass validation
// - Durations after the maximum fail validation
func TestDurationRuleSet_WithMax(t *testing.T) {
	max := 1 * internalTime.Hour
	before := 30 * internalTime.Minute
	after := 2 * internalTime.Hour

	ruleSet := time.Duration().WithMax(max).Any()

	testhelpers.MustApply(t, ruleSet, before)
	testhelpers.MustApply(t, ruleSet, max)
	testhelpers.MustNotApply(t, ruleSet, after, errors.CodeMax)
}

// TestDurationRuleSet_WithMax_Conflict tests:
// - Only one max duration can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
func TestDurationRuleSet_WithMax_Conflict(t *testing.T) {
	max := 1 * internalTime.Hour
	before := 30 * internalTime.Minute
	after := 2 * internalTime.Hour

	// Create an initial rule set with a max and min
	ruleSet := time.Duration().WithMax(max).WithMin(before)

	// Prepare an output variable for Apply
	var output internalTime.Duration

	// Apply with a duration after the max, expecting an error
	err := ruleSet.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a duration exactly at the max, expecting no error
	err = ruleSet.Apply(context.TODO(), max, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different max and validate
	ruleSet2 := ruleSet.WithMax(after)

	// Apply with a duration exactly at the new max, expecting no error
	err = ruleSet2.Apply(context.TODO(), after, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := fmt.Sprintf("DurationRuleSet.WithMax(%s).WithMin(%s)", max, before)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = fmt.Sprintf("DurationRuleSet.WithMin(%s).WithMax(%s)", before, after)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMax and WithMaxExclusive conflict with each other
func TestDurationRuleSet_WithMax_WithMaxExclusiveConflict(t *testing.T) {
	max := 1 * internalTime.Hour
	after := 2 * internalTime.Hour

	ruleSet := time.Duration().WithMax(max)

	// Adding WithMaxExclusive should conflict and replace WithMax
	ruleSet2 := ruleSet.WithMaxExclusive(after)

	var output internalTime.Duration

	// Original rule set should still have WithMax
	err := ruleSet.Apply(context.TODO(), max, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// New rule set should have WithMaxExclusive (exclusive, so after should fail)
	err = ruleSet2.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error for WithMaxExclusive at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := fmt.Sprintf("DurationRuleSet.WithMaxExclusive(%s)", after)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMaxExclusive and WithMax conflict with each other (reverse order)
func TestDurationRuleSet_WithMaxExclusive_WithMaxConflict(t *testing.T) {
	max := 1 * internalTime.Hour
	after := 2 * internalTime.Hour

	ruleSet := time.Duration().WithMaxExclusive(after)

	// Adding WithMax should conflict and replace WithMaxExclusive
	ruleSet2 := ruleSet.WithMax(max)

	var output internalTime.Duration

	// Original rule set should still have WithMaxExclusive
	err := ruleSet.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error for WithMaxExclusive at threshold (exclusive)")
	}

	// New rule set should have WithMax (inclusive, so max should pass)
	err = ruleSet2.Apply(context.TODO(), max, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// Verify serialization shows the conflict resolution
	expected := fmt.Sprintf("DurationRuleSet.WithMax(%s)", max)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
