package net_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
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

	testhelpers.MustRun(t, ruleSet, "hello@example.com")
	testhelpers.MustNotRun(t, ruleSet, "hello@example.bogusbogus", errors.CodePattern)
}

// Requirements:
// - Errors if there isn't any "@"
// - Errors if there is more than one "@"
func TestEmailSplit(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustNotRun(t, ruleSet, "example.com", errors.CodePattern)
	testhelpers.MustNotRun(t, ruleSet, "hello@world@example.com", errors.CodePattern)
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
	mock := testhelpers.NewMockRuleWithErrors[string](1)

	_, err := net.NewEmail().
		WithRuleFunc(mock.Function()).
		Validate("name@example.com")

	if err == nil {
		t.Error("Expected errors to not be empty")
		return
	}

	if mock.CallCount() != 1 {
		t.Errorf("Expected rule to be called 1 times, got %d", mock.CallCount())
		return
	}

	str := net.NewEmail().WithRuleFunc(mock.Function()).String()
	expected := "EmailRuleSet.WithRuleFunc(...)"
	if str != expected {
		t.Errorf("Expected %s, got %s", expected, str)
	}

	rule := testhelpers.NewMockRule[string]()

	_, err = net.NewEmail().
		WithRuleFunc(rule.Function()).
		Validate("name@example.com")

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if c := rule.CallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

// Requirements:
// - Custom domain RuleSet overrides default set.
func TestEmailWithDomain(t *testing.T) {
	domainRuleSet := net.NewDomain().WithSuffix("edu").Any()
	ruleSet := net.NewEmail().WithDomain(domainRuleSet).Any()

	testhelpers.MustRun(t, ruleSet, "hello@example.edu")
	testhelpers.MustNotRun(t, ruleSet, "hello@example.com", errors.CodePattern)
}

// Requirements:
// - Errors when input is not a string
// - errors.CodeType is returned
func TestEmailType(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustNotRun(t, ruleSet, 123, errors.CodeType)
}

// Requirements:
// - No double dots
// - Can't start with a dot
// - Can't end with a dot
func TestEmailDots(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustRun(t, ruleSet, "hello.world@example.com")
	testhelpers.MustNotRun(t, ruleSet, "hello..world@example.com", errors.CodePattern)
	testhelpers.MustNotRun(t, ruleSet, ".helloworld@example.com", errors.CodePattern)
	testhelpers.MustNotRun(t, ruleSet, "helloworld.@example.com", errors.CodePattern)
}

// Requirements:
// - Errors when the local part is empty
func TestEmailEmptyLocal(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustNotRun(t, ruleSet, "@example.com", errors.CodePattern)
}

// Requirements:
// - Serializes to WithRequired()
func TestEmailRequiredString(t *testing.T) {
	ruleSet := net.NewEmail().WithRequired()

	expected := "EmailRuleSet.WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Context is passed to domain
func TestEmailDomainContext(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	ctx := rulecontext.WithPathString(context.Background(), "tests")
	ctx = rulecontext.WithPathString(ctx, "email")

	_, err := ruleSet.Run(ctx, "hello@example.bogusbogus")

	expected := "/tests/email"

	if err == nil {
		t.Error("Expected error to not be nil")
	} else if s := err.First().Path(); s != expected {
		t.Errorf("Expected path to be %s, got: %s", expected, s)
	}
}
