package strings_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestMinLen(t *testing.T) {
	ruleSet := strings.New().WithMinLen(2).Any()

	testhelpers.MustRun(t, ruleSet, "abc")
	testhelpers.MustRun(t, ruleSet, "ab")
	testhelpers.MustNotRun(t, ruleSet, "a", errors.CodeMin)
}

// Requirements:
// - Only one min length can exist on a rule set.
// - Original rule set is not mutated.
// - Most recent minimum is used.
func TestMinLenConflict(t *testing.T) {
	ruleSet := strings.New().WithMinLen(3).WithMaxLen(10)

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
