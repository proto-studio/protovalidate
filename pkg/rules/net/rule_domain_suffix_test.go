package net_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestDomainRuleSet_WithSuffix tests:
// - Domains with the custom suffixes pass.
// - Domains using other suffixes fail.
// - Domains using partial suffixes fail.
// - Domains with no "plus one" fail.
func TestDomainRuleSet_WithSuffix(t *testing.T) {
	ruleSet := net.Domain().WithSuffix("test", "dev.local").Any()

	testhelpers.MustApply(t, ruleSet, "example.test")
	testhelpers.MustApply(t, ruleSet, "example.dev.local")
	testhelpers.MustNotApply(t, ruleSet, "example.local", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "dev.local", errors.CodePattern)
}

// TestDomainRuleSet_WithTLD tests:
// - Domains with a standard TLD pass.
// - Domains using a non-standard TLD do not pass.
// - Domains with no "plus one" fail.
func TestDomainRuleSet_WithTLD(t *testing.T) {
	ruleSet := net.Domain().WithTLD().Any()

	testhelpers.MustApply(t, ruleSet, "example.com")
	testhelpers.MustNotApply(t, ruleSet, "example.bogusbogus", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "com", errors.CodePattern)
}

// TestDomainRuleSet_WithSuffix_PunycodeError tests:
// - Panics when string cannot be encoded as punycode
func TestDomainRuleSet_WithSuffix_PunycodeError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	// idna: invalid label "é"
	str := "example.xn--é.com"
	net.Domain().WithSuffix(str)
}

// TestDomainRuleSet_String_WithSuffix tests:
// - Only one suffix list is preserved (conflict resolution).
// - WithSuffix will serialize up to 3 suffix values.
// - Suffix values are comma separated.
// - Suffix values are quoted.
// - If there are more than 3, the test " ... and X more" is used.
// - Suffix values are normalized.
func TestDomainRuleSet_String_WithSuffix(t *testing.T) {
	values := []string{
		"studio",
		"com",
		"😊",
		"edu",
		"co.uk",
	}

	// WithSuffix uses WithRule, so the label comes from the rule's String() method
	ruleSet := net.Domain().WithSuffix(values[0], values[1]).WithRequired()
	expected := "DomainRuleSet.WithSuffix(\"STUDIO\", \"COM\").WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Calling WithSuffix again should replace the previous suffix rule (conflict resolution)
	ruleSet = ruleSet.WithSuffix(values[0], values[1:3]...)
	expected = "DomainRuleSet.WithRequired().WithSuffix(\"STUDIO\", \"COM\", \"XN--O28H\")"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Calling WithSuffix again should replace the previous suffix rule
	ruleSet = ruleSet.WithSuffix(values[0], values[1:]...)
	expected = "DomainRuleSet.WithRequired().WithSuffix(\"STUDIO\", \"COM\", \"XN--O28H\" ... and 2 more)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
