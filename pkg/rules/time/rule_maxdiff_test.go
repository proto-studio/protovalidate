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

func TestWithMaxDiff(t *testing.T) {
	now := internalTime.Now()
	before14 := now.Add(-14 * internalTime.Minute)
	before16 := now.Add(-16 * internalTime.Minute)

	ruleSet := time.NewTime().WithMaxDiff(-15 * internalTime.Minute).Any()

	testhelpers.MustNotApply(t, ruleSet, before14, errors.CodeMax)
	testhelpers.MustApply(t, ruleSet, before16)
}

// Requirements:
// - Only one max diff can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
func TestWithMaxDiffConflict(t *testing.T) {
	now := internalTime.Now().Add(1 * internalTime.Minute)
	after := now.Add(10 * internalTime.Minute)

	// Create an initial rule set with max and min differences
	ruleSet := time.NewTime().WithMaxDiff(10 * internalTime.Minute).WithMinDiff(0 * internalTime.Minute)

	// Prepare an output variable for Apply
	var output internalTime.Time

	// Apply with a time after the max difference, expecting an error
	err := ruleSet.Apply(context.TODO(), after, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with a time exactly at the max difference, expecting no error
	err = ruleSet.Apply(context.TODO(), now, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different max difference and validate
	ruleSet2 := ruleSet.WithMaxDiff(20 * internalTime.Minute)

	// Apply with a time within the new max difference, expecting no error
	err = ruleSet2.Apply(context.TODO(), after, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := fmt.Sprintf("TimeRuleSet.WithMaxDiff(%s).WithMinDiff(%s)", 10*internalTime.Minute, 0*internalTime.Minute)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = fmt.Sprintf("TimeRuleSet.WithMinDiff(%s).WithMaxDiff(%s)", 0*internalTime.Minute, 20*internalTime.Minute)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
