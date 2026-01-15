package net

import (
	"context"
	"net"
	"testing"
)

// TestIPRuleSet_NoConflict_WithNilParent tests:
//   - noConflict correctly handles the case where a ruleset with parent == nil conflicts
func TestIPRuleSet_NoConflict_WithNilParent(t *testing.T) {
	// Create a ruleset directly (not using factory) with parent == nil and conflictType set
	directRuleSet := &IPRuleSet{
		parent:       nil,
		conflictType: ipConflictTypeRequired,
	}

	// Create a conflict checker that will match the conflictType
	checker := ipConflictTypeRequired

	// Call noConflict - it should return nil because the ruleset conflicts and has no parent
	result := directRuleSet.noConflict(checker)
	if result != nil {
		t.Errorf("Expected noConflict to return nil when ruleset with parent == nil conflicts, got %v", result)
	}
}

// TestIPRuleSet_NoConflict_ParentChanges tests:
//   - noConflict correctly handles when parent changes
func TestIPRuleSet_NoConflict_ParentChanges(t *testing.T) {
	// Create a chain: grandparent -> parent -> child
	// grandparent has a conflict, parent doesn't, child doesn't
	grandparent := &IPRuleSet{
		parent:       nil,
		conflictType: ipConflictTypeRequired,
		label:        "grandparent",
	}
	parent := &IPRuleSet{
		parent:       grandparent,
		conflictType: ipConflictTypeNone,
		label:        "parent",
	}
	child := &IPRuleSet{
		parent:       parent,
		conflictType: ipConflictTypeNone,
		label:        "child",
	}

	// Create a conflict checker that matches grandparent's conflictType
	checker := ipConflictTypeRequired

	// Call noConflict on child - grandparent should be removed, parent changed
	result := child.noConflict(checker)
	if result == nil {
		t.Error("Expected non-nil result")
		return
	}
	// The result should have a different parent than the original
	if result.parent == child.parent {
		t.Error("Expected parent to change after conflict removal")
	}
}

// TestIPRuleSet_Replaces_Success tests:
//   - ipConflictType.Replaces returns true when cast succeeds and conflicts
func TestIPRuleSet_Replaces_Success(t *testing.T) {
	ruleSet := &IPRuleSet{
		conflictType: ipConflictTypeRequired,
	}

	checker := ipConflictTypeRequired
	if !checker.Replaces(ruleSet) {
		t.Error("Expected Replaces to return true for matching conflict type")
	}
}

// TestIPRuleSet_Replaces_NoConflict tests:
//   - ipConflictType.Replaces returns false when conflict types don't match
func TestIPRuleSet_Replaces_NoConflict(t *testing.T) {
	ruleSet := &IPRuleSet{
		conflictType: ipConflictTypeNil,
	}

	checker := ipConflictTypeRequired
	if checker.Replaces(ruleSet) {
		t.Error("Expected Replaces to return false for non-matching conflict type")
	}
}

// TestParseIP_InvalidStringPointer tests:
//   - parseIP handles invalid *string pointer
func TestParseIP_InvalidStringPointer(t *testing.T) {
	invalidIP := "not.a.valid.ip"
	_, err := parseIP(context.TODO(), &invalidIP)
	if err == nil {
		t.Error("Expected error for invalid IP string pointer")
	}
}

// TestSetOutput_InvalidOutputType tests:
//   - setOutput returns error for invalid output type
func TestSetOutput_InvalidOutputType(t *testing.T) {
	ip := net.ParseIP("192.168.1.1")
	var output int // Invalid output type

	err := setOutput(context.TODO(), &output, ip)
	if err == nil {
		t.Error("Expected error for invalid output type")
	}
}

// TestSetOutput_NilPointer tests:
//   - setOutput returns error for nil pointer
func TestSetOutput_NilPointer(t *testing.T) {
	ip := net.ParseIP("192.168.1.1")

	err := setOutput(context.TODO(), nil, ip)
	if err == nil {
		t.Error("Expected error for nil pointer")
	}
}

// TestSetOutput_NonPointer tests:
//   - setOutput returns error for non-pointer
func TestSetOutput_NonPointer(t *testing.T) {
	ip := net.ParseIP("192.168.1.1")
	output := "not a pointer"

	err := setOutput(context.TODO(), output, ip)
	if err == nil {
		t.Error("Expected error for non-pointer")
	}
}

// TestIPVersionRule_Evaluate_NilIP tests:
//   - ipVersionRule.Evaluate handles nil IP
func TestIPVersionRule_Evaluate_NilIP(t *testing.T) {
	rule := &ipVersionRule{allowIPv4: true, allowIPv6: true}
	err := rule.Evaluate(context.TODO(), nil)
	if err != nil {
		t.Error("Expected no error for nil IP")
	}
}

// TestIPVersionRule_String_NoVersion tests:
//   - ipVersionRule.String handles case where neither version is allowed
func TestIPVersionRule_String_NoVersion(t *testing.T) {
	rule := &ipVersionRule{allowIPv4: false, allowIPv6: false}
	s := rule.String()
	if s != "WithIPVersion()" {
		t.Errorf("Expected WithIPVersion(), got %s", s)
	}
}

// TestIPCIDRRule_Evaluate_NilIP tests:
//   - ipCIDRRule.Evaluate handles nil IP
func TestIPCIDRRule_Evaluate_NilIP(t *testing.T) {
	_, ipNet, _ := net.ParseCIDR("192.168.1.0/24")
	rule := &ipCIDRRule{cidrBlocks: []*net.IPNet{ipNet}}
	err := rule.Evaluate(context.TODO(), nil)
	if err != nil {
		t.Error("Expected no error for nil IP")
	}
}

// TestIPCIDRRule_String_Empty tests:
//   - ipCIDRRule.String handles empty CIDR blocks
func TestIPCIDRRule_String_Empty(t *testing.T) {
	rule := &ipCIDRRule{cidrBlocks: []*net.IPNet{}}
	s := rule.String()
	if s != "WithCIDR()" {
		t.Errorf("Expected WithCIDR(), got %s", s)
	}
}

// TestIPSubnetMaskRule_Evaluate_NilIP tests:
//   - ipSubnetMaskRule.Evaluate handles nil IP
func TestIPSubnetMaskRule_Evaluate_NilIP(t *testing.T) {
	rule := &ipSubnetMaskRule{
		network:    net.ParseIP("192.168.1.0"),
		subnetMask: net.IPv4Mask(255, 255, 255, 0),
	}
	err := rule.Evaluate(context.TODO(), nil)
	if err != nil {
		t.Error("Expected no error for nil IP")
	}
}

// TestIPRangeRule_Evaluate_NilIP tests:
//   - ipRangeRule.Evaluate handles nil IP
func TestIPRangeRule_Evaluate_NilIP(t *testing.T) {
	rule := &ipRangeRule{
		startIP: net.ParseIP("192.168.1.1"),
		endIP:   net.ParseIP("192.168.1.100"),
	}
	err := rule.Evaluate(context.TODO(), nil)
	if err != nil {
		t.Error("Expected no error for nil IP")
	}
}

// TestCompareIPs_NilIP tests:
//   - compareIPs handles nil IP
func TestCompareIPs_NilIP(t *testing.T) {
	ip := net.ParseIP("192.168.1.1")

	// Test nil first arg
	result := compareIPs(nil, ip)
	if result != 0 {
		t.Errorf("Expected 0 for nil IP comparison, got %d", result)
	}

	// Test nil second arg
	result = compareIPs(ip, nil)
	if result != 0 {
		t.Errorf("Expected 0 for nil IP comparison, got %d", result)
	}
}

// TestIPPublicPrivateRule_Evaluate_NilIP tests:
//   - ipPublicPrivateRule.Evaluate handles nil IP
func TestIPPublicPrivateRule_Evaluate_NilIP(t *testing.T) {
	rule := &ipPublicPrivateRule{publicOnly: true}
	err := rule.Evaluate(context.TODO(), nil)
	if err != nil {
		t.Error("Expected no error for nil IP")
	}
}

// TestIsPrivateIP_NilIP tests:
//   - isPrivateIP handles nil IP
func TestIsPrivateIP_NilIP(t *testing.T) {
	if isPrivateIP(nil) {
		t.Error("Expected false for nil IP")
	}
}

// TestIsPrivateIP_PublicIPv4 tests:
//   - isPrivateIP returns false for public IPv4
func TestIsPrivateIP_PublicIPv4(t *testing.T) {
	publicIP := net.ParseIP("8.8.8.8")
	if isPrivateIP(publicIP) {
		t.Error("Expected false for public IPv4")
	}
}

// TestIsPrivateIP_PublicIPv6 tests:
//   - isPrivateIP returns false for public IPv6
func TestIsPrivateIP_PublicIPv6(t *testing.T) {
	publicIP := net.ParseIP("2001:4860:4860::8888")
	if isPrivateIP(publicIP) {
		t.Error("Expected false for public IPv6")
	}
}

// TestIPPublicPrivateRule_String tests:
//   - String method returns correct values for all states
func TestIPPublicPrivateRule_String(t *testing.T) {
	// Test default (neither public nor private only)
	rule := &ipPublicPrivateRule{publicOnly: false, privateOnly: false}
	s := rule.String()
	if s != "WithPublicPrivate()" {
		t.Errorf("Expected WithPublicPrivate(), got %s", s)
	}

	// Test public only
	rule = &ipPublicPrivateRule{publicOnly: true, privateOnly: false}
	s = rule.String()
	if s != "WithPublicOnly()" {
		t.Errorf("Expected WithPublicOnly(), got %s", s)
	}

	// Test private only
	rule = &ipPublicPrivateRule{publicOnly: false, privateOnly: true}
	s = rule.String()
	if s != "WithPrivateOnly()" {
		t.Errorf("Expected WithPrivateOnly(), got %s", s)
	}
}

// TestIPRangeRule_String tests:
//   - String method returns correct value
func TestIPRangeRule_String(t *testing.T) {
	rule := &ipRangeRule{
		startIP: net.ParseIP("192.168.1.1"),
		endIP:   net.ParseIP("192.168.1.100"),
	}
	s := rule.String()
	expected := `WithRange("192.168.1.1", "192.168.1.100")`
	if s != expected {
		t.Errorf("Expected %s, got %s", expected, s)
	}
}

// TestIPSubnetMaskRule_String tests:
//   - String method returns correct value
func TestIPSubnetMaskRule_String(t *testing.T) {
	rule := &ipSubnetMaskRule{
		network:    net.ParseIP("192.168.1.0"),
		subnetMask: net.IPv4Mask(255, 255, 255, 0),
	}
	s := rule.String()
	if s == "" {
		t.Error("Expected non-empty string")
	}
}

// TestNoConflict_RuleConflict tests:
//   - noConflict handles conflict via rule.Replaces
func TestNoConflict_RuleConflict(t *testing.T) {
	// Create a mock rule that replaces other rules of same type
	rule := &ipVersionRule{allowIPv4: true}

	parent := &IPRuleSet{
		parent: nil,
		rule:   rule,
		label:  "parent",
	}

	child := &IPRuleSet{
		parent: parent,
		label:  "child",
	}

	// The rule should replace other ipVersionRule instances
	newRule := &ipVersionRule{allowIPv6: true}
	result := child.noConflict(newRule)

	// Result should exist (child doesn't conflict)
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

// TestIPVersionRule_Replaces tests:
//   - Replaces returns correct values for different rule types
func TestIPVersionRule_Replaces(t *testing.T) {
	rule := &ipVersionRule{allowIPv4: true}

	// Should replace another ipVersionRule
	other := &ipVersionRule{allowIPv6: true}
	if !rule.Replaces(other) {
		t.Error("Expected Replaces to return true for same rule type")
	}

	// Should not replace a different rule type (use a mock)
	cidrRule := &ipCIDRRule{}
	if rule.Replaces(cidrRule) {
		t.Error("Expected Replaces to return false for different rule type")
	}
}

// TestIPCIDRRule_Replaces tests:
//   - Replaces returns correct values
func TestIPCIDRRule_Replaces(t *testing.T) {
	rule := &ipCIDRRule{}

	// Should replace another ipCIDRRule
	other := &ipCIDRRule{}
	if !rule.Replaces(other) {
		t.Error("Expected Replaces to return true for same rule type")
	}

	// Should not replace different rule type
	versionRule := &ipVersionRule{}
	if rule.Replaces(versionRule) {
		t.Error("Expected Replaces to return false for different rule type")
	}
}

// TestIPSubnetMaskRule_Replaces tests:
//   - Replaces returns correct values
func TestIPSubnetMaskRule_Replaces(t *testing.T) {
	rule := &ipSubnetMaskRule{}

	// Should replace another ipSubnetMaskRule
	other := &ipSubnetMaskRule{}
	if !rule.Replaces(other) {
		t.Error("Expected Replaces to return true for same rule type")
	}
}

// TestIPRangeRule_Replaces tests:
//   - Replaces returns correct values
func TestIPRangeRule_Replaces(t *testing.T) {
	rule := &ipRangeRule{}

	// Should replace another ipRangeRule
	other := &ipRangeRule{}
	if !rule.Replaces(other) {
		t.Error("Expected Replaces to return true for same rule type")
	}
}

// TestIPPublicPrivateRule_Replaces tests:
//   - Replaces returns correct values
func TestIPPublicPrivateRule_Replaces(t *testing.T) {
	rule := &ipPublicPrivateRule{}

	// Should replace another ipPublicPrivateRule
	other := &ipPublicPrivateRule{}
	if !rule.Replaces(other) {
		t.Error("Expected Replaces to return true for same rule type")
	}
}

// TestIPConflictType_Replaces_NotIPRuleSet tests:
//   - ipConflictType.Replaces returns false when rule is not IPRuleSet
func TestIPConflictType_Replaces_NotIPRuleSet(t *testing.T) {
	checker := ipConflictTypeRequired

	// Use an ipVersionRule which is not an IPRuleSet
	rule := &ipVersionRule{}
	if checker.Replaces(rule) {
		t.Error("Expected Replaces to return false for non-IPRuleSet rule")
	}
}

// TestSetOutput_InterfaceAssignable tests:
//   - setOutput handles interface output where type is assignable to net.IP
func TestSetOutput_InterfaceAssignable(t *testing.T) {
	ip := net.ParseIP("192.168.1.1")

	// Create a pointer to an interface that can hold any value
	var output interface{}
	err := setOutput(context.TODO(), &output, ip)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// The output should be set (as a string since interface{} is not assignable to net.IP)
	if output == nil {
		t.Error("Expected output to be set")
	}
}

// TestIsPrivateIP_ShortIP tests:
//   - isPrivateIP returns false for IPs that are neither IPv4 nor standard IPv6 length
func TestIsPrivateIP_ShortIP(t *testing.T) {
	// Create a short IP (not IPv4 4-byte, not IPv6 16-byte)
	// This is an edge case for malformed IPs
	shortIP := net.IP{192, 168, 1} // Only 3 bytes

	if isPrivateIP(shortIP) {
		t.Error("Expected false for short/malformed IP")
	}
}

// TestIsPrivateIP_IPv4Mapped tests:
//   - isPrivateIP handles IPv4-mapped IPv6 addresses
func TestIsPrivateIP_IPv4Mapped(t *testing.T) {
	// IPv4-mapped IPv6 address for 192.168.1.1 (private)
	// ::ffff:192.168.1.1
	mappedIP := net.ParseIP("::ffff:192.168.1.1")
	if !isPrivateIP(mappedIP) {
		t.Error("Expected true for IPv4-mapped private address")
	}

	// IPv4-mapped IPv6 address for 8.8.8.8 (public)
	mappedPublic := net.ParseIP("::ffff:8.8.8.8")
	if isPrivateIP(mappedPublic) {
		t.Error("Expected false for IPv4-mapped public address")
	}
}
