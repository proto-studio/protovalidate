// Package testhelpers contains utility functions to make testing Rule and RuleSet implementations easier.
package testhelpers

import (
	"context"
	"fmt"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// neverAssignable if an interface with a private method making it impossible to assign to while.
// still satisfying interface{} type checks.
// Used by MustApplyTypes and not exported.
type neverAssignable interface{ priv() }

// neverAssignableI is an implementation of neverAssignable for use in MustApplyTypes
type neverAssignableImpl struct{ privProp int }

func (na *neverAssignableImpl) priv() {}

// CheckRuleSetInterface checks to see if the RuleSet interface is implemented for an interface and returns true if it is.
func CheckRuleSetInterface[T any](v any) bool {
	_, ok := v.(rules.RuleSet[T])
	return ok
}

// CheckRuleInterface checks to see if the Rule interface is implemented for an interface and returns true if it is.
func CheckRuleInterface[T any](v any) bool {
	_, ok := v.(rules.Rule[T])
	return ok
}

// checkEqual is a simple validity function that returns true if both values are equal.
func checkEqual(a, b any) error {
	if a != b {
		return fmt.Errorf("expected output to be %v, got: %v", a, b)
	}
	return nil
}

// checkAlways is a check function that always returns nil.
func checkAlways(_, _ any) error {
	return nil
}

// MustApplyFunc is a test helper that expects a RuleSet to return a nil error.
// If the error is non-nil or the check function returns an error, this function prints the error and returns it.
func MustApplyFunc(t testing.TB, ruleSet rules.RuleSet[any], input, expectedOutput any, fn func(a, b any) error) (any, error) {
	t.Helper()

	// Initialize the actual output variable
	var actualOutput any
	err := ruleSet.Apply(context.TODO(), input, &actualOutput)

	if err != nil {
		str := "Expected error to be nil"

		for _, inner := range err {
			str += fmt.Sprintf("\n  %s at %s", inner, inner.Path())
		}

		t.Errorf(str)
		return actualOutput, err
	} else if err := fn(expectedOutput, actualOutput); err != nil {
		t.Error(err)
		return actualOutput, err
	}

	return actualOutput, nil
}

// MustApply is a test helper that expects a RuleSet to return the input value and nil error.
// If the error is non-nil or the expected output does not match, this function prints the error and returns it.
func MustApply(t testing.TB, ruleSet rules.RuleSet[any], input any) (any, error) {
	t.Helper()
	return MustApplyFunc(t, ruleSet, input, input, checkEqual)
}

// MustApplyAny is a test helper that expects a RuleSet to finish without an error.
func MustApplyAny(t testing.TB, ruleSet rules.RuleSet[any], input any) (any, error) {
	t.Helper()
	return MustApplyFunc(t, ruleSet, input, input, checkAlways)
}

// MustApplyMutation is a test helper that expects a RuleSet to return a specific value and nil error.
// If the error is non-nil or the expected output does not match, this function prints the error and returns it.
func MustApplyMutation(t testing.TB, ruleSet rules.RuleSet[any], input, output any) (any, error) {
	t.Helper()
	return MustApplyFunc(t, ruleSet, input, output, checkEqual)
}

// MustNotApply is a test helper that expects a RuleSet to return an error and checks for a specific error code.
// If the error is nil or the code does not match, a testing error is printed and the function returns false.
//
// This function returns the error on "success" so that you can perform additional comparisons.
func MustNotApply(t testing.TB, ruleSet rules.RuleSet[any], input any, errorCode errors.ErrorCode) error {
	t.Helper()

	var output any
	err := ruleSet.Apply(context.TODO(), input, &output)

	if err == nil {
		t.Error("Expected error to not be nil")
		return nil
	} else if err.First().Code() != errorCode {
		t.Errorf("Expected error code of %s, got %s (%s)", errorCode, err.First().Code(), err)
		return nil
	}

	return err
}

// MustApplyTypes checks to make sure apply supports the various output types expected all rule sets.
// It is recommended all RuleSet implementations pass this assertion.
//
// Output types tested are:
// - Pointer to any.
// - Pointer to correct type.
// - Non-pointer (should error).
// - Pointer to nil (should error).
// - Nil (should error).
//
// Be sure to use an input that should not error if the types are correct.
// Note that Apply may implement output types other than these but these are bare minimum for any public RuleSet.
func MustApplyTypes[T any](t testing.TB, ruleSet rules.RuleSet[T], input T) {
	t.Helper()

	// Do not use MustApply and MustNotApply as these require .Any() which may invalidate the test.

	// Pointer to any
	var outputAny any
	err := ruleSet.Apply(context.TODO(), input, &outputAny)
	if err != nil {
		t.Errorf("Expected error to be nil on `any` output, got: %s", err)
	}

	// Pointer to correct type
	var outputPtr *T = new(T)
	err = ruleSet.Apply(context.TODO(), input, outputPtr)
	if err != nil {
		t.Errorf("Expected error to be nil on `%T` output, got: %s", outputPtr, err)
	}

	// Non-pointer to correct type
	var outputNonPointer T
	err = ruleSet.Apply(context.TODO(), input, outputNonPointer)
	if err == nil {
		t.Errorf("Expected error to not be nil on `%T` output", outputNonPointer)
	} else if code := err.First().Code(); code != errors.CodeInternal {
		t.Errorf("Expected error code to be %s (errors.CodeInternal) on `%T` output, got: %s", errors.CodeInternal, outputNonPointer, code)
	}

	// Pointer to nil
	var outputPointerToNil *T
	err = ruleSet.Apply(context.TODO(), input, outputPointerToNil)
	if err == nil {
		t.Error("Expected error to not be nil on pointer to `nil` output")
	} else if code := err.First().Code(); code != errors.CodeInternal {
		t.Errorf("Expected error code to be %s (errors.CodeInternal) on pointer to `nil` output, got: %s", errors.CodeInternal, code)
	}

	// Incompatible type
	// We must assign a &neverAssignableImpl{} to avoid false errors because the pointer was nil
	var outputIncompatible neverAssignable = &neverAssignableImpl{privProp: 1}
	outputIncompatible.priv()
	err = ruleSet.Apply(context.TODO(), input, outputIncompatible)
	if err == nil {
		t.Error("Expected error to not be nil on incompatible output")
	} else if code := err.First().Code(); code != errors.CodeInternal {
		t.Errorf("Expected error code to be %s (errors.CodeInternal) on incompatible output, got: %s", errors.CodeInternal, code)
	}

	// Nil value
	err = ruleSet.Apply(context.TODO(), input, nil)
	if err == nil {
		t.Error("Expected error to not be nil on `nil` output")
	} else if code := err.First().Code(); code != errors.CodeInternal {
		t.Errorf("Expected error code to be %s (errors.CodeInternal) on `nil` output, got: %s", errors.CodeInternal, code)
	}
}
