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

func TestWithMaxTime(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.NewTime().WithMax(now).Any()

	testhelpers.MustRun(t, ruleSet, before)
	testhelpers.MustRun(t, ruleSet, now)
	testhelpers.MustNotRun(t, ruleSet, after, errors.CodeMax)
}

// Requirements:
// - Only one max length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
func TestWithMaxConflict(t *testing.T) {
	tm, _ := internalTime.Parse(internalTime.RFC3339, "2023-10-05T00:12:12.927Z")
	before := tm.Add(-1 * internalTime.Minute)
	after := tm.Add(1 * internalTime.Minute)

	// Create an initial rule set with a max and min
	ruleSet := time.NewTime().WithMax(tm).WithMin(before)

	// Prepare an output variable for Apply
	var output internalTime.Time

	// Apply with a time after the max, expecting an error
	err := ruleSet.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a time exactly at the max, expecting no error
	err = ruleSet.Apply(context.TODO(), tm, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different max and validate
	ruleSet2 := ruleSet.WithMax(after)

	// Apply with a time exactly at the new max, expecting no error
	err = ruleSet2.Apply(context.TODO(), after, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := fmt.Sprintf("TimeRuleSet.WithMax(%s).WithMin(%s)", tm, before)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = fmt.Sprintf("TimeRuleSet.WithMin(%s).WithMax(%s)", before, after)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
