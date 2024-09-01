package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
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

// Requirements:
// - The required flag defaults to false.
// - WithRequired sets the required flag.
// - Require returns true only when the required flag is set.
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

// Requirements:
// - The inner rule set rules are called.
// - Errors in inner the rule set are passed to the wrapper.
func TestWrapWrapAnyRuleSetInnerError(t *testing.T) {
	innerRuleSet := rules.Any().WithRule(testhelpers.NewMockRuleWithErrors[any](1))

	ruleSet := rules.WrapAny[any](innerRuleSet)

	testhelpers.MustNotRun(t, ruleSet, 123, errors.CodeUnknown)
}

// Requirements:
// - Custom rules are executed.
// - Custom rules can return errors.
// - Mutated values from the custom rules are returned.
func TestWrapAnyCustom(t *testing.T) {
	innerRuleSet := rules.Any()

	ruleSet := rules.WrapAny[any](innerRuleSet).
		WithRule(testhelpers.NewMockRuleWithErrors[any](1))

	testhelpers.MustNotRun(t, ruleSet, "123", errors.CodeUnknown)

	var expected any = "abc"

	ruleSet = rules.WrapAny[any](innerRuleSet).
		WithRule(testhelpers.NewMockRule[any]())

	testhelpers.MustRun(t, ruleSet, expected)
}

// Requirement:
// - Implementations of RuleSet[any] should return themselves when calling the Any method.
func TestWrapAnyReturnsIdentity(t *testing.T) {
	innerRuleSet := rules.Any()

	ruleSet1 := rules.WrapAny[any](innerRuleSet)
	ruleSet2 := ruleSet1.Any()

	if ruleSet1 != ruleSet2 {
		t.Error("Expected Any to be an identity function")
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestWrapAnyRequiredString(t *testing.T) {
	innerRuleSet := rules.Any()
	ruleSet := rules.WrapAny[any](innerRuleSet).WithRequired()

	expected := "AnyRuleSet.Any().WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithRule(...)
func TestWrapAnyRuleString(t *testing.T) {
	innerRuleSet := rules.Any()
	ruleSet := rules.WrapAny[any](innerRuleSet).WithRuleFunc(testhelpers.NewMockRuleWithErrors[any](1).Function())

	expected := "AnyRuleSet.Any().WithRuleFunc(...)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
