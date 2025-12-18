package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithLessInt(t *testing.T) {
	ruleSet := rules.Int().WithLess(10).Any()

	// 9 is less than 10, should pass
	testhelpers.MustApply(t, ruleSet, 9)

	// 10 is equal to 10, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, 10, errors.CodeMax)

	// 11 is greater than 10, should fail
	testhelpers.MustNotApply(t, ruleSet, 11, errors.CodeMax)
}

func TestWithLessFloat(t *testing.T) {
	ruleSet := rules.Float64().WithLess(10.0).Any()

	// 9.9 is less than 10.0, should pass
	testhelpers.MustApply(t, ruleSet, 9.9)

	// 10.0 is equal to 10.0, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, 10.0, errors.CodeMax)

	// 10.1 is greater than 10.0, should fail
	testhelpers.MustNotApply(t, ruleSet, 10.1, errors.CodeMax)
}

// Requirements:
// - Only one WithLess can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithLess is used.
// - Rule is serialized properly.
func TestIntLessConflict(t *testing.T) {
	ruleSet := rules.Int().WithLess(10).WithMore(3)

	var output int

	// Test validation with a value equal to the threshold (should return an error - exclusive)
	err := ruleSet.Apply(context.TODO(), 10, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the threshold (should return an error)
	err = ruleSet.Apply(context.TODO(), 11, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the threshold (should not return an error)
	err = ruleSet.Apply(context.TODO(), 9, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different threshold and test again
	ruleSet2 := ruleSet.WithLess(9)

	// Test validation with a value at the new threshold (should return an error - exclusive)
	err = ruleSet2.Apply(context.TODO(), 9, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the new threshold (should not return an error)
	err = ruleSet2.Apply(context.TODO(), 8, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "IntRuleSet[int].WithLess(10).WithMore(3)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated threshold
	expected = "IntRuleSet[int].WithMore(3).WithLess(9)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Only one WithLess can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithLess is used.
// - Rule is serialized properly.
func TestFloatLessConflict(t *testing.T) {
	ruleSet := rules.Float64().WithLess(10.0).WithMore(3.0)

	var output float64

	// Test validation with a value equal to the threshold (should return an error - exclusive)
	err := ruleSet.Apply(context.TODO(), 10.0, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the threshold (should return an error)
	err = ruleSet.Apply(context.TODO(), 11.0, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the threshold (should not return an error)
	err = ruleSet.Apply(context.TODO(), 9.0, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different threshold and test again
	ruleSet2 := ruleSet.WithLess(9.0)

	// Test validation with a value at the new threshold (should return an error - exclusive)
	err = ruleSet2.Apply(context.TODO(), 9.0, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the new threshold (should not return an error)
	err = ruleSet2.Apply(context.TODO(), 8.0, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "FloatRuleSet[float64].WithLess(10.000000).WithMore(3.000000)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated threshold
	expected = "FloatRuleSet[float64].WithMore(3.000000).WithLess(9.000000)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMax and WithLess conflict with each other
func TestIntMaxLessConflict(t *testing.T) {
	ruleSet := rules.Int().WithMax(10)

	// Adding WithLess should conflict and replace WithMax
	ruleSet2 := ruleSet.WithLess(9)

	var output int

	// Original rule set should still have WithMax
	err := ruleSet.Apply(context.TODO(), 10, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// New rule set should have WithLess (exclusive, so 9 should fail)
	err = ruleSet2.Apply(context.TODO(), 9, &output)
	if err == nil {
		t.Errorf("Expected error for WithLess at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := "IntRuleSet[int].WithLess(9)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithLess and WithMax conflict with each other (reverse order)
func TestIntLessMaxConflict(t *testing.T) {
	ruleSet := rules.Int().WithLess(10)

	// Adding WithMax should conflict and replace WithLess
	ruleSet2 := ruleSet.WithMax(9)

	var output int

	// Original rule set should still have WithLess
	err := ruleSet.Apply(context.TODO(), 10, &output)
	if err == nil {
		t.Errorf("Expected error for WithLess at threshold (exclusive)")
	}

	// New rule set should have WithMax (inclusive, so 9 should pass)
	err = ruleSet2.Apply(context.TODO(), 9, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// Verify serialization shows the conflict resolution
	expected := "IntRuleSet[int].WithMax(9)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
