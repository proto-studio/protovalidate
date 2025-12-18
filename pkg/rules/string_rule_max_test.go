package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestString_WithMax(t *testing.T) {
	ruleSet := rules.String().WithMax("y").Any()

	// "x" is lexicographically less than "y", should pass
	testhelpers.MustApply(t, ruleSet, "x")

	// "y" is equal to "y", should pass (inclusive)
	testhelpers.MustApply(t, ruleSet, "y")

	// "z" is lexicographically greater than "y", should fail
	testhelpers.MustNotApply(t, ruleSet, "z", errors.CodeMax)

	// "ya" is lexicographically greater than "y", should fail
	testhelpers.MustNotApply(t, ruleSet, "ya", errors.CodeMax)
}

// Requirements:
// - Only one max can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
// - Rule is serialized properly.
func TestString_MaxConflict(t *testing.T) {
	ruleSet := rules.String().WithMax("z").WithMin("a")

	var output string

	// Test validation with a value that exceeds the max (should return an error)
	err := ruleSet.Apply(context.TODO(), "zzz", &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Test validation with a value within the range (should not return an error)
	err = ruleSet.Apply(context.TODO(), "m", &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different max and test again
	ruleSet2 := ruleSet.WithMax("y")

	// Test validation with a value that is within the new max (should not return an error)
	err = ruleSet2.Apply(context.TODO(), "x", &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set is not mutated
	expected := "StringRuleSet.WithMax(\"z\").WithMin(\"a\")"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set has the updated max
	expected = "StringRuleSet.WithMin(\"a\").WithMax(\"y\")"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

func TestString_WithMax_Lexicographical(t *testing.T) {
	ruleSet := rules.String().WithMax("banana").Any()

	// "apple" is lexicographically less, should pass
	testhelpers.MustApply(t, ruleSet, "apple")

	// "banana" is equal, should pass
	testhelpers.MustApply(t, ruleSet, "banana")

	// "bananas" is lexicographically greater, should fail
	testhelpers.MustNotApply(t, ruleSet, "bananas", errors.CodeMax)

	// "cherry" is lexicographically greater, should fail
	testhelpers.MustNotApply(t, ruleSet, "cherry", errors.CodeMax)
}

func TestString_WithMax_Truncation(t *testing.T) {
	// Create a very long string (longer than 50 characters)
	longString := "a"
	for i := 0; i < 100; i++ {
		longString += "b"
	}

	ruleSet := rules.String().WithMax(longString).Any()

	// Test that the error message contains truncated string with ellipsis
	// Use a value that exceeds the max
	var output string
	err := ruleSet.Apply(context.TODO(), longString+"z", &output)
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
