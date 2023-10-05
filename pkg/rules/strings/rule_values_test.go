package strings_test

import (
	"fmt"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Allowed values are cumulative.
func TestWithAllowedValues(t *testing.T) {
	ruleSet := strings.New().WithAllowedValues("a", "b").WithMaxLen(1)

	testhelpers.MustBeValid(t, ruleSet.Any(), "a", "a")
	testhelpers.MustBeValid(t, ruleSet.Any(), "b", "b")
	testhelpers.MustBeInvalid(t, ruleSet.Any(), "c", errors.CodeNotAllowed)

	expected := fmt.Sprintf("StringRuleSet.WithAllowedValues(\"a\", \"b\").WithMaxLen(1)")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = ruleSet.WithAllowedValues("c")
	testhelpers.MustBeValid(t, ruleSet.Any(), "a", "a")
	testhelpers.MustBeValid(t, ruleSet.Any(), "c", "c")

	expected = fmt.Sprintf("StringRuleSet.WithMaxLen(1).WithAllowedValues(\"a\", \"b\", \"c\")")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Only the first 3 values are displayed.
// - Values are separated by commas.
// - Values are quoted.
func TestWithAllowedValuesMore(t *testing.T) {
	values := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
	}

	ruleSet := strings.New().WithAllowedValues(values[0], values[1])
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
func TestWithRejectedValues(t *testing.T) {
	ruleSet := strings.New().WithRejectedValues("a", "b")

	testhelpers.MustBeInvalid(t, ruleSet.Any(), "a", errors.CodeForbidden)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), "b", errors.CodeForbidden)
	testhelpers.MustBeValid(t, ruleSet.Any(), "c", "c")

	expected := fmt.Sprintf("StringRuleSet.WithRejectedValues(\"a\", \"b\")")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = ruleSet.WithRejectedValues("c")
	testhelpers.MustBeInvalid(t, ruleSet.Any(), "a", errors.CodeForbidden)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), "c", errors.CodeForbidden)

	expected = fmt.Sprintf("StringRuleSet.WithRejectedValues(\"a\", \"b\").WithRejectedValues(\"c\")")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
