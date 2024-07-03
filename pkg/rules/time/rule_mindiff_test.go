package time_test

import (
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

	testhelpers.MustNotRun(t, ruleSet, before16, errors.CodeMin)
	testhelpers.MustRun(t, ruleSet, before14)
}

func TestStringWithMinDiff(t *testing.T) {
	now := internalTime.Now()
	before14 := now.Add(-14 * internalTime.Minute)
	before16 := now.Add(-16 * internalTime.Minute)

	ruleSet := time.NewTimeString(internalTime.RFC3339).WithMinDiff(-15 * internalTime.Minute).Any()

	testhelpers.MustNotRun(t, ruleSet, before16, errors.CodeMin)
	testhelpers.MustRunMutation(t, ruleSet, before14, before14.Format(internalTime.RFC3339))
}

// Requirements:
// - Only one min diff can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
func TestWithMinDiffConflict(t *testing.T) {
	now := internalTime.Now().Add(1 * internalTime.Minute)
	before := now.Add(-10 * internalTime.Minute)

	ruleSet := time.NewTime().WithMinDiff(0).WithMaxDiff(10 * internalTime.Minute)

	if _, err := ruleSet.Validate(before); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if _, err := ruleSet.Validate(now); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	ruleSet2 := ruleSet.WithMinDiff(-20 * internalTime.Minute)
	if _, err := ruleSet2.Validate(before); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	expected := fmt.Sprintf("TimeRuleSet.WithMinDiff(%s).WithMaxDiff(%s)", 0*internalTime.Minute, 10*internalTime.Minute)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = fmt.Sprintf("TimeRuleSet.WithMaxDiff(%s).WithMinDiff(%s)", 10*internalTime.Minute, -20*internalTime.Minute)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
