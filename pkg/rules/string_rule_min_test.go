package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestStringRuleSet_WithMin tests:
func TestStringRuleSet_WithMin(t *testing.T) {
	ruleSet := rules.String().WithMin("b").Any()

	// "a" is lexicographically less than "b", should fail
	testhelpers.MustNotApply(t, ruleSet, "a", errors.CodeMin)

	// "b" is equal to "b", should pass (inclusive)
	testhelpers.MustApply(t, ruleSet, "b")

	// "c" is lexicographically greater than "b", should pass
	testhelpers.MustApply(t, ruleSet, "c")

	// "ba" is lexicographically greater than "b", should pass
	testhelpers.MustApply(t, ruleSet, "ba")
}

// TestStringRuleSet_WithMin_Conflict tests:
// - Only one min can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
// - Rule is serialized properly.
func TestStringRuleSet_WithMin_Conflict(t *testing.T) {
	ruleSet := rules.String().WithMin("c").WithMax("z")

	var output string

	// Test validation with a value below the min (should return an error)
	err := ruleSet.Apply(context.TODO(), "a", &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value at the min (should not return an error)
	err = ruleSet.Apply(context.TODO(), "c", &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different min and test again
	ruleSet2 := ruleSet.WithMin("b")

	// Test validation with a value at the new min (should not return an error)
	err = ruleSet2.Apply(context.TODO(), "b", &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "StringRuleSet.WithMin(\"c\").WithMax(\"z\")"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated min
	expected = "StringRuleSet.WithMax(\"z\").WithMin(\"b\")"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestStringRuleSet_WithMin_Lexicographical tests:
func TestStringRuleSet_WithMin_Lexicographical(t *testing.T) {
	ruleSet := rules.String().WithMin("apple").Any()

	// "ap" is lexicographically less than "apple"
	testhelpers.MustNotApply(t, ruleSet, "ap", errors.CodeMin)

	// "apple" is equal, should pass
	testhelpers.MustApply(t, ruleSet, "apple")

	// "apples" is lexicographically greater, should pass
	testhelpers.MustApply(t, ruleSet, "apples")

	// "banana" is lexicographically greater, should pass
	testhelpers.MustApply(t, ruleSet, "banana")
}

// TestStringRuleSet_WithMin_Truncation tests:
func TestStringRuleSet_WithMin_Truncation(t *testing.T) {
	// Create a very long string (longer than 50 characters)
	longString := "a"
	for i := 0; i < 100; i++ {
		longString += "b"
	}

	ruleSet := rules.String().WithMin(longString).Any()

	// Test that the error message contains truncated string with ellipsis
	var output string
	err := ruleSet.Apply(context.TODO(), "a", &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
		return
	}

	errMsg := err[0].Error()
	// The error message should contain the truncated string (50 chars + "...")
	// We check that it doesn't contain the full 101-character string
	if len(errMsg) > 200 {
		t.Errorf("Error message seems too long, may not be truncating: %s", errMsg)
	}
	// Should contain ellipsis for long strings
	if len(longString) > 50 {
		// Simple check: error message should be shorter than if it contained the full string
		// and should contain ellipsis
		hasEllipsis := false
		for i := 0; i <= len(errMsg)-3; i++ {
			if errMsg[i:i+3] == "..." {
				hasEllipsis = true
				break
			}
		}
		if !hasEllipsis {
			t.Errorf("Error message should contain ellipsis for long strings: %s", errMsg)
		}
	}
}
