package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMinInt(t *testing.T) {
	ruleSet := rules.Int().WithMin(10).Any()

	testhelpers.MustNotApply(t, ruleSet, 9, errors.CodeMin)
	testhelpers.MustApply(t, ruleSet, 10)
	testhelpers.MustApply(t, ruleSet, 11)
}

func TestWithMinFloat(t *testing.T) {
	ruleSet := rules.Float64().WithMin(10.0).Any()

	testhelpers.MustNotApply(t, ruleSet, 9.9, errors.CodeMin)
	testhelpers.MustApply(t, ruleSet, 10.0)
	testhelpers.MustApply(t, ruleSet, 10.1)
}

// Requirements:
// - Only one min can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
// - Rule is serialized properly.
func TestIntMinConflict(t *testing.T) {
	ruleSet := rules.Int().WithMin(3).WithMax(10)

	var output int

	// Test validation with a value below the min (should return an error)
	err := ruleSet.Apply(context.TODO(), 2, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value at the min (should not return an error)
	err = ruleSet.Apply(context.TODO(), 3, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different min and test again
	ruleSet2 := ruleSet.WithMin(2)

	// Test validation with a value at the new min (should not return an error)
	err = ruleSet2.Apply(context.TODO(), 2, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "IntRuleSet[int].WithMin(3).WithMax(10)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated min
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
	ruleSet := rules.Float64().WithMin(3.0).WithMax(10.0)

	var output float64

	// Test validation with a value below the min (should return an error)
	err := ruleSet.Apply(context.TODO(), 2.0, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value at the min (should not return an error)
	err = ruleSet.Apply(context.TODO(), 3.0, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different min and test again
	ruleSet2 := ruleSet.WithMin(2.0)

	// Test validation with a value at the new min (should not return an error)
	err = ruleSet2.Apply(context.TODO(), 2.0, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "FloatRuleSet[float64].WithMin(3.000000).WithMax(10.000000)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated min
	expected = "FloatRuleSet[float64].WithMax(10.000000).WithMin(2.000000)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
