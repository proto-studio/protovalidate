package numbers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMinInt(t *testing.T) {
	ruleSet := numbers.NewInt().WithMin(10).Any()

	testhelpers.MustNotRun(t, ruleSet, 9, errors.CodeMin)
	testhelpers.MustRun(t, ruleSet, 10)
	testhelpers.MustRun(t, ruleSet, 11)
}

func TestWithMinFloat(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithMin(10.0).Any()

	testhelpers.MustNotRun(t, ruleSet, 9.9, errors.CodeMin)
	testhelpers.MustRun(t, ruleSet, 10.0)
	testhelpers.MustRun(t, ruleSet, 10.1)
}

// Requirements:
// - Only one min can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
// - Rule is serialized properly.
func TestIntMinConflict(t *testing.T) {
	ruleSet := numbers.NewInt().WithMin(3).WithMax(10)

	if _, err := ruleSet.Validate(2); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if _, err := ruleSet.Validate(3); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	ruleSet2 := ruleSet.WithMin(2)
	if _, err := ruleSet2.Validate(2); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	expected := "IntRuleSet[int].WithMin(3).WithMax(10)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "IntRuleSet[int].WithMax(10).WithMin(2)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Only one min can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
// - Rule is serialized properly.
func TestFloatMinConflict(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithMin(3.0).WithMax(10.0)

	if _, err := ruleSet.Validate(2.0); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if _, err := ruleSet.Validate(3.0); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	ruleSet2 := ruleSet.WithMin(2.0)
	if _, err := ruleSet2.Validate(2.0); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	expected := "FloatRuleSet[float64].WithMin(3.000000).WithMax(10.000000)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "FloatRuleSet[float64].WithMax(10.000000).WithMin(2.000000)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
