package rules_test

import (
	"regexp"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Executes valid regular expression provided as string
// - Returns the user supplied error
func TestStringRuleSet_WithRegexpString(t *testing.T) {
	errStr := "test error"
	ruleSet := rules.String().WithRegexpString("^[a-z]+$", errStr).Any()

	testhelpers.MustApply(t, ruleSet, "abc")
	if err := testhelpers.MustNotApply(t, ruleSet, "123", errors.CodePattern); err != nil {
		if err.Error() != errStr {
			t.Errorf("Expected error to be '%s', got: '%s'", errStr, err)
		}
	}
}

// Requirements:
// - Panics on invalid regexp string
func TestStringRuleSet_WithRegexpString_Invalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	rules.String().WithRegexpString("[[[", "")
}

// Requirements:
// - Executes valid regular expression provided as string
// - Returns the user supplied error
func TestStringRuleSet_WithRegexp(t *testing.T) {
	errStr := "test error"
	exp := regexp.MustCompile("^[a-z]+$")
	ruleSet := rules.String().WithRegexp(exp, errStr).Any()

	testhelpers.MustApply(t, ruleSet, "abc")
	if err := testhelpers.MustNotApply(t, ruleSet, "123", errors.CodePattern); err != nil {
		if err.Error() != errStr {
			t.Errorf("Expected error to be '%s', got: '%s'", errStr, err)
		}
	}
}

// Requirements:
// - Serializes to WithRegex(...)
func TestStringRuleSet_String_WithRegexp(t *testing.T) {
	ruleSet := rules.String().WithRegexpString("[a-z]", "").WithRegexpString("[0-9]", "")

	expected := "StringRuleSet.WithRegexp([a-z]).WithRegexp([0-9])"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
