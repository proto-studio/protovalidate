package net

import (
	"context"
	"net"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// ipRangeRule implements the Rule interface for IP range validation.
type ipRangeRule struct {
	startIP net.IP
	endIP   net.IP
}

// Evaluate validates that the IP address is within the specified range (inclusive).
func (rule *ipRangeRule) Evaluate(ctx context.Context, ip net.IP) errors.ValidationErrorCollection {
	if ip == nil {
		return nil
	}

	// Compare IPs byte by byte
	if compareIPs(ip, rule.startIP) < 0 || compareIPs(ip, rule.endIP) > 0 {
		return errors.Collection(errors.Errorf(
			errors.CodePattern, ctx, "invalid format", "IP address is not within the allowed range",
		))
	}

	return nil
}

// compareIPs compares two IP addresses.
// Returns -1 if ip1 < ip2, 0 if ip1 == ip2, 1 if ip1 > ip2.
func compareIPs(ip1, ip2 net.IP) int {
	// Normalize to 16-byte representation
	ip1 = ip1.To16()
	ip2 = ip2.To16()

	if ip1 == nil || ip2 == nil {
		return 0
	}

	for i := 0; i < 16; i++ {
		if ip1[i] < ip2[i] {
			return -1
		}
		if ip1[i] > ip2[i] {
			return 1
		}
	}
	return 0
}

// Replaces returns true for any range rule.
func (rule *ipRangeRule) Replaces(r rules.Rule[net.IP]) bool {
	_, ok := r.(*ipRangeRule)
	return ok
}

// String returns the string representation of the range rule.
func (rule *ipRangeRule) String() string {
	return "WithRange(\"" + rule.startIP.String() + "\", \"" + rule.endIP.String() + "\")"
}

// WithRange returns a new child rule set that validates the IP address is within the specified range (inclusive).
// Both start and end IPs must be the same version (both IPv4 or both IPv6).
func (ruleSet *IPRuleSet) WithRange(startIP, endIP string) *IPRuleSet {
	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)

	if start == nil {
		panic("invalid start IP address: " + startIP)
	}
	if end == nil {
		panic("invalid end IP address: " + endIP)
	}

	// Ensure both IPs are the same version
	if (start.To4() != nil) != (end.To4() != nil) {
		panic("start and end IP addresses must be the same version (both IPv4 or both IPv6)")
	}

	// Ensure start <= end
	if compareIPs(start, end) > 0 {
		panic("start IP address must be less than or equal to end IP address")
	}

	return ruleSet.WithRule(&ipRangeRule{
		startIP: start,
		endIP:   end,
	})
}
