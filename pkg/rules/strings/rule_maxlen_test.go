package strings_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestMaxLen(t *testing.T) {
	ruleSet := strings.New().WithMaxLen(2).Any()

	testhelpers.MustRun(t, ruleSet, "a")
	testhelpers.MustRun(t, ruleSet, "ab")
	testhelpers.MustNotRun(t, ruleSet, "abc", errors.CodeMax)

}

// Requirements:
// - Only one max length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
func TestMaxLenConflict(t *testing.T) {
	ruleSet := strings.New().WithMaxLen(2).WithMinLen(1)

	if _, err := ruleSet.Validate("abc"); err == nil {
		t.Errorf("Expected error to not be nil")
	}
	if _, err := ruleSet.Validate("ab"); err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}

	ruleSet2 := ruleSet.WithMaxLen(3)
	if _, err := ruleSet2.Validate("abc"); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	expected := "StringRuleSet.WithMaxLen(2).WithMinLen(1)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = "StringRuleSet.WithMinLen(1).WithMaxLen(3)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
