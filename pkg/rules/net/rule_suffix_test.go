package net_test

import (
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

	testhelpers.MustBeValid(t, ruleSet, "example.test", "example.test")
	testhelpers.MustBeValid(t, ruleSet, "example.dev.local", "example.dev.local")
	testhelpers.MustBeInvalid(t, ruleSet, "example.local", errors.CodePattern)
	testhelpers.MustBeInvalid(t, ruleSet, "dev.local", errors.CodePattern)
}

// Requirements:
// - Domains with a standard TLD pass.
// - Domains using a non-standard TLD do not pass.
// - Domains with no "plus one" fail.
func TestDomainWithTLD(t *testing.T) {
	ruleSet := net.NewDomain().WithTLD().Any()

	testhelpers.MustBeValid(t, ruleSet, "example.com", "example.com")
	testhelpers.MustBeInvalid(t, ruleSet, "example.bogusbogus", errors.CodePattern)
	testhelpers.MustBeInvalid(t, ruleSet, "com", errors.CodePattern)
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
