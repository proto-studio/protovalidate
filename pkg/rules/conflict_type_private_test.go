package rules

import (
	"testing"
)

// TestStringConflictType_Replaces_WrongType tests:
// - stringConflictType.Replaces returns false for non-StringRuleSet rules
func TestStringConflictType_Replaces_WrongType(t *testing.T) {
	checker := stringConflictTypeRequired

	// Test with a regular rule (not a ruleset) - this should return false
	rule := &stringMinRule{min: "abc"}
	if checker.Replaces(rule) {
		t.Errorf("Expected stringConflictType to return false for stringMinRule")
	}

	// Test that the type check works by verifying it returns false for wrong types
	// We can't directly pass other ruleset types due to Go's type system,
	// but we can verify the type assertion in Replaces works correctly
	// by checking that it returns false for non-StringRuleSet types
	stringRS := String().WithRequired()
	if !checker.Replaces(stringRS) {
		t.Errorf("Expected stringConflictType to return true for StringRuleSet with matching conflictType")
	}

	// Create a StringRuleSet with different conflictType - should still match type but not conflict
	stringRS2 := String().WithNil()
	if checker.Replaces(stringRS2) {
		t.Errorf("Expected stringConflictTypeRequired to return false for StringRuleSet with conflictTypeNil")
	}
}

// TestIntConflictType_Replaces_WrongType tests:
// - conflictTypeReplacesWrapper returns false for non-IntRuleSet rules
func TestIntConflictType_Replaces_WrongType(t *testing.T) {
	checker := conflictTypeReplacesWrapper[int]{ct: intConflictTypeBase}

	// Test with a regular rule (not a ruleset) - should return false
	rule := &minRule[int]{min: 5, fmt: "d"}
	if checker.Replaces(rule) {
		t.Errorf("Expected intConflictType wrapper to return false for minRule (doesn't implement getConflictType)")
	}

	// Test that it works correctly for IntRuleSet
	intRS := Int().WithBase(16)
	if !checker.Replaces(intRS) {
		t.Errorf("Expected intConflictType wrapper to return true for IntRuleSet with matching conflictType")
	}

	// Test with IntRuleSet with different conflictType
	intRS2 := Int().WithRequired()
	if checker.Replaces(intRS2) {
		t.Errorf("Expected intConflictTypeBase wrapper to return false for IntRuleSet with conflictTypeRequired")
	}
}

// TestFloatConflictType_Replaces_WrongType tests:
// - floatConflictTypeReplacesWrapper returns false for non-FloatRuleSet rules
func TestFloatConflictType_Replaces_WrongType(t *testing.T) {
	checker := floatConflictTypeReplacesWrapper[float32]{ct: floatConflictTypeRounding}

	// Test with a regular rule (not a ruleset) - should return false
	rule := &minRule[float32]{min: 5.5, fmt: "f"}
	if checker.Replaces(rule) {
		t.Errorf("Expected floatConflictType wrapper to return false for minRule (doesn't implement getConflictType)")
	}

	// Test that it works correctly for FloatRuleSet
	floatRS := Float32().WithRounding(RoundingUp, 2)
	if !checker.Replaces(floatRS) {
		t.Errorf("Expected floatConflictType wrapper to return true for FloatRuleSet with matching conflictType")
	}

	// Test with FloatRuleSet with different conflictType
	floatRS2 := Float32().WithFixedOutput(2)
	if checker.Replaces(floatRS2) {
		t.Errorf("Expected floatConflictTypeRounding wrapper to return false for FloatRuleSet with floatConflictTypeFixedOutput")
	}
}

// TestSliceConflictType_Replaces_WrongType tests:
// - conflictType.Replaces and sliceConflictTypeReplacesWrapper return false for non-SliceRuleSet rules
func TestSliceConflictType_Replaces_WrongType(t *testing.T) {
	checker1 := conflictType(conflictTypeMinLen)
	checker2 := sliceConflictTypeReplacesWrapper[int]{ct: conflictTypeMinLen}

	// Test with a regular rule (not a ruleset) - should return false
	// minLenRule doesn't implement getConflictType, so both should return false
	rule := &minLenRule[int, []int]{min: 5}
	if checker1.Replaces(rule) {
		t.Errorf("Expected slice conflictType to return false for minLenRule (doesn't implement getConflictType)")
	}
	if checker2.Replaces(rule) {
		t.Errorf("Expected slice wrapper to return false for minLenRule (doesn't implement getConflictType)")
	}

	// Test that it works correctly for SliceRuleSet
	sliceRS := Slice[int]().WithMinLen(5)
	if !checker1.Replaces(sliceRS) {
		t.Errorf("Expected slice conflictType to return true for SliceRuleSet with matching conflictType")
	}
	if !checker2.Replaces(sliceRS) {
		t.Errorf("Expected slice wrapper to return true for SliceRuleSet with matching conflictType")
	}

	// Test with SliceRuleSet with different conflictType
	sliceRS2 := Slice[int]().WithMaxLen(10)
	if checker1.Replaces(sliceRS2) {
		t.Errorf("Expected conflictTypeMinLen to return false for SliceRuleSet with conflictTypeMaxLen")
	}
	if checker2.Replaces(sliceRS2) {
		t.Errorf("Expected conflictTypeMinLen wrapper to return false for SliceRuleSet with conflictTypeMaxLen")
	}
}


