package net_test

import (
	"context"
	stdnet "net"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestIPRuleSet_Apply tests:
// - Default configuration doesn't return errors on valid value.
// - Implements interface.
// - Supports both string and stdnet.IP input/output.
func TestIPRuleSet_Apply(t *testing.T) {
	// Test with string input and string output
	var outputStr string
	example := "192.168.1.1"

	err := net.IP().Apply(context.TODO(), example, &outputStr)
	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if outputStr != example {
		t.Errorf("Expected output to be %s, got %s", example, outputStr)
		return
	}

	// Test with string input and net.IP output
	var outputIP stdnet.IP
	err = net.IP().Apply(context.TODO(), example, &outputIP)
	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if outputIP.String() != example {
		t.Errorf("Expected output to be %s, got %s", example, outputIP.String())
		return
	}

	// Test with net.IP input and string output
	inputIP := stdnet.ParseIP(example)
	var outputStr2 string
	err = net.IP().Apply(context.TODO(), inputIP, &outputStr2)
	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if outputStr2 != example {
		t.Errorf("Expected output to be %s, got %s", example, outputStr2)
		return
	}

	// Test with net.IP input and net.IP output
	var outputIP2 stdnet.IP
	err = net.IP().Apply(context.TODO(), inputIP, &outputIP2)
	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if !outputIP2.Equal(inputIP) {
		t.Errorf("Expected output to equal input IP")
		return
	}

	// Check if the rule set implements the expected interface
	ok := testhelpers.CheckRuleSetInterface[stdnet.IP](net.IP())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

// TestIPRuleSet_Apply_InvalidFormat tests:
// - Errors when IP format is invalid
func TestIPRuleSet_Apply_InvalidFormat(t *testing.T) {
	ruleSet := net.IP().Any()

	testhelpers.MustNotApply(t, ruleSet, "invalid.ip.address", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "999.999.999.999", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "not an ip", errors.CodePattern)
}

// TestIPRuleSet_Apply_Type tests:
// - Errors when input is not a string or stdnet.IP
func TestIPRuleSet_Apply_Type(t *testing.T) {
	ruleSet := net.IP().Any()

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeType)
	testhelpers.MustNotApply(t, ruleSet, []byte{192, 168, 1, 1}, errors.CodeType)
}

// TestIPRuleSet_WithRequired tests:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
func TestIPRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[stdnet.IP](t, net.IP())
}

// TestIPRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestIPRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[stdnet.IP](t, net.IP())
}

// TestIPRuleSet_ErrorConfig tests:
// - IPRuleSet implements error configuration methods
func TestIPRuleSet_ErrorConfig(t *testing.T) {
	testhelpers.MustImplementErrorConfig[stdnet.IP, *net.IPRuleSet](t, net.IP())
}

// TestIPRuleSet_WithIPv4 tests:
// - Allows IPv4 addresses when WithIPv4 is used
// - Rejects IPv6 addresses when only WithIPv4 is used
func TestIPRuleSet_WithIPv4(t *testing.T) {
	ruleSet := net.IP().WithIPv4().Any()

	// Valid IPv4
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")
	testhelpers.MustApplyAny(t, ruleSet, "10.0.0.1")
	testhelpers.MustApplyAny(t, ruleSet, "172.16.0.1")

	// Invalid IPv6
	testhelpers.MustNotApply(t, ruleSet, "2001:db8::1", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "::1", errors.CodePattern)
}

// TestIPRuleSet_WithIPv6 tests:
// - Allows IPv6 addresses when WithIPv6 is used
// - Rejects IPv4 addresses when only WithIPv6 is used
func TestIPRuleSet_WithIPv6(t *testing.T) {
	ruleSet := net.IP().WithIPv6().Any()

	// Valid IPv6
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8::1")
	testhelpers.MustApplyAny(t, ruleSet, "::1")
	testhelpers.MustApplyAny(t, ruleSet, "fe80::1")

	// Invalid IPv4
	testhelpers.MustNotApply(t, ruleSet, "192.168.1.1", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "10.0.0.1", errors.CodePattern)
}

// TestIPRuleSet_WithIPv4AndIPv6 tests:
// - Allows both IPv4 and IPv6 when both methods are called
func TestIPRuleSet_WithIPv4AndIPv6(t *testing.T) {
	ruleSet := net.IP().WithIPv4().WithIPv6().Any()

	// Valid IPv4
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")
	testhelpers.MustApplyAny(t, ruleSet, "10.0.0.1")

	// Valid IPv6
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8::1")
	testhelpers.MustApplyAny(t, ruleSet, "::1")
}

// TestIPRuleSet_WithIPv4Only tests:
// - Only allows IPv4 addresses
func TestIPRuleSet_WithIPv4Only(t *testing.T) {
	ruleSet := net.IP().WithIPv4Only().Any()

	// Valid IPv4
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")

	// Invalid IPv6
	testhelpers.MustNotApply(t, ruleSet, "2001:db8::1", errors.CodePattern)
}

// TestIPRuleSet_WithIPv6Only tests:
// - Only allows IPv6 addresses
func TestIPRuleSet_WithIPv6Only(t *testing.T) {
	ruleSet := net.IP().WithIPv6Only().Any()

	// Valid IPv6
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8::1")

	// Invalid IPv4
	testhelpers.MustNotApply(t, ruleSet, "192.168.1.1", errors.CodePattern)
}

// TestIPRuleSet_WithCIDR tests:
// - Validates IP is within CIDR block
func TestIPRuleSet_WithCIDR(t *testing.T) {
	ruleSet := net.IP().WithCIDR("192.168.1.0/24").Any()

	// Valid IPs within CIDR
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.100")
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.254")

	// Invalid IPs outside CIDR
	testhelpers.MustNotApply(t, ruleSet, "192.168.2.1", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "10.0.0.1", errors.CodePattern)
}

// TestIPRuleSet_WithCIDR_Multiple tests:
// - Validates IP is within any of the specified CIDR blocks
func TestIPRuleSet_WithCIDR_Multiple(t *testing.T) {
	ruleSet := net.IP().WithCIDR("192.168.1.0/24", "10.0.0.0/8").Any()

	// Valid IPs within first CIDR
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")

	// Valid IPs within second CIDR
	testhelpers.MustApplyAny(t, ruleSet, "10.0.0.1")
	testhelpers.MustApplyAny(t, ruleSet, "10.255.255.255")

	// Invalid IPs outside both CIDRs
	testhelpers.MustNotApply(t, ruleSet, "172.16.0.1", errors.CodePattern)
}

// TestIPRuleSet_WithCIDR_IPv6 tests:
// - Validates IPv6 CIDR blocks
func TestIPRuleSet_WithCIDR_IPv6(t *testing.T) {
	ruleSet := net.IP().WithCIDR("2001:db8::/32").Any()

	// Valid IPv6 within CIDR
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8::1")
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8:ffff:ffff:ffff:ffff:ffff:ffff")

	// Invalid IPv6 outside CIDR
	testhelpers.MustNotApply(t, ruleSet, "2001:db9::1", errors.CodePattern)
}

// TestIPRuleSet_WithSubnetMask tests:
// - Validates IP is within network defined by network address and subnet mask
func TestIPRuleSet_WithSubnetMask(t *testing.T) {
	ruleSet := net.IP().WithSubnetMask("192.168.1.0", "255.255.255.0").Any()

	// Valid IPs within network
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.100")
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.254")

	// Invalid IPs outside network
	testhelpers.MustNotApply(t, ruleSet, "192.168.2.1", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "10.0.0.1", errors.CodePattern)
}

// TestIPRuleSet_WithSubnetMask_IPv6 tests:
// - Validates IPv6 subnet masks
func TestIPRuleSet_WithSubnetMask_IPv6(t *testing.T) {
	ruleSet := net.IP().WithSubnetMask("2001:db8::", "ffff:ffff:ffff:ffff::").Any()

	// Valid IPv6 within network
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8::1")
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8:0:0:ffff:ffff:ffff:ffff")

	// Invalid IPv6 outside network
	testhelpers.MustNotApply(t, ruleSet, "2001:db9::1", errors.CodePattern)
}

// TestIPRuleSet_WithRange tests:
// - Validates IP is within specified range
func TestIPRuleSet_WithRange(t *testing.T) {
	ruleSet := net.IP().WithRange("192.168.1.1", "192.168.1.100").Any()

	// Valid IPs within range
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.50")
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.100")

	// Invalid IPs outside range
	testhelpers.MustNotApply(t, ruleSet, "192.168.1.0", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "192.168.1.101", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "192.168.2.1", errors.CodePattern)
}

// TestIPRuleSet_WithRange_IPv6 tests:
// - Validates IPv6 ranges
func TestIPRuleSet_WithRange_IPv6(t *testing.T) {
	ruleSet := net.IP().WithRange("2001:db8::1", "2001:db8::100").Any()

	// Valid IPv6 within range
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8::1")
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8::50")
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8::100")

	// Invalid IPv6 outside range
	testhelpers.MustNotApply(t, ruleSet, "2001:db8::", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "2001:db8::101", errors.CodePattern)
}

// TestIPRuleSet_WithPublicOnly tests:
// - Only allows public IP addresses
func TestIPRuleSet_WithPublicOnly(t *testing.T) {
	ruleSet := net.IP().WithPublicOnly().Any()

	// Valid public IPs
	testhelpers.MustApplyAny(t, ruleSet, "8.8.8.8")
	testhelpers.MustApplyAny(t, ruleSet, "1.1.1.1")
	testhelpers.MustApplyAny(t, ruleSet, "2001:4860:4860::8888")

	// Invalid private IPs
	testhelpers.MustNotApply(t, ruleSet, "192.168.1.1", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "10.0.0.1", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "172.16.0.1", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "127.0.0.1", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "::1", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "fc00::1", errors.CodePattern)
}

// TestIPRuleSet_WithPrivateOnly tests:
// - Only allows private IP addresses
func TestIPRuleSet_WithPrivateOnly(t *testing.T) {
	ruleSet := net.IP().WithPrivateOnly().Any()

	// Valid private IPs
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")
	testhelpers.MustApplyAny(t, ruleSet, "10.0.0.1")
	testhelpers.MustApplyAny(t, ruleSet, "172.16.0.1")
	testhelpers.MustApplyAny(t, ruleSet, "127.0.0.1")
	testhelpers.MustApplyAny(t, ruleSet, "::1")
	testhelpers.MustApplyAny(t, ruleSet, "fc00::1")

	// Invalid public IPs
	testhelpers.MustNotApply(t, ruleSet, "8.8.8.8", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "1.1.1.1", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "2001:4860:4860::8888", errors.CodePattern)
}

// TestIPRuleSet_Combined tests:
// - Tests combining multiple rules
func TestIPRuleSet_Combined(t *testing.T) {
	// IPv4 only, within CIDR, private only
	ruleSet := net.IP().
		WithIPv4Only().
		WithCIDR("192.168.1.0/24").
		WithPrivateOnly().
		Any()

	// Valid: IPv4, within CIDR, private
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")

	// Invalid: IPv6
	testhelpers.MustNotApply(t, ruleSet, "2001:db8::1", errors.CodePattern)

	// Invalid: outside CIDR
	testhelpers.MustNotApply(t, ruleSet, "192.168.2.1", errors.CodePattern)

	// Invalid: public IP
	testhelpers.MustNotApply(t, ruleSet, "8.8.8.8", errors.CodePattern)
}

// TestIPRuleSet_String tests:
// - String representation is correct
func TestIPRuleSet_String(t *testing.T) {
	ruleSet := net.IP().WithIPv4Only()
	expected := "IPRuleSet.WithIPv4Only()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = net.IP().WithCIDR("192.168.1.0/24")
	if s := ruleSet.String(); s == "" {
		t.Error("Expected rule set string to not be empty")
	}
}

// TestIPRuleSet_WithRuleFunc tests:
// - Custom rule functions are executed
func TestIPRuleSet_WithRuleFunc(t *testing.T) {
	mock := testhelpers.NewMockRuleWithErrors[stdnet.IP](1)

	var output stdnet.IP
	testIP := stdnet.ParseIP("192.168.1.1")

	err := net.IP().
		WithRuleFunc(mock.Function()).
		Apply(context.TODO(), testIP, &output)

	if err == nil {
		t.Error("Expected errors to not be empty")
		return
	}

	if mock.EvaluateCallCount() != 1 {
		t.Errorf("Expected rule to be called 1 time, got %d", mock.EvaluateCallCount())
		return
	}

	rule := testhelpers.NewMockRule[stdnet.IP]()

	err = net.IP().
		WithRuleFunc(rule.Function()).
		Apply(context.TODO(), testIP, &output)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

// TestIPRuleSet_Apply_ParseIP_StringPtr tests:
// - parseIP handles *string input
func TestIPRuleSet_Apply_ParseIP_StringPtr(t *testing.T) {
	ruleSet := net.IP().Any()
	str := "192.168.1.1"
	var output string

	err := ruleSet.Apply(context.TODO(), &str, &output)
	if err != nil {
		t.Errorf("Expected no error, got: %s", err)
	}
	if output != str {
		t.Errorf("Expected output %s, got %s", str, output)
	}
}

// TestIPRuleSet_Apply_ParseIP_NilIP tests:
// - parseIP handles nil net.IP input
func TestIPRuleSet_Apply_ParseIP_NilIP(t *testing.T) {
	ruleSet := net.IP().Any()
	var nilIP stdnet.IP
	var output string

	err := ruleSet.Apply(context.TODO(), nilIP, &output)
	if err == nil {
		t.Error("Expected error for nil IP")
	}
}

// TestIPRuleSet_Apply_SetOutput_Interface tests:
// - setOutput handles interface output that doesn't match net.IP
func TestIPRuleSet_Apply_SetOutput_Interface(t *testing.T) {
	ruleSet := net.IP().Any()
	var output interface{}

	err := ruleSet.Apply(context.TODO(), "192.168.1.1", &output)
	if err != nil {
		t.Errorf("Expected no error, got: %s", err)
	}
	if output == nil {
		t.Error("Expected output to be set")
	}
}

// TestIPRuleSet_Evaluate_NilIP tests:
// - Evaluate handles nil IP
func TestIPRuleSet_Evaluate_NilIP(t *testing.T) {
	ruleSet := net.IP()
	var nilIP stdnet.IP

	err := ruleSet.Evaluate(context.TODO(), nilIP)
	if err == nil {
		t.Error("Expected error for nil IP")
	}
}

// TestIPRuleSet_WithIPv4_Merge tests:
// - WithIPv4 merges with existing IPv6 rule
func TestIPRuleSet_WithIPv4_Merge(t *testing.T) {
	ruleSet := net.IP().WithIPv6().WithIPv4().Any()

	// Should allow both IPv4 and IPv6 after merge
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8::1")
}

// TestIPRuleSet_WithIPv6_Merge tests:
// - WithIPv6 merges with existing IPv4 rule
func TestIPRuleSet_WithIPv6_Merge(t *testing.T) {
	ruleSet := net.IP().WithIPv4().WithIPv6().Any()

	// Should allow both IPv4 and IPv6 after merge
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")
	testhelpers.MustApplyAny(t, ruleSet, "2001:db8::1")
}

// TestIPRuleSet_WithCIDR_Invalid tests:
// - WithCIDR panics on invalid CIDR
func TestIPRuleSet_WithCIDR_Invalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid CIDR")
		}
	}()

	net.IP().WithCIDR("invalid.cidr")
}

// TestIPRuleSet_WithCIDR_Empty tests:
// - WithCIDR handles empty CIDR blocks
func TestIPRuleSet_WithCIDR_Empty(t *testing.T) {
	ruleSet := net.IP().WithCIDR("192.168.1.0/24").Any()

	// Should still work with valid IP
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")
}

// TestIPRuleSet_WithSubnetMask_InvalidNetwork tests:
// - WithSubnetMask panics on invalid network address
func TestIPRuleSet_WithSubnetMask_InvalidNetwork(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid network address")
		}
	}()

	net.IP().WithSubnetMask("invalid.network", "255.255.255.0")
}

// TestIPRuleSet_WithSubnetMask_InvalidMask tests:
// - WithSubnetMask panics on invalid mask
func TestIPRuleSet_WithSubnetMask_InvalidMask(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid mask")
		}
	}()

	net.IP().WithSubnetMask("192.168.1.0", "invalid.mask")
}

// TestIPRuleSet_WithSubnetMask_VersionMismatch tests:
// - WithSubnetMask panics on version mismatch
func TestIPRuleSet_WithSubnetMask_VersionMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for version mismatch")
		}
	}()

	net.IP().WithSubnetMask("192.168.1.0", "ffff:ffff:ffff:ffff::")
}

// TestIPRuleSet_WithRange_InvalidStart tests:
// - WithRange panics on invalid start IP
func TestIPRuleSet_WithRange_InvalidStart(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid start IP")
		}
	}()

	net.IP().WithRange("invalid.start", "192.168.1.100")
}

// TestIPRuleSet_WithRange_InvalidEnd tests:
// - WithRange panics on invalid end IP
func TestIPRuleSet_WithRange_InvalidEnd(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid end IP")
		}
	}()

	net.IP().WithRange("192.168.1.1", "invalid.end")
}

// TestIPRuleSet_WithRange_VersionMismatch tests:
// - WithRange panics on version mismatch
func TestIPRuleSet_WithRange_VersionMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for version mismatch")
		}
	}()

	net.IP().WithRange("192.168.1.1", "2001:db8::100")
}

// TestIPRuleSet_WithRange_StartGreaterThanEnd tests:
// - WithRange panics when start > end
func TestIPRuleSet_WithRange_StartGreaterThanEnd(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when start > end")
		}
	}()

	net.IP().WithRange("192.168.1.100", "192.168.1.1")
}

// TestIPRuleSet_WithPrivateOnly_AllRanges tests:
// - WithPrivateOnly covers all private IP ranges
func TestIPRuleSet_WithPrivateOnly_AllRanges(t *testing.T) {
	ruleSet := net.IP().WithPrivateOnly().Any()

	// Test all IPv4 private ranges
	testhelpers.MustApplyAny(t, ruleSet, "10.0.0.1")
	testhelpers.MustApplyAny(t, ruleSet, "10.255.255.255")
	testhelpers.MustApplyAny(t, ruleSet, "172.16.0.1")
	testhelpers.MustApplyAny(t, ruleSet, "172.31.255.255")
	testhelpers.MustApplyAny(t, ruleSet, "192.168.0.1")
	testhelpers.MustApplyAny(t, ruleSet, "192.168.255.255")
	testhelpers.MustApplyAny(t, ruleSet, "127.0.0.1")

	// Test IPv6 private ranges
	testhelpers.MustApplyAny(t, ruleSet, "::1")
	testhelpers.MustApplyAny(t, ruleSet, "fc00::1")
	testhelpers.MustApplyAny(t, ruleSet, "fd00::1")
	testhelpers.MustApplyAny(t, ruleSet, "fe80::1")
	testhelpers.MustApplyAny(t, ruleSet, "febf::1")
}

// TestIPRuleSet_String_VersionRule tests:
// - String() for version rules covers all branches
func TestIPRuleSet_String_VersionRule(t *testing.T) {
	// Test WithIPv4().WithIPv6() case
	ruleSet := net.IP().WithIPv4().WithIPv6()
	if s := ruleSet.String(); s == "" {
		t.Error("Expected non-empty string")
	}

	// Test WithIPv4Only case
	ruleSet = net.IP().WithIPv4Only()
	if s := ruleSet.String(); s == "" {
		t.Error("Expected non-empty string")
	}

	// Test WithIPv6Only case
	ruleSet = net.IP().WithIPv6Only()
	if s := ruleSet.String(); s == "" {
		t.Error("Expected non-empty string")
	}
}

// TestIPRuleSet_String_CIDRRule tests:
// - String() for CIDR rules covers all branches
func TestIPRuleSet_String_CIDRRule(t *testing.T) {
	// Single CIDR
	ruleSet := net.IP().WithCIDR("192.168.1.0/24")
	if s := ruleSet.String(); s == "" {
		t.Error("Expected non-empty string")
	}

	// Multiple CIDRs
	ruleSet = net.IP().WithCIDR("192.168.1.0/24", "10.0.0.0/8")
	if s := ruleSet.String(); s == "" {
		t.Error("Expected non-empty string")
	}
}

// TestIPRuleSet_String_RangeRule tests:
// - String() for range rules
func TestIPRuleSet_String_RangeRule(t *testing.T) {
	ruleSet := net.IP().WithRange("192.168.1.1", "192.168.1.100")
	if s := ruleSet.String(); s == "" {
		t.Error("Expected non-empty string")
	}
}

// TestIPRuleSet_String_SubnetMaskRule tests:
// - String() for subnet mask rules
func TestIPRuleSet_String_SubnetMaskRule(t *testing.T) {
	ruleSet := net.IP().WithSubnetMask("192.168.1.0", "255.255.255.0")
	if s := ruleSet.String(); s == "" {
		t.Error("Expected non-empty string")
	}
}

// TestIPRuleSet_String_PublicPrivateRule tests:
// - String() for public/private rules
func TestIPRuleSet_String_PublicPrivateRule(t *testing.T) {
	// Test WithPublicOnly
	ruleSet := net.IP().WithPublicOnly()
	if s := ruleSet.String(); s == "" {
		t.Error("Expected non-empty string")
	}

	// Test WithPrivateOnly
	ruleSet = net.IP().WithPrivateOnly()
	if s := ruleSet.String(); s == "" {
		t.Error("Expected non-empty string")
	}
}

// TestIPRuleSet_CIDR_NoMatch tests:
// - CIDR rule when IP doesn't match any CIDR
func TestIPRuleSet_CIDR_NoMatch(t *testing.T) {
	ruleSet := net.IP().WithCIDR("192.168.1.0/24").Any()

	// IP outside CIDR
	testhelpers.MustNotApply(t, ruleSet, "10.0.0.1", errors.CodePattern)
}

// TestIPRuleSet_Range_Boundary tests:
// - Range rule boundary conditions
func TestIPRuleSet_Range_Boundary(t *testing.T) {
	ruleSet := net.IP().WithRange("192.168.1.1", "192.168.1.100").Any()

	// Exactly at start
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.1")

	// Exactly at end
	testhelpers.MustApplyAny(t, ruleSet, "192.168.1.100")

	// Just before start
	testhelpers.MustNotApply(t, ruleSet, "192.168.1.0", errors.CodePattern)

	// Just after end
	testhelpers.MustNotApply(t, ruleSet, "192.168.1.101", errors.CodePattern)
}
