package rules_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestAnyRuleSet(t *testing.T) {
	anyval, err := rules.Any().Validate(123)

	ok := testhelpers.CheckRuleSetInterface[any](rules.Any())
	if !ok {
		t.Error("Expected rule set to be implemented")
	}

	if err != nil {
		t.Fatal("Expected errors to be empty")
		return
	}

	if anyval != 123 {
		t.Errorf("Expected 123 to be returned. Got: %v", anyval)
	}
}

func TestAnyForbidden(t *testing.T) {
	_, err := rules.Any().WithForbidden().Validate(123)

	if err.Size() == 0 {
		t.Error("Expected errors to not be empty")
	}
}

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

func TestAnyCustom(t *testing.T) {
	_, err := rules.Any().
		WithRuleFunc(testhelpers.MockCustomRule[any]("123", 1)).
		Validate("123")

	if err.Size() == 0 {
		t.Fatal("Expected errors to not be empty")
	}

	expected := "abc"

	actual, err := rules.Any().
		WithRuleFunc(testhelpers.MockCustomRule[any](expected, 0)).
		Validate("123")

	if err != nil {
		t.Fatal("Expected errors to be empty")
	}

	if expected != actual {
		t.Errorf("Expected '%s' to equal '%s'", actual, expected)
	}
}

func TestAnyReturnsIdentity(t *testing.T) {
	ruleSet1 := rules.Any()
	ruleSet2 := ruleSet1.Any()

	if ruleSet1 != ruleSet2 {
		t.Error("Expected Any to be an identity function")
	}
}
