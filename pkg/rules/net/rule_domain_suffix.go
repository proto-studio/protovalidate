package net

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/net/idna"
	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

//go:generate go run ../../../_scripts/get-tlds.go -o tlds.go

// Implements the Rule interface for domain validation.
type domainSuffixRule struct {
	suffix [][]string
}

// Evaluate takes a context and string value and returns an error if it does not appear to be a valid domain.
func (rule *domainSuffixRule) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	// Convert to punycode
	punycode, _ := idna.ToASCII(value)

	parts := strings.Split(strings.ToUpper(punycode), ".")

	// Check against the suffix list.
	for _, suffix := range rule.suffix {
		if len(suffix) < len(parts) && compareSuffix(parts[len(parts)-len(suffix):], suffix) {
			return nil
		}
	}

	return errors.Collection(
		errors.Errorf(errors.CodePattern, ctx, "invalid format", "domain suffix is not valid"),
	)
}

// Replaces returns true for any suffix rule.
func (rule *domainSuffixRule) Replaces(x rules.Rule[string]) bool {
	_, ok := x.(*domainSuffixRule)
	return ok
}

// String returns the string representation of the domain suffix rule.
// Example: WithSuffix("com", "org")
func (rule *domainSuffixRule) String() string {
	l := len(rule.suffix)

	var sb strings.Builder
	sb.WriteString("WithSuffix(")

	// Append up to the first 3 strings or the total number of strings if less than 3
	for i := 0; i < 3 && i < l; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		suffixStr := strings.Join(rule.suffix[i], ".")
		truncated := util.TruncateString(suffixStr)
		sb.WriteString(fmt.Sprintf(`"%s"`, truncated))
	}

	// If there are more than 3 strings, append the "... and X more" message
	if l > 3 {
		sb.WriteString(fmt.Sprintf(" ... and %d more", l-3))
	}

	sb.WriteString(")")

	return sb.String()
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
// This rule only performs tests against the text of the domain. It does not check if the domain is actually
// registered or if the DNS is correctly configured. Network access is not required.
//
// WithSuffix will panic is any of the suffix values are not valid domains themselves.
func (v *DomainRuleSet) WithSuffix(suffix string, rest ...string) *DomainRuleSet {
	list := make([]string, 0, 1+len(rest))
	list = append(list, suffix)
	list = append(list, rest...)

	suffixList := toSuffixList(list)

	return v.WithRule(&domainSuffixRule{
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
