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

func TestWithMinDiff(t *testing.T) {
	now := internalTime.Now()
	before14 := now.Add(-14 * internalTime.Minute)
	before16 := now.Add(-16 * internalTime.Minute)

	ruleSet := time.NewTime().WithMinDiff(-15 * internalTime.Minute).Any()

	testhelpers.MustNotApply(t, ruleSet, before16, errors.CodeMin)
	testhelpers.MustApply(t, ruleSet, before14)
}

// Requirements:
// - Only one min diff can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
func TestWithMinDiffConflict(t *testing.T) {
	now := internalTime.Now().Add(1 * internalTime.Minute)
	before := now.Add(-10 * internalTime.Minute)

	// Create an initial rule set with min and max differences
	ruleSet := time.NewTime().WithMinDiff(0).WithMaxDiff(10 * internalTime.Minute)

	// Prepare an output variable for Apply
	var output internalTime.Time

	// Apply with a time before the min difference, expecting an error
	err := ruleSet.Apply(context.TODO(), before, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a time exactly at the max difference, expecting no error
	err = ruleSet.Apply(context.TODO(), now, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different min difference and validate
	ruleSet2 := ruleSet.WithMinDiff(-20 * internalTime.Minute)

	// Apply with a time within the new min difference, expecting no error
	err = ruleSet2.Apply(context.TODO(), before, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := fmt.Sprintf("TimeRuleSet.WithMinDiff(%s).WithMaxDiff(%s)", 0*internalTime.Minute, 10*internalTime.Minute)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = fmt.Sprintf("TimeRuleSet.WithMaxDiff(%s).WithMinDiff(%s)", 10*internalTime.Minute, -20*internalTime.Minute)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
