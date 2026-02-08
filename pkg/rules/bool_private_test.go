package rules

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
)

// TestBoolRuleSet_NoConflict_WithNilParent tests:
//   - noConflict correctly handles the case where a ruleset with parent == nil conflicts
//   - This tests the edge case where someone directly creates a BoolRuleSet (not using factory)
//     with a conflict type set, then calls a conflicting method
func TestBoolRuleSet_NoConflict_WithNilParent(t *testing.T) {
	// Create a ruleset directly (not using factory) with parent == nil and conflictType set
	directRuleSet := &BoolRuleSet{
		parent:       nil,
		conflictType: boolConflictTypeRequired,
	}

	// Create a conflict checker that will match the conflictType
	checker := boolConflictTypeReplacesWrapper{ct: boolConflictTypeRequired}

	// Call noConflict - it should return nil because the ruleset conflicts and has no parent
	result := directRuleSet.noConflict(checker)
	if result != nil {
		t.Errorf("Expected noConflict to return nil when ruleset with parent == nil conflicts, got %v", result)
	}
}

// TestBoolRuleSet_NoConflict_ParentChanged tests:
//   - noConflict correctly handles the case where parent changes
//   - This tests the edge case where parent.noConflict returns a different parent
func TestBoolRuleSet_NoConflict_ParentChanged(t *testing.T) {
	// Create a parent ruleset with a conflict type
	parent := &BoolRuleSet{
		parent:       nil,
		conflictType: boolConflictTypeRequired,
	}

	// Create a child ruleset
	child := &BoolRuleSet{
		parent:       parent,
		conflictType: boolConflictTypeNone,
	}

	// Create a conflict checker that will match the parent's conflictType
	checker := boolConflictTypeReplacesWrapper{ct: boolConflictTypeRequired}

	// Call noConflict - parent should be removed, so child should become the base
	result := child.noConflict(checker)
	if result == nil {
		t.Error("Expected noConflict to return a ruleset when parent conflicts")
		return
	}

	// Result should have no parent (parent was removed due to conflict)
	if result.parent != nil {
		t.Errorf("Expected result to have nil parent after parent conflict, got %v", result.parent)
	}
}

// TestBoolRuleSet_NoConflict_ParentChanged_WithRule tests:
//   - noConflict correctly handles the case where parent changes and child has a rule
//   - This tests the edge case where parent.noConflict returns a different parent and child has a rule
func TestBoolRuleSet_NoConflict_ParentChanged_WithRule(t *testing.T) {
	// Create a grandparent ruleset
	grandparent := &BoolRuleSet{
		parent:       nil,
		conflictType: boolConflictTypeNone,
	}

	// Create a parent ruleset with a conflict type
	parent := &BoolRuleSet{
		parent:       grandparent,
		conflictType: boolConflictTypeRequired,
	}

	// Create a child ruleset with a rule
	mockRule := &mockRuleWithoutConflictType{}
	child := &BoolRuleSet{
		parent:       parent,
		conflictType: boolConflictTypeNone,
		rule:         mockRule,
		label:        "WithRule",
	}

	// Create a conflict checker that will match the parent's conflictType
	checker := boolConflictTypeReplacesWrapper{ct: boolConflictTypeRequired}

	// Call noConflict - parent should be removed, so child should have grandparent as parent
	result := child.noConflict(checker)
	if result == nil {
		t.Error("Expected noConflict to return a ruleset when parent conflicts")
		return
	}

	// Result should have grandparent as parent (parent was removed due to conflict)
	if result.parent != grandparent {
		t.Errorf("Expected result to have grandparent as parent after parent conflict, got %v", result.parent)
	}

	// Result should still have the rule
	if result.rule != mockRule {
		t.Errorf("Expected result to have the same rule, got %v", result.rule)
	}
}

// TestBoolRuleSet_NoConflict_ParentUnchanged tests:
//   - noConflict correctly handles the case where parent doesn't change
//   - This tests the edge case where parent.noConflict returns the same parent
func TestBoolRuleSet_NoConflict_ParentUnchanged(t *testing.T) {
	// Create a parent ruleset without a conflict type
	parent := &BoolRuleSet{
		parent:       nil,
		conflictType: boolConflictTypeNone,
	}

	// Create a child ruleset
	child := &BoolRuleSet{
		parent:       parent,
		conflictType: boolConflictTypeNone,
	}

	// Create a conflict checker that won't match anything
	checker := boolConflictTypeReplacesWrapper{ct: boolConflictTypeStrict}

	// Call noConflict - parent should not change, so child should be returned unchanged
	result := child.noConflict(checker)
	if result == nil {
		t.Error("Expected noConflict to return a ruleset when nothing conflicts")
		return
	}

	// Result should be the same as child (no changes)
	if result != child {
		t.Errorf("Expected result to be the same as child when parent doesn't change, got %v", result)
	}

	// Result should have the same parent
	if result.parent != parent {
		t.Errorf("Expected result to have the same parent, got %v", result.parent)
	}
}

// TestBoolRuleSet_NoConflict_RuleConflicts tests:
//   - noConflict correctly handles the case where the rule conflicts
//   - This tests the edge case where checker.Replaces(ruleSet.rule) returns true
func TestBoolRuleSet_NoConflict_RuleConflicts(t *testing.T) {
	// Create a parent ruleset
	parent := &BoolRuleSet{
		parent:       nil,
		conflictType: boolConflictTypeNone,
	}

	// Create a child ruleset with a conflicting rule
	// We'll use a rule that implements getConflictType to trigger the conflict
	conflictingRuleSet := &BoolRuleSet{
		parent:       nil,
		conflictType: boolConflictTypeRequired,
	}

	child := &BoolRuleSet{
		parent:       parent,
		conflictType: boolConflictTypeNone,
		rule:         conflictingRuleSet,
	}

	// Create a conflict checker that will match the rule's conflictType
	checker := boolConflictTypeReplacesWrapper{ct: boolConflictTypeRequired}

	// Call noConflict - rule should conflict, so we should skip this node and go to parent
	result := child.noConflict(checker)
	if result == nil {
		t.Error("Expected noConflict to return a ruleset when rule conflicts")
		return
	}

	// Result should be the parent (child was skipped due to rule conflict)
	if result != parent {
		t.Errorf("Expected result to be parent when rule conflicts, got %v", result)
	}
}

// TestBoolConflictTypeReplacesWrapper_Replaces_False tests:
//   - Replaces returns false when rule doesn't implement getConflictType
func TestBoolConflictTypeReplacesWrapper_Replaces_False(t *testing.T) {
	// Create a mock rule that doesn't implement getConflictType
	mockRule := &mockRuleWithoutConflictType{}

	// Create a conflict checker
	checker := boolConflictTypeReplacesWrapper{ct: boolConflictTypeRequired}

	// Replaces should return false
	if checker.Replaces(mockRule) {
		t.Error("Expected Replaces to return false for rule without getConflictType")
	}
}

// mockRuleWithoutConflictType is a mock rule that doesn't implement getConflictType
type mockRuleWithoutConflictType struct{}

func (m *mockRuleWithoutConflictType) Evaluate(ctx context.Context, value bool) errors.ValidationError {
	return nil
}

func (m *mockRuleWithoutConflictType) Replaces(_ Rule[bool]) bool {
	return false
}

func (m *mockRuleWithoutConflictType) String() string {
	return "MockRule"
}
