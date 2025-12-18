package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMoreInt(t *testing.T) {
	ruleSet := rules.Int().WithMore(10).Any()

	// 9 is less than 10, should fail
	testhelpers.MustNotApply(t, ruleSet, 9, errors.CodeMin)

	// 10 is equal to 10, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, 10, errors.CodeMin)

	// 11 is greater than 10, should pass
	testhelpers.MustApply(t, ruleSet, 11)
}

func TestWithMoreFloat(t *testing.T) {
	ruleSet := rules.Float64().WithMore(10.0).Any()

	// 9.9 is less than 10.0, should fail
	testhelpers.MustNotApply(t, ruleSet, 9.9, errors.CodeMin)

	// 10.0 is equal to 10.0, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, 10.0, errors.CodeMin)

	// 10.1 is greater than 10.0, should pass
	testhelpers.MustApply(t, ruleSet, 10.1)
}

// Requirements:
// - Only one WithMore can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMore is used.
// - Rule is serialized properly.
func TestIntMoreConflict(t *testing.T) {
	ruleSet := rules.Int().WithMore(3).WithLess(10)

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
	ruleSet2 := ruleSet.WithMore(2)

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
	expected := "IntRuleSet[int].WithMore(3).WithLess(10)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated threshold
	expected = "IntRuleSet[int].WithLess(10).WithMore(2)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Only one WithMore can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMore is used.
// - Rule is serialized properly.
func TestFloatMoreConflict(t *testing.T) {
	ruleSet := rules.Float64().WithMore(3.0).WithLess(10.0)

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
	ruleSet2 := ruleSet.WithMore(2.0)

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
	expected := "FloatRuleSet[float64].WithMore(3.000000).WithLess(10.000000)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated threshold
	expected = "FloatRuleSet[float64].WithLess(10.000000).WithMore(2.000000)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMin and WithMore conflict with each other
func TestIntMinMoreConflict(t *testing.T) {
	ruleSet := rules.Int().WithMin(2)

	// Adding WithMore should conflict and replace WithMin
	ruleSet2 := ruleSet.WithMore(3)

	var output int

	// Original rule set should still have WithMin
	err := ruleSet.Apply(context.TODO(), 2, &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMin at threshold, got %s", err)
	}

	// New rule set should have WithMore (exclusive, so 3 should fail)
	err = ruleSet2.Apply(context.TODO(), 3, &output)
	if err == nil {
		t.Errorf("Expected error for WithMore at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := "IntRuleSet[int].WithMore(3)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMore and WithMin conflict with each other (reverse order)
func TestIntMoreMinConflict(t *testing.T) {
	ruleSet := rules.Int().WithMore(2)

	// Adding WithMin should conflict and replace WithMore
	ruleSet2 := ruleSet.WithMin(3)

	var output int

	// Original rule set should still have WithMore
	err := ruleSet.Apply(context.TODO(), 2, &output)
	if err == nil {
		t.Errorf("Expected error for WithMore at threshold (exclusive)")
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
