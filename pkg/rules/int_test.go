package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestIntRuleSet(t *testing.T) {
	var intval int
	err := rules.NewInt().Apply(context.Background(), 123, &intval)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if intval != 123 {
		t.Errorf("Expected int 123 to be returned. Got: %d", intval)
		return
	}

	ok := testhelpers.CheckRuleSetInterface[int](rules.NewInt())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	testhelpers.MustApplyTypes[int](t, rules.NewInt(), 123)
}

func TestIntStrictError(t *testing.T) {
	var out int
	err := rules.NewInt().WithStrict().Apply(context.Background(), "123", &out)

	if err == nil || len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func tryIntCoercion(t *testing.T, val interface{}, expected int) {
	var actual int
	err := rules.NewInt().Apply(context.Background(), val, &actual)

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
	err := rules.NewInt().WithBase(16).Apply(context.Background(), "BeEf", &actual)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if expected != actual {
		t.Errorf("Expected '%d' and got '%d'", expected, actual)
		return
	}

	err = rules.NewInt().WithBase(16).Apply(context.Background(), "XYZ", &actual)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func TestIntCoercionFromFloatWithError(t *testing.T) {
	var out int
	err := rules.NewInt().Apply(context.Background(), 1.000001, &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func TestIntRequired(t *testing.T) {
	ruleSet := rules.NewInt()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestIntCustom(t *testing.T) {
	var out int
	err := rules.NewInt().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[int](1).Function()).
		Apply(context.Background(), "123", &out)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}

	rule := testhelpers.NewMockRule[int]()
	err = rules.NewInt().
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
	ruleSet := rules.NewInt().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestIntRequiredString(t *testing.T) {
	ruleSet := rules.NewInt().WithRequired()

	expected := "IntRuleSet[int].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithStrict()
func TestIntStrictString(t *testing.T) {
	ruleSet := rules.NewInt().WithStrict()

	expected := "IntRuleSet[int].WithStrict()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithBase(16)
func TestIntBaseString(t *testing.T) {
	ruleSet := rules.NewInt().WithBase(16)

	expected := "IntRuleSet[int].WithBase(16)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithRounding(...)
func TestIntRoundingString(t *testing.T) {
	ruleSet := rules.NewInt().WithRounding(rules.RoundingHalfEven)

	expected := "IntRuleSet[int].WithRounding(HalfEven)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Evaluate behaves like Apply.
func TestInt_Evaluate(t *testing.T) {
	ruleSet := rules.NewInt().WithMin(5)
	testhelpers.MustEvaluate[int](t, ruleSet, 10)
	testhelpers.MustNotEvaluate[int](t, ruleSet, 1, errors.CodeMin)
}
