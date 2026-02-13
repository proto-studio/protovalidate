package time_test

import (
	"context"
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/time"
)

// TestDurationRuleSet_WithRounding_RoundingDown tests:
// - RoundingDown floors the value when there's a remainder
func TestDurationRuleSet_WithRounding_RoundingDown(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingDown)

	// 5.5 seconds should round down to 5 seconds
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}

	// 5.9 seconds should round down to 5 seconds
	output, err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+900*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}
}

// TestDurationRuleSet_WithRounding_RoundingUp tests:
// - RoundingUp ceils the value when there's a remainder
func TestDurationRuleSet_WithRounding_RoundingUp(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingUp)

	// 5.1 seconds should round up to 6 seconds
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+100*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6*internalTime.Second {
		t.Errorf("Expected 6s, got %s", output)
	}

	// 5.9 seconds should round up to 6 seconds
	output, err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+900*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6*internalTime.Second {
		t.Errorf("Expected 6s, got %s", output)
	}
}

// TestDurationRuleSet_WithRounding_RoundingHalfUp tests:
// - RoundingHalfUp rounds to nearest, ties round up
func TestDurationRuleSet_WithRounding_RoundingHalfUp(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfUp)

	// 5.4 seconds should round down to 5 seconds
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+400*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}

	// 5.5 seconds should round up to 6 seconds
	output, err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6*internalTime.Second {
		t.Errorf("Expected 6s, got %s", output)
	}

	// 5.6 seconds should round up to 6 seconds
	output, err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+600*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6*internalTime.Second {
		t.Errorf("Expected 6s, got %s", output)
	}
}

// TestDurationRuleSet_WithRounding_RoundingHalfEven tests:
// - RoundingHalfEven rounds to nearest, ties round to even
func TestDurationRuleSet_WithRounding_RoundingHalfEven(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfEven)

	// 5.4 seconds should round down to 5 seconds
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+400*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}

	// 5.5 seconds (odd) should round up to 6 (even)
	output, err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6*internalTime.Second {
		t.Errorf("Expected 6s, got %s", output)
	}

	// 6.5 seconds (even) should round down to 6 (even)
	output, err = ruleSet.Apply(context.TODO(), 6*internalTime.Second+500*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6*internalTime.Second {
		t.Errorf("Expected 6s, got %s", output)
	}
}

// TestDurationRuleSet_WithRounding_NoRounding tests:
// - Default (RoundingNone) errors when there's a remainder
func TestDurationRuleSet_WithRounding_NoRounding(t *testing.T) {
	// Don't set rounding explicitly - default is RoundingNone
	ruleSet := time.Duration().WithUnit(internalTime.Second)

	// 5.5 seconds should error when not evenly divisible (no rounding)
	_, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond)
	if err == nil {
		t.Error("Expected error for non-evenly divisible duration")
	} else if err.Code() != errors.CodeRange {
		t.Errorf("Expected CodeRange, got %s", err.Code())
	}

	// 5 seconds (exact) should work
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}
}

// TestDurationRuleSet_WithRounding_RangeError tests:
// - Apply returns duration; very large duration is still returned (no int8 output type)
func TestDurationRuleSet_WithRounding_RangeError(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Nanosecond).WithRounding(rules.RoundingHalfUp)

	// 200 seconds converts to 200e9 ns; Apply returns time.Duration
	output, err := ruleSet.Apply(context.TODO(), 200*internalTime.Second)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 200*internalTime.Second {
		t.Errorf("Expected 200s, got %s", output)
	}
}

// TestDurationRuleSet_WithRounding_Conflict tests:
// - Most recent rounding setting is used
func TestDurationRuleSet_WithRounding_Conflict(t *testing.T) {
	ruleSet := time.Duration().WithRounding(rules.RoundingDown)
	ruleSet2 := ruleSet.WithRounding(rules.RoundingUp)

	// Verify string representation shows the most recent rounding
	expected := "DurationRuleSet.WithRounding(Up)"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestDurationRuleSet_WithRounding_InterfaceWithNumeric tests Apply returns duration
func TestDurationRuleSet_WithRounding_InterfaceWithNumeric(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second)

	// 5.5 seconds should error when no rounding (not evenly divisible)
	_, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond)
	if err == nil {
		t.Error("Expected error for non-evenly divisible duration")
	} else if err.Code() != errors.CodeRange {
		t.Errorf("Expected CodeRange, got %s", err.Code())
	}

	// 5 seconds (exact) should work
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}
}

// TestDurationRuleSet_WithRounding_HalfEvenRoundUp tests:
// - RoundingHalfEven rounds up when remainder > halfUnit
func TestDurationRuleSet_WithRounding_HalfEvenRoundUp(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfEven)

	// 5.6 seconds - remainder (600ms) > halfUnit (500ms), should round up to 6 seconds
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+600*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6*internalTime.Second {
		t.Errorf("Expected 6s, got %s", output)
	}
}

// TestDurationRuleSet_WithRounding_NilPointerInput tests:
// - Nil *time.Duration input returns type error
func TestDurationRuleSet_WithRounding_NilPointerInput(t *testing.T) {
	ruleSet := time.Duration()

	var input *internalTime.Duration
	_, err := ruleSet.Apply(context.TODO(), input)
	if err == nil {
		t.Error("Expected error for nil *time.Duration input")
	} else if err.Code() != errors.CodeType {
		t.Errorf("Expected CodeType, got %s", err.Code())
	}
}

// TestDurationRuleSet_WithRounding_DurationOutput tests:
// - Without rounding, duration is returned directly
// - With rounding, duration is rounded to the unit
func TestDurationRuleSet_WithRounding_DurationOutput(t *testing.T) {
	// Without rounding, duration not evenly divisible by unit returns error
	ruleSet := time.Duration().WithUnit(internalTime.Second)
	_, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond)
	if err == nil {
		t.Error("Expected error for 5.5s when unit is second and no rounding")
	}

	// With rounding, duration should be rounded to the nearest unit
	ruleSet = time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfUp)
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6*internalTime.Second {
		t.Errorf("Expected 6s, got %s", output)
	}

	// Test rounding down
	ruleSet = time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingDown)
	output, err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+900*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}

	// Same with different input
	ruleSet = time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfUp)
	output, err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+600*internalTime.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6*internalTime.Second {
		t.Errorf("Expected 6s duration, got %s", output)
	}
}

// TestDurationRuleSet_WithRounding_UnsignedOutput tests Apply returns duration for positive and negative
func TestDurationRuleSet_WithRounding_UnsignedOutput(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfUp)

	// Positive duration
	output, err := ruleSet.Apply(context.TODO(), 5*internalTime.Second)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}

	// Negative duration is returned as-is (no unsigned output type)
	output, err = ruleSet.Apply(context.TODO(), -5*internalTime.Second)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != -5*internalTime.Second {
		t.Errorf("Expected -5s, got %s", output)
	}
}
