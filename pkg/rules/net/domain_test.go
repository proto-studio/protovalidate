package net_test

import (
	"context"
	"strings"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Default configuration doesn't return errors on valid value.
// - Implements interface.
func TestDomainRuleSet_Apply(t *testing.T) {
	// Prepare the output variable for Apply
	var output string

	example := "example.com"

	// Apply with a valid domain string
	err := net.Domain().Apply(context.TODO(), example, &output)

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if output != example {
		t.Error("Expected test domain to be returned")
		return
	}

	// Check if the rule set implements the expected interface
	ok := testhelpers.CheckRuleSetInterface[string](net.Domain())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	testhelpers.MustApplyTypes[string](t, net.Domain(), example)
}

// Requirements:
// - Segments (labels) cannot exceed 63 characters
// See: RFC 1035
func TestDomainRuleSet_Apply_SegmentLength(t *testing.T) {
	ruleSet := net.Domain().Any()

	okLabel := strings.Repeat("a", 63)
	badLabel := strings.Repeat("a", 64)

	testhelpers.MustApply(t, ruleSet, okLabel+".com")
	testhelpers.MustNotApply(t, ruleSet, badLabel+".com", errors.CodePattern)
}

// Requirements:
// - Errors when string cannot be encoded as punycode
func TestDomainRuleSet_Apply_PunycodeError(t *testing.T) {
	ruleSet := net.Domain().Any()

	// idna: invalid label "é"
	str := "example.xn--é.com"
	testhelpers.MustNotApply(t, ruleSet, str+".com", errors.CodePattern)
}

// Requirements:
// - Errors when domain is too long
// - errors.CodeMax is returned
func TestDomainRuleSet_Apply_Length(t *testing.T) {
	ruleSet := net.Domain().Any()

	str := strings.Repeat(strings.Repeat("a", 32), 9)
	testhelpers.MustNotApply(t, ruleSet, str+".com", errors.CodeMax)
}

// Requirements:
// - Errors when input is not a string
// - errors.CodeType is returned
func TestDomainRuleSet_Apply_Type(t *testing.T) {
	ruleSet := net.Domain().Any()

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeType)
}

// Requirements:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
func TestDomainRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[string](t, net.Domain())
}

func TestDomainRuleSet_WithRuleFunc(t *testing.T) {
	mock := testhelpers.NewMockRuleWithErrors[string](1)

	// Prepare the output variable for Apply
	var output string

	// Apply with a mock rule that should trigger an error
	err := net.Domain().
		WithRuleFunc(mock.Function()).
		Apply(context.TODO(), "example.com", &output)

	if err == nil {
		t.Error("Expected errors to not be empty")
		return
	}

	if mock.EvaluateCallCount() != 1 {
		t.Errorf("Expected rule to be called 1 time, got %d", mock.EvaluateCallCount())
		return
	}

	rule := testhelpers.NewMockRule[string]()

	// Apply with a mock rule that should pass without errors
	err = net.Domain().
		WithRuleFunc(rule.Function()).
		Apply(context.TODO(), "example.com", &output)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestDomainRuleSet_String_WithRequired(t *testing.T) {
	ruleSet := net.Domain().WithRequired()

	expected := "DomainRuleSet.WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestDomainRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[string](t, net.Domain())
}
