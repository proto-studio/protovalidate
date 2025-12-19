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

// TestTimeRuleSet_WithMin tests:
// - Minimum time validation works correctly
func TestTimeRuleSet_WithMin(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.Time().WithMin(now).Any()

	testhelpers.MustNotApply(t, ruleSet, before, errors.CodeMin)

	testhelpers.MustApply(t, ruleSet, now)
	testhelpers.MustApply(t, ruleSet, after)
}

// TestTimeRuleSet_WithMin_Conflict tests:
// - Only one min length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
func TestTimeRuleSet_WithMin_Conflict(t *testing.T) {
	tm, _ := internalTime.Parse(internalTime.RFC3339, "2023-10-05T00:12:12.927Z")
	before := tm.Add(-1 * internalTime.Minute)
	after := tm.Add(1 * internalTime.Minute)

	// Create an initial rule set with min and max values
	ruleSet := time.Time().WithMin(tm).WithMax(after)

	// Prepare an output variable for Apply
	var output internalTime.Time

	// Apply with a time before the min, expecting an error
	err := ruleSet.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a time exactly at the min, expecting no error
	err = ruleSet.Apply(context.TODO(), tm, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different min value and validate
	ruleSet2 := ruleSet.WithMin(before)

	// Apply with a time exactly at the new min, expecting no error
	err = ruleSet2.Apply(context.TODO(), before, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := fmt.Sprintf("TimeRuleSet.WithMin(%s).WithMax(%s)", tm, after)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = fmt.Sprintf("TimeRuleSet.WithMax(%s).WithMin(%s)", after, before)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
