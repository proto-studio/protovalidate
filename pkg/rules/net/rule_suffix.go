package net

import (
	"context"
	"strings"

	"golang.org/x/net/idna"
	"proto.zip/studio/validate/pkg/errors"
)

//go:generate go run ../../../scripts/get-tlds.go -o tlds.go

// Implements the Rule interface for domain validation.
type domainRule struct {
	suffix [][]string
}

// Evaluate takes a context and string value and returns an error if it does not appear to be a valid domain.
func (rule *domainRule) Evaluate(ctx context.Context, value string) (string, errors.ValidationErrorCollection) {
	// Convert to punycode
	punycode, _ := idna.ToASCII(value)

	parts := strings.Split(strings.ToUpper(punycode), ".")

	// Check against the suffix list.
	for _, suffix := range rule.suffix {
		if len(suffix) < len(parts) && compareSuffix(parts[len(parts)-len(suffix):], suffix) {
			return value, nil
		}
	}

	return value, errors.Collection(
		errors.Errorf(errors.CodePattern, ctx, "domain suffix does not match any valid suffixes"),
	)
}

// compareSuffix checks if two slices of strings are equal.
func compareSuffix(a, b []string) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// toSuffixList converts an array of strings into a list of suffixes.
func toSuffixList(input []string) [][]string {
	suffixList := make([][]string, len(input))

	for i := range input {
		punycode, err := idna.ToASCII(input[i])

		if err != nil {
			panic(err)
		}

		parts := strings.Split(strings.ToUpper(punycode), ".")
		suffixList[i] = parts
	}

	return suffixList
}

// WithSuffix returns a new child RuleSet that test to see if the domain has a matching suffix.
//
// This method takes one or more domain suffixes which will be used to validate against the domain.
// Suffix matching is case insensitive.
//
// The validated domain cannot be only the suffix, at least one additional subdomain must be included.
//
// This rule only performs tests against the pattern of the domain. It does not check if the domain is actually
// registered or if the DNS is correctly configured. Network access is not required.
//
// WithSuffix will panic is any of the suffix values are not valid domains themselves.
func (v *DomainRuleSet) WithSuffix(suffix string, rest ...string) *DomainRuleSet {
	list := make([]string, 0, 1+len(rest))
	list = append(list, suffix)
	list = append(list, rest...)

	suffixList := toSuffixList(list)

	return v.WithRule(&domainRule{
		suffixList,
	})
}

// WithTLD returns a new child RuleSet that ensures that the domain ends in a valid Top Level Domain (TLD).
//
// The domain is validated against the IANA list of Top Level Domains:
// http://data.iana.org/TLD/tlds-alpha-by-domain.txt
//
// Actively maintained versions will receive minor updates when the list of TLDs changes so if you use this
// method it is recommended that you periodically check for updates.
func (v *DomainRuleSet) WithTLD() *DomainRuleSet {
	return v.WithSuffix(TLDs[0], TLDs[1:]...)
}
