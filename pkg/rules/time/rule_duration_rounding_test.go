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

	// 5.5 seconds should round down to 5
	var output int64
	err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5 {
		t.Errorf("Expected 5, got %d", output)
	}

	// 5.9 seconds should round down to 5
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+900*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5 {
		t.Errorf("Expected 5, got %d", output)
	}
}

// TestDurationRuleSet_WithRounding_RoundingUp tests:
// - RoundingUp ceils the value when there's a remainder
func TestDurationRuleSet_WithRounding_RoundingUp(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingUp)

	// 5.1 seconds should round up to 6
	var output int64
	err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+100*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6 {
		t.Errorf("Expected 6, got %d", output)
	}

	// 5.9 seconds should round up to 6
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+900*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6 {
		t.Errorf("Expected 6, got %d", output)
	}
}

// TestDurationRuleSet_WithRounding_RoundingHalfUp tests:
// - RoundingHalfUp rounds to nearest, ties round up
func TestDurationRuleSet_WithRounding_RoundingHalfUp(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfUp)

	// 5.4 seconds should round down to 5
	var output int64
	err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+400*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5 {
		t.Errorf("Expected 5, got %d", output)
	}

	// 5.5 seconds should round up to 6
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6 {
		t.Errorf("Expected 6, got %d", output)
	}

	// 5.6 seconds should round up to 6
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+600*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6 {
		t.Errorf("Expected 6, got %d", output)
	}
}

// TestDurationRuleSet_WithRounding_RoundingHalfEven tests:
// - RoundingHalfEven rounds to nearest, ties round to even
func TestDurationRuleSet_WithRounding_RoundingHalfEven(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfEven)

	// 5.4 seconds should round down to 5
	var output int64
	err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+400*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5 {
		t.Errorf("Expected 5, got %d", output)
	}

	// 5.5 seconds (odd) should round up to 6 (even)
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6 {
		t.Errorf("Expected 6, got %d", output)
	}

	// 6.5 seconds (even) should round down to 6 (even)
	err = ruleSet.Apply(context.TODO(), 6*internalTime.Second+500*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6 {
		t.Errorf("Expected 6, got %d", output)
	}
}

// TestDurationRuleSet_WithRounding_NoRounding tests:
// - Default (RoundingNone) errors when there's a remainder
func TestDurationRuleSet_WithRounding_NoRounding(t *testing.T) {
	// Don't set rounding explicitly - default is RoundingNone
	ruleSet := time.Duration().WithUnit(internalTime.Second)

	// 5.5 seconds should error when output is numeric
	var output int64
	err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond, &output)
	if err == nil {
		t.Error("Expected error for non-evenly divisible duration")
	} else if err.Code() != errors.CodeRange {
		t.Errorf("Expected CodeRange, got %s", err.Code())
	}

	// 5 seconds (exact) should work
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 5 {
		t.Errorf("Expected 5, got %d", output)
	}
}

// TestDurationRuleSet_WithRounding_RangeError tests:
// - Range error when value doesn't fit in output type
func TestDurationRuleSet_WithRounding_RangeError(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Nanosecond).WithRounding(rules.RoundingHalfUp)

	// Try to convert a very large duration (200 seconds = 200000000000 nanoseconds) to int8 (max 127)
	// This should fail with a range error
	var output int8
	err := ruleSet.Apply(context.TODO(), 200*internalTime.Second, &output)
	if err == nil {
		t.Error("Expected range error for value that doesn't fit in int8")
	} else if err.Code() != errors.CodeRange {
		t.Errorf("Expected CodeRange, got %s", err.Code())
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

// TestDurationRuleSet_WithRounding_InterfaceWithNumeric tests:
// - An int wrapped in any should work the same as a regular int
func TestDurationRuleSet_WithRounding_InterfaceWithNumeric(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second)

	// 5.5 seconds should error when output is an interface containing int64
	var output any = int64(0)
	err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond, &output)
	if err == nil {
		t.Error("Expected error for non-evenly divisible duration with interface output")
	} else if err.Code() != errors.CodeRange {
		t.Errorf("Expected CodeRange, got %s", err.Code())
	}

	// 5 seconds (exact) should work
	output = int64(0)
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != int64(5) {
		t.Errorf("Expected 5, got %v", output)
	}
}

// TestDurationRuleSet_WithRounding_HalfEvenRoundUp tests:
// - RoundingHalfEven rounds up when remainder > halfUnit
func TestDurationRuleSet_WithRounding_HalfEvenRoundUp(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfEven)

	// 5.6 seconds - remainder (600ms) > halfUnit (500ms), should round up to 6
	var output int64
	err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+600*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if output != 6 {
		t.Errorf("Expected 6, got %d", output)
	}
}

// TestDurationRuleSet_WithRounding_NilPointerInput tests:
// - Nil *time.Duration input returns type error
func TestDurationRuleSet_WithRounding_NilPointerInput(t *testing.T) {
	ruleSet := time.Duration()

	var input *internalTime.Duration
	var output internalTime.Duration
	err := ruleSet.Apply(context.TODO(), input, &output)
	if err == nil {
		t.Error("Expected error for nil *time.Duration input")
	} else if err.Code() != errors.CodeType {
		t.Errorf("Expected CodeType, got %s", err.Code())
	}
}

// TestDurationRuleSet_WithRounding_DurationOutput tests:
// - time.Duration output without rounding is set directly
// - time.Duration output with rounding is rounded to the unit
func TestDurationRuleSet_WithRounding_DurationOutput(t *testing.T) {
	// Without rounding, duration is set directly
	ruleSet := time.Duration().WithUnit(internalTime.Second)
	var output internalTime.Duration
	err := ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	// Output should be 5.5 seconds exactly (no rounding)
	if output != 5*internalTime.Second+500*internalTime.Millisecond {
		t.Errorf("Expected 5.5s, got %s", output)
	}

	// With rounding, duration should be rounded to the nearest unit
	ruleSet = time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfUp)
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+500*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	// Output should be 6 seconds (rounded up)
	if output != 6*internalTime.Second {
		t.Errorf("Expected 6s, got %s", output)
	}

	// Test rounding down
	ruleSet = time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingDown)
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+900*internalTime.Millisecond, &output)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	// Output should be 5 seconds (rounded down)
	if output != 5*internalTime.Second {
		t.Errorf("Expected 5s, got %s", output)
	}

	// Test with interface containing time.Duration
	ruleSet = time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfUp)
	var outputInterface any = internalTime.Duration(0)
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second+600*internalTime.Millisecond, &outputInterface)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if d, ok := outputInterface.(internalTime.Duration); !ok || d != 6*internalTime.Second {
		t.Errorf("Expected 6s duration, got %v", outputInterface)
	}
}

// TestDurationRuleSet_WithRounding_UnsignedOutput tests:
// - Unsigned integer output types work correctly
// - Negative durations error for unsigned output
// - Overflow errors for unsigned types
func TestDurationRuleSet_WithRounding_UnsignedOutput(t *testing.T) {
	ruleSet := time.Duration().WithUnit(internalTime.Second).WithRounding(rules.RoundingHalfUp)

	// Positive duration to uint64 should work
	var outputUint64 uint64
	err := ruleSet.Apply(context.TODO(), 5*internalTime.Second, &outputUint64)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if outputUint64 != 5 {
		t.Errorf("Expected 5, got %d", outputUint64)
	}

	// Negative duration to unsigned should error with CodeRange
	err = ruleSet.Apply(context.TODO(), -5*internalTime.Second, &outputUint64)
	if err == nil {
		t.Error("Expected error for negative duration to unsigned type")
	} else if err.Code() != errors.CodeRange {
		t.Errorf("Expected CodeRange, got %s", err.Code())
	}

	// Overflow for uint8
	var outputUint8 uint8
	err = ruleSet.Apply(context.TODO(), 300*internalTime.Second, &outputUint8)
	if err == nil {
		t.Error("Expected range error for overflow")
	} else if err.Code() != errors.CodeRange {
		t.Errorf("Expected CodeRange, got %s", err.Code())
	}

	// Test interface containing unsigned int
	var outputInterface any = uint64(0)
	err = ruleSet.Apply(context.TODO(), 5*internalTime.Second, &outputInterface)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if outputInterface != uint64(5) {
		t.Errorf("Expected uint64(5), got %v", outputInterface)
	}

	// Test negative duration to interface containing unsigned int
	outputInterface = uint64(0)
	err = ruleSet.Apply(context.TODO(), -5*internalTime.Second, &outputInterface)
	if err == nil {
		t.Error("Expected error for negative duration to interface containing unsigned type")
	} else if err.Code() != errors.CodeRange {
		t.Errorf("Expected CodeRange, got %s", err.Code())
	}
}
