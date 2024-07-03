package net_test

import (
	"fmt"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Domains with the custom suffixes pass.
// - Domains using other suffixes fail.
// - Domains using partial suffixes fail.
// - Domains with no "plus one" fail.
func TestDomainWithSuffix(t *testing.T) {
	ruleSet := net.NewDomain().WithSuffix("test", "dev.local").Any()

	testhelpers.MustRun(t, ruleSet, "example.test")
	testhelpers.MustRun(t, ruleSet, "example.dev.local")
	testhelpers.MustNotRun(t, ruleSet, "example.local", errors.CodePattern)
	testhelpers.MustNotRun(t, ruleSet, "dev.local", errors.CodePattern)
}

// Requirements:
// - Domains with a standard TLD pass.
// - Domains using a non-standard TLD do not pass.
// - Domains with no "plus one" fail.
func TestDomainWithTLD(t *testing.T) {
	ruleSet := net.NewDomain().WithTLD().Any()

	testhelpers.MustRun(t, ruleSet, "example.com")
	testhelpers.MustNotRun(t, ruleSet, "example.bogusbogus", errors.CodePattern)
	testhelpers.MustNotRun(t, ruleSet, "com", errors.CodePattern)
}

// Requirements:
// - Panics when string cannot be encoded as punycode
func TestSuffixPunycodeError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	// idna: invalid label "é"
	str := "example.xn--é.com"
	net.NewDomain().WithSuffix(str)
}

// Requirements:
// - Only one suffix list is preserved.
// - WithSuffix will serialize up to 3 suffix values.
// - Suffix values are comma separated.
// - Suffix values are quoted.
// - If there are more than 3, the test " ... and X more" is used.
// - Suffix values are normalized.
func TestLayoutsSerialize(t *testing.T) {
	values := []string{
		"studio",
		"com",
		"😊",
		"edu",
		"co.uk",
	}

	ruleSet := net.NewDomain().WithSuffix(values[0], values[1]).WithRequired()
	expected := fmt.Sprintf("DomainRuleSet.WithSuffix(\"STUDIO\", \"COM\").WithRequired()")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = ruleSet.WithSuffix(values[0], values[1:3]...)
	expected = fmt.Sprintf("DomainRuleSet.WithRequired().WithSuffix(\"STUDIO\", \"COM\", \"XN--O28H\")")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = ruleSet.WithSuffix(values[0], values[1:]...)
	expected = fmt.Sprintf("DomainRuleSet.WithRequired().WithSuffix(\"STUDIO\", \"COM\", \"XN--O28H\" ... and 2 more)")
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
