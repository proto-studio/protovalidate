package time

import (
	"context"
	"testing"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// TestTimeRuleSet_NoConflict_WithNilParent tests:
//   - noConflict correctly handles the case where a ruleset with parent == nil conflicts
//   - This tests the edge case where someone directly creates a TimeRuleSet (not using factory)
//     with a conflict type set, then calls a conflicting method
func TestTimeRuleSet_NoConflict_WithNilParent(t *testing.T) {
	// Create a ruleset directly (not using factory) with parent == nil and conflictType set
	directRuleSet := &TimeRuleSet{
		parent:       nil,
		conflictType: conflictTypeRequired,
	}

	// Create a conflict checker that will match the conflictType
	checker := conflictType(conflictTypeRequired)

	// Call noConflict - it should return nil because the ruleset conflicts and has no parent
	result := directRuleSet.noConflict(checker)
	if result != nil {
		t.Errorf("Expected noConflict to return nil when ruleset with parent == nil conflicts, got %v", result)
	}
}

// TestTimeConflictType_Replaces_WrongType tests:
//   - conflictType.Replaces returns false when the cast to *TimeRuleSet fails
//   - This tests that the type assertion in Replaces correctly handles non-TimeRuleSet types
func TestTimeConflictType_Replaces_WrongType(t *testing.T) {
	checker := conflictType(conflictTypeRequired)

	// Test with a regular rule (not a ruleset) - cast should fail
	// RuleFunc doesn't implement the TimeRuleSet interface, so the cast in Replaces should fail
	rule := rules.RuleFunc[time.Time](func(ctx context.Context, value time.Time) errors.ValidationErrorCollection {
		return nil
	})
	if checker.Replaces(rule) {
		t.Errorf("Expected time conflictType to return false for RuleFunc (cast should fail)")
	}

	// Test with TimeRuleSet (correct type) - should work
	timeRS := Time().WithRequired()
	if !checker.Replaces(timeRS) {
		t.Errorf("Expected time conflictType to return true for TimeRuleSet with matching conflictType")
	}

	// Test with TimeRuleSet with different conflictType - should match type but not conflict
	timeRS2 := Time().WithNil()
	if checker.Replaces(timeRS2) {
		t.Errorf("Expected conflictTypeRequired to return false for TimeRuleSet with conflictTypeNil")
	}

	// Test that the type assertion works correctly by verifying it returns false for wrong types
	// We can't directly pass other ruleset types due to Go's type system, but we can verify
	// the type assertion in Replaces works by checking that it returns false for non-TimeRuleSet types
	// The internal implementation uses: rs, ok := r.(*TimeRuleSet); if !ok { return false }
	// This is tested implicitly by the RuleFunc test above, which should fail the type assertion
}
