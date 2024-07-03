package numbers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestIntRuleSet(t *testing.T) {
	intval, err := numbers.NewInt().Validate(123)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if intval != 123 {
		t.Errorf("Expected int 123 to be returned. Got: %d", intval)
		return
	}

	ok := testhelpers.CheckRuleSetInterface[int](numbers.NewInt())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

func TestIntStrictError(t *testing.T) {
	_, err := numbers.NewInt().WithStrict().Validate("123")

	if err == nil || len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func tryIntCoercion(t *testing.T, val interface{}, expected int) {
	actual, err := numbers.NewInt().Validate(val)

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
	actual, err := numbers.NewInt().WithBase(16).Validate("BeEf")

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if expected != actual {
		t.Errorf("Expected '%d' and got '%d'", expected, actual)
		return
	}

	_, err = numbers.NewInt().WithBase(16).Validate("XYZ")

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func TestIntCoercionFromFloatWithError(t *testing.T) {
	_, err := numbers.NewInt().Validate(1.000001)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func TestIntRequired(t *testing.T) {
	ruleSet := numbers.NewInt()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestIntCustom(t *testing.T) {
	_, err := numbers.NewInt().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[int](1).Function()).
		Validate("123")

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}

	rule := testhelpers.NewMockRule[int]()
	_, err = numbers.NewInt().
		WithRuleFunc(rule.Function()).
		Validate(123)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if c := rule.CallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

func TestAnyInt(t *testing.T) {
	ruleSet := numbers.NewInt().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestIntRequiredString(t *testing.T) {
	ruleSet := numbers.NewInt().WithRequired()

	expected := "IntRuleSet[int].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithStrict()
func TestIntStrictString(t *testing.T) {
	ruleSet := numbers.NewInt().WithStrict()

	expected := "IntRuleSet[int].WithStrict()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithBase(16)
func TestIntBaseString(t *testing.T) {
	ruleSet := numbers.NewInt().WithBase(16)

	expected := "IntRuleSet[int].WithBase(16)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithRounding(...)
func TestIntRoundingString(t *testing.T) {
	ruleSet := numbers.NewInt().WithRounding(numbers.RoundingHalfEven)

	expected := "IntRuleSet[int].WithRounding(HalfEven)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
