package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestIntRuleSet_WithMinExclusive tests:
func TestIntRuleSet_WithMinExclusive(t *testing.T) {
	ruleSet := rules.Int().WithMinExclusive(10).Any()

	// 9 is less than 10, should fail
	testhelpers.MustNotApply(t, ruleSet, 9, errors.CodeMin)

	// 10 is equal to 10, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, 10, errors.CodeMin)

	// 11 is greater than 10, should pass
	testhelpers.MustApply(t, ruleSet, 11)
}

// TestFloatRuleSet_WithMinExclusive tests:
func TestFloatRuleSet_WithMinExclusive(t *testing.T) {
	ruleSet := rules.Float64().WithMinExclusive(10.0).Any()

	// 9.9 is less than 10.0, should fail
	testhelpers.MustNotApply(t, ruleSet, 9.9, errors.CodeMin)

	// 10.0 is equal to 10.0, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, 10.0, errors.CodeMin)

	// 10.1 is greater than 10.0, should pass
	testhelpers.MustApply(t, ruleSet, 10.1)
}

// TestIntRuleSet_WithMinExclusive_Conflict tests:
// - Only one WithMinExclusive can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMinExclusive is used.
// - Rule is serialized properly.
func TestIntRuleSet_WithMinExclusive_Conflict(t *testing.T) {
	ruleSet := rules.Int().WithMinExclusive(3).WithMaxExclusive(10)

	var output int

	// Test validation with a value equal to the threshold (should return an error - exclusive)
	err := ruleSet.Apply(context.TODO(), 3, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the threshold (should return an error)
	err = ruleSet.Apply(context.TODO(), 2, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the threshold (should not return an error)
	err = ruleSet.Apply(context.TODO(), 4, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different threshold and test again
	ruleSet2 := ruleSet.WithMinExclusive(2)

	// Test validation with a value at the new threshold (should return an error - exclusive)
	err = ruleSet2.Apply(context.TODO(), 2, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the new threshold (should not return an error)
	err = ruleSet2.Apply(context.TODO(), 3, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "IntRuleSet[int].WithMinExclusive(3).WithMaxExclusive(10)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated threshold
	expected = "IntRuleSet[int].WithMaxExclusive(10).WithMinExclusive(2)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestFloatRuleSet_WithMinExclusive_Conflict tests:
// - Only one WithMinExclusive can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMinExclusive is used.
// - Rule is serialized properly.
func TestFloatRuleSet_WithMinExclusive_Conflict(t *testing.T) {
	ruleSet := rules.Float64().WithMinExclusive(3.0).WithMaxExclusive(10.0)

	var output float64

	// Test validation with a value equal to the threshold (should return an error - exclusive)
	err := ruleSet.Apply(context.TODO(), 3.0, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the threshold (should return an error)
	err = ruleSet.Apply(context.TODO(), 2.0, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the threshold (should not return an error)
	err = ruleSet.Apply(context.TODO(), 4.0, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different threshold and test again
	ruleSet2 := ruleSet.WithMinExclusive(2.0)

	// Test validation with a value at the new threshold (should return an error - exclusive)
	err = ruleSet2.Apply(context.TODO(), 2.0, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the new threshold (should not return an error)
	err = ruleSet2.Apply(context.TODO(), 3.0, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "FloatRuleSet[float64].WithMinExclusive(3.000000).WithMaxExclusive(10.000000)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated threshold
	expected = "FloatRuleSet[float64].WithMaxExclusive(10.000000).WithMinExclusive(2.000000)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMin and WithMinExclusive conflict with each other
func TestIntRuleSet_WithMin_WithMinExclusiveConflict(t *testing.T) {
	ruleSet := rules.Int().WithMin(2)

	// Adding WithMinExclusive should conflict and replace WithMin
	ruleSet2 := ruleSet.WithMinExclusive(3)

	var output int

	// Original rule set should still have WithMin
	err := ruleSet.Apply(context.TODO(), 2, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMin at threshold, got %s", err)
	}

	// New rule set should have WithMinExclusive (exclusive, so 3 should fail)
	err = ruleSet2.Apply(context.TODO(), 3, &output)
	if err == nil {
		t.Errorf("Expected error for WithMinExclusive at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := "IntRuleSet[int].WithMinExclusive(3)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMinExclusive and WithMin conflict with each other (reverse order)
func TestIntRuleSet_WithMinExclusive_WithMinConflict(t *testing.T) {
	ruleSet := rules.Int().WithMinExclusive(2)

	// Adding WithMin should conflict and replace WithMinExclusive
	ruleSet2 := ruleSet.WithMin(3)

	var output int

	// Original rule set should still have WithMinExclusive
	err := ruleSet.Apply(context.TODO(), 2, &output)
	if err == nil {
		t.Errorf("Expected error for WithMinExclusive at threshold (exclusive)")
	}

	// New rule set should have WithMin (inclusive, so 3 should pass)
	err = ruleSet2.Apply(context.TODO(), 3, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMin at threshold, got %s", err)
	}

	// Verify serialization shows the conflict resolution
	expected := "IntRuleSet[int].WithMin(3)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
