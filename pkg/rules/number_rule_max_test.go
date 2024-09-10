package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMaxInt(t *testing.T) {
	ruleSet := rules.NewInt().WithMax(10).Any()

	testhelpers.MustApply(t, ruleSet, 9)
	testhelpers.MustApply(t, ruleSet, 10)
	testhelpers.MustNotApply(t, ruleSet, 11, errors.CodeMax)
}

func TestWithMaxFloat(t *testing.T) {
	ruleSet := rules.NewFloat64().WithMax(10.0).Any()

	testhelpers.MustApply(t, ruleSet, 9.9)
	testhelpers.MustApply(t, ruleSet, 10.0)
	testhelpers.MustNotApply(t, ruleSet, 10.1, errors.CodeMax)
}

// Requirements:
// - Only one max can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
// - Rule is serialized properly.
func TestIntMaxConflict(t *testing.T) {
	ruleSet := rules.NewInt().WithMax(10).WithMin(3)

	var output int

	// Test validation with a value that exceeds the max (should return an error)
	err := ruleSet.Apply(context.TODO(), 15, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value within the range (should not return an error)
	err = ruleSet.Apply(context.TODO(), 5, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different max and test again
	ruleSet2 := ruleSet.WithMax(20)

	// Test validation with a value that is within the new max (should not return an error)
	err = ruleSet2.Apply(context.TODO(), 15, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "IntRuleSet[int].WithMax(10).WithMin(3)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated max
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
	ruleSet := rules.NewFloat64().WithMax(10.0).WithMin(3.0)

	var output float64

	// Test validation with a value that exceeds the max (should return an error)
	err := ruleSet.Apply(context.TODO(), 15.0, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value within the range (should not return an error)
	err = ruleSet.Apply(context.TODO(), 5.0, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different max and test again
	ruleSet2 := ruleSet.WithMax(20.0)

	// Test validation with a value that is within the new max (should not return an error)
	err = ruleSet2.Apply(context.TODO(), 15.0, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "FloatRuleSet[float64].WithMax(10.000000).WithMin(3.000000)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated max
	expected = "FloatRuleSet[float64].WithMin(3.000000).WithMax(20.000000)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
