package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestSliceRuleSet_WithRule_NoConflict tests:
// - Adding a non-conflicting rule preserves existing rules
// - Rules are applied in the correct order
func TestSliceRuleSet_WithRule_NoConflict(t *testing.T) {
	mockRule1 := testhelpers.NewMockRule[[]int]()
	mockRule1.ConflictKey = "rule1"

	mockRule2 := testhelpers.NewMockRule[[]int]()
	mockRule2.ConflictKey = "rule2"

	// Create rule set with first rule
	ruleSet1 := rules.Slice[int]().WithRule(mockRule1)

	// Add second non-conflicting rule
	ruleSet2 := ruleSet1.WithRule(mockRule2)

	// Both rules should be in the chain
	expected := "SliceRuleSet[int].WithMock().WithMock()"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Apply should call both rules
	var output []int
	err := ruleSet2.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err != nil {
		t.Errorf("Expected no errors, got %s", err)
	}

	// Both rules should have been evaluated
	if mockRule1.EvaluateCallCount() != 1 {
		t.Errorf("Expected mockRule1 to be evaluated 1 time, got %d", mockRule1.EvaluateCallCount())
	}
	if mockRule2.EvaluateCallCount() != 1 {
		t.Errorf("Expected mockRule2 to be evaluated 1 time, got %d", mockRule2.EvaluateCallCount())
	}
}

// TestSliceRuleSet_WithRule_Conflict tests:
// - Adding a conflicting rule removes the old conflicting rule
// - Non-conflicting rules are preserved
// - Original rule set is not mutated
func TestSliceRuleSet_WithRule_Conflict(t *testing.T) {
	mockRule1 := testhelpers.NewMockRule[[]int]()
	mockRule1.ConflictKey = "conflict"

	mockRule2 := testhelpers.NewMockRule[[]int]()
	mockRule2.ConflictKey = "conflict" // Same conflict key

	mockRule3 := testhelpers.NewMockRule[[]int]()
	mockRule3.ConflictKey = "rule3" // Different conflict key

	// Create rule set with first rule
	ruleSet1 := rules.Slice[int]().WithRule(mockRule1)

	// Add non-conflicting rule
	ruleSet2 := ruleSet1.WithRule(mockRule3)

	// Add conflicting rule (should remove mockRule1)
	ruleSet3 := ruleSet2.WithRule(mockRule2)

	// Verify original rule set is not mutated
	expected1 := "SliceRuleSet[int].WithMock().WithMock()"
	if s := ruleSet2.String(); s != expected1 {
		t.Errorf("Expected original rule set to be %s, got %s", expected1, s)
	}

	// New rule set should have mockRule2 and mockRule3, but not mockRule1
	expected2 := "SliceRuleSet[int].WithMock().WithMock()"
	if s := ruleSet3.String(); s != expected2 {
		t.Errorf("Expected new rule set to be %s, got %s", expected2, s)
	}

	// Apply should call mockRule2 and mockRule3, but not mockRule1
	var output []int
	err := ruleSet3.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err != nil {
		t.Errorf("Expected no errors, got %s", err)
	}

	// mockRule1 should not have been evaluated (conflict removed it)
	if mockRule1.EvaluateCallCount() != 0 {
		t.Errorf("Expected mockRule1 to not be evaluated (conflict removed), got %d", mockRule1.EvaluateCallCount())
	}

	// mockRule2 should have been evaluated
	if mockRule2.EvaluateCallCount() != 1 {
		t.Errorf("Expected mockRule2 to be evaluated 1 time, got %d", mockRule2.EvaluateCallCount())
	}

	// mockRule3 should have been evaluated
	if mockRule3.EvaluateCallCount() != 1 {
		t.Errorf("Expected mockRule3 to be evaluated 1 time, got %d", mockRule3.EvaluateCallCount())
	}
}

// TestSliceRuleSet_WithRule_ConflictMultiple tests:
// - Multiple conflicting rules in the chain are all removed
// - Only the most recent non-conflicting rule remains
func TestSliceRuleSet_WithRule_ConflictMultiple(t *testing.T) {
	mockRule1 := testhelpers.NewMockRule[[]int]()
	mockRule1.ConflictKey = "conflict"

	mockRule2 := testhelpers.NewMockRule[[]int]()
	mockRule2.ConflictKey = "conflict" // Same conflict key

	mockRule3 := testhelpers.NewMockRule[[]int]()
	mockRule3.ConflictKey = "conflict" // Same conflict key

	// Create rule set with multiple conflicting rules
	ruleSet1 := rules.Slice[int]().WithRule(mockRule1)
	ruleSet2 := ruleSet1.WithRule(mockRule2)
	ruleSet3 := ruleSet2.WithRule(mockRule3)

	// Only the most recent rule should remain
	expected := "SliceRuleSet[int].WithMock()"
	if s := ruleSet3.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Apply should only call mockRule3
	var output []int
	err := ruleSet3.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err != nil {
		t.Errorf("Expected no errors, got %s", err)
	}

	// Only mockRule3 should have been evaluated
	if mockRule1.EvaluateCallCount() != 0 {
		t.Errorf("Expected mockRule1 to not be evaluated, got %d", mockRule1.EvaluateCallCount())
	}
	if mockRule2.EvaluateCallCount() != 0 {
		t.Errorf("Expected mockRule2 to not be evaluated, got %d", mockRule2.EvaluateCallCount())
	}
	if mockRule3.EvaluateCallCount() != 1 {
		t.Errorf("Expected mockRule3 to be evaluated 1 time, got %d", mockRule3.EvaluateCallCount())
	}
}

// TestSliceRuleSet_WithRule_ConflictWithNonConflicting tests:
// - Conflicting rules are removed, but non-conflicting rules in between are preserved
func TestSliceRuleSet_WithRule_ConflictWithNonConflicting(t *testing.T) {
	mockRule1 := testhelpers.NewMockRule[[]int]()
	mockRule1.ConflictKey = "conflict"

	mockRule2 := testhelpers.NewMockRule[[]int]()
	mockRule2.ConflictKey = "rule2" // Non-conflicting

	mockRule3 := testhelpers.NewMockRule[[]int]()
	mockRule3.ConflictKey = "conflict" // Conflicting with mockRule1

	// Create rule set: conflict -> non-conflict -> conflict
	ruleSet1 := rules.Slice[int]().WithRule(mockRule1)
	ruleSet2 := ruleSet1.WithRule(mockRule2)
	ruleSet3 := ruleSet2.WithRule(mockRule3)

	// mockRule1 should be removed, but mockRule2 should remain
	expected := "SliceRuleSet[int].WithMock().WithMock()"
	if s := ruleSet3.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Apply should call mockRule2 and mockRule3, but not mockRule1
	var output []int
	err := ruleSet3.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err != nil {
		t.Errorf("Expected no errors, got %s", err)
	}

	// mockRule1 should not have been evaluated (conflict removed it)
	if mockRule1.EvaluateCallCount() != 0 {
		t.Errorf("Expected mockRule1 to not be evaluated, got %d", mockRule1.EvaluateCallCount())
	}

	// mockRule2 should have been evaluated (non-conflicting, preserved)
	if mockRule2.EvaluateCallCount() != 1 {
		t.Errorf("Expected mockRule2 to be evaluated 1 time, got %d", mockRule2.EvaluateCallCount())
	}

	// mockRule3 should have been evaluated
	if mockRule3.EvaluateCallCount() != 1 {
		t.Errorf("Expected mockRule3 to be evaluated 1 time, got %d", mockRule3.EvaluateCallCount())
	}
}

// TestSliceRuleSet_WithRule_ConflictRoot tests:
// - Conflicting with root rule set (no parent) is handled correctly
func TestSliceRuleSet_WithRule_ConflictRoot(t *testing.T) {
	mockRule1 := testhelpers.NewMockRule[[]int]()
	mockRule1.ConflictKey = "conflict"

	mockRule2 := testhelpers.NewMockRule[[]int]()
	mockRule2.ConflictKey = "conflict" // Same conflict key

	// Create rule set with first rule (at root)
	ruleSet1 := rules.Slice[int]().WithRule(mockRule1)

	// Add conflicting rule (should remove mockRule1, leaving just root)
	ruleSet2 := ruleSet1.WithRule(mockRule2)

	// Only the new rule should remain
	expected := "SliceRuleSet[int].WithMock()"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Apply should only call mockRule2
	var output []int
	err := ruleSet2.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err != nil {
		t.Errorf("Expected no errors, got %s", err)
	}

	// mockRule1 should not have been evaluated
	if mockRule1.EvaluateCallCount() != 0 {
		t.Errorf("Expected mockRule1 to not be evaluated, got %d", mockRule1.EvaluateCallCount())
	}

	// mockRule2 should have been evaluated
	if mockRule2.EvaluateCallCount() != 1 {
		t.Errorf("Expected mockRule2 to be evaluated 1 time, got %d", mockRule2.EvaluateCallCount())
	}
}

// TestSliceRuleSet_WithRule_NoConflictKey tests:
// - Rules without ConflictKey don't conflict with each other
func TestSliceRuleSet_WithRule_NoConflictKey(t *testing.T) {
	mockRule1 := testhelpers.NewMockRule[[]int]()
	// No ConflictKey set

	mockRule2 := testhelpers.NewMockRule[[]int]()
	// No ConflictKey set

	// Create rule set with both rules
	ruleSet1 := rules.Slice[int]().WithRule(mockRule1)
	ruleSet2 := ruleSet1.WithRule(mockRule2)

	// Both rules should be in the chain
	expected := "SliceRuleSet[int].WithMock().WithMock()"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Apply should call both rules
	var output []int
	err := ruleSet2.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err != nil {
		t.Errorf("Expected no errors, got %s", err)
	}

	// Both rules should have been evaluated
	if mockRule1.EvaluateCallCount() != 1 {
		t.Errorf("Expected mockRule1 to be evaluated 1 time, got %d", mockRule1.EvaluateCallCount())
	}
	if mockRule2.EvaluateCallCount() != 1 {
		t.Errorf("Expected mockRule2 to be evaluated 1 time, got %d", mockRule2.EvaluateCallCount())
	}
}

// TestSliceRuleSet_WithRule_ConflictPreservesProperties tests:
// - When conflicts are resolved, other properties (maxLen, minLen, required, withNil) are preserved
func TestSliceRuleSet_WithRule_ConflictPreservesProperties(t *testing.T) {
	mockRule1 := testhelpers.NewMockRule[[]int]()
	mockRule1.ConflictKey = "conflict"

	mockRule2 := testhelpers.NewMockRule[[]int]()
	mockRule2.ConflictKey = "conflict"

	// Create rule set with properties and conflicting rule
	ruleSet1 := rules.Slice[int]().
		WithMaxLen(10).
		WithMinLen(2).
		WithRequired().
		WithRule(mockRule1)

	// Add conflicting rule
	ruleSet2 := ruleSet1.WithRule(mockRule2)

	// Properties should be preserved
	if !ruleSet2.Required() {
		t.Error("Expected Required to be true")
	}

	// Verify maxLen and minLen are preserved by checking String representation
	// (they should still be in the chain)
	expected := "SliceRuleSet[int].WithMaxLen(10).WithMinLen(2).WithRequired().WithMock()"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	// Apply should work correctly with preserved properties
	var output []int
	err := ruleSet2.Apply(context.TODO(), []int{1, 2, 3}, &output)
	if err != nil {
		t.Errorf("Expected no errors, got %s", err)
	}
}
