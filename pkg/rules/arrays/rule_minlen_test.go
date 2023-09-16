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
	} else if err.Size() != 1 {
		t.Errorf("Expected 1 error got %d", err.Size())
	}
}
