package strings_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules/strings"
)

func TestMinLen(t *testing.T) {
	ruleSet := strings.New().WithMinLen(2)

	_, err := ruleSet.Validate("abc")
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	_, err = ruleSet.Validate("ab")
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	_, err = ruleSet.Validate("a")
	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if err.Size() != 1 {
		t.Errorf("Expected 1 error got %d", err.Size())
	}
}
