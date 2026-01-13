package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestStringRuleSet_WithMaxExclusive tests:
// - Strings less than maximum pass validation
// - Strings equal to maximum fail validation (exclusive)
// - Strings greater than maximum fail validation
func TestStringRuleSet_WithMaxExclusive(t *testing.T) {
	ruleSet := rules.String().WithMaxExclusive("y").Any()

	// "x" is lexicographically less than "y", should pass
	testhelpers.MustApply(t, ruleSet, "x")

	// "y" is equal to "y", should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, "y", errors.CodeMaxExclusive)

	// "z" is lexicographically greater than "y", should fail
	testhelpers.MustNotApply(t, ruleSet, "z", errors.CodeMaxExclusive)

	// "ya" is lexicographically greater than "y", should fail
	testhelpers.MustNotApply(t, ruleSet, "ya", errors.CodeMaxExclusive)
}

// TestStringRuleSet_WithMaxExclusive_Conflict tests:
// - Only one WithMaxExclusive can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMaxExclusive is used.
// - Rule is serialized properly.
func TestStringRuleSet_WithMaxExclusive_Conflict(t *testing.T) {
	ruleSet := rules.String().WithMaxExclusive("z").WithMinExclusive("a")

	var output string

	// Test validation with a value equal to the threshold (should return an error)
	err := ruleSet.Apply(context.TODO(), "z", &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the threshold (should return an error)
	err = ruleSet.Apply(context.TODO(), "zzz", &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the threshold (should not return an error)
	err = ruleSet.Apply(context.TODO(), "y", &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different threshold and test again
	ruleSet2 := ruleSet.WithMaxExclusive("y")

	// Test validation with a value at the new threshold (should return an error - exclusive)
	err = ruleSet2.Apply(context.TODO(), "y", &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the new threshold (should not return an error)
	err = ruleSet2.Apply(context.TODO(), "x", &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "StringRuleSet.WithMaxExclusive(\"z\").WithMinExclusive(\"a\")"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated threshold
	expected = "StringRuleSet.WithMinExclusive(\"a\").WithMaxExclusive(\"y\")"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestStringRuleSet_WithMaxExclusive_Lexicographical tests:
// - Uses lexicographical comparison for string maximum exclusive
func TestStringRuleSet_WithMaxExclusive_Lexicographical(t *testing.T) {
	ruleSet := rules.String().WithMaxExclusive("banana").Any()

	// "apple" is lexicographically less, should pass
	testhelpers.MustApply(t, ruleSet, "apple")

	// "banana" is equal, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, "banana", errors.CodeMaxExclusive)

	// "bananas" is lexicographically greater, should fail
	testhelpers.MustNotApply(t, ruleSet, "bananas", errors.CodeMaxExclusive)

	// "cherry" is lexicographically greater, should fail
	testhelpers.MustNotApply(t, ruleSet, "cherry", errors.CodeMaxExclusive)
}
