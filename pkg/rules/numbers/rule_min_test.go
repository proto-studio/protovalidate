package numbers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMinInt(t *testing.T) {
	ruleSet := numbers.NewInt().WithMin(10).Any()

	testhelpers.MustBeInvalid(t, ruleSet, 9, errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet, 10, 10)
	testhelpers.MustBeValid(t, ruleSet, 11, 11)
}

func TestWithMinFloat(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithMin(10.0).Any()

	testhelpers.MustBeInvalid(t, ruleSet, 9.9, errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet, 10.0, 10.0)
	testhelpers.MustBeValid(t, ruleSet, 10.1, 10.1)
}

// Requirements:
// - Only one min can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
// - Rule is serialized properly.
func TestIntMinConflict(t *testing.T) {
	ruleSet := numbers.NewInt().WithMin(3).WithMax(10)

	testhelpers.MustBeInvalid(t, ruleSet.Any(), 2, errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet.Any(), 3, 3)

	ruleSet2 := ruleSet.WithMin(2)
	testhelpers.MustBeValid(t, ruleSet2.Any(), 2, 2)

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

	testhelpers.MustBeInvalid(t, ruleSet.Any(), 2.0, errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet.Any(), 3.0, 3.0)

	ruleSet2 := ruleSet.WithMin(2.0)
	testhelpers.MustBeValid(t, ruleSet2.Any(), 2.0, 2.0)

	expected := "FloatRuleSet[float64].WithMin(3.000000).WithMax(10.000000)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "FloatRuleSet[float64].WithMax(10.000000).WithMin(2.000000)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
