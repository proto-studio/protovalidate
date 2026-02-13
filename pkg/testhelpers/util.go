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

// neverAssignableImpl is an implementation of neverAssignable for use in MustApplyTypes
type neverAssignableImpl struct{ privProp int }

func (na *neverAssignableImpl) priv() { _ = na.privProp }

var _ neverAssignable = (*neverAssignableImpl)(nil)

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

	actualOutput, err := ruleSet.Apply(context.TODO(), input)
	if err != nil {
		str := "Expected error to be nil"
		for _, inner := range errors.Unwrap(err) {
			if ve, ok := inner.(errors.ValidationError); ok {
				str += fmt.Sprintf("\n  %s at %s", ve, ve.Path())
			} else {
				str += fmt.Sprintf("\n  %s", inner)
			}
		}

		t.Error(str)
		return actualOutput, err
	}
	if err := fn(expectedOutput, actualOutput); err != nil {
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

	_, err := ruleSet.Apply(context.TODO(), input)
	if err == nil {
		t.Error("Expected error to not be nil")
		return nil
	}
	if err.Code() != errorCode {
		t.Errorf("Expected error code of %s, got %s (%s)", errorCode, err.Code(), err)
		return nil
	}

	return err
}

// MustApplyTypes checks that Apply returns (value, nil) for valid input.
// Use an input that should not error if the types are correct.
func MustApplyTypes[T any](t testing.TB, ruleSet rules.RuleSet[T], input T) {
	t.Helper()

	output, err := ruleSet.Apply(context.TODO(), input)
	if err != nil {
		t.Errorf("Expected error to be nil for valid input, got: %s", err)
		return
	}
	// Basic sanity: returned value should be usable (type matches)
	_ = output
}

// MustEvaluate is a test helper that expects a Rule to return nil for an error.
// If the error is non-nil, this function prints the error and returns it.
//
// WrapAnyRuleSet calls Apply because of the less strict input type, so when testing Evaluate it is important
// to not use .Any() with this test.
func MustEvaluate[T any](t testing.TB, rule rules.Rule[T], input T) error {
	t.Helper()

	err := rule.Evaluate(context.TODO(), input)

	if err != nil {
		str := "Expected error to be nil"
		for _, inner := range errors.Unwrap(err) {
			if ve, ok := inner.(errors.ValidationError); ok {
				str += fmt.Sprintf("\n  %s at %s", ve, ve.Path())
			} else {
				str += fmt.Sprintf("\n  %s", inner)
			}
		}

		t.Error(str)
		return err
	}

	return nil
}

// MustNotEvaluate is a test helper that expects a Rule to return an error and checks for a specific error code.
// If the error is nil or the code does not match, a testing error is printed and the function returns false.
//
// This function returns the error on "success" so that you can perform additional comparisons.
//
// WrapAnyRuleSet calls Apply because of the less strict input type, so when testing Evaluate it is important
// to not use .Any() with this test.
func MustNotEvaluate[T any](t testing.TB, rule rules.Rule[T], input T, errorCode errors.ErrorCode) error {
	t.Helper()
	err := rule.Evaluate(context.TODO(), input)

	if err == nil {
		t.Error("Expected error to not be nil")
		return nil
	} else if err.Code() != errorCode {
		t.Errorf("Expected error code of %s, got %s (%s)", errorCode, err.Code(), err)
		return nil
	}

	return err
}
