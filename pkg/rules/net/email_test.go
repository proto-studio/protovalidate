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
	// Prepare the output variable for Apply
	var output string

	example := "hello@example.com"

	// Use Apply instead of Validate
	err := net.NewEmail().Apply(context.TODO(), example, &output)

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if output != example {
		t.Error("Expected test email to be returned")
		return
	}

	// Check if the rule set implements the expected interface
	ok := testhelpers.CheckRuleSetInterface[string](net.NewEmail())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	testhelpers.MustApplyTypes[string](t, net.NewEmail(), example)
}

// Requirements:
// - Default validator requires a TLD
// - Unknown TLDs error
func TestEmailDefaultDomain(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustApply(t, ruleSet, "hello@example.com")
	testhelpers.MustNotApply(t, ruleSet, "hello@example.bogusbogus", errors.CodePattern)
}

// Requirements:
// - Errors if there isn't any "@"
// - Errors if there is more than one "@"
func TestEmailSplit(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustNotApply(t, ruleSet, "example.com", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "hello@world@example.com", errors.CodePattern)
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

	// Prepare the output variable for Apply
	var output string

	// Apply with a mock rule that should trigger an error
	err := net.NewEmail().
		WithRuleFunc(mock.Function()).
		Apply(context.TODO(), "name@example.com", &output)

	if err == nil {
		t.Error("Expected errors to not be empty")
		return
	}

	if mock.EvaluateCallCount() != 1 {
		t.Errorf("Expected rule to be called 1 time, got %d", mock.EvaluateCallCount())
		return
	}

	// Check the string representation of the rule set
	str := net.NewEmail().WithRuleFunc(mock.Function()).String()
	expected := "EmailRuleSet.WithRuleFunc(...)"
	if str != expected {
		t.Errorf("Expected %s, got %s", expected, str)
	}

	rule := testhelpers.NewMockRule[string]()

	// Apply with a mock rule that should pass without errors
	err = net.NewEmail().
		WithRuleFunc(rule.Function()).
		Apply(context.TODO(), "name@example.com", &output)

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

// Requirements:
// - Custom domain RuleSet overrides default set.
func TestEmailWithDomain(t *testing.T) {
	domainRuleSet := net.NewDomain().WithSuffix("edu")
	ruleSet := net.NewEmail().WithDomain(domainRuleSet).Any()

	testhelpers.MustApply(t, ruleSet, "hello@example.edu")
	testhelpers.MustNotApply(t, ruleSet, "hello@example.com", errors.CodePattern)
}

// Requirements:
// - Errors when input is not a string
// - errors.CodeType is returned
func TestEmailType(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeType)
}

// Requirements:
// - No double dots
// - Can't start with a dot
// - Can't end with a dot
func TestEmailDots(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustApply(t, ruleSet, "hello.world@example.com")
	testhelpers.MustNotApply(t, ruleSet, "hello..world@example.com", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, ".helloworld@example.com", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "helloworld.@example.com", errors.CodePattern)
}

// Requirements:
// - Errors when the local part is empty
func TestEmailEmptyLocal(t *testing.T) {
	ruleSet := net.NewEmail().Any()

	testhelpers.MustNotApply(t, ruleSet, "@example.com", errors.CodePattern)
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

	// Prepare the output variable for Apply
	var output string

	// Use Apply instead of Run
	err := ruleSet.Apply(ctx, "hello@example.bogusbogus", &output)

	expected := "/tests/email"

	if err == nil {
		t.Error("Expected error to not be nil")
	} else if s := err.First().Path(); s != expected {
		t.Errorf("Expected path to be %s, got: %s", expected, s)
	}
}
