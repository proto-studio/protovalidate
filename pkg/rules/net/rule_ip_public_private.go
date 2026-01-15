package net

import (
	"context"
	"net"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// isPrivateIP checks if an IP address is in a private network range.
func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	// IPv4 private ranges
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
			(ip4[0] == 192 && ip4[1] == 168) ||
			(ip4[0] == 127) // loopback
	}

	// IPv6 private ranges
	// fc00::/7 - Unique Local Addresses
	// fe80::/10 - Link-Local Addresses
	// ::1 - loopback
	if len(ip) == net.IPv6len {
		return (ip[0] == 0xfc || ip[0] == 0xfd) ||
			(ip[0] == 0xfe && (ip[1]&0xc0) == 0x80) ||
			ip.Equal(net.IPv6loopback)
	}

	return false
}

// ipPublicPrivateRule implements the Rule interface for public/private IP validation.
type ipPublicPrivateRule struct {
	publicOnly  bool
	privateOnly bool
}

// Evaluate validates that the IP address matches the public/private requirement.
func (rule *ipPublicPrivateRule) Evaluate(ctx context.Context, ip net.IP) errors.ValidationErrorCollection {
	if ip == nil {
		return nil
	}

	isPrivate := isPrivateIP(ip)

	if rule.publicOnly && isPrivate {
		return errors.Collection(errors.Errorf(
			errors.CodePattern, ctx, "invalid format", "private IP addresses are not allowed",
		))
	}

	if rule.privateOnly && !isPrivate {
		return errors.Collection(errors.Errorf(
			errors.CodePattern, ctx, "invalid format", "public IP addresses are not allowed",
		))
	}

	return nil
}

// Replaces returns true for any public/private rule.
func (rule *ipPublicPrivateRule) Replaces(r rules.Rule[net.IP]) bool {
	_, ok := r.(*ipPublicPrivateRule)
	return ok
}

// String returns the string representation of the public/private rule.
func (rule *ipPublicPrivateRule) String() string {
	if rule.publicOnly {
		return "WithPublicOnly()"
	}
	if rule.privateOnly {
		return "WithPrivateOnly()"
	}
	return "WithPublicPrivate()"
}

// WithPublicOnly returns a new child rule set that only allows public IP addresses.
// Private IP addresses (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, loopback, etc.) are rejected.
func (ruleSet *IPRuleSet) WithPublicOnly() *IPRuleSet {
	return ruleSet.WithRule(&ipPublicPrivateRule{
		publicOnly:  true,
		privateOnly: false,
	})
}

// WithPrivateOnly returns a new child rule set that only allows private IP addresses.
// Public IP addresses are rejected.
func (ruleSet *IPRuleSet) WithPrivateOnly() *IPRuleSet {
	return ruleSet.WithRule(&ipPublicPrivateRule{
		publicOnly:  false,
		privateOnly: true,
	})
}
