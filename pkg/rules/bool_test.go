package rules_test

import (
	"context"
	"reflect"
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestBoolRuleSet_Apply tests:
// - Implements the RuleSet interface
// - Correctly applies boolean validation
// - Returns the correct value
func TestBoolRuleSet_Apply(t *testing.T) {
	var boolval bool
	err := rules.Bool().Apply(context.Background(), true, &boolval)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if !boolval {
		t.Error("Expected bool true to be returned. Got: false")
		return
	}

	ok := testhelpers.CheckRuleSetInterface[bool](rules.Bool())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	testhelpers.MustApplyTypes[bool](t, rules.Bool(), true)
}

// TestBoolRuleSet_Apply_StrictError tests:
// - Returns error when strict mode is enabled and input is not a boolean
func TestBoolRuleSet_Apply_StrictError(t *testing.T) {
	var out bool
	err := rules.Bool().WithStrict().Apply(context.Background(), "true", &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func tryBoolCoercion(t *testing.T, val interface{}, expected bool) {
	var actual bool
	err := rules.Bool().Apply(context.Background(), val, &actual)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}
	if expected != actual {
		t.Errorf("Expected '%v' and got '%v'", expected, actual)
		return
	}
}

// TestBoolRuleSet_Apply_CoerceFromString tests:
// - Coerces string values to booleans
func TestBoolRuleSet_Apply_CoerceFromString(t *testing.T) {
	tryBoolCoercion(t, "true", true)
	tryBoolCoercion(t, "false", false)
	tryBoolCoercion(t, "TRUE", true)
	tryBoolCoercion(t, "FALSE", false)
	tryBoolCoercion(t, "True", true)
	tryBoolCoercion(t, "False", false)
	tryBoolCoercion(t, "1", true)
	tryBoolCoercion(t, "0", false)
	tryBoolCoercion(t, "t", true)
	tryBoolCoercion(t, "f", false)
	tryBoolCoercion(t, "T", true)
	tryBoolCoercion(t, "F", false)
}

// TestBoolRuleSet_Apply_CoerceFromString_Invalid tests:
// - Returns error for invalid string values
func TestBoolRuleSet_Apply_CoerceFromString_Invalid(t *testing.T) {
	var out bool
	err := rules.Bool().Apply(context.Background(), "invalid", &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

// TestBoolRuleSet_Apply_CoerceFromInt tests:
// - Coerces integer values to booleans (non-zero = true, zero = false)
func TestBoolRuleSet_Apply_CoerceFromInt(t *testing.T) {
	tryBoolCoercion(t, int(1), true)
	tryBoolCoercion(t, int(0), false)
	tryBoolCoercion(t, int(-1), true)
	tryBoolCoercion(t, int(42), true)
	tryBoolCoercion(t, int8(1), true)
	tryBoolCoercion(t, int8(0), false)
	tryBoolCoercion(t, int16(1), true)
	tryBoolCoercion(t, int16(0), false)
	tryBoolCoercion(t, int32(1), true)
	tryBoolCoercion(t, int32(0), false)
	tryBoolCoercion(t, int64(1), true)
	tryBoolCoercion(t, int64(0), false)
	tryBoolCoercion(t, uint(1), true)
	tryBoolCoercion(t, uint(0), false)
	tryBoolCoercion(t, uint8(1), true)
	tryBoolCoercion(t, uint8(0), false)
	tryBoolCoercion(t, uint16(1), true)
	tryBoolCoercion(t, uint16(0), false)
	tryBoolCoercion(t, uint32(1), true)
	tryBoolCoercion(t, uint32(0), false)
	tryBoolCoercion(t, uint64(1), true)
	tryBoolCoercion(t, uint64(0), false)
}

// TestBoolRuleSet_Apply_CoerceFromFloat tests:
// - Coerces float values to booleans (non-zero = true, zero = false)
func TestBoolRuleSet_Apply_CoerceFromFloat(t *testing.T) {
	tryBoolCoercion(t, float32(1.0), true)
	tryBoolCoercion(t, float32(0.0), false)
	tryBoolCoercion(t, float32(-1.0), true)
	tryBoolCoercion(t, float32(0.5), true)
	tryBoolCoercion(t, float64(1.0), true)
	tryBoolCoercion(t, float64(0.0), false)
	tryBoolCoercion(t, float64(-1.0), true)
	tryBoolCoercion(t, float64(0.5), true)
}

// TestBoolRuleSet_Apply_CoerceFromBool tests:
// - Direct bool values work
func TestBoolRuleSet_Apply_CoerceFromBool(t *testing.T) {
	tryBoolCoercion(t, true, true)
	tryBoolCoercion(t, false, false)
}

// TestBoolRuleSet_Apply_StrictMode_Bool tests:
// - Strict mode allows bool values
func TestBoolRuleSet_Apply_StrictMode_Bool(t *testing.T) {
	var out bool
	err := rules.Bool().WithStrict().Apply(context.Background(), true, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if !out {
		t.Error("Expected true")
		return
	}
}

// TestBoolRuleSet_Apply_StrictMode_Int tests:
// - Strict mode rejects integer values
func TestBoolRuleSet_Apply_StrictMode_Int(t *testing.T) {
	var out bool
	err := rules.Bool().WithStrict().Apply(context.Background(), 1, &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

// TestBoolRuleSet_Apply_StrictMode_Float tests:
// - Strict mode rejects float values
func TestBoolRuleSet_Apply_StrictMode_Float(t *testing.T) {
	var out bool
	err := rules.Bool().WithStrict().Apply(context.Background(), 1.0, &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

// TestBoolRuleSet_WithRequired tests:
// - WithRequired is correctly implemented
func TestBoolRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[bool](t, rules.Bool())
}

// TestBoolRuleSet_WithRuleFunc tests:
// - Custom rule functions are executed
// - Custom rules can return errors
// - Rule evaluation is called correctly
func TestBoolRuleSet_WithRuleFunc(t *testing.T) {
	var out bool
	err := rules.Bool().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[bool](1).Function()).
		Apply(context.Background(), true, &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}

	rule := testhelpers.NewMockRule[bool]()
	err = rules.Bool().
		WithRuleFunc(rule.Function()).
		Apply(context.Background(), true, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

// TestBoolRuleSet_Any tests:
// - Any returns a RuleSet[any] implementation
func TestBoolRuleSet_Any(t *testing.T) {
	ruleSet := rules.Bool().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	}
}

// TestBoolRuleSet_String tests:
// - Serializes correctly
func TestBoolRuleSet_String(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.BoolRuleSet
		expected string
	}{
		{"Base", rules.Bool(), "BoolRuleSet"},
		{"WithRequired", rules.Bool().WithRequired(), "BoolRuleSet.WithRequired()"},
		{"WithStrict", rules.Bool().WithStrict(), "BoolRuleSet.WithStrict()"},
		{"WithNil", rules.Bool().WithNil(), "BoolRuleSet.WithNil()"},
		{"Chained", rules.Bool().WithRequired().WithStrict(), "BoolRuleSet.WithRequired().WithStrict()"},
		{"ChainedAll", rules.Bool().WithRequired().WithStrict().WithNil(), "BoolRuleSet.WithRequired().WithStrict().WithNil()"},
		{"ConflictResolution_Required", rules.Bool().WithRequired().WithRequired(), "BoolRuleSet.WithRequired()"},
		{"ConflictResolution_Strict", rules.Bool().WithStrict().WithStrict(), "BoolRuleSet.WithStrict()"},
		{"ConflictResolution_Nil", rules.Bool().WithNil().WithNil(), "BoolRuleSet.WithNil()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ruleSet.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestBoolRuleSet_Evaluate tests:
// - Evaluate behaves like Apply.
func TestBoolRuleSet_Evaluate(t *testing.T) {
	ruleSet := rules.Bool()
	testhelpers.MustEvaluate[bool](t, ruleSet, true)
	testhelpers.MustEvaluate[bool](t, ruleSet, false)
}

// TestBoolRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestBoolRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[bool](t, rules.Bool())
}

// TestBoolRuleSet_Apply_StringOutput tests:
// - Outputs string values when output is a string type
func TestBoolRuleSet_Apply_StringOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.BoolRuleSet
		input    interface{}
		expected string
	}{
		{"True", rules.Bool(), true, "true"},
		{"False", rules.Bool(), false, "false"},
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

// TestBoolRuleSet_Apply_PointerToStringOutput tests:
// - Outputs string values when output is a pointer to string type
// - Handles nil pointer by creating a new string
func TestBoolRuleSet_Apply_PointerToStringOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.BoolRuleSet
		input    interface{}
		expected string
	}{
		{"True", rules.Bool(), true, "true"},
		{"False", rules.Bool(), false, "false"},
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

// TestBoolRuleSet_Apply_VariantTypes tests:
// - Applies correctly to various output types
func TestBoolRuleSet_Apply_VariantTypes(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  rules.RuleSet[any]
		input    interface{}
		expected interface{}
	}{
		{"Bool", rules.Bool().Any(), true, true},
		{"Bool_False", rules.Bool().Any(), false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out any
			err := tt.ruleSet.Apply(context.Background(), tt.input, &out)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if reflect.TypeOf(out) != reflect.TypeOf(tt.expected) {
				t.Errorf("Expected type %T, got %T", tt.expected, out)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected value %v, got %v", tt.expected, out)
				return
			}
		})
	}
}

// TestBoolRuleSet_ErrorConfig tests:
// - BoolRuleSet implements error configuration methods
func TestBoolRuleSet_ErrorConfig(t *testing.T) {
	testhelpers.MustImplementErrorConfig[bool, *rules.BoolRuleSet](t, rules.Bool())
}

// TestBoolRuleSet_Apply_PointerToBool tests:
// - Handles pointer to bool input
func TestBoolRuleSet_Apply_PointerToBool(t *testing.T) {
	trueVal := true
	falseVal := false

	var out bool
	err := rules.Bool().Apply(context.Background(), &trueVal, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if !out {
		t.Error("Expected true")
		return
	}

	err = rules.Bool().Apply(context.Background(), &falseVal, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out {
		t.Error("Expected false")
		return
	}
}

// TestBoolRuleSet_Apply_PointerToBool_Nil tests:
// - Handles nil pointer to bool input
func TestBoolRuleSet_Apply_PointerToBool_Nil(t *testing.T) {
	var out bool
	var nilBool *bool
	err := rules.Bool().Apply(context.Background(), nilBool, &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

// TestBoolRuleSet_Evaluate_WithErrors tests:
// - Evaluate returns errors when rules fail
func TestBoolRuleSet_Evaluate_WithErrors(t *testing.T) {
	ruleSet := rules.Bool().WithRuleFunc(testhelpers.NewMockRuleWithErrors[bool](1).Function())
	err := ruleSet.Evaluate(context.Background(), true)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

// TestBoolRuleSet_String_WithRule tests:
// - String() includes rule string when label is empty
func TestBoolRuleSet_String_WithRule(t *testing.T) {
	rule := testhelpers.NewMockRule[bool]()
	ruleSet := rules.Bool().WithRule(rule)
	
	// The String() method should include the rule's string representation
	str := ruleSet.String()
	if str == "" {
		t.Error("Expected non-empty string")
		return
	}
}

// TestBoolRuleSet_noConflict_EdgeCases tests:
// - noConflict handles various edge cases
func TestBoolRuleSet_noConflict_EdgeCases(t *testing.T) {
	// Test with a rule that doesn't implement getConflictType
	rule := testhelpers.NewMockRule[bool]()
	ruleSet := rules.Bool().WithRule(rule)
	
	// This should not panic and should work correctly
	str := ruleSet.String()
	if str == "" {
		t.Error("Expected non-empty string")
		return
	}
}

// TestBoolRuleSet_coerceBool_EdgeCases tests:
// - coerceBool handles various edge cases
func TestBoolRuleSet_coerceBool_EdgeCases(t *testing.T) {
	// Test with unsupported type in strict mode
	var out bool
	err := rules.Bool().WithStrict().Apply(context.Background(), []int{1, 2, 3}, &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}

	// Test with unsupported type in non-strict mode
	err = rules.Bool().Apply(context.Background(), []int{1, 2, 3}, &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}
