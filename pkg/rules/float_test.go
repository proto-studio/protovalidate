package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

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

func TestFloatRuleSet_Apply_StrictError(t *testing.T) {
	var out float64
	err := rules.Float64().WithStrict().Apply(context.Background(), "123.0", &out)

	if err == nil || len(err) == 0 {
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

func TestFloatRuleSet_Apply_CoerceFromString(t *testing.T) {
	tryFloatCoercion(t, "123.0", 123.0)
}

func TestFloatRuleSet_Apply_CoerceFromFloat(t *testing.T) {
	tryFloatCoercion(t, float32(123.0), 123.0)
}

func TestFloatRuleSet_Apply_CoerceFromFloat64(t *testing.T) {
	tryFloatCoercion(t, float64(123.0), 123.0)
}

func TestFloatRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[float64](t, rules.Float64())
}

func TestFloatRuleSet_WithRuleFunc(t *testing.T) {
	var out float64
	err := rules.Float64().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[float64](1).Function()).
		Apply(context.Background(), "123.0", &out)

	if err == nil || len(err) == 0 {
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

func TestFloatRuleSet_Any(t *testing.T) {
	ruleSet := rules.Float64().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestFloatRuleSet_String_WithRequired(t *testing.T) {
	ruleSet := rules.Float64().WithRequired()

	expected := "FloatRuleSet[float64].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithStrict()
func TestFloatRuleSet_String_WithStrict(t *testing.T) {
	ruleSet := rules.Float64().WithStrict()

	expected := "FloatRuleSet[float64].WithStrict()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithRounding(...)
func TestFloatRuleSet_String_WithRounding(t *testing.T) {
	ruleSet := rules.Float64().WithRounding(rules.RoundingHalfEven, 5)

	expected := "FloatRuleSet[float64].WithRounding(HalfEven, 5)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Evaluate behaves like Apply.
func TestFloatRuleSet_Evaluate(t *testing.T) {
	ruleSet := rules.Float64().WithMin(5)
	testhelpers.MustEvaluate[float64](t, ruleSet, 10)
	testhelpers.MustNotEvaluate[float64](t, ruleSet, 1, errors.CodeMin)
}

// Requirements:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestFloatRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[float64](t, rules.Float64())
}
