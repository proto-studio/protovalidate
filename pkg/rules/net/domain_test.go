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
func TestDomainRuleSet(t *testing.T) {
	// Prepare the output variable for Apply
	var output string

	// Apply with a valid domain string
	err := net.NewDomain().Apply(context.TODO(), "example.com", &output)

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if output != "example.com" {
		t.Error("Expected test domain to be returned")
		return
	}

	// Check if the rule set implements the expected interface
	ok := testhelpers.CheckRuleSetInterface[string](net.NewDomain())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

// Requirements:
// - Segments (labels) cannot exceed 63 characters
// See: RFC 1035
func TestDomainSegmentLength(t *testing.T) {
	ruleSet := net.NewDomain().Any()

	okLabel := strings.Repeat("a", 63)
	badLabel := strings.Repeat("a", 64)

	testhelpers.MustRun(t, ruleSet, okLabel+".com")
	testhelpers.MustNotRun(t, ruleSet, badLabel+".com", errors.CodePattern)
}

// Requirements:
// - Errors when string cannot be encoded as punycode
func TestDomainPunycodeError(t *testing.T) {
	ruleSet := net.NewDomain().Any()

	// idna: invalid label "é"
	str := "example.xn--é.com"
	testhelpers.MustNotRun(t, ruleSet, str+".com", errors.CodePattern)
}

// Requirements:
// - Errors when domain is too long
// - errors.CodeMax is returned
func TestDomainLength(t *testing.T) {
	ruleSet := net.NewDomain().Any()

	str := strings.Repeat(strings.Repeat("a", 32), 9)
	testhelpers.MustNotRun(t, ruleSet, str+".com", errors.CodeMax)
}

// Requirements:
// - Errors when input is not a string
// - errors.CodeType is returned
func TestDomainType(t *testing.T) {
	ruleSet := net.NewDomain().Any()

	testhelpers.MustNotRun(t, ruleSet, 123, errors.CodeType)
}

// Requirements:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
func TestDomainRequired(t *testing.T) {
	ruleSet := net.NewDomain()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestDomainCustom(t *testing.T) {
	mock := testhelpers.NewMockRuleWithErrors[string](1)

	// Prepare the output variable for Apply
	var output string

	// Apply with a mock rule that should trigger an error
	err := net.NewDomain().
		WithRuleFunc(mock.Function()).
		Apply(context.TODO(), "example.com", &output)

	if err == nil {
		t.Error("Expected errors to not be empty")
		return
	}

	if mock.CallCount() != 1 {
		t.Errorf("Expected rule to be called 1 time, got %d", mock.CallCount())
		return
	}

	rule := testhelpers.NewMockRule[string]()

	// Apply with a mock rule that should pass without errors
	err = net.NewDomain().
		WithRuleFunc(rule.Function()).
		Apply(context.TODO(), "example.com", &output)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if c := rule.CallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestDomainRequiredString(t *testing.T) {
	ruleSet := net.NewDomain().WithRequired()

	expected := "DomainRuleSet.WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
