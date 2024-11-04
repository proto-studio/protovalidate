package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestSlice_MinLen(t *testing.T) {
	ruleSet := rules.Slice[int]().WithMinLen(2)

	// Prepare an output variable for Apply
	var output []int

	// Apply with an array that exceeds the minimum length, expecting no error
	err := ruleSet.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Apply with an array that matches the minimum length, expecting no error
	err = ruleSet.Apply(context.TODO(), []int{1, 2}, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Apply with an array that is below the minimum length, expecting an error
	err = ruleSet.Apply(context.TODO(), []int{1}, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if len(err) != 1 {
		t.Errorf("Expected 1 error, got %d", len(err))
	}
}

// Requirements:
// - Only one min length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
func TestSlice_MinLen_Conflict(t *testing.T) {
	ruleSet := rules.Slice[int]().WithMinLen(3).WithMaxLen(10)

	// Prepare an output variable for Apply
	var output []int

	// Apply with an array that is below the minimum length, expecting an error
	err := ruleSet.Apply(context.TODO(), []int{1, 2}, &output)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	}

	// Apply with an array that matches the minimum length, expecting no error
	err = ruleSet.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Create a new rule set with a different minimum length and validate
	ruleSet2 := ruleSet.WithMinLen(2)

	// Apply with an array that matches the new minimum length, expecting no error
	err = ruleSet2.Apply(context.TODO(), []int{1, 2}, &output)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Verify that the original rule set's string representation is correct
	expected := "SliceRuleSet[int].WithMinLen(3).WithMaxLen(10)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = "SliceRuleSet[int].WithMaxLen(10).WithMinLen(2)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

func TestString_WithMinLen(t *testing.T) {
	ruleSet := rules.String().WithMinLen(2).Any()

	testhelpers.MustApply(t, ruleSet, "abc")
	testhelpers.MustApply(t, ruleSet, "ab")
	testhelpers.MustNotApply(t, ruleSet, "a", errors.CodeMin)
}

// Requirements:
// - Only one min length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
func TestString_WithMinLen_Conflict(t *testing.T) {
	ruleSet := rules.String().WithMinLen(3).WithMaxLen(10)

	// Prepare the output variable for Apply
	var out string

	// First validation with min length 3
	if err := ruleSet.Apply(context.TODO(), "ab", &out); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if err := ruleSet.Apply(context.TODO(), "abc", &out); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	// Update the rule set with min length 2 and validate
	ruleSet2 := ruleSet.WithMinLen(2)
	if err := ruleSet2.Apply(context.TODO(), "ab", &out); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	// Check the string representation of the rule sets
	expected := "StringRuleSet.WithMinLen(3).WithMaxLen(10)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "StringRuleSet.WithMaxLen(10).WithMinLen(2)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
