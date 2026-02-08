package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestFloatRuleSet_Apply tests:
// - Implements the RuleSet interface
// - Correctly applies float validation
// - Returns the correct value
func TestFloatRuleSet_Apply(t *testing.T) {
	var floatval float64
	err := rules.Float64().Apply(context.Background(), 123.0, &floatval)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if floatval != 123.0 {
		t.Errorf("Expected float 123.0.0 to be returned. Got: %f", floatval)
		return
	}

	ok := testhelpers.CheckRuleSetInterface[float64](rules.Float64())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	testhelpers.MustApplyTypes[float64](t, rules.Float64(), 123.0)
}

// TestFloatRuleSet_Apply_StrictError tests:
func TestFloatRuleSet_Apply_StrictError(t *testing.T) {
	var out float64
	err := rules.Float64().WithStrict().Apply(context.Background(), "123.0", &out)

	if len(errors.Unwrap(err)) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func tryFloatCoercion(t *testing.T, val interface{}, expected float64) {
	var actual float64
	err := rules.Float64().Apply(context.Background(), val, &actual)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}
	if expected != actual {
		t.Errorf("Expected '%f' and got '%f'", expected, actual)
		return
	}
}

// TestFloatRuleSet_Apply_CoerceFromString tests:
// - Coerces string values to floats
func TestFloatRuleSet_Apply_CoerceFromString(t *testing.T) {
	tryFloatCoercion(t, "123.0", 123.0)
}

// TestFloatRuleSet_Apply_CoerceFromFloat tests:
// - Coerces float32 values to float64
func TestFloatRuleSet_Apply_CoerceFromFloat(t *testing.T) {
	tryFloatCoercion(t, float32(123.0), 123.0)
}

// TestFloatRuleSet_Apply_CoerceFromFloat64 tests:
// - Coerces float64 values to float64
func TestFloatRuleSet_Apply_CoerceFromFloat64(t *testing.T) {
	tryFloatCoercion(t, float64(123.0), 123.0)
}

// TestFloatRuleSet_WithRequired tests:
// - WithRequired is correctly implemented
func TestFloatRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[float64](t, rules.Float64())
}

// TestFloatRuleSet_WithRuleFunc tests:
// - Custom rule functions are executed
// - Custom rules can return errors
// - Rule evaluation is called correctly
func TestFloatRuleSet_WithRuleFunc(t *testing.T) {
	var out float64
	err := rules.Float64().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[float64](1).Function()).
		Apply(context.Background(), "123.0", &out)

	if len(errors.Unwrap(err)) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}

	rule := testhelpers.NewMockRule[float64]()

	err = rules.Float64().
		WithRuleFunc(rule.Function()).
		Apply(context.Background(), 123.0, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

// TestFloatRuleSet_Any tests:
// - Any returns a RuleSet[any] implementation
func TestFloatRuleSet_Any(t *testing.T) {
	ruleSet := rules.Float64().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	}
}

// TestFloatRuleSet_String_WithRequired tests:
// - Serializes to WithRequired()
func TestFloatRuleSet_String(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.FloatRuleSet[float64]
		expected string
	}{
		{"Base", rules.Float64(), "FloatRuleSet[float64]"},
		{"WithRequired", rules.Float64().WithRequired(), "FloatRuleSet[float64].WithRequired()"},
		{"WithStrict", rules.Float64().WithStrict(), "FloatRuleSet[float64].WithStrict()"},
		{"WithNil", rules.Float64().WithNil(), "FloatRuleSet[float64].WithNil()"},
		{"WithRounding", rules.Float64().WithRounding(rules.RoundingHalfEven, 5), "FloatRuleSet[float64].WithRounding(HalfEven, 5)"},
		{"WithFixedOutput", rules.Float64().WithFixedOutput(2), "FloatRuleSet[float64].WithFixedOutput(2)"},
		{"Chained", rules.Float64().WithRequired().WithStrict(), "FloatRuleSet[float64].WithRequired().WithStrict()"},
		{"ChainedWithRounding", rules.Float64().WithRequired().WithRounding(rules.RoundingUp, 3), "FloatRuleSet[float64].WithRequired().WithRounding(Up, 3)"},
		{"ChainedAll", rules.Float64().WithRequired().WithStrict().WithRounding(rules.RoundingDown, 2), "FloatRuleSet[float64].WithRequired().WithStrict().WithRounding(Down, 2)"},
		{"ConflictResolution_Rounding", rules.Float64().WithRounding(rules.RoundingUp, 3).WithRounding(rules.RoundingDown, 2), "FloatRuleSet[float64].WithRounding(Down, 2)"},
		{"ConflictResolution_FixedOutput", rules.Float64().WithFixedOutput(2).WithFixedOutput(4), "FloatRuleSet[float64].WithFixedOutput(4)"},
		{"WithMin", rules.Float64().WithMin(5.5), "FloatRuleSet[float64].WithMin(5.5)"},
		{"ChainedWithRule", rules.Float64().WithRequired().WithMin(5.5), "FloatRuleSet[float64].WithRequired().WithMin(5.5)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ruleSet.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestFloatRuleSet_String_WithRounding tests:
// - Serializes to WithRounding(...)
func TestFloatRuleSet_String_WithRounding(t *testing.T) {
	ruleSet := rules.Float64().WithRounding(rules.RoundingHalfEven, 5)

	expected := "FloatRuleSet[float64].WithRounding(HalfEven, 5)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestFloatRuleSet_Evaluate tests:
// - Evaluate behaves like Apply.
func TestFloatRuleSet_Evaluate(t *testing.T) {
	ruleSet := rules.Float64().WithMin(5)
	testhelpers.MustEvaluate[float64](t, ruleSet, 10)
	testhelpers.MustNotEvaluate[float64](t, ruleSet, 1, errors.CodeMin)
}

// TestFloatRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestFloatRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[float64](t, rules.Float64())
}

// TestFloatRuleSet_Apply_StringOutput tests:
// - Outputs string values when output is a string type
// - Uses appropriate precision
func TestFloatRuleSet_Apply_StringOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.FloatRuleSet[float64]
		input    interface{}
		expected string
	}{
		{"DefaultPrecision", rules.Float64(), 123.456789012345, "123.456789012345"},
		{"WithFixedOutput2", rules.Float64().WithFixedOutput(2), 123.456789, "123.46"},
		{"WithFixedOutput0", rules.Float64().WithFixedOutput(0), 123.456789, "123"},
		{"WithFixedOutput5", rules.Float64().WithFixedOutput(5), 123.456789, "123.45679"},
		{"IntegerValue", rules.Float64(), 42.0, "42"},
		{"Negative", rules.Float64(), -123.456, "-123.456"},
		{"Zero", rules.Float64(), 0.0, "0"},
		{"SmallValue", rules.Float64(), 0.001, "0.001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out string
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected string %q, got %q", tt.expected, out)
			}
		})
	}
}

// TestFloatRuleSet_Apply_PointerToStringOutput tests:
// - Outputs string values when output is a pointer to string type
// - Handles nil pointer by creating a new string
func TestFloatRuleSet_Apply_PointerToStringOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.FloatRuleSet[float64]
		input    interface{}
		expected string
	}{
		{"DefaultPrecision", rules.Float64(), 123.456, "123.456"},
		{"WithFixedOutput2", rules.Float64().WithFixedOutput(2), 123.456, "123.46"},
		{"Negative", rules.Float64(), -42.5, "-42.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_NilPointer", func(t *testing.T) {
			var out *string
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out == nil {
				t.Error("Expected pointer to be non-nil")
				return
			}

			if *out != tt.expected {
				t.Errorf("Expected string %q, got %q", tt.expected, *out)
			}
		})

		t.Run(tt.name+"_ExistingPointer", func(t *testing.T) {
			existing := "existing"
			out := &existing
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out == nil {
				t.Error("Expected pointer to be non-nil")
				return
			}

			if *out != tt.expected {
				t.Errorf("Expected string %q, got %q", tt.expected, *out)
			}
		})
	}
}

// TestFloatRuleSet_Apply_StringOutput_Float32 tests:
// - String output works with float32 type
func TestFloatRuleSet_Apply_StringOutput_Float32(t *testing.T) {
	var out string
	err := rules.Float32().Apply(context.Background(), float32(123.456), &out)

	if err != nil {
		t.Errorf("Expected no errors, got: %v", err)
		return
	}

	// Float32 should format with appropriate precision
	if out == "" {
		t.Error("Expected non-empty string")
	}
}

// TestFloatRuleSet_Apply_StringOutput_WithRounding tests:
// - When rounding is applied, the value is rounded first
// - Output formatting uses 'g' format by default (no trailing zeros)
func TestFloatRuleSet_Apply_StringOutput_WithRounding(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.FloatRuleSet[float64]
		input    float64
		expected string
	}{
		{"RoundingHalfEven_Precision2", rules.Float64().WithRounding(rules.RoundingHalfEven, 2), 123.456, "123.46"},
		{"RoundingHalfUp_Precision2", rules.Float64().WithRounding(rules.RoundingHalfUp, 2), 123.455, "123.46"},
		{"RoundingDown_Precision2", rules.Float64().WithRounding(rules.RoundingDown, 2), 123.456, "123.45"},
		{"RoundingUp_Precision2", rules.Float64().WithRounding(rules.RoundingUp, 2), 123.451, "123.46"},
		{"RoundingHalfEven_Precision0", rules.Float64().WithRounding(rules.RoundingHalfEven, 0), 123.5, "124"},
		{"RoundingDown_Precision0", rules.Float64().WithRounding(rules.RoundingDown, 0), 123.9, "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out string
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected string %q, got %q", tt.expected, out)
			}
		})
	}
}

// TestFloatRuleSet_WithFixedOutput tests:
// - WithFixedOutput controls string output precision
// - Values are zero-padded to the specified precision
func TestFloatRuleSet_WithFixedOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.FloatRuleSet[float64]
		input    float64
		expected string
	}{
		{"Precision0", rules.Float64().WithFixedOutput(0), 123.456, "123"},
		{"Precision1", rules.Float64().WithFixedOutput(1), 123.456, "123.5"},
		{"Precision2", rules.Float64().WithFixedOutput(2), 123.456, "123.46"},
		{"Precision3", rules.Float64().WithFixedOutput(3), 123.456, "123.456"},
		{"Precision4_ZeroPad", rules.Float64().WithFixedOutput(4), 123.4, "123.4000"},
		{"Precision2_Integer", rules.Float64().WithFixedOutput(2), 42.0, "42.00"},
		{"Precision2_Negative", rules.Float64().WithFixedOutput(2), -123.456, "-123.46"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out string
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected string %q, got %q", tt.expected, out)
			}
		})
	}
}

// TestFloatRuleSet_WithFixedOutput_WithRounding tests:
// - WithFixedOutput and WithRounding can be combined
// - Rounding is applied first, then output is formatted with fixed precision
func TestFloatRuleSet_WithFixedOutput_WithRounding(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.FloatRuleSet[float64]
		input    float64
		expected string
	}{
		// Rounding to 2 decimal places, output with 2 decimal places
		{"Round2_Output2", rules.Float64().WithRounding(rules.RoundingHalfEven, 2).WithFixedOutput(2), 123.456, "123.46"},
		// Rounding to 2 decimal places, output with 4 decimal places (zero-padded)
		{"Round2_Output4", rules.Float64().WithRounding(rules.RoundingHalfEven, 2).WithFixedOutput(4), 123.456, "123.4600"},
		// Rounding to 0 decimal places, output with 2 decimal places (zero-padded)
		{"Round0_Output2", rules.Float64().WithRounding(rules.RoundingHalfEven, 0).WithFixedOutput(2), 123.456, "123.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out string
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected string %q, got %q", tt.expected, out)
			}
		})
	}
}

// TestFloatRuleSet_String_WithFixedOutput tests:
// - Serializes to WithFixedOutput(...)
func TestFloatRuleSet_String_WithFixedOutput(t *testing.T) {
	ruleSet := rules.Float64().WithFixedOutput(2)

	expected := "FloatRuleSet[float64].WithFixedOutput(2)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestFloatRuleSet_WithFixedOutput_EdgeCases tests edge cases for WithFixedOutput:
// - Zero-padding when value has no decimals
// - Precision 0 with integer values
// - Values that already have exact precision
func TestFloatRuleSet_WithFixedOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.FloatRuleSet[float64]
		input    float64
		expected string
	}{
		// Zero-padding integer values
		{"ZeroPad_Integer", rules.Float64().WithFixedOutput(3), 100.0, "100.000"},
		// Precision 0 with integer (no decimal point in formatted output)
		{"Precision0_Integer", rules.Float64().WithFixedOutput(0), 100.0, "100"},
		// Value already has exact precision needed (no padding required)
		{"ExactPrecision_NoPadding", rules.Float64().WithFixedOutput(2), 123.45, "123.45"},
		// Value with exact precision from rounding
		{"ExactPrecision_FromFormat", rules.Float64().WithFixedOutput(6), 123.456789, "123.456789"},
		// Zero value with padding
		{"Zero_WithPadding", rules.Float64().WithFixedOutput(3), 0.0, "0.000"},
		// Small value with extra padding
		{"Small_WithPadding", rules.Float64().WithFixedOutput(5), 0.1, "0.10000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out string
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected string %q, got %q", tt.expected, out)
			}
		})
	}
}

// TestFloatRuleSet_WithRounding_TrailingZeros tests that trailing zeros are trimmed after rounding:
// - Rounding to a precision where result ends in zeros
// - Integer results after rounding
func TestFloatRuleSet_WithRounding_TrailingZeros(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.FloatRuleSet[float64]
		input    float64
		expected string
	}{
		// Rounding results in trailing zeros that should be trimmed
		{"TrailingZeros_Trimmed", rules.Float64().WithRounding(rules.RoundingHalfEven, 3), 123.400, "123.4"},
		// Rounding to integer (all decimals become zero)
		{"AllZeros_Trimmed", rules.Float64().WithRounding(rules.RoundingHalfEven, 2), 100.00, "100"},
		// No trailing zeros to trim
		{"NoTrailingZeros", rules.Float64().WithRounding(rules.RoundingHalfEven, 2), 123.45, "123.45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out string
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected string %q, got %q", tt.expected, out)
			}
		})
	}
}

// TestFloatRuleSet_Float32_StringOutput tests string output with float32 type:
// - Exercises the float32 branch in formatFloat
func TestFloatRuleSet_Float32_StringOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.FloatRuleSet[float32]
		input    float32
		expected string
	}{
		{"Default", rules.Float32(), float32(123.456), "123.456"},
		{"WithFixedOutput", rules.Float32().WithFixedOutput(2), float32(123.456), "123.46"},
		{"WithRounding", rules.Float32().WithRounding(rules.RoundingHalfEven, 1), float32(123.456), "123.5"},
		{"Integer", rules.Float32(), float32(42.0), "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out string
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected string %q, got %q", tt.expected, out)
			}
		})
	}
}

// TestFloatRuleSet_WithFixedOutput_PointerToString tests pointer to string output with fixed precision
func TestFloatRuleSet_WithFixedOutput_PointerToString(t *testing.T) {
	var out *string
	err := rules.Float64().WithFixedOutput(2).Apply(context.Background(), 123.4, &out)

	if err != nil {
		t.Errorf("Expected no errors, got: %v", err)
		return
	}

	if out == nil {
		t.Error("Expected pointer to be non-nil")
		return
	}

	if *out != "123.40" {
		t.Errorf("Expected string %q, got %q", "123.40", *out)
	}
}

// TestFloatRuleSet_ErrorConfig tests:
// - FloatRuleSet implements error configuration methods
func TestFloatRuleSet_ErrorConfig(t *testing.T) {
	testhelpers.MustImplementErrorConfig[float64, *rules.FloatRuleSet[float64]](t, rules.Float64())
}

// TestFloatRuleSet_Apply_CoerceFromBool tests:
// - Coerces bool values to floats (true -> 1.0, false -> 0.0)
func TestFloatRuleSet_Apply_CoerceFromBool(t *testing.T) {
	tryFloatCoercion(t, true, 1.0)
	tryFloatCoercion(t, false, 0.0)
}

// TestFloatRuleSet_Apply_CoerceFromBool_Strict tests:
// - Strict mode rejects bool values
func TestFloatRuleSet_Apply_CoerceFromBool_Strict(t *testing.T) {
	var out float64
	err := rules.Float64().WithStrict().Apply(context.Background(), true, &out)

	if len(errors.Unwrap(err)) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

// TestFloatRuleSet_Apply_BoolOutput tests:
// - Outputs bool values when output is a bool type (non-zero = true, zero = false)
func TestFloatRuleSet_Apply_BoolOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.FloatRuleSet[float64]
		input    interface{}
		expected bool
	}{
		{"NonZero", rules.Float64(), 42.5, true},
		{"Zero", rules.Float64(), 0.0, false},
		{"Negative", rules.Float64(), -1.0, true},
		{"One", rules.Float64(), 1.0, true},
		{"SmallPositive", rules.Float64(), 0.001, true},
		{"SmallNegative", rules.Float64(), -0.001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bool
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected bool %v, got %v", tt.expected, out)
			}
		})
	}
}
