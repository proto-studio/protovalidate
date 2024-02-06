package numbers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithRangeInt(t *testing.T) {
	ruleSet := numbers.NewInt().WithRange(10, 15).Any()

	testhelpers.MustBeInvalid(t, ruleSet, 9, errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet, 10, 10)
	testhelpers.MustBeValid(t, ruleSet, 15, 15)
	testhelpers.MustBeInvalid(t, ruleSet, 16, errors.CodeMax)
}

func TestWithRangeFloat(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRange(10.0, 15.0).Any()

	testhelpers.MustBeInvalid(t, ruleSet, 9.9, errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet, 10.0, 10.0)
	testhelpers.MustBeValid(t, ruleSet, 15.0, 15.0)
	testhelpers.MustBeInvalid(t, ruleSet, 15.1, errors.CodeMax)
}

// Requirements:
// - Only one min/max can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum and maximum is used.
// - Rule is serialized properly (only contains the most recent min/max).
func TestIntRangeConflict(t *testing.T) {
	ruleSet := numbers.NewInt().WithRange(3, 10)

	testhelpers.MustBeValid(t, ruleSet.Any(), 3, 3)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), 2, errors.CodeMin)

	testhelpers.MustBeValid(t, ruleSet.Any(), 10, 10)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), 11, errors.CodeMax)

	ruleSet2 := ruleSet.WithMin(2)
	testhelpers.MustBeValid(t, ruleSet2.Any(), 2, 2)

	ruleSet3 := ruleSet.WithMax(11)
	testhelpers.MustBeValid(t, ruleSet3.Any(), 11, 11)

	expected := "IntRuleSet[int].WithMin(3).WithMax(10)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "IntRuleSet[int].WithMax(10).WithMin(2)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "IntRuleSet[int].WithMin(3).WithMax(11)"
	if s := ruleSet3.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Only one min/max can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum and maximum is used.
// - Rule is serialized properly (only contains the most recent min/max).
func TestFloatRangeConflict(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithMin(3.0).WithMax(10.0)

	testhelpers.MustBeValid(t, ruleSet.Any(), 3.0, 3.0)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), 2.0, errors.CodeMin)

	testhelpers.MustBeValid(t, ruleSet.Any(), 10.0, 10.0)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), 11.0, errors.CodeMax)

	ruleSet2 := ruleSet.WithMin(2)
	testhelpers.MustBeValid(t, ruleSet2.Any(), 2.0, 2.0)

	ruleSet3 := ruleSet.WithMax(11)
	testhelpers.MustBeValid(t, ruleSet3.Any(), 11.0, 11.0)

	expected := "FloatRuleSet[float64].WithMin(3.000000).WithMax(10.000000)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "FloatRuleSet[float64].WithMax(10.000000).WithMin(2.000000)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "FloatRuleSet[float64].WithMin(3.000000).WithMax(11.000000)"
	if s := ruleSet3.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Int should panic when min is not less than max
func TestPanicIntMin(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	numbers.NewInt().WithRange(10, 10)
}

// Requirements:
// - Float should panic when min is not less than max
func TestPanicFloatMin(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	numbers.NewFloat64().WithRange(10.0, 10.0)
}
