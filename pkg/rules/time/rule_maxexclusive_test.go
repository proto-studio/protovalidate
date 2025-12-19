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

func TestTimeRuleSet_WithMaxExclusive(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.Time().WithMaxExclusive(now).Any()

	// before is before now, should pass
	testhelpers.MustApply(t, ruleSet, before)

	// now is equal to now, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, now, errors.CodeMax)

	// after is after now, should fail
	testhelpers.MustNotApply(t, ruleSet, after, errors.CodeMax)
}

// Requirements:
// - Only one WithMaxExclusive can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMaxExclusive is used.
func TestTimeRuleSet_WithMaxExclusive_Conflict(t *testing.T) {
	tm, _ := internalTime.Parse(internalTime.RFC3339, "2023-10-05T00:12:12.927Z")
	before := tm.Add(-1 * internalTime.Minute)
	after := tm.Add(1 * internalTime.Minute)

	// Create an initial rule set with before and after values
	ruleSet := time.Time().WithMaxExclusive(tm).WithMinExclusive(before)

	// Prepare an output variable for Apply
	var output internalTime.Time

	// Apply with a time equal to the threshold, expecting an error (exclusive)
	err := ruleSet.Apply(context.TODO(), tm, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a time after the threshold, expecting an error
	err = ruleSet.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a time before the threshold, expecting no error
	err = ruleSet.Apply(context.TODO(), before.Add(30*internalTime.Second), &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different before value and validate
	ruleSet2 := ruleSet.WithMaxExclusive(after)

	// Apply with a time exactly at the new threshold, expecting an error (exclusive)
	err = ruleSet2.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a time before the new threshold, expecting no error
	err = ruleSet2.Apply(context.TODO(), tm, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := fmt.Sprintf("TimeRuleSet.WithMaxExclusive(%s).WithMinExclusive(%s)", tm, before)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = fmt.Sprintf("TimeRuleSet.WithMinExclusive(%s).WithMaxExclusive(%s)", before, after)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMax and WithMaxExclusive conflict with each other
func TestTimeRuleSet_WithMax_WithMaxExclusiveConflict(t *testing.T) {
	tm, _ := internalTime.Parse(internalTime.RFC3339, "2023-10-05T00:12:12.927Z")
	after := tm.Add(1 * internalTime.Minute)

	ruleSet := time.Time().WithMax(tm)

	// Adding WithMaxExclusive should conflict and replace WithMax
	ruleSet2 := ruleSet.WithMaxExclusive(after)

	var output internalTime.Time

	// Original rule set should still have WithMax
	err := ruleSet.Apply(context.TODO(), tm, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// New rule set should have WithMaxExclusive (exclusive, so after should fail)
	err = ruleSet2.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error for WithMaxExclusive at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := fmt.Sprintf("TimeRuleSet.WithMaxExclusive(%s)", after)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMaxExclusive and WithMax conflict with each other (reverse order)
func TestTimeRuleSet_WithMaxExclusive_WithMaxConflict(t *testing.T) {
	tm, _ := internalTime.Parse(internalTime.RFC3339, "2023-10-05T00:12:12.927Z")
	after := tm.Add(1 * internalTime.Minute)

	ruleSet := time.Time().WithMaxExclusive(after)

	// Adding WithMax should conflict and replace WithMaxExclusive
	ruleSet2 := ruleSet.WithMax(tm)

	var output internalTime.Time

	// Original rule set should still have WithMaxExclusive
	err := ruleSet.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error for WithMaxExclusive at threshold (exclusive)")
	}

	// New rule set should have WithMax (inclusive, so tm should pass)
	err = ruleSet2.Apply(context.TODO(), tm, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// Verify serialization shows the conflict resolution
	expected := fmt.Sprintf("TimeRuleSet.WithMax(%s)", tm)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
