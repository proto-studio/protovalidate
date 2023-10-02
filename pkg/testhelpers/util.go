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

// MockCustomRule is a mock implementation of the Rule interface that can be used for testing.
// It accepts a return value to return from the rule and also a number of errors to return.
//
// If errorCount is 0 than nil is returned for errors.
// The return value of this function is a Rule function.
func MockCustomRule[T any](retval T, errorCount int) func(_ context.Context, _ T) (T, errors.ValidationErrorCollection) {
	var errs errors.ValidationErrorCollection

	if errorCount > 0 {
		errs = make(errors.ValidationErrorCollection, errorCount)

		for i := 0; i < errorCount; i++ {
			errs[i] = errors.Errorf(errors.CodeUnknown, context.Background(), "test")
		}
	}

	return func(_ context.Context, _ T) (T, errors.ValidationErrorCollection) {
		return retval, errs
	}
}

// checkEqual is a simple validity function that returns true if both values are equal.
func checkEqual(a, b any) error {
	if a != b {
		return fmt.Errorf("Expected output to be %v, got: %v", a, b)
	}
	return nil
}

// MustBeValidFunc is a test helper that expects a RuleSet to a nil error.
// If the error is non-nil or the check function returns an error, this function prints the error and returns it.
func MustBeValidFunc(t testing.TB, ruleSet rules.RuleSet[any], input, expectedOutput any, fn func(a, b any) error) error {
	t.Helper()

	actualOutput, err := ruleSet.Validate(input)

	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
		return err
	} else if err := fn(expectedOutput, actualOutput); err != nil {
		t.Error(err)
		return err
	}

	return nil
}

// MustBeValid is a test helper that expects a RuleSet to return a specific value and nil error.
// If the error is non-nil or the expected output does not match, this function prints the error and returns it.
func MustBeValid(t testing.TB, ruleSet rules.RuleSet[any], input, expectedOutput any) error {
	t.Helper()
	return MustBeValidFunc(t, ruleSet, input, expectedOutput, checkEqual)
}

// MustBeInvalid is a test helper that expects a RuleSet to return an error and checks for a specific error code.
// If the error is nil or the code does not match, a testing error is printed and the function returns false.
//
// This function returns the error on "success" so that you can perform additional comparisons.
func MustBeInvalid(t testing.TB, ruleSet rules.RuleSet[any], input any, errorCode errors.ErrorCode) error {
	t.Helper()

	_, err := ruleSet.Validate(input)

	if err == nil {
		t.Error("Expected error to not be nil")
		return nil
	} else if err.First().Code() != errorCode {
		t.Errorf("Expected error code of %d got %d (%s)", errorCode, err.First().Code(), err)
		return nil
	}

	return err
}
