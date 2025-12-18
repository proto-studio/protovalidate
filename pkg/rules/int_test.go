package rules_test

import (
	"context"
	"reflect"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestIntRuleSet(t *testing.T) {
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

func TestIntStrictError(t *testing.T) {
	var out int
	err := rules.Int().WithStrict().Apply(context.Background(), "123", &out)

	if err == nil || len(err) == 0 {
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

func TestIntCoercionFromString(t *testing.T) {
	tryIntCoercion(t, "123", 123)
}

func TestIntCoercionFromFloat(t *testing.T) {
	tryIntCoercion(t, float32(123.0), 123)
}

func TestIntCoercionFromInt64(t *testing.T) {
	tryIntCoercion(t, float64(123.0), 123)
}

func TestIntCoercionFromHex(t *testing.T) {
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

func TestIntCoercionFromFloatWithError(t *testing.T) {
	var out int
	err := rules.Int().Apply(context.Background(), 1.000001, &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func TestIntRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[int](t, rules.Int())
}

func TestIntCustom(t *testing.T) {
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

func TestAnyInt(t *testing.T) {
	ruleSet := rules.Int().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestIntRequiredString(t *testing.T) {
	ruleSet := rules.Int().WithRequired()

	expected := "IntRuleSet[int].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithStrict()
func TestIntStrictString(t *testing.T) {
	ruleSet := rules.Int().WithStrict()

	expected := "IntRuleSet[int].WithStrict()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithBase(16)
func TestIntBaseString(t *testing.T) {
	ruleSet := rules.Int().WithBase(16)

	expected := "IntRuleSet[int].WithBase(16)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithRounding(...)
func TestIntRoundingString(t *testing.T) {
	ruleSet := rules.Int().WithRounding(rules.RoundingHalfEven)

	expected := "IntRuleSet[int].WithRounding(HalfEven)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Evaluate behaves like Apply.
func TestInt_Evaluate(t *testing.T) {
	ruleSet := rules.Int().WithMin(5)
	testhelpers.MustEvaluate[int](t, ruleSet, 10)
	testhelpers.MustNotEvaluate[int](t, ruleSet, 1, errors.CodeMin)
}

func TestIntVariantTypes(t *testing.T) {
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

// Requirements:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestIntWithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[int](t, rules.Int())
}
