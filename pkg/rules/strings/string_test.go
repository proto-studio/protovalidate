package strings_test

import (
	"testing"

	"proto.zip/studio/validate"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestStringRuleSet(t *testing.T) {
	str, err := strings.New().Validate("test")

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if str != "test" {
		t.Error("Expected test string to be returned")
		return
	}

	ok := testhelpers.CheckRuleSetInterface[string](strings.New())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

func TestStringRuleSetTypeError(t *testing.T) {
	_, err := strings.New().WithStrict().Validate(123)

	if err == nil || err.Size() == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func TestStringNilOnSuccess(t *testing.T) {
	ruleSet := validate.String().
		WithMinLen(3).
		WithMaxLen(7)
	_, err := ruleSet.Validate("abc")

	if err != nil {
		t.Errorf("Expected errors to be nil, got: %s", err)
		return
	}
}

func tryStringCoercion(t *testing.T, val interface{}, expected string) {
	actual, err := strings.New().Validate(val)

	if err != nil {
		t.Errorf("Expected errors to be empty and got '%s'", err)
		return
	}
	if expected != actual {
		t.Errorf("Expected '%s' and got '%s'", expected, actual)
		return
	}
}

func TestStringCoercionFromInt(t *testing.T) {
	tryStringCoercion(t, 123, "123")
}

func TestStringCoercionFromIntPointer(t *testing.T) {
	x := 123
	tryStringCoercion(t, &x, "123")
}

func TestStringCoercionFromFloat(t *testing.T) {
	tryStringCoercion(t, 123.123, "123.123")
}

func TestStringCoercionFromFloatPointer(t *testing.T) {
	x := 123.123
	tryStringCoercion(t, &x, "123.123")
}

func TestStringCoercionFromInt64(t *testing.T) {
	tryStringCoercion(t, int64(123), "123")
}

func TestStringCoercionFromInt64Pointer(t *testing.T) {
	var x int64 = 123
	tryStringCoercion(t, &x, "123")
}

func TestStringCoercionFromStringPointer(t *testing.T) {
	s := "hello"
	tryStringCoercion(t, &s, s)
}

func TestStringCoercionFromUknown(t *testing.T) {
	val := new(struct {
		x int
	})

	testhelpers.MustBeInvalid(t, strings.New().Any(), &val, errors.CodeType)
}

func TestStringRequired(t *testing.T) {
	ruleSet := strings.New()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestStringCustom(t *testing.T) {
	_, err := strings.New().
		WithRuleFunc(testhelpers.MockCustomRule("123", 1)).
		Validate("123")

	if err == nil {
		t.Error("Expected errors to not be empty")
		return
	}

	expected := "abc"

	actual, err := strings.New().
		WithRuleFunc(testhelpers.MockCustomRule(expected, 0)).
		Validate("123")

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if expected != actual {
		t.Errorf("Expected '%s' to equal '%s'", actual, expected)
		return
	}
}

func TestAny(t *testing.T) {
	ruleSet := strings.New().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}
