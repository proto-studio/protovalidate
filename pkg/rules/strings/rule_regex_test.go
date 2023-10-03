package strings_test

import (
	"regexp"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Executes valid regular expression provided as string
// - Returns the user supplied error
func TestRegexString(t *testing.T) {
	errStr := "test error"
	ruleSet := strings.New().WithRegexpString("^[a-z]+$", errStr).Any()

	testhelpers.MustBeValid(t, ruleSet, "abc", "abc")
	if err := testhelpers.MustBeInvalid(t, ruleSet, "123", errors.CodePattern); err != nil {
		if err.Error() != errStr {
			t.Errorf("Expected error to be '%s', got: '%s'", errStr, err)
		}
	}
}

// Requirements:
// - Panics on invalid regexp string
func TestInvalidRegex(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	strings.New().WithRegexpString("[[[", "")
}

// Requirements:
// - Executes valid regular expression provided as string
// - Returns the user supplied error
func TestRegex(t *testing.T) {
	errStr := "test error"
	exp := regexp.MustCompile("^[a-z]+$")
	ruleSet := strings.New().WithRegexp(exp, errStr).Any()

	testhelpers.MustBeValid(t, ruleSet, "abc", "abc")
	if err := testhelpers.MustBeInvalid(t, ruleSet, "123", errors.CodePattern); err != nil {
		if err.Error() != errStr {
			t.Errorf("Expected error to be '%s', got: '%s'", errStr, err)
		}
	}
}
