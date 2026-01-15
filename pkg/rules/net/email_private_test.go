package net

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestEmailRuleSet_NoConflict_WithNilParent tests:
// - noConflict correctly handles the case where a ruleset with parent == nil conflicts
// - This tests the edge case where someone directly creates an EmailRuleSet (not using factory)
//   with a conflict type set, then calls a conflicting method
func TestEmailRuleSet_NoConflict_WithNilParent(t *testing.T) {
	// Create a ruleset directly (not using factory) with parent == nil and conflictType set
	directRuleSet := &EmailRuleSet{
		parent:       nil,
		conflictType: emailConflictTypeRequired,
	}

	// Create a conflict checker that will match the conflictType
	checker := emailConflictTypeRequired

	// Call noConflict - it should return nil because the ruleset conflicts and has no parent
	result := directRuleSet.noConflict(checker)
	if result != nil {
		t.Errorf("Expected noConflict to return nil when ruleset with parent == nil conflicts, got %v", result)
	}
}

// TestEmailRuleSet_WithRuleFunc_Conflict tests:
// - Conflicting rules are deduplicated
func TestEmailRuleSet_WithRuleFunc_Conflict(t *testing.T) {
	testVal := "hello@example.com"

	mockA := testhelpers.NewMockRule[string]()
	mockA.ConflictKey = "test"

	mockB := testhelpers.NewMockRule[string]()

	var output string
	err := Email().
		WithRule(mockB).
		WithRule(mockA).
		WithRule(mockB).
		WithRule(mockA).
		WithRule(mockB).
		Apply(context.TODO(), testVal, &output)

	if err != nil {
		t.Errorf("Expected errors to be nil, got: %s", err)
	}

	if mockA.EvaluateCallCount() != 1 {
		t.Errorf("Expected 1 call to Evaluate for mockA, got: %d", mockA.EvaluateCallCount())
	}

	if mockB.EvaluateCallCount() != 3 {
		t.Errorf("Expected 3 calls to Evaluate for mockB, got: %d", mockB.EvaluateCallCount())
	}
}
