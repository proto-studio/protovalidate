// Package testhelpers contains utility functions to make testing Rule and RuleSet implementations easier.
package testhelpers

import (
	"context"
	"fmt"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

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

// MustBeValidFunc is a test helper that expects a RuleSet to a nil error.
// If the error is non-nil or the check function returns an error, this function prints the error and returns it.
//
// Deprecated: MustBeValidFunc is deprecated and will be removed in v1.0.0. Use MustRunFunc instead.
func MustBeValidFunc(t testing.TB, ruleSet rules.RuleSet[any], input, expectedOutput any, fn func(a, b any) error) error {
	t.Helper()
	return MustRunFunc(t, ruleSet, input, expectedOutput, fn)
}

// MustBeValid is a test helper that expects a RuleSet to return a specific value and nil error.
// If the error is non-nil or the expected output does not match, this function prints the error and returns it.
//
// Deprecated: MustBeValid is deprecated and will be removed in v1.0.0. Use MustRun instead.
func MustBeValid(t testing.TB, ruleSet rules.RuleSet[any], input, expectedOutput any) error {
	t.Helper()
	return MustRunFunc(t, ruleSet, input, expectedOutput, checkEqual)
}

// MustRunFunc is a test helper that expects a RuleSet to a nil error.
// If the error is non-nil or the check function returns an error, this function prints the error and returns it.
func MustRunFunc(t testing.TB, ruleSet rules.RuleSet[any], input, expectedOutput any, fn func(a, b any) error) error {
	t.Helper()

	actualOutput, err := ruleSet.Run(context.TODO(), input)

	if err != nil {
		str := "Expected error to be nil"

		for _, inner := range err {
			str = "\n  " + fmt.Sprintf("%s at %s", inner, inner.Path())
		}

		t.Errorf(str)
		return err
	} else if err := fn(expectedOutput, actualOutput); err != nil {
		t.Error(err)
		return err
	}

	return nil
}

// MustRun is a test helper that expects a RuleSet to return the input value and nil error.
// If the error is non-nil or the expected output does not match, this function prints the error and returns it.
func MustRun(t testing.TB, ruleSet rules.RuleSet[any], input any) error {
	t.Helper()
	return MustRunFunc(t, ruleSet, input, input, checkEqual)
}

// MustRunAny is a test helper that expects a RuleSet to finish without an error.
func MustRunAny(t testing.TB, ruleSet rules.RuleSet[any], input any) error {
	t.Helper()
	return MustRunFunc(t, ruleSet, input, input, checkAlways)
}

// MustRunMutation is a test helper that expects a RuleSet to return a specific value and nil error.
// If the error is non-nil or the expected output does not match, this function prints the error and returns it.
func MustRunMutation(t testing.TB, ruleSet rules.RuleSet[any], input, output any) error {
	t.Helper()
	return MustRunFunc(t, ruleSet, input, output, checkEqual)
}

// MustBeValidAny is a test helper that expects a RuleSet to finish without an error.
// It does not check the return value.
func MustBeValidAny(t testing.TB, ruleSet rules.RuleSet[any], input any) error {
	t.Helper()
	return MustBeValidFunc(t, ruleSet, input, nil, checkAlways)
}

// MustBeInvalid is a test helper that expects a RuleSet to return an error and checks for a specific error code.
// If the error is nil or the code does not match, a testing error is printed and the function returns false.
//
// This function returns the error on "success" so that you can perform additional comparisons.
func MustBeInvalid(t testing.TB, ruleSet rules.RuleSet[any], input any, errorCode errors.ErrorCode) error {
	return MustNotRun(t, ruleSet, input, errorCode)
}

// MustNotRun is a test helper that expects a RuleSet to return an error and checks for a specific error code.
// If the error is nil or the code does not match, a testing error is printed and the function returns false.
//
// This function returns the error on "success" so that you can perform additional comparisons.
func MustNotRun(t testing.TB, ruleSet rules.RuleSet[any], input any, errorCode errors.ErrorCode) error {
	t.Helper()

	_, err := ruleSet.Run(context.TODO(), input)

	if err == nil {
		t.Error("Expected error to not be nil")
		return nil
	} else if err.First().Code() != errorCode {
		t.Errorf("Expected error code of %s got %s (%s)", errorCode, err.First().Code(), err)
		return nil
	}

	return err
}
