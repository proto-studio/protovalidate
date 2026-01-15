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

// TestDurationRuleSet_WithMin tests:
// - Minimum duration validation works correctly
func TestDurationRuleSet_WithMin(t *testing.T) {
	min := 1 * internalTime.Hour
	before := 30 * internalTime.Minute
	after := 2 * internalTime.Hour

	ruleSet := time.Duration().WithMin(min).Any()

	testhelpers.MustNotApply(t, ruleSet, before, errors.CodeMin)

	testhelpers.MustApply(t, ruleSet, min)
	testhelpers.MustApply(t, ruleSet, after)
}

// TestDurationRuleSet_WithMin_Conflict tests:
// - Only one min duration can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
func TestDurationRuleSet_WithMin_Conflict(t *testing.T) {
	min := 1 * internalTime.Hour
	before := 30 * internalTime.Minute
	after := 2 * internalTime.Hour

	// Create an initial rule set with min and max values
	ruleSet := time.Duration().WithMin(min).WithMax(after)

	// Prepare an output variable for Apply
	var output internalTime.Duration

	// Apply with a duration before the min, expecting an error
	err := ruleSet.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a duration exactly at the min, expecting no error
	err = ruleSet.Apply(context.TODO(), min, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different min value and validate
	ruleSet2 := ruleSet.WithMin(before)

	// Apply with a duration exactly at the new min, expecting no error
	err = ruleSet2.Apply(context.TODO(), before, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := fmt.Sprintf("DurationRuleSet.WithMin(%s).WithMax(%s)", min, after)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = fmt.Sprintf("DurationRuleSet.WithMax(%s).WithMin(%s)", after, before)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMin and WithMinExclusive conflict with each other
func TestDurationRuleSet_WithMin_WithMinExclusiveConflict(t *testing.T) {
	min := 1 * internalTime.Hour
	before := 30 * internalTime.Minute

	ruleSet := time.Duration().WithMin(min)

	// Adding WithMinExclusive should conflict and replace WithMin
	ruleSet2 := ruleSet.WithMinExclusive(before)

	var output internalTime.Duration

	// Original rule set should still have WithMin
	err := ruleSet.Apply(context.TODO(), min, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMin at threshold, got %s", err)
	}

	// New rule set should have WithMinExclusive (exclusive, so before should fail)
	err = ruleSet2.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error for WithMinExclusive at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := fmt.Sprintf("DurationRuleSet.WithMinExclusive(%s)", before)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMinExclusive and WithMin conflict with each other (reverse order)
func TestDurationRuleSet_WithMinExclusive_WithMinConflict(t *testing.T) {
	min := 1 * internalTime.Hour
	before := 30 * internalTime.Minute

	ruleSet := time.Duration().WithMinExclusive(before)

	// Adding WithMin should conflict and replace WithMinExclusive
	ruleSet2 := ruleSet.WithMin(min)

	var output internalTime.Duration

	// Original rule set should still have WithMinExclusive
	err := ruleSet.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error for WithMinExclusive at threshold (exclusive)")
	}

	// New rule set should have WithMin (inclusive, so min should pass)
	err = ruleSet2.Apply(context.TODO(), min, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMin at threshold, got %s", err)
	}

	// Verify serialization shows the conflict resolution
	expected := fmt.Sprintf("DurationRuleSet.WithMin(%s)", min)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
