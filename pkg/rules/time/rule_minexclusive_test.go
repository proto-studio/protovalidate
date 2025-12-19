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

func TestTimeRuleSet_WithMinExclusive(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.Time().WithMinExclusive(now).Any()

	// before is before now, should fail
	testhelpers.MustNotApply(t, ruleSet, before, errors.CodeMin)

	// now is equal to now, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, now, errors.CodeMin)

	// after is after now, should pass
	testhelpers.MustApply(t, ruleSet, after)
}

// Requirements:
// - Only one WithMinExclusive can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMinExclusive is used.
func TestTimeRuleSet_WithMinExclusive_Conflict(t *testing.T) {
	tm, _ := internalTime.Parse(internalTime.RFC3339, "2023-10-05T00:12:12.927Z")
	before := tm.Add(-1 * internalTime.Minute)
	after := tm.Add(1 * internalTime.Minute)

	// Create an initial rule set with after and before values
	ruleSet := time.Time().WithMinExclusive(tm).WithMaxExclusive(after)

	// Prepare an output variable for Apply
	var output internalTime.Time

	// Apply with a time equal to the threshold, expecting an error (exclusive)
	err := ruleSet.Apply(context.TODO(), tm, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a time before the threshold, expecting an error
	err = ruleSet.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a time after the threshold, expecting no error
	err = ruleSet.Apply(context.TODO(), after.Add(-30*internalTime.Second), &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different after value and validate
	ruleSet2 := ruleSet.WithMinExclusive(before)

	// Apply with a time exactly at the new threshold, expecting an error (exclusive)
	err = ruleSet2.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a time after the new threshold, expecting no error
	err = ruleSet2.Apply(context.TODO(), tm, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := fmt.Sprintf("TimeRuleSet.WithMinExclusive(%s).WithMaxExclusive(%s)", tm, after)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = fmt.Sprintf("TimeRuleSet.WithMaxExclusive(%s).WithMinExclusive(%s)", after, before)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMin and WithMinExclusive conflict with each other
func TestTimeRuleSet_WithMin_WithMinExclusiveConflict(t *testing.T) {
	tm, _ := internalTime.Parse(internalTime.RFC3339, "2023-10-05T00:12:12.927Z")
	before := tm.Add(-1 * internalTime.Minute)

	ruleSet := time.Time().WithMin(tm)

	// Adding WithMinExclusive should conflict and replace WithMin
	ruleSet2 := ruleSet.WithMinExclusive(before)

	var output internalTime.Time

	// Original rule set should still have WithMin
	err := ruleSet.Apply(context.TODO(), tm, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMin at threshold, got %s", err)
	}

	// New rule set should have WithMinExclusive (exclusive, so before should fail)
	err = ruleSet2.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error for WithMinExclusive at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := fmt.Sprintf("TimeRuleSet.WithMinExclusive(%s)", before)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMinExclusive and WithMin conflict with each other (reverse order)
func TestTimeRuleSet_WithMinExclusive_WithMinConflict(t *testing.T) {
	tm, _ := internalTime.Parse(internalTime.RFC3339, "2023-10-05T00:12:12.927Z")
	before := tm.Add(-1 * internalTime.Minute)

	ruleSet := time.Time().WithMinExclusive(before)

	// Adding WithMin should conflict and replace WithMinExclusive
	ruleSet2 := ruleSet.WithMin(tm)

	var output internalTime.Time

	// Original rule set should still have WithMinExclusive
	err := ruleSet.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error for WithMinExclusive at threshold (exclusive)")
	}

	// New rule set should have WithMin (inclusive, so tm should pass)
	err = ruleSet2.Apply(context.TODO(), tm, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMin at threshold, got %s", err)
	}

	// Verify serialization shows the conflict resolution
	expected := fmt.Sprintf("TimeRuleSet.WithMin(%s)", tm)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
