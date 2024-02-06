package numbers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMaxInt(t *testing.T) {
	ruleSet := numbers.NewInt().WithMax(10).Any()

	testhelpers.MustBeValid(t, ruleSet, 9, 9)
	testhelpers.MustBeValid(t, ruleSet, 10, 10)
	testhelpers.MustBeInvalid(t, ruleSet, 11, errors.CodeMax)
}

func TestWithMaxFloat(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithMax(10.0).Any()

	testhelpers.MustBeValid(t, ruleSet, 9.9, 9.9)
	testhelpers.MustBeValid(t, ruleSet, 10.0, 10.0)
	testhelpers.MustBeInvalid(t, ruleSet, 10.1, errors.CodeMax)
}

// Requirements:
// - Only one max can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
// - Rule is serialized properly.
func TestIntMaxConflict(t *testing.T) {
	ruleSet := numbers.NewInt().WithMax(10).WithMin(3)

	testhelpers.MustBeInvalid(t, ruleSet.Any(), 15, errors.CodeMax)
	testhelpers.MustBeValid(t, ruleSet.Any(), 5, 5)

	ruleSet2 := ruleSet.WithMax(20)
	testhelpers.MustBeValid(t, ruleSet2.Any(), 15, 15)

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

	testhelpers.MustBeInvalid(t, ruleSet.Any(), 15.0, errors.CodeMax)
	testhelpers.MustBeValid(t, ruleSet.Any(), 5.0, 5.0)

	ruleSet2 := ruleSet.WithMax(20.0)
	testhelpers.MustBeValid(t, ruleSet2.Any(), 15.0, 15.0)

	expected := "FloatRuleSet[float64].WithMax(10.000000).WithMin(3.000000)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "FloatRuleSet[float64].WithMin(3.000000).WithMax(20.000000)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
