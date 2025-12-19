package rules_test

import (
	"fmt"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Allowed values are cumulative.
func TestStringRuleSet_WithAllowedValues(t *testing.T) {
	ruleSet := rules.String().WithAllowedValues("a", "b").WithMaxLen(1)

	testhelpers.MustApply(t, ruleSet.Any(), "a")
	testhelpers.MustApply(t, ruleSet.Any(), "b")
	testhelpers.MustNotApply(t, ruleSet.Any(), "c", errors.CodeNotAllowed)

	expected := fmt.Sprintf("StringRuleSet.WithAllowedValues(\"a\", \"b\").WithMaxLen(1)")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = ruleSet.WithAllowedValues("c")
	testhelpers.MustApply(t, ruleSet.Any(), "a")
	testhelpers.MustApply(t, ruleSet.Any(), "c")

	expected = fmt.Sprintf("StringRuleSet.WithMaxLen(1).WithAllowedValues(\"a\", \"b\", \"c\")")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Only the first 3 values are displayed.
// - Values are separated by commas.
// - Values are quoted.
func TestStringRuleSet_WithAllowedValues_More(t *testing.T) {
	values := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
	}

	ruleSet := rules.String().WithAllowedValues(values[0], values[1])
	expected := fmt.Sprintf("StringRuleSet.WithAllowedValues(\"%s\", \"%s\")", values[0], values[1])
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = ruleSet.WithAllowedValues(values[2])
	expected = fmt.Sprintf("StringRuleSet.WithAllowedValues(\"%s\", \"%s\", \"%s\")", values[0], values[1], values[2])
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = ruleSet.WithAllowedValues(values[3], values[4:]...)
	expected = fmt.Sprintf("StringRuleSet.WithAllowedValues(\"%s\", \"%s\", \"%s\" ... and 2 more)", values[0], values[1], values[2])
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Rejected values are cumulative.
// - Rejected values causes a validation error.
func TestStringRuleSet_WithRejectedValues(t *testing.T) {
	ruleSet := rules.String().WithRejectedValues("a", "b")

	testhelpers.MustNotApply(t, ruleSet.Any(), "a", errors.CodeForbidden)
	testhelpers.MustNotApply(t, ruleSet.Any(), "b", errors.CodeForbidden)
	testhelpers.MustApply(t, ruleSet.Any(), "c")

	expected := fmt.Sprintf("StringRuleSet.WithRejectedValues(\"a\", \"b\")")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = ruleSet.WithRejectedValues("c")
	testhelpers.MustNotApply(t, ruleSet.Any(), "a", errors.CodeForbidden)
	testhelpers.MustNotApply(t, ruleSet.Any(), "c", errors.CodeForbidden)

	expected = fmt.Sprintf("StringRuleSet.WithRejectedValues(\"a\", \"b\").WithRejectedValues(\"c\")")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
