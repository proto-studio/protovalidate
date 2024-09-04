package rules_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Implements the RuleSet interface.
// - Does not error when default configured.
// - Returns the value with the correct type.
func TestAnyRuleSet(t *testing.T) {
	ruleSet := rules.Any()

	ok := testhelpers.CheckRuleSetInterface[any](ruleSet)
	if !ok {
		t.Error("Expected rule set to be implemented")
	}

	testhelpers.MustApply(t, ruleSet, 123)

	testhelpers.MustApplyTypes[any](t, ruleSet, 123)
}

// Requirements:
// - Sets the required flag when calling WithForbidden.
// - Returns error when forbidden.
func TestAnyForbidden(t *testing.T) {
	ruleSet := rules.Any().WithForbidden()

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeForbidden)
}

// Requirements:
// - Required defaults to false.
// - Calling WithRequired sets the required flag.
func TestAnyRequired(t *testing.T) {
	ruleSet := rules.Any()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

// Requirements:
// - Custom rules are executed.
// - Custom rules can return errors.
func TestAnyCustom(t *testing.T) {
	ruleSet := rules.Any().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[any](1).Function())

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeUnknown)

	rule := testhelpers.NewMockRule[any]()

	ruleSet = rules.Any().
		WithRuleFunc(rule.Function())

	testhelpers.MustApply(t, ruleSet, "123")

	if c := rule.CallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
	}
}

// Requirement:
// - Implementations of RuleSet[any] should return themselves when calling the Any method.
func TestAnyReturnsIdentity(t *testing.T) {
	ruleSet1 := rules.Any()
	ruleSet2 := ruleSet1.Any()

	if ruleSet1 != ruleSet2 {
		t.Error("Expected Any to be an identity function")
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestAnyRequiredString(t *testing.T) {
	ruleSet := rules.Any().WithRequired()

	expected := "AnyRuleSet.WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithForbidden()
func TestAnyForbiddenString(t *testing.T) {
	ruleSet := rules.Any().WithForbidden()

	expected := "AnyRuleSet.WithForbidden()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithRule(...)
func TestAnyRuleString(t *testing.T) {
	ruleSet := rules.Any().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[any](1).Function())

	expected := "AnyRuleSet.WithRuleFunc(...)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirement:
// - RuleSets are usable as Rules for the same type
func TestAnyComposition(t *testing.T) {
	innerRuleSet := rules.Any().
		WithRule(testhelpers.NewMockRuleWithErrors[any](1))

	ruleSet := rules.Any().WithRule(innerRuleSet)

	testhelpers.MustNotApply(t, ruleSet, 123, errors.CodeUnknown)
}
