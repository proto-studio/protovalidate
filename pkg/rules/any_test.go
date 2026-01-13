package rules_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestAnyRuleSet_Apply tests:
// - Implements the RuleSet interface.
// - Does not error when default configured.
// - Returns the value with the correct type.
func TestAnyRuleSet_Apply(t *testing.T) {
	ruleSet := rules.Any()

	ok := testhelpers.CheckRuleSetInterface[any](ruleSet)
	if !ok {
		t.Error("Expected rule set to be implemented")
	}

	testhelpers.MustApply(t, ruleSet, 123)

	testhelpers.MustApplyTypes[any](t, ruleSet, 123)
}

// TestAnyRuleSet_WithForbidden tests:
// - Sets the required flag when calling WithForbidden.
// - Returns error when forbidden.
func TestAnyRuleSet_WithForbidden(t *testing.T) {
	ruleSet := rules.Any().WithForbidden()

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeForbidden)
}

// TestAnyRuleSet_WithRequired tests:
// - Required defaults to false.
// - Calling WithRequired sets the required flag.
func TestAnyRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[any](t, rules.Any())
}

// TestAnyRuleSet_WithRuleFunc tests:
// - Custom rules are executed.
// - Custom rules can return errors.
func TestAnyRuleSet_WithRuleFunc(t *testing.T) {
	ruleSet := rules.Any().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[any](1).Function())

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeUnknown)

	rule := testhelpers.NewMockRule[any]()

	ruleSet = rules.Any().
		WithRuleFunc(rule.Function())

	testhelpers.MustApply(t, ruleSet, "123")

	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
	}
}

// Requirement:
// - Implementations of RuleSet[any] should return themselves when calling the Any method.
func TestAnyRuleSet_Any_ReturnsIdentity(t *testing.T) {
	ruleSet1 := rules.Any()
	ruleSet2 := ruleSet1.Any()

	if ruleSet1 != ruleSet2 {
		t.Error("Expected Any to be an identity function")
	}
}

// TestAnyRuleSet_String_WithRequired tests:
// - Serializes to WithRequired()
func TestAnyRuleSet_String_WithRequired(t *testing.T) {
	ruleSet := rules.Any().WithRequired()

	expected := "AnyRuleSet.WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestAnyRuleSet_String_WithForbidden tests:
// - Serializes to WithForbidden()
func TestAnyRuleSet_String_WithForbidden(t *testing.T) {
	ruleSet := rules.Any().WithForbidden()

	expected := "AnyRuleSet.WithForbidden()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestAnyRuleSet_String_WithRuleFunc tests:
// - Serializes to WithRule(...)
func TestAnyRuleSet_String_WithRuleFunc(t *testing.T) {
	ruleSet := rules.Any().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[any](1).Function())

	expected := "AnyRuleSet.WithRuleFunc(...)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestAnyRuleSet_Composition tests:
// - RuleSets are usable as Rules for the same type
func TestAnyRuleSet_Composition(t *testing.T) {
	innerRuleSet := rules.Any().
		WithRule(testhelpers.NewMockRuleWithErrors[any](1))

	ruleSet := rules.Any().WithRule(innerRuleSet)

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeUnknown)
}

// TestAnyRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestAnyRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[any](t, rules.Any())
}

// TestAnyRuleSet_ErrorConfig tests:
// - AnyRuleSet implements error configuration methods
func TestAnyRuleSet_ErrorConfig(t *testing.T) {
	testhelpers.MustImplementErrorConfig[any, *rules.AnyRuleSet](t, rules.Any())
}
