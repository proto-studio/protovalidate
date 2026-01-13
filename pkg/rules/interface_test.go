package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// MyTestInterface is a test interface used for interface validation testing.
type MyTestInterface interface {
	Test() int
}

// MyTestImpl is a test implementation of MyTestInterface.
type MyTestImpl struct{}

func (x MyTestImpl) Test() int { return 1 }

// MyTestImplInt is a test implementation of MyTestInterface using int.
type MyTestImplInt int

func (x MyTestImplInt) Test() int { return int(x) }

// MyTestImplStr is a test implementation of MyTestInterface using string.
type MyTestImplStr string

func (x MyTestImplStr) Test() int { return len(x) }

// TestInterfaceRuleSet_Apply tests:
// - Implements the RuleSet interface.
// - Does not error when default configured.
// - Returns the value with the correct type.
// - Errors if input cannot be implicitly cast to the interface.
func TestInterfaceRuleSet_Apply(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]()

	ok := testhelpers.CheckRuleSetInterface[MyTestInterface](ruleSet)
	if !ok {
		t.Error("Expected rule set to be implemented")
	}

	testhelpers.MustApply(t, ruleSet.Any(), MyTestImpl{})
	testhelpers.MustNotApply(t, ruleSet.Any(), 123, errors.CodeType)

	testhelpers.MustApplyTypes[MyTestInterface](t, ruleSet, MyTestImpl{})
}

// TestInterfaceRuleSet_WithRequired tests:
// - Required defaults to false.
// - Calling WithRequired sets the required flag.
func TestInterfaceRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[MyTestInterface](t, rules.Interface[MyTestInterface]())

	// Test idempotency
	ruleSet := rules.Interface[MyTestInterface]().WithRequired()
	if ruleSet.WithRequired() != ruleSet {
		t.Error("Expected WithRequired to be idempotent")
	}
}

// TestInterfaceRuleSet_WithRuleFunc tests:
// - Custom rules are executed.
// - Custom rules can return errors.
func TestInterfaceRuleSet_WithRuleFunc(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[MyTestInterface](1).Function())

	testhelpers.MustNotApply(t, ruleSet.Any(), MyTestImpl{}, errors.CodeUnknown)

	rule := testhelpers.NewMockRule[MyTestInterface]()

	ruleSet = rules.Interface[MyTestInterface]().
		WithRuleFunc(rule.Function())

	testhelpers.MustApply(t, ruleSet.Any(), MyTestImpl{})

	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
	}
}

// TestInterfaceRuleSet_String_WithRequired tests:
// - Serializes to WithRequired()
func TestInterfaceRuleSet_String_WithRequired(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]().WithRequired()

	expected := "InterfaceRuleSet[MyTestInterface].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestInterfaceRuleSet_String_WithRuleFunc tests:
// - Serializes to WithRule(...)
func TestInterfaceRuleSet_String_WithRuleFunc(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[MyTestInterface](1).Function())

	expected := "InterfaceRuleSet[MyTestInterface].WithRuleFunc(...)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestInterfaceRuleSet_Composition tests:
// - RuleSets are usable as Rules for the same type
func TestInterfaceRuleSet_Composition(t *testing.T) {
	innerRuleSet := rules.Interface[MyTestInterface]().
		WithRule(testhelpers.NewMockRuleWithErrors[MyTestInterface](1))

	ruleSet := rules.Interface[MyTestInterface]().WithRule(innerRuleSet)

	testhelpers.MustNotApply(t, ruleSet.Any(), MyTestImpl{}, errors.CodeUnknown)
}

// Requirement:
// - Cast functions work.
// - Cast functions can stack.
func TestInterfaceRuleSet_WithCast(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]()

	testhelpers.MustNotApply(t, ruleSet.Any(), 123, errors.CodeType)

	ruleSet = ruleSet.WithCast(func(ctx context.Context, v any) (MyTestInterface, errors.ValidationErrorCollection) {
		if intval, ok := v.(int); ok {
			return MyTestImplInt(intval), nil
		}
		return nil, nil
	})

	testhelpers.MustApplyMutation(t, ruleSet.Any(), 123, MyTestImplInt(123))
	testhelpers.MustNotApply(t, ruleSet.Any(), "abc", errors.CodeType)

	ruleSetWithString := ruleSet.WithCast(func(ctx context.Context, v any) (MyTestInterface, errors.ValidationErrorCollection) {
		if strval, ok := v.(string); ok {
			return MyTestImplStr(strval), nil
		}
		return nil, nil
	})

	testhelpers.MustApplyMutation(t, ruleSetWithString.Any(), 123, MyTestImplInt(123))
	testhelpers.MustApplyMutation(t, ruleSetWithString.Any(), "abc", MyTestImplStr("abc"))

	// If a cast returns an error that error is returned
	ruleSetWithError := ruleSet.WithCast(func(ctx context.Context, v any) (MyTestInterface, errors.ValidationErrorCollection) {
		if _, ok := v.(string); ok {
			return nil, errors.Collection(
				errors.Errorf(errors.CodeUnexpected, ctx, "unexpected", "test"),
			)
		}
		return nil, nil
	})
	testhelpers.MustApplyMutation(t, ruleSetWithError.Any(), 123, MyTestImplInt(123))
	testhelpers.MustNotApply(t, ruleSetWithError.Any(), "abc", errors.CodeUnexpected)
}

// TestInterfaceRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestInterfaceRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[MyTestInterface](t, rules.Interface[MyTestInterface]())
}

// TestInterfaceRuleSet_ErrorConfig tests:
// - InterfaceRuleSet implements error configuration methods
func TestInterfaceRuleSet_ErrorConfig(t *testing.T) {
	testhelpers.MustImplementErrorConfig[MyTestInterface, *rules.InterfaceRuleSet[MyTestInterface]](t, rules.Interface[MyTestInterface]())
}
