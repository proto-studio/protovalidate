package time_test

import (
	"context"
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/time"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestDurationRuleSet_Apply tests:
// - Default configuration doesn't return errors on valid value.
// - Implements interface.
func TestDurationRuleSet_Apply(t *testing.T) {
	dur := 1 * internalTime.Hour

	// Prepare an output variable for Apply
	var output internalTime.Duration

	// Use Apply to validate the duration
	err := time.Duration().Apply(context.TODO(), dur, &output)

	if err != nil {
		t.Fatal("Expected errors to be empty")
	}

	if output != dur {
		t.Fatal("Expected test duration to be returned")
	}

	// Check if the rule set implements the expected interface
	ok := testhelpers.CheckRuleSetInterface[internalTime.Duration](time.Duration())
	if !ok {
		t.Fatal("Expected rule set to be implemented")
	}

	testhelpers.MustApplyTypes[internalTime.Duration](t, time.Duration(), dur)
}

// TestDurationRuleSet_Apply_String tests:
// - Will coerce duration from string
func TestDurationRuleSet_Apply_String(t *testing.T) {
	s := "1h30m"
	expected := 1*internalTime.Hour + 30*internalTime.Minute

	ruleSet := time.Duration()
	testhelpers.MustApplyMutation(t, ruleSet.Any(), s, expected)
}

// TestDurationRuleSet_Apply_Int64 tests:
// - Will coerce duration from int64 (nanoseconds)
func TestDurationRuleSet_Apply_Int64(t *testing.T) {
	nanos := int64(3600000000000) // 1 hour in nanoseconds
	expected := 1 * internalTime.Hour

	ruleSet := time.Duration()
	testhelpers.MustApplyMutation(t, ruleSet.Any(), nanos, expected)
}

// TestDurationRuleSet_Apply_Int tests:
// - Will coerce duration from int (nanoseconds)
func TestDurationRuleSet_Apply_Int(t *testing.T) {
	nanos := int(3600000000000) // 1 hour in nanoseconds
	expected := 1 * internalTime.Hour

	ruleSet := time.Duration()
	testhelpers.MustApplyMutation(t, ruleSet.Any(), nanos, expected)
}

// TestDurationRuleSet_Apply_Pointer tests:
// - Correctly handles pointer duration values
func TestDurationRuleSet_Apply_Pointer(t *testing.T) {
	dur := 1 * internalTime.Hour

	ruleSet := time.Duration()
	testhelpers.MustApplyMutation(t, ruleSet.Any(), &dur, dur)
}

// TestDurationRuleSet_Apply_BadType tests:
// - Returns error for types that cannot be coerced to duration
func TestDurationRuleSet_Apply_BadType(t *testing.T) {
	ruleSet := time.Duration()
	type x struct{}

	testhelpers.MustNotApply(t, ruleSet.Any(), new(x), errors.CodeType)
}

// TestDurationRuleSet_Apply_InvalidString tests:
// - Returns error for invalid duration strings
func TestDurationRuleSet_Apply_InvalidString(t *testing.T) {
	ruleSet := time.Duration()
	invalid := "not a duration"

	testhelpers.MustNotApply(t, ruleSet.Any(), invalid, errors.CodePattern)
}

// TestDurationRuleSet_WithRequired tests:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
func TestDurationRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[internalTime.Duration](t, time.Duration())
}

// TestDurationCustom tests:
// - Custom rule functions are executed
// - Custom rules can return errors
// - Rule evaluation is called correctly
func TestDurationCustom(t *testing.T) {
	dur := 1 * internalTime.Hour

	ruleSet := time.Duration().WithRuleFunc(testhelpers.NewMockRuleWithErrors[internalTime.Duration](1).Function()).Any()
	testhelpers.MustNotApply(t, ruleSet, dur, errors.CodeUnknown)

	rule := testhelpers.NewMockRule[internalTime.Duration]()
	ruleSet = time.Duration().WithRuleFunc(rule.Function()).Any()
	testhelpers.MustApply(t, ruleSet, dur)

	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

// TestDurationRuleSet_Any tests:
// - Any returns a RuleSet[any] implementation
func TestDurationRuleSet_Any(t *testing.T) {
	ruleSet := time.Duration().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	}
}

// TestDurationRuleSet_String tests:
// - String representation is correct
func TestDurationRuleSet_String(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *time.DurationRuleSet
		expected string
	}{
		{"Base", time.Duration(), "DurationRuleSet"},
		{"WithRequired", time.Duration().WithRequired(), "DurationRuleSet.WithRequired()"},
		{"WithNil", time.Duration().WithNil(), "DurationRuleSet.WithNil()"},
		{"WithMin", time.Duration().WithMin(1 * internalTime.Hour), "DurationRuleSet.WithMin(1h0m0s)"},
		{"WithMax", time.Duration().WithMax(24 * internalTime.Hour), "DurationRuleSet.WithMax(24h0m0s)"},
		{"WithMinExclusive", time.Duration().WithMinExclusive(1 * internalTime.Hour), "DurationRuleSet.WithMinExclusive(1h0m0s)"},
		{"WithMaxExclusive", time.Duration().WithMaxExclusive(24 * internalTime.Hour), "DurationRuleSet.WithMaxExclusive(24h0m0s)"},
		{"WithUnit", time.Duration().WithUnit(internalTime.Second), "DurationRuleSet.WithUnit(\"1s\")"},
		{"WithRounding", time.Duration().WithRounding(rules.RoundingHalfEven), "DurationRuleSet.WithRounding(HalfEven)"},
		{"Chained", time.Duration().WithRequired().WithNil(), "DurationRuleSet.WithRequired().WithNil()"},
		{"ChainedWithRule", time.Duration().WithRequired().WithMin(1 * internalTime.Hour), "DurationRuleSet.WithRequired().WithMin(1h0m0s)"},
		{"ChainedWithUnit", time.Duration().WithUnit(internalTime.Minute).WithMin(1 * internalTime.Hour), "DurationRuleSet.WithUnit(\"1m0s\").WithMin(1h0m0s)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ruleSet.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestDurationRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestDurationRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[internalTime.Duration](t, time.Duration())
}

// TestDurationRuleSet_ErrorConfig tests:
// - DurationRuleSet implements error configuration methods
func TestDurationRuleSet_ErrorConfig(t *testing.T) {
	testhelpers.MustImplementErrorConfig[internalTime.Duration, *time.DurationRuleSet](t, time.Duration())
}

// TestDurationRuleSet_NoConflict_ParentNil tests:
// - noConflict returns nil when parent is nil and there's a conflict
func TestDurationRuleSet_NoConflict_ParentNil(t *testing.T) {
	// Create a rule set with a min rule
	ruleSet := time.Duration().WithMin(1 * internalTime.Hour)
	
	// Add a conflicting min rule - this should replace the first one
	// When the base rule set (with nil parent) has a conflict, noConflict should handle it
	ruleSet2 := ruleSet.WithMin(2 * internalTime.Hour)
	
	// Verify the new min is in effect
	var output internalTime.Duration
	err := ruleSet2.Apply(context.TODO(), 1*internalTime.Hour, &output)
	if err == nil {
		t.Error("Expected error for 1h (below new min of 2h)")
	}
	
	err = ruleSet2.Apply(context.TODO(), 2*internalTime.Hour, &output)
	if err != nil {
		t.Errorf("Expected no error for 2h (at new min), got %s", err)
	}
}

// TestDurationRuleSet_NoConflict_ParentChanged tests:
// - noConflict handles the case where parent changes
func TestDurationRuleSet_NoConflict_ParentChanged(t *testing.T) {
	// Create a chain: base -> WithMin(1h) -> WithMax(2h)
	base := time.Duration()
	withMin := base.WithMin(1 * internalTime.Hour)
	withMax := withMin.WithMax(2 * internalTime.Hour)
	
	// Now add a conflicting min rule, which should replace the first WithMin
	// This should trigger the parent changed path in noConflict
	newMin := withMax.WithMin(30 * internalTime.Minute)
	
	// Verify the new rule set has the new min
	var output internalTime.Duration
	err := newMin.Apply(context.TODO(), 30*internalTime.Minute, &output)
	if err != nil {
		t.Errorf("Expected no error for new min threshold, got %s", err)
	}
	
	// Verify the old min is gone (30m should pass with new min of 30m)
	err = newMin.Apply(context.TODO(), 1*internalTime.Hour, &output)
	if err != nil {
		t.Errorf("Expected no error for 1h (above new min), got %s", err)
	}
	
	// Verify the max is still there
	err = newMin.Apply(context.TODO(), 3*internalTime.Hour, &output)
	if err == nil {
		t.Error("Expected error for 3h (above max of 2h)")
	}
}

// TestDurationRuleSet_Replaces_NonDurationRuleSet tests:
// - Replaces returns false for non-DurationRuleSet rules
func TestDurationRuleSet_Replaces_NonDurationRuleSet(t *testing.T) {
	// Create a rule set and add a custom rule (not a DurationRuleSet)
	mockRule := testhelpers.NewMockRule[internalTime.Duration]()
	ruleSet := time.Duration().WithRuleFunc(mockRule.Function())
	
	// Add WithRequired which creates a conflictTypeDurationRequired
	// This should not conflict with the mock rule since mock rule is not a DurationRuleSet
	// The Replaces method should return false for non-DurationRuleSet rules
	ruleSet2 := ruleSet.WithRequired()
	
	// Both rules should be present - the custom rule and WithRequired
	if ruleSet2 == nil {
		t.Error("Expected rule set not to be nil")
	}
	
	// Verify both are evaluated
	dur := 1 * internalTime.Hour
	var output internalTime.Duration
	err := ruleSet2.Apply(context.TODO(), dur, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	
	// The mock rule should have been called
	if mockRule.EvaluateCallCount() != 1 {
		t.Errorf("Expected mock rule to be called once, got %d", mockRule.EvaluateCallCount())
	}
}
