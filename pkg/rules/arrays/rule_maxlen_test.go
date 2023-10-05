package arrays_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules/arrays"
)

func TestMaxLen(t *testing.T) {
	ruleSet := arrays.New[int]().WithMaxLen(2)

	_, err := ruleSet.Validate([]int{1})
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	_, err = ruleSet.Validate([]int{1, 2})
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	_, err = ruleSet.Validate([]int{1, 2, 3})
	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if len(err) != 1 {
		t.Errorf("Expected 1 error got %d", len(err))
	}
}

// Requirements:
// - Only one max length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
func TestMaxLenConflict(t *testing.T) {
	ruleSet := arrays.New[int]().WithMaxLen(3).WithMinLen(1)

	if _, err := ruleSet.Validate([]int{1, 2, 3, 4}); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if _, err := ruleSet.Validate([]int{1, 2, 3}); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	ruleSet2 := ruleSet.WithMaxLen(4)
	if _, err := ruleSet2.Validate([]int{1, 2, 3, 4}); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	expected := "ArrayRuleSet[int].WithMaxLen(3).WithMinLen(1)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "ArrayRuleSet[int].WithMinLen(1).WithMaxLen(4)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
