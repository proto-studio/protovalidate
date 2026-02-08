package net

import (
	"context"
	"net"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// ipVersionRule implements the Rule interface for IP version validation.
type ipVersionRule struct {
	allowIPv4 bool
	allowIPv6 bool
}

// Evaluate validates that the IP address matches the allowed IP version(s).
func (rule *ipVersionRule) Evaluate(ctx context.Context, ip net.IP) errors.ValidationError {
	if ip == nil {
		return nil
	}

	isIPv4 := ip.To4() != nil
	isIPv6 := !isIPv4 && ip.To16() != nil

	if isIPv4 && !rule.allowIPv4 {
		return errors.Errorf(errors.CodePattern, ctx, "invalid format", "IPv4 addresses are not allowed")
	}
	if isIPv6 && !rule.allowIPv6 {
		return errors.Errorf(errors.CodePattern, ctx, "invalid format", "IPv6 addresses are not allowed")
	}

	return nil
}

// Replaces returns true for any IP version rule.
func (rule *ipVersionRule) Replaces(r rules.Rule[net.IP]) bool {
	_, ok := r.(*ipVersionRule)
	return ok
}

// String returns the string representation of the IP version rule.
func (rule *ipVersionRule) String() string {
	if rule.allowIPv4 && rule.allowIPv6 {
		return "WithIPv4().WithIPv6()"
	} else if rule.allowIPv4 {
		return "WithIPv4Only()"
	} else if rule.allowIPv6 {
		return "WithIPv6Only()"
	}
	return "WithIPVersion()"
}

// WithIPv4 returns a new child rule set that allows IPv4 addresses.
// This method can be combined with WithIPv6 to allow both IPv4 and IPv6.
func (ruleSet *IPRuleSet) WithIPv4() *IPRuleSet {
	// Check if there's already an IPv6 rule in the parent chain
	current := ruleSet
	for current != nil {
		if current.rule != nil {
			if vRule, ok := current.rule.(*ipVersionRule); ok && vRule.allowIPv6 {
				// Merge: allow both IPv4 and IPv6
				return ruleSet.WithRule(&ipVersionRule{
					allowIPv4: true,
					allowIPv6: true,
				})
			}
		}
		current = current.parent
	}
	return ruleSet.WithRule(&ipVersionRule{
		allowIPv4: true,
		allowIPv6: false,
	})
}

// WithIPv6 returns a new child rule set that allows IPv6 addresses.
// This method can be combined with WithIPv4 to allow both IPv4 and IPv6.
func (ruleSet *IPRuleSet) WithIPv6() *IPRuleSet {
	// Check if there's already an IPv4 rule in the parent chain
	current := ruleSet
	for current != nil {
		if current.rule != nil {
			if vRule, ok := current.rule.(*ipVersionRule); ok && vRule.allowIPv4 {
				// Merge: allow both IPv4 and IPv6
				return ruleSet.WithRule(&ipVersionRule{
					allowIPv4: true,
					allowIPv6: true,
				})
			}
		}
		current = current.parent
	}
	return ruleSet.WithRule(&ipVersionRule{
		allowIPv4: false,
		allowIPv6: true,
	})
}

// WithIPv4Only returns a new child rule set that only allows IPv4 addresses.
func (ruleSet *IPRuleSet) WithIPv4Only() *IPRuleSet {
	return ruleSet.WithRule(&ipVersionRule{
		allowIPv4: true,
		allowIPv6: false,
	})
}

// WithIPv6Only returns a new child rule set that only allows IPv6 addresses.
func (ruleSet *IPRuleSet) WithIPv6Only() *IPRuleSet {
	return ruleSet.WithRule(&ipVersionRule{
		allowIPv4: false,
		allowIPv6: true,
	})
}
