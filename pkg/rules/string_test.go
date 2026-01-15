package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestStringRuleSet_Apply tests:
// - Implements the RuleSet interface
// - Correctly applies string validation
// - Returns the correct value
func TestStringRuleSet_Apply(t *testing.T) {
	// Prepare the output variable for Apply
	var str string

	// Use Apply instead of Validate
	err := rules.String().Apply(context.TODO(), "test", &str)

	if err != nil {
		t.Fatal("Expected errors to be empty")
	}

	if str != "test" {
		t.Fatal("Expected test string to be returned")
	}

	ok := testhelpers.CheckRuleSetInterface[string](rules.String())
	if !ok {
		t.Fatal("Expected rule set to be implemented")
	}

	testhelpers.MustApplyTypes(t, rules.String(), "abc")
}

// TestStringRuleSet_RuleInterface tests:
// - Should be usable as a rule
// - Must implement the Rule[string] interface
func TestStringRuleSet_RuleInterface(t *testing.T) {
	ok := testhelpers.CheckRuleInterface[string](rules.String())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

// TestStringRuleSet_Apply_TypeError tests:
// - Returns error when strict mode is enabled and input is not a string
func TestStringRuleSet_Apply_TypeError(t *testing.T) {
	// Prepare the output variable for Apply
	var str string

	// Use Apply instead of Validate
	err := rules.String().WithStrict().Apply(context.TODO(), 123, &str)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
	}
}

func tryStringCoercion(t testing.TB, val interface{}, expected string) {
	ruleSet := rules.String()
	testhelpers.MustApplyMutation(t, ruleSet.Any(), val, expected)
}

// TestStringRuleSet_Apply_CoerceFromInt tests:
// - Coerces integer values to strings
func TestStringRuleSet_Apply_CoerceFromInt(t *testing.T) {
	tryStringCoercion(t, 123, "123")
}

// TestStringRuleSet_Apply_CoerceFromIntPointer tests:
// - Coerces integer pointer values to strings
func TestStringRuleSet_Apply_CoerceFromIntPointer(t *testing.T) {
	x := 123
	tryStringCoercion(t, &x, "123")
}

// TestStringRuleSet_Apply_CoerceFromFloat tests:
// - Coerces float values to strings
func TestStringRuleSet_Apply_CoerceFromFloat(t *testing.T) {
	tryStringCoercion(t, 123.123, "123.123")
}

// TestStringRuleSet_Apply_CoerceFromFloatPointer tests:
// - Coerces float pointer values to strings
func TestStringRuleSet_Apply_CoerceFromFloatPointer(t *testing.T) {
	x := 123.123
	tryStringCoercion(t, &x, "123.123")
}

// TestStringRuleSet_Apply_CoerceFromInt64 tests:
// - Coerces int64 values to strings
func TestStringRuleSet_Apply_CoerceFromInt64(t *testing.T) {
	tryStringCoercion(t, int64(123), "123")
}

// TestStringRuleSet_Apply_CoerceFromInt64Pointer tests:
// - Coerces int64 pointer values to strings
func TestStringRuleSet_Apply_CoerceFromInt64Pointer(t *testing.T) {
	var x int64 = 123
	tryStringCoercion(t, &x, "123")
}

// TestStringRuleSet_Apply_CoerceFromStringPointer tests:
// - Coerces string pointer values to strings
func TestStringRuleSet_Apply_CoerceFromStringPointer(t *testing.T) {
	s := "hello"
	tryStringCoercion(t, &s, s)
}

// TestStringRuleSet_Apply_CoerceFromUnknown tests:
// - Returns error for unknown types that cannot be coerced
func TestStringRuleSet_Apply_CoerceFromUnknown(t *testing.T) {
	val := new(struct {
		x int
	})

	testhelpers.MustNotApply(t, rules.String().Any(), &val, errors.CodeType)
}

// TestStringRuleSet_WithRequired tests:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
func TestStringRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired(t, rules.String())
}

// TestStringRuleSet_WithRuleFunc tests:
// - Custom rule functions are executed
// - Custom rules can return errors
// - Rule evaluation is called correctly
func TestStringRuleSet_WithRuleFunc(t *testing.T) {
	// Prepare the output variable for Apply
	var out string

	// Test with a rule that is expected to produce an error
	err := rules.String().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[string](1).Function()).
		Apply(context.TODO(), "123", &out)

	if err == nil {
		t.Error("Expected errors to not be empty")
		return
	}

	// Test with a rule that is not expected to produce an error
	rule := testhelpers.NewMockRule[string]()

	err = rules.String().
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

// TestStringRuleSet_Any tests:
// - Any returns a RuleSet[any] implementation
func TestStringRuleSet_Any(t *testing.T) {
	ruleSet := rules.String().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	}
}

// TestStringRuleSet_String_WithRequired tests:
// - Serializes to WithRequired()
func TestStringRuleSet_String(t *testing.T) {
	tests := []struct {
		name     string
		ruleSet  *rules.StringRuleSet
		expected string
	}{
		{"Base", rules.String(), "StringRuleSet"},
		{"WithRequired", rules.String().WithRequired(), "StringRuleSet.WithRequired()"},
		{"WithStrict", rules.String().WithStrict(), "StringRuleSet.WithStrict()"},
		{"WithNil", rules.String().WithNil(), "StringRuleSet.WithNil()"},
		{"Chained", rules.String().WithRequired().WithStrict(), "StringRuleSet.WithRequired().WithStrict()"},
		{"ChainedWithNil", rules.String().WithRequired().WithNil(), "StringRuleSet.WithRequired().WithNil()"},
		{"ChainedAll", rules.String().WithRequired().WithStrict().WithNil(), "StringRuleSet.WithRequired().WithStrict().WithNil()"},
		{"ConflictResolution_Required", rules.String().WithRequired().WithRequired(), "StringRuleSet.WithRequired()"},
		{"ConflictResolution_Nil", rules.String().WithNil().WithNil(), "StringRuleSet.WithNil()"},
		{"ConflictResolution_Strict", rules.String().WithStrict().WithStrict(), "StringRuleSet.WithStrict()"},
		{"WithMin", rules.String().WithMin("abc"), "StringRuleSet.WithMin(\"abc\")"},
		{"WithMinLong", rules.String().WithMin("this is a very long string that should be truncated when displayed in the rule set string representation"), "StringRuleSet.WithMin(\"this is a very long string that should be truncate...\")"},
		{"WithMax", rules.String().WithMax("xyz"), "StringRuleSet.WithMax(\"xyz\")"},
		{"WithRegexp", rules.String().WithRegexpString("^[a-z]+$", "error"), "StringRuleSet.WithRegexp(\"^[a-z]+$\")"},
		{"WithRegexpLong", rules.String().WithRegexpString("^[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*$", "error"), "StringRuleSet.WithRegexp(\"^[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[A-Z]*[a-z]+[0-9]*[...\")"},
		{"ChainedWithRule", rules.String().WithRequired().WithMin("abc"), "StringRuleSet.WithRequired().WithMin(\"abc\")"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ruleSet.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestStringRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestStringRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil(t, rules.String())
}

// TestStringRuleSet_ErrorConfig tests:
// - All error customization methods work correctly
func TestStringRuleSet_ErrorConfig(t *testing.T) {
	testhelpers.MustImplementErrorConfig[string, *rules.StringRuleSet](t, rules.String())
}

// TestStringRuleSet_ErrorConfig_WithRule tests:
// - ErrorConfig is applied to errors from custom rules added via WithRule
func TestStringRuleSet_ErrorConfig_WithRule(t *testing.T) {
	ruleSet := rules.String().
		WithRule(&testhelpers.ErrorConfigTestRule[string]{}).
		WithDocsURI("https://example.com/rule-error")

	testhelpers.MustApplyErrorConfigWithCustomRule(t, ruleSet, "test", "https://example.com/rule-error")
}

// TestStringRuleSet_ErrorConfig_WithRuleFunc tests:
// - ErrorConfig is applied to errors from custom rules added via WithRuleFunc
func TestStringRuleSet_ErrorConfig_WithRuleFunc(t *testing.T) {
	ruleSet := rules.String().
		WithRuleFunc(testhelpers.ErrorConfigTestRuleFunc[string]()).
		WithErrorMeta("source", "rulefunc")

	testhelpers.MustApplyErrorConfigWithMetaOnInput(t, ruleSet, "test", "source", "rulefunc")
}

// TestStringRuleSet_ErrorConfig_CoercionError tests:
// - ErrorConfig is applied to coercion errors
func TestStringRuleSet_ErrorConfig_CoercionError(t *testing.T) {
	var out string
	ruleSet := rules.String().
		WithStrict(). // Strict mode disables coercion
		WithErrorMessage("type error", "expected a string")

	errs := ruleSet.Apply(context.Background(), 123, &out)

	if len(errs) == 0 {
		t.Fatal("Expected coercion error")
	}

	if errs[0].ShortError() != "type error" {
		t.Errorf("Expected short error 'type error', got: %s", errs[0].ShortError())
	}
}

// TestStringRuleSet_ErrorConfig_WithMinLen tests:
// - ErrorConfig is applied to errors from built-in WithMinLen rule
func TestStringRuleSet_ErrorConfig_WithMinLen(t *testing.T) {
	var out string
	ruleSet := rules.String().
		WithMinLen(5).
		WithErrorMessage("custom short", "custom long")

	errs := ruleSet.Apply(context.Background(), "ab", &out)

	if len(errs) == 0 {
		t.Fatal("Expected validation error")
	}

	if errs[0].ShortError() != "custom short" {
		t.Errorf("Expected short error 'custom short', got: %s", errs[0].ShortError())
	}
}
