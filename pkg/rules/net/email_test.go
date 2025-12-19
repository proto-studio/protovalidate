package net_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestEmailRuleSet_Apply tests:
// - Default configuration doesn't return errors on valid value.
// - Implements interface.
func TestEmailRuleSet_Apply(t *testing.T) {
	// Prepare the output variable for Apply
	var output string

	example := "hello@example.com"

	// Use Apply instead of Validate
	err := net.Email().Apply(context.TODO(), example, &output)

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if output != example {
		t.Error("Expected test email to be returned")
		return
	}

	// Check if the rule set implements the expected interface
	ok := testhelpers.CheckRuleSetInterface[string](net.Email())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	testhelpers.MustApplyTypes[string](t, net.Email(), example)
}

// TestEmailRuleSet_Apply_DefaultDomain tests:
// - Default validator requires a TLD
// - Unknown TLDs error
func TestEmailRuleSet_Apply_DefaultDomain(t *testing.T) {
	ruleSet := net.Email().Any()

	testhelpers.MustApply(t, ruleSet, "hello@example.com")
	testhelpers.MustNotApply(t, ruleSet, "hello@example.bogusbogus", errors.CodePattern)
}

// TestEmailRuleSet_Apply_Split tests:
// - Errors if there isn't any "@"
// - Errors if there is more than one "@"
func TestEmailRuleSet_Apply_Split(t *testing.T) {
	ruleSet := net.Email().Any()

	testhelpers.MustNotApply(t, ruleSet, "example.com", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "hello@world@example.com", errors.CodePattern)
}

// TestEmailRuleSet_WithRequired tests:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
func TestEmailRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[string](t, net.Email())
}

// TestEmailRuleSet_WithRuleFunc tests:
// - Custom rule functions are executed
func TestEmailRuleSet_WithRuleFunc(t *testing.T) {
	mock := testhelpers.NewMockRuleWithErrors[string](1)

	// Prepare the output variable for Apply
	var output string

	// Apply with a mock rule that should trigger an error
	err := net.Email().
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
	str := net.Email().WithRuleFunc(mock.Function()).String()
	expected := "EmailRuleSet.WithRuleFunc(...)"
	if str != expected {
		t.Errorf("Expected %s, got %s", expected, str)
	}

	rule := testhelpers.NewMockRule[string]()

	// Apply with a mock rule that should pass without errors
	err = net.Email().
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

// TestEmailRuleSet_WithDomain tests:
// - Custom domain RuleSet overrides default set.
func TestEmailRuleSet_WithDomain(t *testing.T) {
	domainRuleSet := net.Domain().WithSuffix("edu")
	ruleSet := net.Email().WithDomain(domainRuleSet).Any()

	testhelpers.MustApply(t, ruleSet, "hello@example.edu")
	testhelpers.MustNotApply(t, ruleSet, "hello@example.com", errors.CodePattern)
}

// TestEmailRuleSet_Apply_Type tests:
// - Errors when input is not a string
// - errors.CodeType is returned
func TestEmailRuleSet_Apply_Type(t *testing.T) {
	ruleSet := net.Email().Any()

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeType)
}

// TestEmailRuleSet_Apply_Dots tests:
// - No double dots
// - Can't start with a dot
// - Can't end with a dot
func TestEmailRuleSet_Apply_Dots(t *testing.T) {
	ruleSet := net.Email().Any()

	testhelpers.MustApply(t, ruleSet, "hello.world@example.com")
	testhelpers.MustNotApply(t, ruleSet, "hello..world@example.com", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, ".helloworld@example.com", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "helloworld.@example.com", errors.CodePattern)
}

// TestEmailRuleSet_Apply_EmptyLocal tests:
// - Errors when the local part is empty
func TestEmailRuleSet_Apply_EmptyLocal(t *testing.T) {
	ruleSet := net.Email().Any()

	testhelpers.MustNotApply(t, ruleSet, "@example.com", errors.CodePattern)
}

// TestEmailRuleSet_String_WithRequired tests:
// - Serializes to WithRequired()
func TestEmailRuleSet_String_WithRequired(t *testing.T) {
	ruleSet := net.Email().WithRequired()

	expected := "EmailRuleSet.WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestEmailRuleSet_DomainContext tests:
// - Context is passed to domain
func TestEmailRuleSet_DomainContext(t *testing.T) {
	ruleSet := net.Email().Any()

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

// TestEmailRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestEmailRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[string](t, net.Email())
}
