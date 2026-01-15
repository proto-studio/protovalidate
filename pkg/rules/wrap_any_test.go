package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestWrapWrapAnyRuleSet tests:
// - Does not error when default configured.
// - Returns the value with the correct type.
// - Implements the RuleSet interface.
func TestWrapWrapAnyRuleSet(t *testing.T) {
	innerRuleSet := rules.Any()

	// Prepare the output variable for Apply
	var anyval any

	// Use Apply instead of Validate
	err := rules.WrapAny[any](innerRuleSet).Apply(context.TODO(), 123, &anyval)

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if anyval != 123 {
		t.Errorf("Expected 123 to be returned. Got: %v", anyval)
		return
	}

	ok := testhelpers.CheckRuleSetInterface[any](rules.WrapAny[any](innerRuleSet))
	if !ok {
		t.Error("Expected rule set to be implemented")
	}
}

// TestWrapAnyRequired tests:
// - The required flag defaults to false.
// - WithRequired sets the required flag.
// - Require returns true only when the required flag is set.
func TestWrapAnyRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[any](t, rules.WrapAny[any](rules.Any()))

	// Test that wrapping a required rule set preserves the required flag
	innerRuleSet1 := rules.Any().WithRequired()
	ruleSet1 := rules.WrapAny[any](innerRuleSet1)

	if !ruleSet1.Required() {
		t.Error("Expected rule set to be required")
	}
}

// TestWrapWrapAnyRuleSetInnerError tests:
// - The inner rule set rules are called.
// - Errors in inner the rule set are passed to the wrapper.
func TestWrapWrapAnyRuleSetInnerError(t *testing.T) {
	innerRuleSet := rules.Any().WithRule(testhelpers.NewMockRuleWithErrors[any](1))

	ruleSet := rules.WrapAny[any](innerRuleSet)

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeUnknown)
}

// TestWrapAnyCustom tests:
// - Custom rules are executed.
// - Custom rules can return errors.
// - Mutated values from the custom rules are returned.
func TestWrapAnyCustom(t *testing.T) {
	innerRuleSet := rules.Any()

	ruleSet := rules.WrapAny[any](innerRuleSet).
		WithRule(testhelpers.NewMockRuleWithErrors[any](1))

	testhelpers.MustNotApply(t, ruleSet, "123", errors.CodeUnknown)

	var expected any = "abc"

	ruleSet = rules.WrapAny[any](innerRuleSet).
		WithRule(testhelpers.NewMockRule[any]())

	testhelpers.MustApply(t, ruleSet, expected)
}

// TestWrapAnyReturnsIdentity tests:
// - Implementations of RuleSet[any] should return themselves when calling the Any method.
func TestWrapAnyReturnsIdentity(t *testing.T) {
	innerRuleSet := rules.Any()

	ruleSet1 := rules.WrapAny[any](innerRuleSet)
	ruleSet2 := ruleSet1.Any()

	if ruleSet1 != ruleSet2 {
		t.Error("Expected Any to be an identity function")
	}
}

// TestWrapAnyRequiredString tests:
// - Serializes to WithRequired()
func TestWrapAnyRequiredString(t *testing.T) {
	innerRuleSet := rules.Any()
	ruleSet := rules.WrapAny[any](innerRuleSet).WithRequired()

	expected := "AnyRuleSet.Any().WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestWrapAnyRuleString tests:
// - Serializes to WithRule(...)
func TestWrapAnyRuleString(t *testing.T) {
	innerRuleSet := rules.Any()
	ruleSet := rules.WrapAny[any](innerRuleSet).WithRuleFunc(testhelpers.NewMockRuleWithErrors[any](1).Function())

	expected := "AnyRuleSet.Any().WithRuleFunc(<function>)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirement:
// - Evaluate calls Apply and not Evaluate on the wrapped RuleSets that do not implement Rule[any].
// - Evaluate calls Evaluate on the wrapped RuleSets that do implement Rule[any].
// - In both cases the custom rules gets called exactly once.
//
// This is the only exception to the policy that testhelper.MustEvaluate/MustNotEvaluate
// cannot be used with WrapAnyRuleSet.
func TestWrapAnyEvaluate(t *testing.T) {
	v := 123

	// Both these should call Evaluate on the underlying rule but not Apply

	innerRuleSet := testhelpers.NewMockRuleSet[int]()
	innerRuleSet.OutputValue = &v
	ruleSet := rules.WrapAny[int](innerRuleSet)
	testhelpers.MustEvaluate[any](t, ruleSet, 123)

	if a := innerRuleSet.ApplyCallCount(); a != 0 {
		t.Errorf("Expected ApplyCallCount to be 0, got: %d", a)
	} else if e := innerRuleSet.EvaluateCallCount(); e != 1 {
		t.Errorf("Expected EvaluateCallCount to be 1, got: %d", a)
	}

	innerRuleSetWithErrors := testhelpers.NewMockRuleSetWithErrors[int](1)
	innerRuleSetWithErrors.OutputValue = &v
	ruleSetWithErrors := rules.WrapAny[int](innerRuleSetWithErrors)
	testhelpers.MustNotEvaluate[any](t, ruleSetWithErrors, 123, errors.CodeUnknown)

	if a := innerRuleSetWithErrors.ApplyCallCount(); a != 0 {
		t.Errorf("Expected ApplyCallCount to be 0, got: %d", a)
	} else if e := innerRuleSetWithErrors.EvaluateCallCount(); e != 1 {
		t.Errorf("Expected EvaluateCallCount to be 1, got: %d", a)
	}

	// Both of these should call Apply since the input type cannot be cast to int

	innerRuleSet.Reset()
	testhelpers.MustEvaluate[any](t, ruleSet, "123")

	if e := innerRuleSet.EvaluateCallCount(); e != 0 {
		t.Errorf("Expected EvaluateCallCount to be 0, got: %d", e)
	} else if a := innerRuleSet.ApplyCallCount(); a != 1 {
		t.Errorf("Expected ApplyCallCount to be 1, got: %d", a)
	}

	innerRuleSetWithErrors.Reset()
	testhelpers.MustNotEvaluate[any](t, ruleSetWithErrors, "123", errors.CodeUnknown)

	if e := innerRuleSetWithErrors.EvaluateCallCount(); e != 0 {
		t.Errorf("Expected EvaluateCallCount to be 0, got: %d", e)
	} else if a := innerRuleSetWithErrors.ApplyCallCount(); a != 1 {
		t.Errorf("Expected ApplyCallCount to be 1, got: %d", a)
	}
}

// TestWrapAnyWithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestWrapAnyWithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[any](t, rules.WrapAny[any](rules.Any()))
}
