package arrays_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/rules/arrays"
)

func TestMaxLen(t *testing.T) {
	ruleSet := arrays.New[int]().WithMaxLen(2)

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
func TestMaxLenConflict(t *testing.T) {
	ruleSet := arrays.New[int]().WithMaxLen(3).WithMinLen(1)

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
	expected := "ArrayRuleSet[int].WithMaxLen(3).WithMinLen(1)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Verify that the new rule set's string representation is correct
	expected = "ArrayRuleSet[int].WithMinLen(1).WithMaxLen(4)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
