package net_test

import (
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
	d, err := net.NewDomain().Validate("example.com")

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if d != "example.com" {
		t.Error("Expected test domain to be returned")
		return
	}

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

	testhelpers.MustBeValid(t, ruleSet, okLabel+".com", okLabel+".com")
	testhelpers.MustBeInvalid(t, ruleSet, badLabel+".com", errors.CodePattern)
}

// Requirements:
// - Errors when string cannot be encoded as punycode
func TestDomainPunycodeError(t *testing.T) {
	ruleSet := net.NewDomain().Any()

	// idna: invalid label "é"
	str := "example.xn--é.com"
	testhelpers.MustBeInvalid(t, ruleSet, str+".com", errors.CodePattern)
}

// Requirements:
// - Errors when domain is too long
// - errors.CodeMax is returned
func TestDomainLength(t *testing.T) {
	ruleSet := net.NewDomain().Any()

	str := strings.Repeat(strings.Repeat("a", 32), 9)
	testhelpers.MustBeInvalid(t, ruleSet, str+".com", errors.CodeMax)
}

// Requirements:
// - Errors when input is not a string
// - errors.CodeType is returned
func TestDomainType(t *testing.T) {
	ruleSet := net.NewDomain().Any()

	testhelpers.MustBeInvalid(t, ruleSet, 123, errors.CodeType)
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
	_, err := net.NewDomain().
		WithRuleFunc(testhelpers.MockCustomRule("example.com", 1)).
		Validate("example.com")

	if err == nil {
		t.Error("Expected errors to not be empty")
		return
	}

	expected := "example.com"

	actual, err := net.NewDomain().
		WithRuleFunc(testhelpers.MockCustomRule(expected, 0)).
		Validate("example.com")

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if expected != actual {
		t.Errorf("Expected '%s' to equal '%s'", actual, expected)
		return
	}
}
