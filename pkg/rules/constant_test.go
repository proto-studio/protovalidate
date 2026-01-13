package rules_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestConstantRuleSet_Apply tests:
// - Implements the RuleSet interface.
// - Errors when the constant does not match.
// - Returns the value with the correct type.
func TestConstantRuleSet_Apply(t *testing.T) {
	ruleSet := rules.Constant[string]("abc")

	ok := testhelpers.CheckRuleSetInterface[string](ruleSet)
	if !ok {
		t.Error("Expected rule set to be implemented")
	}

	testhelpers.MustApply(t, ruleSet.Any(), "abc")
	testhelpers.MustNotApply(t, ruleSet.Any(), "x", errors.CodePattern)

	testhelpers.MustApplyTypes[string](t, ruleSet, "abc")
}

// TestConstantRuleSet_Apply_Coerce tests:
// - Returns a coercion error if the type does not match.
func TestConstantRuleSet_Apply_Coerce(t *testing.T) {
	ruleSet := rules.Constant[string]("abc")
	testhelpers.MustNotApply(t, ruleSet.Any(), 123, errors.CodeType)
}

// TestConstantRuleSet_WithRequired tests:
// - Required defaults to false.
// - Calling WithRequired sets the required flag.
// - Value is carried over.
// - Returns identity if called more than once.
func TestConstantRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[string](t, rules.Constant("abc"))

	// Test value is carried over and idempotency
	ruleSet := rules.Constant("abc").WithRequired()
	testhelpers.MustApply(t, ruleSet.Any(), "abc")
	testhelpers.MustNotApply(t, ruleSet.Any(), "x", errors.CodePattern)

	if ruleSet != ruleSet.WithRequired() {
		t.Error("Expected the same rule set to be returned")
	}
}

// TestConstantRuleSet_String_WithRequired tests:
// - Serializes to WithRequired()
func TestConstantRuleSet_String_WithRequired(t *testing.T) {
	ruleSet := rules.Constant("x").WithRequired()

	expected := `ConstantRuleSet(x).WithRequired()`
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestConstantRuleSet_Conflict tests:
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

// TestConstantRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestConstantRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[string](t, rules.Constant[string]("abc"))
}

// TestConstantRuleSet_ErrorConfig tests:
// - ConstantRuleSet implements error configuration methods
func TestConstantRuleSet_ErrorConfig(t *testing.T) {
	testhelpers.MustImplementErrorConfig[string, *rules.ConstantRuleSet[string]](t, rules.Constant[string]("test"))
}
