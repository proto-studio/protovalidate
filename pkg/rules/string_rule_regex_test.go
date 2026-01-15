package rules_test

import (
	"regexp"
	"strings"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestStringRuleSet_WithRegexpString tests:
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

// TestStringRuleSet_WithRegexpString_Invalid tests:
// - Panics on invalid regexp string
func TestStringRuleSet_WithRegexpString_Invalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	rules.String().WithRegexpString("[[[", "")
}

// TestStringRuleSet_WithRegexp tests:
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

// TestStringRuleSet_String_WithRegexp tests:
// - Serializes to WithRegex(...)
func TestStringRuleSet_String_WithRegexp(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.StringRuleSet
		expected string
	}{
		{"Simple", rules.String().WithRegexpString("[a-z]", ""), "StringRuleSet.WithRegexp(\"[a-z]\")"},
		{"Multiple", rules.String().WithRegexpString("[a-z]", "").WithRegexpString("[0-9]", ""), "StringRuleSet.WithRegexp(\"[a-z]\").WithRegexp(\"[0-9]\")"},
		{"LongPattern", rules.String().WithRegexpString("^[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*$", ""), "StringRuleSet.WithRegexp(\"^[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[...\")"},
		{"VeryLongPattern", rules.String().WithRegexpString("^[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*$", ""), "StringRuleSet.WithRegexp(\"^[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[...\")"},
		{"Chained", rules.String().WithRequired().WithRegexpString("[a-z]", ""), "StringRuleSet.WithRequired().WithRegexp(\"[a-z]\")"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ruleSet.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRegexpRule_String_Truncation tests that the rule String() method truncates long regex patterns.
func TestRegexpRule_String_Truncation(t *testing.T) {
	// Create a very long regex pattern (longer than 50 characters)
	longPattern := "^[a-z]+"
	for i := 0; i < 20; i++ {
		longPattern += "[0-9]*[A-Z]*"
	}
	longPattern += "$"

	ruleSet := rules.String().WithRegexpString(longPattern, "error")
	ruleStr := ruleSet.String()

	// The String() method should truncate the regex pattern if it's too long
	if len(longPattern) > 50 {
		if !strings.Contains(ruleStr, "...") {
			t.Errorf("String() should contain ellipsis for truncated patterns, got: %s", ruleStr)
		}
	}
}
