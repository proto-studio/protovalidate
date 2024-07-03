package time_test

import (
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

func TestWithMaxTimeString(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.NewTimeString(internalTime.RFC3339).WithMax(now).Any()

	testhelpers.MustRunMutation(t, ruleSet, before, before.Format(internalTime.RFC3339))
	testhelpers.MustRunMutation(t, ruleSet, now, now.Format(internalTime.RFC3339))
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

	ruleSet := time.NewTime().WithMax(tm).WithMin(before)

	if _, err := ruleSet.Validate(after); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if _, err := ruleSet.Validate(tm); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	ruleSet2 := ruleSet.WithMax(after)
	if _, err := ruleSet2.Validate(after); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	expected := fmt.Sprintf("TimeRuleSet.WithMax(%s).WithMin(%s)", tm, before)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = fmt.Sprintf("TimeRuleSet.WithMin(%s).WithMax(%s)", before, after)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
