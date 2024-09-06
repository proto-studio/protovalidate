package numbers_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestFloatRuleSet(t *testing.T) {
	var floatval float64
	err := numbers.NewFloat64().Apply(context.Background(), 123.0, &floatval)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if floatval != 123.0 {
		t.Errorf("Expected float 123.0.0 to be returned. Got: %f", floatval)
		return
	}

	ok := testhelpers.CheckRuleSetInterface[float64](numbers.NewFloat64())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	testhelpers.MustApplyTypes[float64](t, numbers.NewFloat64(), 123.0)
}

func TestFloatStrictError(t *testing.T) {
	var out float64
	err := numbers.NewFloat64().WithStrict().Apply(context.Background(), "123.0", &out)

	if err == nil || len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func tryFloatCoercion(t *testing.T, val interface{}, expected float64) {
	var actual float64
	err := numbers.NewFloat64().Apply(context.Background(), "123.0", &actual)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}
	if expected != actual {
		t.Errorf("Expected '%f' and got '%f'", expected, actual)
		return
	}
}

func TestFloatCoercionFromString(t *testing.T) {
	tryFloatCoercion(t, "123.0", 123.0)
}

func TestFloatCoercionFromFloat(t *testing.T) {
	tryFloatCoercion(t, float32(123.0), 123.0)
}

func TestFloatCoercionFromFloat64(t *testing.T) {
	tryFloatCoercion(t, float64(123.0), 123.0)
}

func TestFloatRequired(t *testing.T) {
	ruleSet := numbers.NewFloat64()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestFloatCustom(t *testing.T) {
	var out float64
	err := numbers.NewFloat64().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[float64](1).Function()).
		Apply(context.Background(), "123.0", &out)

	if err == nil || len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}

	rule := testhelpers.NewMockRule[float64]()

	err = numbers.NewFloat64().
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

func TestAnyFloat(t *testing.T) {
	ruleSet := numbers.NewFloat64().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestFloatRequiredString(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRequired()

	expected := "FloatRuleSet[float64].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithStrict()
func TestFloatStrictString(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithStrict()

	expected := "FloatRuleSet[float64].WithStrict()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithRounding(...)
func TestFloatRoundingString(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingHalfEven, 5)

	expected := "FloatRuleSet[float64].WithRounding(HalfEven, 5)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Evaluate behaves like Apply.
func TestFloat_Evaluate(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithMin(5)
	testhelpers.MustEvaluate[float64](t, ruleSet, 10)
	testhelpers.MustNotEvaluate[float64](t, ruleSet, 1, errors.CodeMin)
}
