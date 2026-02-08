package net

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// TestDomainConflictType_Replaces_WrongType tests:
// - domainConflictType.Replaces returns false for non-DomainRuleSet rules
func TestDomainConflictType_Replaces_WrongType(t *testing.T) {
	checker := domainConflictTypeRequired

	// Test with StringRuleSet (wrong type)
	stringRS := rules.String().WithRequired()
	if checker.Replaces(stringRS) {
		t.Errorf("Expected domainConflictType to return false for StringRuleSet")
	}

	// Test with EmailRuleSet (wrong type)
	emailRS := Email().WithRequired()
	if checker.Replaces(emailRS) {
		t.Errorf("Expected domainConflictType to return false for EmailRuleSet")
	}

	// Test with URIRuleSet (wrong type)
	uriRS := URI().WithRequired()
	if checker.Replaces(uriRS) {
		t.Errorf("Expected domainConflictType to return false for URIRuleSet")
	}

	// Test with a regular rule (not a ruleset) - use RuleFunc which doesn't implement getConflictType
	rule := rules.RuleFunc[string](func(ctx context.Context, value string) errors.ValidationError {
		return nil
	})
	if checker.Replaces(rule) {
		t.Errorf("Expected domainConflictType to return false for RuleFunc")
	}
}

// TestEmailConflictType_Replaces_WrongType tests:
// - emailConflictType.Replaces returns false for non-EmailRuleSet rules
func TestEmailConflictType_Replaces_WrongType(t *testing.T) {
	checker := emailConflictTypeRequired

	// Test with StringRuleSet (wrong type)
	stringRS := rules.String().WithRequired()
	if checker.Replaces(stringRS) {
		t.Errorf("Expected emailConflictType to return false for StringRuleSet")
	}

	// Test with DomainRuleSet (wrong type)
	domainRS := Domain().WithRequired()
	if checker.Replaces(domainRS) {
		t.Errorf("Expected emailConflictType to return false for DomainRuleSet")
	}

	// Test with URIRuleSet (wrong type)
	uriRS := URI().WithRequired()
	if checker.Replaces(uriRS) {
		t.Errorf("Expected emailConflictType to return false for URIRuleSet")
	}

	// Test with a regular rule (not a ruleset) - use RuleFunc which doesn't implement getConflictType
	rule := rules.RuleFunc[string](func(ctx context.Context, value string) errors.ValidationError {
		return nil
	})
	if checker.Replaces(rule) {
		t.Errorf("Expected emailConflictType to return false for RuleFunc")
	}
}

// TestURIConflictType_Replaces_WrongType tests:
// - uriConflictType.Replaces returns false for non-URIRuleSet rules
func TestURIConflictType_Replaces_WrongType(t *testing.T) {
	checker := uriConflictTypeRequired

	// Test with StringRuleSet (wrong type)
	stringRS := rules.String().WithRequired()
	if checker.Replaces(stringRS) {
		t.Errorf("Expected uriConflictType to return false for StringRuleSet")
	}

	// Test with DomainRuleSet (wrong type)
	domainRS := Domain().WithRequired()
	if checker.Replaces(domainRS) {
		t.Errorf("Expected uriConflictType to return false for DomainRuleSet")
	}

	// Test with EmailRuleSet (wrong type)
	emailRS := Email().WithRequired()
	if checker.Replaces(emailRS) {
		t.Errorf("Expected uriConflictType to return false for EmailRuleSet")
	}

	// Test with a regular rule (not a ruleset) - use RuleFunc which doesn't implement getConflictType
	rule := rules.RuleFunc[string](func(ctx context.Context, value string) errors.ValidationError {
		return nil
	})
	if checker.Replaces(rule) {
		t.Errorf("Expected uriConflictType to return false for RuleFunc")
	}
}
