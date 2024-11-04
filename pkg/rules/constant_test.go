package rules_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Implements the RuleSet interface.
// - Errors when the constant does not match.
// - Returns the value with the correct type.
func TestConstantRuleSet(t *testing.T) {
	ruleSet := rules.Constant[string]("abc")

	ok := testhelpers.CheckRuleSetInterface[string](ruleSet)
	if !ok {
		t.Error("Expected rule set to be implemented")
	}

	testhelpers.MustApply(t, ruleSet.Any(), "abc")
	testhelpers.MustNotApply(t, ruleSet.Any(), "x", errors.CodePattern)

	testhelpers.MustApplyTypes[string](t, ruleSet, "abc")
}

// Requirements:
// - Returns a coercion error if the type does not match.
func TestConstantCoerce(t *testing.T) {
	ruleSet := rules.Constant[string]("abc")
	testhelpers.MustNotApply(t, ruleSet.Any(), 123, errors.CodeType)
}

// Requirements:
// - Required defaults to false.
// - Calling WithRequired sets the required flag.
// - Value is carried over.
// - Returns identity if called more than once.
func TestConstantRequired(t *testing.T) {
	ruleSet := rules.Constant("abc")

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}

	testhelpers.MustApply(t, ruleSet.Any(), "abc")
	testhelpers.MustNotApply(t, ruleSet.Any(), "x", errors.CodePattern)

	if ruleSet != ruleSet.WithRequired() {
		t.Error("Expected the same rule set to be returned")
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestConstantRequiredString(t *testing.T) {
	ruleSet := rules.Constant("x").WithRequired()

	expected := `ConstantRuleSet(x).WithRequired()`
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Conflict always returns true for ConstantRuleSet.
func TestConstantRuleSet_Conflict(t *testing.T) {
	abc := rules.Constant("abc")
	xyz := rules.Constant("xyz")

	if !abc.Conflict(xyz) {
		t.Error("Expected Conflict to be true for abc -> xyz")
	}
	if !xyz.Conflict(abc) {
		t.Error("Expected Conflict to be true for xyz -> abc")
	}
}
