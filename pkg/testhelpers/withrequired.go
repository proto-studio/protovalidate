package testhelpers

import (
	"reflect"
	"testing"

	"proto.zip/studio/validate/pkg/rules"
)

// MustImplementWithRequired tests that a rule set properly implements WithRequired and Required methods.
// It checks that:
// - The rule set has a WithRequired method
// - The rule set has a Required method
// - Required() returns false by default
// - WithRequired() returns a new rule set with Required() returning true
// - The returned rule set from WithRequired() implements RuleSet[T]
//
// The function uses reflection to check for method presence, making it package-agnostic.
func MustImplementWithRequired[T any](t testing.TB, ruleSet rules.RuleSet[T]) {
	t.Helper()

	// Check if the rule set has a WithRequired method using reflection
	ruleSetValue := reflect.ValueOf(ruleSet)
	requiredMethod := ruleSetValue.MethodByName("Required")
	// Note: requiredMethod will always be valid because Required() is part of RuleSet interface

	// Check if the rule set has a WithRequired method using reflection
	withRequiredMethod := ruleSetValue.MethodByName("WithRequired")
	if !withRequiredMethod.IsValid() {
		t.Error("Expected rule set to have a WithRequired method")
		return
	}

	// Test that Required() returns false by default
	requiredResult := requiredMethod.Call(nil)
	// Note: len(requiredResult) will always be 1 because Required() is part of RuleSet interface
	required := requiredResult[0].Interface().(bool)
	if required {
		t.Error("Expected rule set to not be required by default")
	}

	// Test WithRequired() returns a new rule set
	withRequiredResult := withRequiredMethod.Call(nil)
	if len(withRequiredResult) != 1 {
		t.Errorf("Expected WithRequired to return one value, got %d", len(withRequiredResult))
		return
	}
	ruleSetWithRequired, ok := withRequiredResult[0].Interface().(rules.RuleSet[T])
	if !ok {
		t.Error("Expected WithRequired to return a RuleSet[T]")
		return
	}

	// Test that the returned rule set has Required() returning true
	ruleSetWithRequiredValue := reflect.ValueOf(ruleSetWithRequired)
	requiredMethodWithRequired := ruleSetWithRequiredValue.MethodByName("Required")
	// Note: requiredMethodWithRequired will always be valid because Required() is part of RuleSet interface
	requiredResultWithRequired := requiredMethodWithRequired.Call(nil)
	// Note: len(requiredResultWithRequired) will always be 1 because Required() is part of RuleSet interface
	requiredWithRequired := requiredResultWithRequired[0].Interface().(bool)
	if !requiredWithRequired {
		t.Error("Expected rule set returned from WithRequired to be required")
	}
}
