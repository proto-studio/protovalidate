package rules_test

import (
	"context"
	"reflect"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestIntRuleSet_Apply tests:
// - Implements the RuleSet interface
// - Correctly applies integer validation
// - Returns the correct value
func TestIntRuleSet_Apply(t *testing.T) {
	var intval int
	err := rules.Int().Apply(context.Background(), 123, &intval)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if intval != 123 {
		t.Errorf("Expected int 123 to be returned. Got: %d", intval)
		return
	}

	ok := testhelpers.CheckRuleSetInterface[int](rules.Int())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	testhelpers.MustApplyTypes[int](t, rules.Int(), 123)
}

// TestIntRuleSet_Apply_StrictError tests:
// - Returns error when strict mode is enabled and input is not an integer
func TestIntRuleSet_Apply_StrictError(t *testing.T) {
	var out int
	err := rules.Int().WithStrict().Apply(context.Background(), "123", &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func tryIntCoercion(t *testing.T, val interface{}, expected int) {
	var actual int
	err := rules.Int().Apply(context.Background(), val, &actual)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}
	if expected != actual {
		t.Errorf("Expected '%d' and got '%d'", expected, actual)
		return
	}
}

// TestIntRuleSet_Apply_CoerceFromString tests:
// - Coerces string values to integers
func TestIntRuleSet_Apply_CoerceFromString(t *testing.T) {
	tryIntCoercion(t, "123", 123)
}

// TestIntRuleSet_Apply_CoerceFromFloat tests:
// - Coerces float32 values to integers
func TestIntRuleSet_Apply_CoerceFromFloat(t *testing.T) {
	tryIntCoercion(t, float32(123.0), 123)
}

// TestIntRuleSet_Apply_CoerceFromInt64 tests:
// - Coerces float64 values to integers
func TestIntRuleSet_Apply_CoerceFromInt64(t *testing.T) {
	tryIntCoercion(t, float64(123.0), 123)
}

// TestIntRuleSet_Apply_CoerceFromHex tests:
// - Coerces hexadecimal string values to integers when base is set
// - Returns error for invalid hexadecimal strings
func TestIntRuleSet_Apply_CoerceFromHex(t *testing.T) {
	expected := 0xBEEF
	var actual int
	err := rules.Int().WithBase(16).Apply(context.Background(), "BeEf", &actual)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if expected != actual {
		t.Errorf("Expected '%d' and got '%d'", expected, actual)
		return
	}

	err = rules.Int().WithBase(16).Apply(context.Background(), "XYZ", &actual)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

// TestIntRuleSet_Apply_CoerceFromFloatWithError tests:
// - Returns error when float value cannot be exactly represented as integer
func TestIntRuleSet_Apply_CoerceFromFloatWithError(t *testing.T) {
	var out int
	err := rules.Int().Apply(context.Background(), 1.000001, &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

// TestIntRuleSet_WithRequired tests:
// - WithRequired is correctly implemented
func TestIntRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[int](t, rules.Int())
}

// TestIntRuleSet_WithRuleFunc tests:
// - Custom rule functions are executed
// - Custom rules can return errors
// - Rule evaluation is called correctly
func TestIntRuleSet_WithRuleFunc(t *testing.T) {
	var out int
	err := rules.Int().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[int](1).Function()).
		Apply(context.Background(), "123", &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}

	rule := testhelpers.NewMockRule[int]()
	err = rules.Int().
		WithRuleFunc(rule.Function()).
		Apply(context.Background(), 123, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

// TestIntRuleSet_Any tests:
// - Any returns a RuleSet[any] implementation
func TestIntRuleSet_Any(t *testing.T) {
	ruleSet := rules.Int().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	}
}

// TestIntRuleSet_String_WithRequired tests:
// - Serializes to WithRequired()
func TestIntRuleSet_String_WithRequired(t *testing.T) {
	ruleSet := rules.Int().WithRequired()

	expected := "IntRuleSet[int].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestIntRuleSet_String_WithStrict tests:
// - Serializes to WithStrict()
func TestIntRuleSet_String_WithStrict(t *testing.T) {
	ruleSet := rules.Int().WithStrict()

	expected := "IntRuleSet[int].WithStrict()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestIntRuleSet_String_WithBase tests:
// - Serializes to WithBase(16)
func TestIntRuleSet_String_WithBase(t *testing.T) {
	ruleSet := rules.Int().WithBase(16)

	expected := "IntRuleSet[int].WithBase(16)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestIntRuleSet_String_WithRounding tests:
// - Serializes to WithRounding(...)
func TestIntRuleSet_String_WithRounding(t *testing.T) {
	ruleSet := rules.Int().WithRounding(rules.RoundingHalfEven)

	expected := "IntRuleSet[int].WithRounding(HalfEven)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestIntRuleSet_Evaluate tests:
// - Evaluate behaves like Apply.
func TestIntRuleSet_Evaluate(t *testing.T) {
	ruleSet := rules.Int().WithMin(5)
	testhelpers.MustEvaluate[int](t, ruleSet, 10)
	testhelpers.MustNotEvaluate[int](t, ruleSet, 1, errors.CodeMin)
}

// TestIntRuleSet_Apply_VariantTypes tests:
// - Applies correctly to various integer types
func TestIntRuleSet_Apply_VariantTypes(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  rules.RuleSet[any]
		input    interface{}
		expected interface{}
	}{
		{"Int", rules.Int().Any(), int(42), int(42)},
		{"Uint", rules.Uint().Any(), uint(42), uint(42)},
		{"Int8", rules.Int8().Any(), int8(42), int8(42)},
		{"Uint8", rules.Uint8().Any(), uint8(42), uint8(42)},
		{"Int16", rules.Int16().Any(), int16(42), int16(42)},
		{"Uint16", rules.Uint16().Any(), uint16(42), uint16(42)},
		{"Int32", rules.Int32().Any(), int32(42), int32(42)},
		{"Uint32", rules.Uint32().Any(), uint32(42), uint32(42)},
		{"Int64", rules.Int64().Any(), int64(42), int64(42)},
		{"Uint64", rules.Uint64().Any(), uint64(42), uint64(42)},
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

// TestIntRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestIntRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[int](t, rules.Int())
}

// TestIntRuleSet_Apply_StringOutput tests:
// - Outputs string values when output is a string type
// - Uses the same base as input parsing
func TestIntRuleSet_Apply_StringOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.IntRuleSet[int]
		input    interface{}
		expected string
	}{
		{"Base10", rules.Int(), 123, "123"},
		{"Base16", rules.Int().WithBase(16), 0xBEEF, "beef"},
		{"Base16Hex", rules.Int().WithBase(16), 0xFF, "ff"},
		{"Base8", rules.Int().WithBase(8), 0777, "777"},
		{"Base2", rules.Int().WithBase(2), 0b1010, "1010"},
		{"Negative", rules.Int(), -42, "-42"},
		{"Zero", rules.Int(), 0, "0"},
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

// TestIntRuleSet_Apply_PointerToStringOutput tests:
// - Outputs string values when output is a pointer to string type
// - Handles nil pointer by creating a new string
func TestIntRuleSet_Apply_PointerToStringOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.IntRuleSet[int]
		input    interface{}
		expected string
	}{
		{"Base10", rules.Int(), 123, "123"},
		{"Base16", rules.Int().WithBase(16), 0xBEEF, "beef"},
		{"Negative", rules.Int(), -42, "-42"},
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

// TestIntRuleSet_Apply_StringOutput_VariousTypes tests:
// - String output works with various integer types
func TestIntRuleSet_Apply_StringOutput_VariousTypes(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  rules.RuleSet[any]
		input    interface{}
		expected string
	}{
		{"Int", rules.Int().Any(), int(42), "42"},
		{"Uint", rules.Uint().Any(), uint(42), "42"},
		{"Int8", rules.Int8().Any(), int8(42), "42"},
		{"Uint8", rules.Uint8().Any(), uint8(42), "42"},
		{"Int16", rules.Int16().Any(), int16(42), "42"},
		{"Uint16", rules.Uint16().Any(), uint16(42), "42"},
		{"Int32", rules.Int32().Any(), int32(42), "42"},
		{"Uint32", rules.Uint32().Any(), uint32(42), "42"},
		{"Int64", rules.Int64().Any(), int64(42), "42"},
		{"Uint64", rules.Uint64().Any(), uint64(42), "42"},
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
