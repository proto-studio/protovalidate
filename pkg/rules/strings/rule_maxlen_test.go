package strings_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules/strings"
)

func TestMaxLen(t *testing.T) {
	ruleSet := strings.New().WithMaxLen(2)

	_, err := ruleSet.Validate("a")
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	_, err = ruleSet.Validate("ab")
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	_, err = ruleSet.Validate("abc")
	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if s := len(err); s != 1 {
		t.Errorf("Expected 1 error got %d", s)
	}
}
