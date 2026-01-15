package time

import (
	"context"
	"testing"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// TestDurationRuleSet_NoConflict_WithNilParent tests:
//   - noConflict correctly handles the case where a ruleset with parent == nil conflicts
//   - This tests the edge case where someone directly creates a DurationRuleSet (not using factory)
//     with a conflict type set, then calls a conflicting method
func TestDurationRuleSet_NoConflict_WithNilParent(t *testing.T) {
	// Create a ruleset directly (not using factory) with parent == nil and conflictType set
	directRuleSet := &DurationRuleSet{
		parent:       nil,
		conflictType: conflictTypeDurationRequired,
	}

	// Create a conflict checker that will match the conflictType
	checker := conflictTypeDuration(conflictTypeDurationRequired)

	// Call noConflict - it should return nil because the ruleset conflicts and has no parent
	result := directRuleSet.noConflict(checker)
	if result != nil {
		t.Errorf("Expected noConflict to return nil when ruleset with parent == nil conflicts, got %v", result)
	}
}

// TestDurationConflictType_Replaces_WrongType tests:
//   - conflictTypeDuration.Replaces returns false when the cast to *DurationRuleSet fails
//   - This tests that the type assertion in Replaces correctly handles non-DurationRuleSet types
func TestDurationConflictType_Replaces_WrongType(t *testing.T) {
	checker := conflictTypeDuration(conflictTypeDurationRequired)

	// Test with a regular rule (not a ruleset) - cast should fail
	// RuleFunc doesn't implement the DurationRuleSet interface, so the cast in Replaces should fail
	rule := rules.RuleFunc[time.Duration](func(ctx context.Context, value time.Duration) errors.ValidationErrorCollection {
		return nil
	})
	if checker.Replaces(rule) {
		t.Errorf("Expected duration conflictType to return false for RuleFunc (cast should fail)")
	}

	// Test with DurationRuleSet (correct type) - should work
	durationRS := Duration().WithRequired()
	if !checker.Replaces(durationRS) {
		t.Errorf("Expected duration conflictType to return true for DurationRuleSet with matching conflictType")
	}

	// Test with DurationRuleSet with different conflictType - should match type but not conflict
	durationRS2 := Duration().WithNil()
	if checker.Replaces(durationRS2) {
		t.Errorf("Expected conflictTypeDurationRequired to return false for DurationRuleSet with conflictTypeDurationNil")
	}

	// Test that the type assertion works correctly by verifying it returns false for wrong types
	// We can't directly pass other ruleset types due to Go's type system, but we can verify
	// the type assertion in Replaces works by checking that it returns false for non-DurationRuleSet types
	// The internal implementation uses: rs, ok := r.(*DurationRuleSet); if !ok { return false }
	// This is tested implicitly by the RuleFunc test above, which should fail the type assertion
}
