package rules_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

type MyTestInterface interface {
	Test() int
}

type MyTestImpl struct{}

func (x MyTestImpl) Test() int { return 1 }

type MyTestImplInt int

func (x MyTestImplInt) Test() int { return int(x) }

type MyTestImplStr string

func (x MyTestImplStr) Test() int { return len(x) }

// Requirements:
// - Implements the RuleSet interface.
// - Does not error when default configured.
// - Returns the value with the correct type.
// - Errors if input cannot be implicitly cast to the interface.
func TestInterfaceRuleSet(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]()

	ok := testhelpers.CheckRuleSetInterface[MyTestInterface](ruleSet)
	if !ok {
		t.Error("Expected rule set to be implemented")
	}

	testhelpers.MustRun(t, ruleSet.Any(), MyTestImpl{})
	testhelpers.MustNotRun(t, ruleSet.Any(), 123, errors.CodeType)
}

// Requirements:
// - Required defaults to false.
// - Calling WithRequired sets the required flag.
func TestInterfaceRequired(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]()

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
func TestInterfaceCustom(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[MyTestInterface](1).Function())

	testhelpers.MustNotRun(t, ruleSet.Any(), MyTestImpl{}, errors.CodeUnknown)

	rule := testhelpers.NewMockRule[MyTestInterface]()

	ruleSet = rules.Interface[MyTestInterface]().
		WithRuleFunc(rule.Function())

	testhelpers.MustRun(t, ruleSet.Any(), MyTestImpl{})

	if c := rule.CallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestInterfaceRequiredString(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]().WithRequired()

	expected := "InterfaceRuleSet[MyTestInterface].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithRule(...)
func TestInterfaceRuleString(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]().
		WithRuleFunc(testhelpers.NewMockRuleWithErrors[MyTestInterface](1).Function())

	expected := "InterfaceRuleSet[MyTestInterface].WithRuleFunc(...)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirement:
// - RuleSets are usable as Rules for the same type
func TestInterfaceComposition(t *testing.T) {
	innerRuleSet := rules.Interface[MyTestInterface]().
		WithRule(testhelpers.NewMockRuleWithErrors[MyTestInterface](1))

	ruleSet := rules.Interface[MyTestInterface]().WithRule(innerRuleSet)

	testhelpers.MustNotRun(t, ruleSet.Any(), MyTestImpl{}, errors.CodeUnknown)
}

// Requirement:
// - Cast functions work.
// - Cast functions can stack.
func TestInterfaceWithCast(t *testing.T) {
	ruleSet := rules.Interface[MyTestInterface]()

	testhelpers.MustNotRun(t, ruleSet.Any(), 123, errors.CodeType)

	ruleSet = ruleSet.WithCast(func(v any) (MyTestInterface, bool) {
		if intval, ok := v.(int); ok {
			return MyTestImplInt(intval), true
		}
		return nil, false
	})

	testhelpers.MustRunMutation(t, ruleSet.Any(), 123, MyTestImplInt(123))
	testhelpers.MustNotRun(t, ruleSet.Any(), "abc", errors.CodeType)

	ruleSet = ruleSet.WithCast(func(v any) (MyTestInterface, bool) {
		if strval, ok := v.(string); ok {
			return MyTestImplStr(strval), true
		}
		return nil, false
	})

	testhelpers.MustRunMutation(t, ruleSet.Any(), 123, MyTestImplInt(123))
	testhelpers.MustRunMutation(t, ruleSet.Any(), "abc", MyTestImplStr("abc"))
}
