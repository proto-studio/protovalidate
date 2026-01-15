package net

import (
	"testing"
)

// TestURIRuleSet_NoConflict_WithNilParent tests:
// - noConflict correctly handles the case where a ruleset with parent == nil conflicts
// - This tests the edge case where someone directly creates a URIRuleSet (not using factory)
//   with a conflict type set, then calls a conflicting method
func TestURIRuleSet_NoConflict_WithNilParent(t *testing.T) {
	// Create a ruleset directly (not using factory) with parent == nil and conflictType set
	directRuleSet := &URIRuleSet{
		parent:       nil,
		conflictType: uriConflictTypeRequired,
	}

	// Create a conflict checker that will match the conflictType
	checker := uriConflictTypeRequired

	// Call noConflict - it should return nil because the ruleset conflicts and has no parent
	result := directRuleSet.noConflict(checker)
	if result != nil {
		t.Errorf("Expected noConflict to return nil when ruleset with parent == nil conflicts, got %v", result)
	}
}
