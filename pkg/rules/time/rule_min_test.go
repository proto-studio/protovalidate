package time_test

import (
	"fmt"
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/time"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMinTime(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.NewTime().WithMin(now).Any()

	testhelpers.MustNotRun(t, ruleSet, before, errors.CodeMin)

	testhelpers.MustRun(t, ruleSet, now)
	testhelpers.MustRun(t, ruleSet, after)
}

func TestWithMinTimeString(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.NewTimeString(internalTime.RFC3339).WithMin(now).Any()

	testhelpers.MustNotRun(t, ruleSet, before, errors.CodeMin)

	testhelpers.MustRunMutation(t, ruleSet, now, now.Format(internalTime.RFC3339))
	testhelpers.MustRunMutation(t, ruleSet, after, after.Format(internalTime.RFC3339))
}

// Requirements:
// - Only one min length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
func TestWithMinConflict(t *testing.T) {
	tm, _ := internalTime.Parse(internalTime.RFC3339, "2023-10-05T00:12:12.927Z")
	before := tm.Add(-1 * internalTime.Minute)
	after := tm.Add(1 * internalTime.Minute)

	ruleSet := time.NewTime().WithMin(tm).WithMax(after)

	if _, err := ruleSet.Validate(before); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if _, err := ruleSet.Validate(tm); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	ruleSet2 := ruleSet.WithMin(before)
	if _, err := ruleSet2.Validate(before); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	expected := fmt.Sprintf("TimeRuleSet.WithMin(%s).WithMax(%s)", tm, after)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = fmt.Sprintf("TimeRuleSet.WithMax(%s).WithMin(%s)", after, before)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
