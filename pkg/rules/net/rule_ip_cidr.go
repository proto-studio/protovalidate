package net

import (
	"context"
	"net"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// ipCIDRRule implements the Rule interface for CIDR block validation.
type ipCIDRRule struct {
	cidrBlocks []*net.IPNet
}

// Evaluate validates that the IP address is within one of the allowed CIDR blocks.
func (rule *ipCIDRRule) Evaluate(ctx context.Context, ip net.IP) errors.ValidationErrorCollection {
	if ip == nil {
		return nil
	}

	for _, cidr := range rule.cidrBlocks {
		if cidr.Contains(ip) {
			return nil
		}
	}

	return errors.Collection(errors.Errorf(
		errors.CodePattern, ctx, "invalid format", "IP address is not within the allowed CIDR block(s)",
	))
}

// Replaces returns true for any CIDR rule.
func (rule *ipCIDRRule) Replaces(r rules.Rule[net.IP]) bool {
	_, ok := r.(*ipCIDRRule)
	return ok
}

// String returns the string representation of the CIDR rule.
func (rule *ipCIDRRule) String() string {
	if len(rule.cidrBlocks) == 0 {
		return "WithCIDR()"
	}
	if len(rule.cidrBlocks) == 1 {
		return "WithCIDR(\"" + rule.cidrBlocks[0].String() + "\")"
	}
	return "WithCIDR(...)"
}

// WithCIDR returns a new child rule set that validates the IP address is within one of the specified CIDR blocks.
// The CIDR blocks can be specified as strings (e.g., "192.168.1.0/24" or "2001:db8::/32").
func (ruleSet *IPRuleSet) WithCIDR(cidr string, rest ...string) *IPRuleSet {
	cidrBlocks := make([]*net.IPNet, 0, 1+len(rest))
	allCIDRs := append([]string{cidr}, rest...)

	for _, cidrStr := range allCIDRs {
		_, ipNet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			panic("invalid CIDR block: " + cidrStr)
		}
		cidrBlocks = append(cidrBlocks, ipNet)
	}

	return ruleSet.WithRule(&ipCIDRRule{
		cidrBlocks: cidrBlocks,
	})
}
