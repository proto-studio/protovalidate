package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestIntRuleSet_WithMaxExclusive tests:
// - Integers less than maximum pass validation
// - Integers equal to maximum fail validation (exclusive)
// - Integers greater than maximum fail validation
func TestIntRuleSet_WithMaxExclusive(t *testing.T) {
	ruleSet := rules.Int().WithMaxExclusive(10).Any()

	// 9 is less than 10, should pass
	testhelpers.MustApply(t, ruleSet, 9)

	// 10 is equal to 10, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, 10, errors.CodeMaxExclusive)

	// 11 is greater than 10, should fail
	testhelpers.MustNotApply(t, ruleSet, 11, errors.CodeMaxExclusive)
}

// TestFloatRuleSet_WithMaxExclusive tests:
// - Floats less than maximum pass validation
// - Floats equal to maximum fail validation (exclusive)
// - Floats greater than maximum fail validation
func TestFloatRuleSet_WithMaxExclusive(t *testing.T) {
	ruleSet := rules.Float64().WithMaxExclusive(10.0).Any()

	// 9.9 is less than 10.0, should pass
	testhelpers.MustApply(t, ruleSet, 9.9)

	// 10.0 is equal to 10.0, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, 10.0, errors.CodeMaxExclusive)

	// 10.1 is greater than 10.0, should fail
	testhelpers.MustNotApply(t, ruleSet, 10.1, errors.CodeMaxExclusive)
}

// TestIntRuleSet_WithMaxExclusive_Conflict tests:
// - Only one WithMaxExclusive can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMaxExclusive is used.
// - Rule is serialized properly.
func TestIntRuleSet_WithMaxExclusive_Conflict(t *testing.T) {
	ruleSet := rules.Int().WithMaxExclusive(10).WithMinExclusive(3)

	// Test validation with a value equal to the threshold (should return an error - exclusive)
	_, err := ruleSet.Apply(context.TODO(), 10)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the threshold (should return an error)
	_, err = ruleSet.Apply(context.TODO(), 11)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the threshold (should not return an error)
	_, err = ruleSet.Apply(context.TODO(), 9)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different threshold and test again
	ruleSet2 := ruleSet.WithMaxExclusive(9)

	// Test validation with a value at the new threshold (should return an error - exclusive)
	_, err = ruleSet2.Apply(context.TODO(), 9)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the new threshold (should not return an error)
	_, err = ruleSet2.Apply(context.TODO(), 8)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "IntRuleSet[int].WithMaxExclusive(10).WithMinExclusive(3)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated threshold
	expected = "IntRuleSet[int].WithMinExclusive(3).WithMaxExclusive(9)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestFloatRuleSet_WithMaxExclusive_Conflict tests:
// - Only one WithMaxExclusive can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMaxExclusive is used.
// - Rule is serialized properly.
func TestFloatRuleSet_WithMaxExclusive_Conflict(t *testing.T) {
	ruleSet := rules.Float64().WithMaxExclusive(10.0).WithMinExclusive(3.0)

	// Test validation with a value equal to the threshold (should return an error - exclusive)
	_, err := ruleSet.Apply(context.TODO(), 10.0)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the threshold (should return an error)
	_, err = ruleSet.Apply(context.TODO(), 11.0)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the threshold (should not return an error)
	_, err = ruleSet.Apply(context.TODO(), 9.0)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different threshold and test again
	ruleSet2 := ruleSet.WithMaxExclusive(9.0)

	// Test validation with a value at the new threshold (should return an error - exclusive)
	_, err = ruleSet2.Apply(context.TODO(), 9.0)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the new threshold (should not return an error)
	_, err = ruleSet2.Apply(context.TODO(), 8.0)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "FloatRuleSet[float64].WithMaxExclusive(10).WithMinExclusive(3)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated threshold
	expected = "FloatRuleSet[float64].WithMinExclusive(3).WithMaxExclusive(9)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMax and WithMaxExclusive conflict with each other
func TestIntRuleSet_WithMax_WithMaxExclusiveConflict(t *testing.T) {
	ruleSet := rules.Int().WithMax(10)

	// Adding WithMaxExclusive should conflict and replace WithMax
	ruleSet2 := ruleSet.WithMaxExclusive(9)

	// Original rule set should still have WithMax
	_, err := ruleSet.Apply(context.TODO(), 10)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// New rule set should have WithMaxExclusive (exclusive, so 9 should fail)
	_, err = ruleSet2.Apply(context.TODO(), 9)
	if err == nil {
		t.Errorf("Expected error for WithMaxExclusive at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := "IntRuleSet[int].WithMaxExclusive(9)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMaxExclusive and WithMax conflict with each other (reverse order)
func TestIntRuleSet_WithMaxExclusive_WithMaxConflict(t *testing.T) {
	ruleSet := rules.Int().WithMaxExclusive(10)

	// Adding WithMax should conflict and replace WithMaxExclusive
	ruleSet2 := ruleSet.WithMax(9)

	// Original rule set should still have WithMaxExclusive
	_, err := ruleSet.Apply(context.TODO(), 10)
	if err == nil {
		t.Errorf("Expected error for WithMaxExclusive at threshold (exclusive)")
	}

	// New rule set should have WithMax (inclusive, so 9 should pass)
	_, err = ruleSet2.Apply(context.TODO(), 9)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// Verify serialization shows the conflict resolution
	expected := "IntRuleSet[int].WithMax(9)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
