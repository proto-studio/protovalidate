package net_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Default configuration doesn't return errors on valid value.
// - Implements interface.
func TestEmailRuleSet(t *testing.T) {
	d, err := net.NewEmail().Validate("hello@example.com")

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if d != "hello@example.com" {
		t.Error("Expected test email to be returned")
		return
	}

	ok := testhelpers.CheckRuleSetInterface[string](net.NewEmail())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

// Requirements:
// - Default validator requires a TLD
// - Unknown TLDs error
func TestEmailDefaultDomain(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustBeValid(t, ruleSet, "hello@example.com", "hello@example.com")
	testhelpers.MustBeInvalid(t, ruleSet, "hello@example.bogusbogus", errors.CodePattern)
}

// Requirements:
// - Errors if there isn't any "@"
// - Errors if there is more than one "@"
func TestEmailSplit(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustBeInvalid(t, ruleSet, "example.com", errors.CodePattern)
	testhelpers.MustBeInvalid(t, ruleSet, "hello@world@example.com", errors.CodePattern)
}

// Requirements:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
func TestEmailRequired(t *testing.T) {
	ruleSet := net.NewEmail()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestEmailCustom(t *testing.T) {
	_, err := net.NewEmail().
		WithRuleFunc(testhelpers.MockCustomRule("name@example.com", 1)).
		Validate("name@example.com")

	if err == nil {
		t.Error("Expected errors to not be empty")
		return
	}

	expected := "name@example.com"

	actual, err := net.NewEmail().
		WithRuleFunc(testhelpers.MockCustomRule(expected, 0)).
		Validate("name@example.com")

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if expected != actual {
		t.Errorf("Expected '%s' to equal '%s'", actual, expected)
		return
	}
}

// Requirements:
// - Custom domain RuleSet overrides default set.
func TestEmailWithDomain(t *testing.T) {
	domainRuleSet := net.NewDomain().WithSuffix("edu").Any()
	ruleSet := net.NewEmail().WithDomain(domainRuleSet).Any()

	testhelpers.MustBeValid(t, ruleSet, "hello@example.edu", "hello@example.edu")
	testhelpers.MustBeInvalid(t, ruleSet, "hello@example.com", errors.CodePattern)
}

// Requirements:
// - Errors when input is not a string
// - errors.CodeType is returned
func TestEmailType(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustBeInvalid(t, ruleSet, 123, errors.CodeType)
}

// Requirements:
// - No double dots
// - Can't start with a dot
// - Can't end with a dot
func TestEmailDots(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustBeValid(t, ruleSet, "hello.world@example.com", "hello.world@example.com")
	testhelpers.MustBeInvalid(t, ruleSet, "hello..world@example.com", errors.CodePattern)
	testhelpers.MustBeInvalid(t, ruleSet, ".helloworld@example.com", errors.CodePattern)
	testhelpers.MustBeInvalid(t, ruleSet, "helloworld.@example.com", errors.CodePattern)
}

// Requirements:
// - Errors when the local part is empty
func TestEmailEmptyLocal(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustBeInvalid(t, ruleSet, "@example.com", errors.CodePattern)
}
