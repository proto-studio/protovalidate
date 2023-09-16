package rules_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWrapWrapAnyRuleSet(t *testing.T) {
	innerRuleSet := rules.Any()
	anyval, err := rules.WrapAny[any](innerRuleSet).Validate(123)

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

func TestWrapAnyRequired(t *testing.T) {
	innerRuleSet1 := rules.Any().WithRequired()
	ruleSet1 := rules.WrapAny[any](innerRuleSet1)

	if !ruleSet1.Required() {
		t.Error("Expected rule set to be required")
	}

	innerRuleSet2 := rules.Any()
	ruleSet2 := rules.WrapAny[any](innerRuleSet2)

	if ruleSet2.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet2 = ruleSet2.WithRequired()

	if !ruleSet2.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestWrapWrapAnyRuleSetInnerError(t *testing.T) {
	innerRuleSet := rules.Any().WithRuleFunc(testhelpers.MockCustomRule[any]("123", 1))

	_, err := rules.WrapAny[any](innerRuleSet).Validate(123)

	if err == nil {
		t.Error("Expected errors to not be empty")
	}
}

func TestWrapAnyCustom(t *testing.T) {
	innerRuleSet := rules.Any()

	_, err := rules.WrapAny[any](innerRuleSet).
		WithRuleFunc(testhelpers.MockCustomRule[any]("123", 1)).
		Validate("123")

	if err.Size() == 0 {
		t.Error("Expected errors to not be empty")
		return
	}

	expected := "abc"

	actual, err := rules.WrapAny[any](innerRuleSet).
		WithRuleFunc(testhelpers.MockCustomRule[any](expected, 0)).
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

func TestWrapAnyReturnsIdentity(t *testing.T) {
	innerRuleSet := rules.Any()

	ruleSet1 := rules.WrapAny[any](innerRuleSet)
	ruleSet2 := ruleSet1.Any()

	if ruleSet1 != ruleSet2 {
		t.Error("Expected Any to be an identity function")
	}
}
