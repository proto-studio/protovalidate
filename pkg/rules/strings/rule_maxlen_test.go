package strings_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestMaxLen(t *testing.T) {
	ruleSet := strings.New().WithMaxLen(2).Any()

	testhelpers.MustApply(t, ruleSet, "a")
	testhelpers.MustApply(t, ruleSet, "ab")
	testhelpers.MustNotApply(t, ruleSet, "abc", errors.CodeMax)

}

// Requirements:
// - Only one max length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent maximum is used.
func TestMaxLenConflict(t *testing.T) {
	ruleSet := strings.New().WithMaxLen(2).WithMinLen(1)

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
