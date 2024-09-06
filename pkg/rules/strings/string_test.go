package strings_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestStringRuleSet(t *testing.T) {
	// Prepare the output variable for Apply
	var str string

	// Use Apply instead of Validate
	err := strings.New().Apply(context.TODO(), "test", &str)

	if err != nil {
		t.Fatal("Expected errors to be empty")
	}

	if str != "test" {
		t.Fatal("Expected test string to be returned")
	}

	ok := testhelpers.CheckRuleSetInterface[string](strings.New())
	if !ok {
		t.Fatal("Expected rule set to be implemented")
	}

	testhelpers.MustApplyTypes[string](t, strings.New(), "abc")
}

// Requirements:
// - Should be usable as a rule
// - Must implement the Rule[string] interface
func TestRuleImplementation(t *testing.T) {
	ok := testhelpers.CheckRuleInterface[string](strings.New())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

func TestStringRuleSetTypeError(t *testing.T) {
	// Prepare the output variable for Apply
	var str string

	// Use Apply instead of Validate
	err := strings.New().WithStrict().Apply(context.TODO(), 123, &str)

	if err == nil || len(err) == 0 {
		t.Error("Expected errors to not be empty")
	}
}

func tryStringCoercion(t testing.TB, val interface{}, expected string) {
	ruleSet := strings.New()
	testhelpers.MustApplyMutation(t, ruleSet.Any(), val, expected)
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

func TestStringCoercionFromUnknown(t *testing.T) {
	val := new(struct {
		x int
	})

	testhelpers.MustNotApply(t, strings.New().Any(), &val, errors.CodeType)
}

// Requirements:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
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
	// Prepare the output variable for Apply
	var out string

	// Test with a rule that is expected to produce an error
	err := strings.New().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[string](1).Function()).
		Apply(context.TODO(), "123", &out)

	if err == nil {
		t.Error("Expected errors to not be empty")
		return
	}

	// Test with a rule that is not expected to produce an error
	rule := testhelpers.NewMockRule[string]()

	err = strings.New().
		WithRuleFunc(rule.Function()).
		Apply(context.TODO(), "123", &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	// Verify that the rule was called exactly once
	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
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

// Requirements:
// - Serializes to WithRequired()
func TestRequiredString(t *testing.T) {
	ruleSet := strings.New().WithRequired()

	expected := "StringRuleSet.WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithStrict()
func TestStrictString(t *testing.T) {
	ruleSet := strings.New().WithStrict()

	expected := "StringRuleSet.WithStrict()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
