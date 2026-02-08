package testhelpers

import (
	"context"
	"reflect"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// MustImplementWithNil tests that a rule set properly handles nil values with and without WithNil.
// It checks that:
// - The rule set has a WithNil method
// - Without WithNil, nil input returns CodeNull error
// - With WithNil, nil input succeeds and sets output to nil
// - With both WithNil and WithRequired, nil input succeeds (WithNil takes precedence)
//
// The function uses a zero value of type T for the output pointer type.
func MustImplementWithNil[T any](t testing.TB, ruleSet rules.RuleSet[T]) {
	t.Helper()

	// Check if the rule set has a WithNil method using reflection
	ruleSetValue := reflect.ValueOf(ruleSet)
	withNilMethod := ruleSetValue.MethodByName("WithNil")
	if !withNilMethod.IsValid() {
		t.Error("Expected rule set to have a WithNil method")
		return
	}

	ctx := context.TODO()

	// Test without WithNil - should error with CodeNull
	var output *T
	err := ruleSet.Apply(ctx, nil, &output)
	if err == nil {
		t.Error("Expected error when nil is provided without WithNil")
	} else if err.Code() != errors.CodeNull {
		t.Errorf("Expected error code to be CodeNull, got: %s", err.Code())
	}

	// Test with WithNil - should not error
	// Initialize output2 to a non-nil value so we can verify it was set to nil
	var zeroVal T
	output2 := &zeroVal
	// Call WithNil using reflection
	withNilResult := withNilMethod.Call(nil)
	if len(withNilResult) != 1 {
		t.Errorf("Expected WithNil to return one value, got %d", len(withNilResult))
		return
	}
	ruleSetWithNil, ok := withNilResult[0].Interface().(rules.RuleSet[T])
	if !ok {
		t.Error("Expected WithNil to return a RuleSet[T]")
		return
	}
	err = ruleSetWithNil.Apply(ctx, nil, &output2)
	if err != nil {
		t.Errorf("Expected no error when nil is provided with WithNil, got: %s", err)
	}
	if output2 != nil {
		t.Error("Expected output to be nil")
	}

	// Test with both WithNil and WithRequired - should not error (WithNil takes precedence)
	// Check if the rule set has a WithRequired method using reflection
	withRequiredMethod := ruleSetValue.MethodByName("WithRequired")
	if withRequiredMethod.IsValid() {
		// Call WithRequired on the original rule set
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

		// Now call WithNil on the rule set that has WithRequired
		ruleSetWithRequiredValue := reflect.ValueOf(ruleSetWithRequired)
		withNilMethodOnRequired := ruleSetWithRequiredValue.MethodByName("WithNil")

		withNilOnRequiredResult := withNilMethodOnRequired.Call(nil)
		if len(withNilOnRequiredResult) != 1 {
			t.Errorf("Expected WithNil to return one value, got %d", len(withNilOnRequiredResult))
			return
		}
		ruleSetWithBoth, ok := withNilOnRequiredResult[0].Interface().(rules.RuleSet[T])
		if !ok {
			t.Error("Expected WithNil to return a RuleSet[T]")
			return
		}

		// Test that nil input succeeds when both WithNil and WithRequired are set
		var zeroVal3 T
		output3 := &zeroVal3
		err = ruleSetWithBoth.Apply(ctx, nil, &output3)
		if err != nil {
			t.Errorf("Expected no error when nil is provided with both WithNil and WithRequired, got: %s", err)
		}
		if output3 != nil {
			t.Error("Expected output to be nil when both WithNil and WithRequired are set")
		}
	}
}
