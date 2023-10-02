package numbers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestFloatRuleSet(t *testing.T) {
	floatval, err := numbers.NewFloat64().Validate(123.0)

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
}

func TestFloatStrictError(t *testing.T) {
	_, err := numbers.NewFloat64().WithStrict().Validate("123.0")

	if err == nil || len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func tryFloatCoercion(t *testing.T, val interface{}, expected float64) {
	actual, err := numbers.NewFloat64().Validate(val)

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
	_, err := numbers.NewFloat64().
		WithRuleFunc(testhelpers.MockCustomRule(123.0, 1)).
		Validate("123.0")

	if err == nil || len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}

	expected := 456.0

	actual, err := numbers.NewFloat64().
		WithRuleFunc(testhelpers.MockCustomRule(expected, 0)).
		Validate(123.0)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if expected != actual {
		t.Errorf("Expected '%f' to equal '%f'", actual, expected)
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
