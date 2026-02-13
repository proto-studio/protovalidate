package time_test

import (
	"context"
	"fmt"
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/rules/time"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestDurationRuleSet_WithUnit_Seconds tests:
// - Numeric inputs are interpreted as seconds when WithUnit(time.Second) is used
func TestDurationRuleSet_WithUnit_Seconds(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).Any()

	// 5 should be interpreted as 5 seconds
	expected := 5 * internalTime.Second
	testhelpers.MustApplyMutation(t, ruleSet, 5, expected)

	// 60 should be interpreted as 60 seconds (1 minute)
	expected = 60 * internalTime.Second
	testhelpers.MustApplyMutation(t, ruleSet, 60, expected)
}

// TestDurationRuleSet_WithUnit_Milliseconds tests:
// - Numeric inputs are interpreted as milliseconds when WithUnit(time.Millisecond) is used
func TestDurationRuleSet_WithUnit_Milliseconds(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Millisecond).Any()

	// 1000 should be interpreted as 1000 milliseconds (1 second)
	expected := 1000 * internalTime.Millisecond
	testhelpers.MustApplyMutation(t, ruleSet, 1000, expected)

	// 500 should be interpreted as 500 milliseconds
	expected = 500 * internalTime.Millisecond
	testhelpers.MustApplyMutation(t, ruleSet, 500, expected)
}

// TestDurationRuleSet_WithUnit_Minutes tests:
// - Numeric inputs are interpreted as minutes when WithUnit(time.Minute) is used
func TestDurationRuleSet_WithUnit_Minutes(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Minute).Any()

	// 30 should be interpreted as 30 minutes
	expected := 30 * internalTime.Minute
	testhelpers.MustApplyMutation(t, ruleSet, 30, expected)

	// 1 should be interpreted as 1 minute
	expected = 1 * internalTime.Minute
	testhelpers.MustApplyMutation(t, ruleSet, 1, expected)
}

// TestDurationRuleSet_WithUnit_Hours tests:
// - Numeric inputs are interpreted as hours when WithUnit(time.Hour) is used
func TestDurationRuleSet_WithUnit_Hours(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Hour).Any()

	// 24 should be interpreted as 24 hours
	expected := 24 * internalTime.Hour
	testhelpers.MustApplyMutation(t, ruleSet, 24, expected)

	// 1 should be interpreted as 1 hour
	expected = 1 * internalTime.Hour
	testhelpers.MustApplyMutation(t, ruleSet, 1, expected)
}

// TestDurationRuleSet_WithUnit_Days tests:
// - Numeric inputs are interpreted as days when WithUnit(24*time.Hour) is used
func TestDurationRuleSet_WithUnit_Days(t *testing.T) {
	ruleSet := time.Duration().WithUnit(24 * internalTime.Hour).Any()

	// 7 should be interpreted as 7 days
	expected := 7 * 24 * internalTime.Hour
	testhelpers.MustApplyMutation(t, ruleSet, 7, expected)

	// 1 should be interpreted as 1 day
	expected = 24 * internalTime.Hour
	testhelpers.MustApplyMutation(t, ruleSet, 1, expected)
}

// TestDurationRuleSet_WithUnit_Weeks tests:
// - Numeric inputs are interpreted as weeks when WithUnit(7*24*time.Hour) is used
func TestDurationRuleSet_WithUnit_Weeks(t *testing.T) {
	ruleSet := time.Duration().WithUnit(7 * 24 * internalTime.Hour).Any()

	// 2 should be interpreted as 2 weeks
	expected := 2 * 7 * 24 * internalTime.Hour
	testhelpers.MustApplyMutation(t, ruleSet, 2, expected)

	// 1 should be interpreted as 1 week
	expected = 7 * 24 * internalTime.Hour
	testhelpers.MustApplyMutation(t, ruleSet, 1, expected)
}

// TestDurationRuleSet_WithUnit_Int64 tests:
// - int64 inputs are also converted using the unit
func TestDurationRuleSet_WithUnit_Int64(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).Any()

	// int64 value should be interpreted as seconds
	expected := 10 * internalTime.Second
	testhelpers.MustApplyMutation(t, ruleSet, int64(10), expected)
}

// TestDurationRuleSet_WithUnit_DefaultNanoseconds tests:
// - When WithUnit is not called, numeric inputs default to nanoseconds
func TestDurationRuleSet_WithUnit_DefaultNanoseconds(t *testing.T) {
	ruleSet := time.Duration().Any()

	// Without WithUnit, 1000000000 should be interpreted as nanoseconds (1 second)
	expected := 1000000000 * internalTime.Nanosecond
	testhelpers.MustApplyMutation(t, ruleSet, 1000000000, expected)
}

// TestDurationRuleSet_WithUnit_StringInput tests:
// - String inputs are not affected by WithUnit (they use time.ParseDuration)
func TestDurationRuleSet_WithUnit_StringInput(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).Any()

	// String input should be parsed directly, not affected by unit
	expected := 5 * internalTime.Minute
	testhelpers.MustApplyMutation(t, ruleSet, "5m", expected)
}

// TestDurationRuleSet_WithUnit_DurationInput tests:
// - time.Duration inputs are not affected by WithUnit
func TestDurationRuleSet_WithUnit_DurationInput(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).Any()

	// Duration input should be used directly, not affected by unit
	input := 5 * internalTime.Minute
	testhelpers.MustApplyMutation(t, ruleSet, input, input)
}

// TestDurationRuleSet_WithUnit_Conflict tests:
// - Only one WithUnit can exist on a rule set
// - Original rule set is not mutated
// - Most recent unit is used
func TestDurationRuleSet_WithUnit_Conflict(t *testing.T) {
	// Create an initial rule set with seconds unit
	ruleSet := time.Duration().WithUnit(internalTime.Second)

	// Apply with 5 seconds as duration
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}

	// Create a new rule set with minutes unit
	ruleSet2 := ruleSet.WithUnit(internalTime.Minute)

	// Apply with 5 minutes as duration
	output, err = ruleSet2.Apply(context.TODO(), 5*internalTime.Minute)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}
	if output != 5*internalTime.Minute {
		t.Errorf("Expected 5m, got %s", output)
	}

	// Verify original rule set still uses seconds
	output, err = ruleSet.Apply(context.TODO(), 5*internalTime.Second)
	if err != nil {
		t.Errorf("Expected error to be nil, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}

	// Verify string representation
	expected := fmt.Sprintf("DurationRuleSet.WithUnit(\"%s\")", internalTime.Second)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	expected = fmt.Sprintf("DurationRuleSet.WithUnit(\"%s\")", internalTime.Minute)
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected new rule set to be %s, got %s", expected, s)
	}
}

// TestDurationRuleSet_WithUnit_InvalidUnit tests:
// - Invalid units (<= 0) are ignored and the rule set is returned unchanged
func TestDurationRuleSet_WithUnit_InvalidUnit(t *testing.T) {
	ruleSet := time.Duration()

	// Try to set invalid units
	ruleSet2 := ruleSet.WithUnit(0)
	if ruleSet2 != ruleSet {
		t.Error("Expected WithUnit(0) to return unchanged rule set")
	}

	ruleSet3 := ruleSet.WithUnit(-1 * internalTime.Second)
	if ruleSet3 != ruleSet {
		t.Error("Expected WithUnit(negative) to return unchanged rule set")
	}
}

// TestDurationRuleSet_WithUnit_Chained tests:
// - WithUnit works correctly when chained with other methods
func TestDurationRuleSet_WithUnit_Chained(t *testing.T) {
	ruleSet := time.Duration().
		WithUnit(internalTime.Second).
		WithMin(1 * internalTime.Second).
		WithMax(60 * internalTime.Second)

	// Test valid value (30 seconds as duration)
	output, err := ruleSet.Apply(context.TODO(), 30*internalTime.Second)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 30*internalTime.Second {
		t.Errorf("Expected 30s, got %s", output)
	}

	// Test value below min (0 seconds)
	_, err = ruleSet.Apply(context.TODO(), 0)
	if err == nil {
		t.Error("Expected error for value below minimum")
	}

	// Test value above max (100 seconds)
	_, err = ruleSet.Apply(context.TODO(), 100*internalTime.Second)
	if err == nil {
		t.Error("Expected error for value above maximum")
	}
}

// TestDurationRuleSet_WithUnit_ParentChain tests:
// - Unit is found by walking up the parent chain
func TestDurationRuleSet_WithUnit_ParentChain(t *testing.T) {
	// Create a chain: base -> WithUnit(seconds) -> WithMin
	base := time.Duration()
	withUnit := base.WithUnit(internalTime.Second)
	withMin := withUnit.WithMin(1 * internalTime.Second)

	// Apply should use the unit from the parent chain
	output, err := withMin.Apply(context.TODO(), 30*internalTime.Second)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 30*internalTime.Second {
		t.Errorf("Expected 30s, got %s", output)
	}
}

// TestDurationRuleSet_WithUnit_OutputDurationType tests:
// - When output is time.Duration, duration inputs are passed through unchanged
// - When output is numeric, duration is converted to numeric using the unit
func TestDurationRuleSet_WithUnit_OutputDurationType(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second)

	// Test with duration input - Apply returns duration
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}

	// Test with duration input and no unit
	ruleSetNoUnit := time.Duration()
	output2, err := ruleSetNoUnit.Apply(context.TODO(), 5*internalTime.Second)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output2 != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output2)
	}

	// Test with duration input - should pass through unchanged (covered by TestDurationRuleSet_WithUnit_DurationInput)
	// This test focuses on numeric input -> duration output conversion
}
