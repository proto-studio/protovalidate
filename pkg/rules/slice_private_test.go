package rules

import (
	"testing"
)

// TestSliceRuleSet_NoConflict_WithNilParent tests:
//   - noConflict correctly handles the case where a ruleset with parent == nil conflicts
//   - This tests the edge case where someone directly creates a SliceRuleSet (not using factory)
//     with a conflict type set, then calls a conflicting method
func TestSliceRuleSet_NoConflict_WithNilParent(t *testing.T) {
	// Create a ruleset directly (not using factory) with parent == nil and conflictType set
	directRuleSet := &SliceRuleSet[int]{
		parent:       nil,
		conflictType: conflictTypeRequired,
	}

	// Create a conflict checker that will match the conflictType
	checker := sliceConflictTypeReplacesWrapper[int]{ct: conflictTypeRequired}

	// Call noConflict - it should return nil because the ruleset conflicts and has no parent
	result := directRuleSet.noConflict(checker)
	if result != nil {
		t.Errorf("Expected noConflict to return nil when ruleset with parent == nil conflicts, got %v", result)
	}
}
