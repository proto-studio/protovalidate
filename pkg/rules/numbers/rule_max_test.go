package numbers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMaxInt(t *testing.T) {
	ruleSet := numbers.NewInt().WithMax(10).Any()

	testhelpers.MustRun(t, ruleSet, 9)
	testhelpers.MustRun(t, ruleSet, 10)
	testhelpers.MustNotRun(t, ruleSet, 11, errors.CodeMax)
}

func TestWithMaxFloat(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithMax(10.0).Any()

	testhelpers.MustRun(t, ruleSet, 9.9)
	testhelpers.MustRun(t, ruleSet, 10.0)
	testhelpers.MustNotRun(t, ruleSet, 10.1, errors.CodeMax)
}

// Requirements:
// - Only one max can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
// - Rule is serialized properly.
func TestIntMaxConflict(t *testing.T) {
	ruleSet := numbers.NewInt().WithMax(10).WithMin(3)

	if _, err := ruleSet.Validate(15); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if _, err := ruleSet.Validate(5); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	ruleSet2 := ruleSet.WithMax(20)
	if _, err := ruleSet2.Validate(15); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	expected := "IntRuleSet[int].WithMax(10).WithMin(3)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "IntRuleSet[int].WithMin(3).WithMax(20)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Only one max can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
// - Rule is serialized properly.
func TestFloatMaxConflict(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithMax(10.0).WithMin(3.0)

	if _, err := ruleSet.Validate(15.0); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if _, err := ruleSet.Validate(5.0); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	ruleSet2 := ruleSet.WithMax(20.0)
	if _, err := ruleSet2.Validate(15.0); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	expected := "FloatRuleSet[float64].WithMax(10.000000).WithMin(3.000000)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "FloatRuleSet[float64].WithMin(3.000000).WithMax(20.000000)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
