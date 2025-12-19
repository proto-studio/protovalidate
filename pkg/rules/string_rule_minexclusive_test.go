package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestStringRuleSet_WithMinExclusive(t *testing.T) {
	ruleSet := rules.String().WithMinExclusive("b").Any()

	// "a" is lexicographically less than "b", should fail
	testhelpers.MustNotApply(t, ruleSet, "a", errors.CodeMin)

	// "b" is equal to "b", should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, "b", errors.CodeMin)

	// "c" is lexicographically greater than "b", should pass
	testhelpers.MustApply(t, ruleSet, "c")

	// "ba" is lexicographically greater than "b", should pass
	testhelpers.MustApply(t, ruleSet, "ba")
}

// Requirements:
// - Only one WithMinExclusive can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent WithMinExclusive is used.
// - Rule is serialized properly.
func TestStringRuleSet_WithMinExclusive_Conflict(t *testing.T) {
	ruleSet := rules.String().WithMinExclusive("c").WithMaxExclusive("z")

	var output string

	// Test validation with a value equal to the threshold (should return an error)
	err := ruleSet.Apply(context.TODO(), "c", &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value below the threshold (should return an error)
	err = ruleSet.Apply(context.TODO(), "b", &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the threshold (should not return an error)
	err = ruleSet.Apply(context.TODO(), "d", &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different threshold and test again
	ruleSet2 := ruleSet.WithMinExclusive("b")

	// Test validation with a value at the new threshold (should return an error - exclusive)
	err = ruleSet2.Apply(context.TODO(), "b", &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value above the new threshold (should not return an error)
	err = ruleSet2.Apply(context.TODO(), "c", &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "StringRuleSet.WithMinExclusive(\"c\").WithMaxExclusive(\"z\")"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated threshold
	expected = "StringRuleSet.WithMaxExclusive(\"z\").WithMinExclusive(\"b\")"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

func TestStringRuleSet_WithMinExclusive_Lexicographical(t *testing.T) {
	ruleSet := rules.String().WithMinExclusive("apple").Any()

	// "ap" is lexicographically less than "apple", should fail
	testhelpers.MustNotApply(t, ruleSet, "ap", errors.CodeMin)

	// "apple" is equal, should fail (exclusive)
	testhelpers.MustNotApply(t, ruleSet, "apple", errors.CodeMin)

	// "apples" is lexicographically greater, should pass
	testhelpers.MustApply(t, ruleSet, "apples")

	// "banana" is lexicographically greater, should pass
	testhelpers.MustApply(t, ruleSet, "banana")
}
