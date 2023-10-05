package arrays_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules/arrays"
)

func TestMinLen(t *testing.T) {
	ruleSet := arrays.New[int]().WithMinLen(2)

	_, err := ruleSet.Validate([]int{1, 2, 3})
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	_, err = ruleSet.Validate([]int{1, 2})
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	_, err = ruleSet.Validate([]int{1})
	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if len(err) != 1 {
		t.Errorf("Expected 1 error got %d", len(err))
	}
}

// Requirements:
// - Only one min length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
func TestMinLenConflict(t *testing.T) {
	ruleSet := arrays.New[int]().WithMinLen(3).WithMaxLen(10)

	if _, err := ruleSet.Validate([]int{1, 2}); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if _, err := ruleSet.Validate([]int{1, 2, 3}); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	ruleSet2 := ruleSet.WithMinLen(2)
	if _, err := ruleSet2.Validate([]int{1, 2}); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	expected := "ArrayRuleSet[int].WithMinLen(3).WithMaxLen(10)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "ArrayRuleSet[int].WithMaxLen(10).WithMinLen(2)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
