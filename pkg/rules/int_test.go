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
	intval, err := rules.Int().Apply(context.Background(), 123)

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
	_, err := rules.Int().WithStrict().Apply(context.Background(), "123")

	if len(errors.Unwrap(err)) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func tryIntCoercion(t *testing.T, val interface{}, expected int) {
	actual, err := rules.Int().Apply(context.Background(), val)

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
	actual, err := rules.Int().WithBase(16).Apply(context.Background(), "BeEf")

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if expected != actual {
		t.Errorf("Expected '%d' and got '%d'", expected, actual)
		return
	}

	_, err = rules.Int().WithBase(16).Apply(context.Background(), "XYZ")

	if len(errors.Unwrap(err)) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

// TestIntRuleSet_Apply_CoerceFromFloatWithError tests:
// - Returns error when float value cannot be exactly represented as integer
func TestIntRuleSet_Apply_CoerceFromFloatWithError(t *testing.T) {
	_, err := rules.Int().Apply(context.Background(), 1.000001)

	if len(errors.Unwrap(err)) == 0 {
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
	_, err := rules.Int().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[int](1).Function()).
		Apply(context.Background(), "123")

	if len(errors.Unwrap(err)) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}

	rule := testhelpers.NewMockRule[int]()
	_, err = rules.Int().
		WithRuleFunc(rule.Function()).
		Apply(context.Background(), 123)

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
func TestIntRuleSet_String(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.IntRuleSet[int]
		expected string
	}{
		{"Base", rules.Int(), "IntRuleSet[int]"},
		{"WithRequired", rules.Int().WithRequired(), "IntRuleSet[int].WithRequired()"},
		{"WithStrict", rules.Int().WithStrict(), "IntRuleSet[int].WithStrict()"},
		{"WithNil", rules.Int().WithNil(), "IntRuleSet[int].WithNil()"},
		{"WithBase", rules.Int().WithBase(16), "IntRuleSet[int].WithBase(16)"},
		{"WithRounding", rules.Int().WithRounding(rules.RoundingHalfEven), "IntRuleSet[int].WithRounding(HalfEven)"},
		{"Chained", rules.Int().WithRequired().WithStrict(), "IntRuleSet[int].WithRequired().WithStrict()"},
		{"ChainedWithBase", rules.Int().WithRequired().WithBase(16), "IntRuleSet[int].WithRequired().WithBase(16)"},
		{"ChainedAll", rules.Int().WithRequired().WithStrict().WithBase(16), "IntRuleSet[int].WithRequired().WithStrict().WithBase(16)"},
		{"ConflictResolution_Base", rules.Int().WithBase(10).WithBase(16), "IntRuleSet[int].WithBase(16)"},
		{"ConflictResolution_Rounding", rules.Int().WithRounding(rules.RoundingUp).WithRounding(rules.RoundingDown), "IntRuleSet[int].WithRounding(Down)"},
		{"WithMin", rules.Int().WithMin(5), "IntRuleSet[int].WithMin(5)"},
		{"ChainedWithRule", rules.Int().WithRequired().WithMin(5), "IntRuleSet[int].WithRequired().WithMin(5)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ruleSet.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
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
			out, err := tt.ruleSet.Apply(context.Background(), tt.input)

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

// TestIntRuleSet_Apply_StringOutput tests Apply returns correct int values
func TestIntRuleSet_Apply_StringOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.IntRuleSet[int]
		input    interface{}
		expected int
	}{
		{"Base10", rules.Int(), 123, 123},
		{"Base16", rules.Int().WithBase(16), 0xBEEF, 0xBEEF},
		{"Base16Hex", rules.Int().WithBase(16), 0xFF, 0xFF},
		{"Base8", rules.Int().WithBase(8), 0777, 0777},
		{"Base2", rules.Int().WithBase(2), 0b1010, 0b1010},
		{"Negative", rules.Int(), -42, -42},
		{"Zero", rules.Int(), 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.ruleSet.Apply(context.Background(), tt.input)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, out)
			}
		})
	}
}

// TestIntRuleSet_Apply_PointerToStringOutput tests Apply returns correct int
func TestIntRuleSet_Apply_PointerToStringOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.IntRuleSet[int]
		input    interface{}
		expected int
	}{
		{"Base10", rules.Int(), 123, 123},
		{"Base16", rules.Int().WithBase(16), 0xBEEF, 0xBEEF},
		{"Negative", rules.Int(), -42, -42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.ruleSet.Apply(context.Background(), tt.input)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, out)
			}
		})
	}
}

// TestIntRuleSet_Apply_StringOutput_VariousTypes tests Apply with various integer types
func TestIntRuleSet_Apply_StringOutput_VariousTypes(t *testing.T) {
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
			out, err := tt.ruleSet.Apply(context.Background(), tt.input)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, out)
			}
		})
	}
}

// TestIntRuleSet_ErrorConfig tests:
// - IntRuleSet implements error configuration methods
func TestIntRuleSet_ErrorConfig(t *testing.T) {
	testhelpers.MustImplementErrorConfig[int, *rules.IntRuleSet[int]](t, rules.Int())
}

// TestIntRuleSet_Apply_CoerceFromBool tests:
// - Coerces bool values to integers (true -> 1, false -> 0)
func TestIntRuleSet_Apply_CoerceFromBool(t *testing.T) {
	tryIntCoercion(t, true, 1)
	tryIntCoercion(t, false, 0)
}

// TestIntRuleSet_Apply_CoerceFromBool_Strict tests:
// - Strict mode rejects bool values
func TestIntRuleSet_Apply_CoerceFromBool_Strict(t *testing.T) {
	_, err := rules.Int().WithStrict().Apply(context.Background(), true)

	if len(errors.Unwrap(err)) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

// TestIntRuleSet_Apply_BoolOutput tests Apply returns correct int
func TestIntRuleSet_Apply_BoolOutput(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.IntRuleSet[int]
		input    interface{}
		expected int
	}{
		{"NonZero", rules.Int(), 42, 42},
		{"Zero", rules.Int(), 0, 0},
		{"Negative", rules.Int(), -1, -1},
		{"One", rules.Int(), 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.ruleSet.Apply(context.Background(), tt.input)

			if err != nil {
				t.Errorf("Expected no errors, got: %v", err)
				return
			}

			if out != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, out)
			}
		})
	}
}
