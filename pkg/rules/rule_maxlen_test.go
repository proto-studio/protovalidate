package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestSlice_MaxLen(t *testing.T) {
	ruleSet := rules.Slice[int]().WithMaxLen(2)

	// Prepare an output variable for Apply
	var output []int

	// Apply with an array that is under the maximum length, expecting no error
	err := ruleSet.Apply(context.TODO(), []int{1}, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Apply with an array that is exactly the maximum length, expecting no error
	err = ruleSet.Apply(context.TODO(), []int{1, 2}, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Apply with an array that exceeds the maximum length, expecting an error
	err = ruleSet.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if len(err) != 1 {
		t.Errorf("Expected 1 error, got %d", len(err))
	}
}

// Requirements:
// - Only one max length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
func TestSlice_MaxLen_Conflict(t *testing.T) {
	ruleSet := rules.Slice[int]().WithMaxLen(3).WithMinLen(1)

	// Prepare an output variable for Apply
	var output []int

	// Apply with an array that exceeds the maximum length, expecting an error
	err := ruleSet.Apply(context.TODO(), []int{1, 2, 3, 4}, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with an array that matches the maximum length, expecting no error
	err = ruleSet.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different maximum length and validate
	ruleSet2 := ruleSet.WithMaxLen(4)

	// Apply with an array that matches the new maximum length, expecting no error
	err = ruleSet2.Apply(context.TODO(), []int{1, 2, 3, 4}, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := "SliceRuleSet[int].WithMaxLen(3).WithMinLen(1)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = "SliceRuleSet[int].WithMinLen(1).WithMaxLen(4)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

func TestString_WithMaxLen(t *testing.T) {
	ruleSet := rules.String().WithMaxLen(2).Any()

	testhelpers.MustApply(t, ruleSet, "a")
	testhelpers.MustApply(t, ruleSet, "ab")
	testhelpers.MustNotApply(t, ruleSet, "abc", errors.CodeMax)

}

// Requirements:
// - Only one max length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
func TestString_WithMaxLen_Conflict(t *testing.T) {
	ruleSet := rules.String().WithMaxLen(2).WithMinLen(1)

	// Prepare the output variable for Apply
	var out string

	// First validation with max length 2
	if err := ruleSet.Apply(context.TODO(), "abc", &out); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if err := ruleSet.Apply(context.TODO(), "ab", &out); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Update the rule set with max length 3 and validate
	ruleSet2 := ruleSet.WithMaxLen(3)
	if err := ruleSet2.Apply(context.TODO(), "abc", &out); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Check the string representation of the rule sets
	expected := "StringRuleSet.WithMaxLen(2).WithMinLen(1)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "StringRuleSet.WithMinLen(1).WithMaxLen(3)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
