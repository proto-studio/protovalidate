package numbers_test

import (
	"fmt"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Allowed values are cumulative.
func TestWithAllowedValues(t *testing.T) {
	ruleSet := numbers.NewInt().WithAllowedValues(1, 5).WithMax(100)

	testhelpers.MustBeValid(t, ruleSet.Any(), 1, 1)
	testhelpers.MustBeValid(t, ruleSet.Any(), 5, 5)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), 10, errors.CodeNotAllowed)

	expected := fmt.Sprintf("IntRuleSet[int].WithAllowedValues(1, 5).WithMax(100)")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = ruleSet.WithAllowedValues(10)
	testhelpers.MustBeValid(t, ruleSet.Any(), 1, 1)
	testhelpers.MustBeValid(t, ruleSet.Any(), 10, 10)

	expected = fmt.Sprintf("IntRuleSet[int].WithMax(100).WithAllowedValues(1, 5, 10)")
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
	ruleSet := numbers.NewInt().WithRejectedValues(1, 5)

	testhelpers.MustBeInvalid(t, ruleSet.Any(), 1, errors.CodeForbidden)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), 5, errors.CodeForbidden)
	testhelpers.MustBeValid(t, ruleSet.Any(), 10, 10)

	expected := fmt.Sprintf("IntRuleSet[int].WithRejectedValues(1, 5)")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = ruleSet.WithRejectedValues(10)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), 1, errors.CodeForbidden)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), 10, errors.CodeForbidden)

	expected = fmt.Sprintf("IntRuleSet[int].WithRejectedValues(1, 5).WithRejectedValues(10)")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
