package net

import (
	"context"
	"net"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// ipSubnetMaskRule implements the Rule interface for subnet mask validation.
type ipSubnetMaskRule struct {
	network    net.IP
	subnetMask net.IPMask
}

// Evaluate validates that the IP address is within the network defined by the network address and subnet mask.
func (rule *ipSubnetMaskRule) Evaluate(ctx context.Context, ip net.IP) errors.ValidationError {
	if ip == nil {
		return nil
	}

	// Create the network from the base network address and mask
	network := net.IPNet{
		IP:   rule.network,
		Mask: rule.subnetMask,
	}

	if !network.Contains(ip) {
		return errors.Errorf(errors.CodePattern, ctx, "invalid format", "IP address is not within the specified network")
	}

	return nil
}

// Replaces returns true for any subnet mask rule.
func (rule *ipSubnetMaskRule) Replaces(r rules.Rule[net.IP]) bool {
	_, ok := r.(*ipSubnetMaskRule)
	return ok
}

// String returns the string representation of the subnet mask rule.
func (rule *ipSubnetMaskRule) String() string {
	return "WithSubnetMask(\"" + rule.network.String() + "\", \"" + net.IP(rule.subnetMask).String() + "\")"
}

// WithSubnetMask returns a new child rule set that validates the IP address is within the network
// defined by the network address and subnet mask.
// The network address and subnet mask can be specified as strings
// (e.g., network "192.168.1.0" with mask "255.255.255.0" for IPv4,
// or network "2001:db8::" with mask "ffff:ffff:ffff:ffff::" for IPv6).
func (ruleSet *IPRuleSet) WithSubnetMask(networkAddr, mask string) *IPRuleSet {
	networkIP := net.ParseIP(networkAddr)
	if networkIP == nil {
		panic("invalid network address: " + networkAddr)
	}

	maskIP := net.ParseIP(mask)
	if maskIP == nil {
		panic("invalid subnet mask: " + mask)
	}

	var ipMask net.IPMask
	if maskIP.To4() != nil {
		ipMask = net.IPv4Mask(maskIP[12], maskIP[13], maskIP[14], maskIP[15])
	} else {
		ipMask = net.IPMask(maskIP)
	}

	// Ensure network and mask are the same version
	if (networkIP.To4() != nil) != (maskIP.To4() != nil) {
		panic("network address and subnet mask must be the same version (both IPv4 or both IPv6)")
	}

	return ruleSet.WithRule(&ipSubnetMaskRule{
		network:    networkIP,
		subnetMask: ipMask,
	})
}
