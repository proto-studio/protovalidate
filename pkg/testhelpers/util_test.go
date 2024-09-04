package testhelpers_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

type MockT struct {
	testing.T

	errorCount  int
	errorValues []any
}

func (t *MockT) Error(err ...any) {
	t.errorCount++
	t.errorValues = append(t.errorValues, err...)
}

func (t *MockT) Errorf(msg string, params ...any) {
	t.errorCount++
	t.errorValues = append(t.errorValues, fmt.Sprintf(msg, params...))
}

func TestMustApply(t *testing.T) {
	ruleSet := rules.Any()

	mockT := &MockT{}
	if _, err := testhelpers.MustApply(mockT, ruleSet, 10); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0, got: %d", mockT.errorCount)
	}

	ruleSet = ruleSet.WithRule(testhelpers.NewMockRuleWithErrors[any](1))

	mockT = &MockT{}
	if _, err := testhelpers.MustApply(mockT, ruleSet, 10); err == nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}
}

func TestMustApplyFunc(t *testing.T) {
	ruleSet := rules.Any()
	callCount := 0

	checkValid := func(a, b any) error {
		callCount++
		return nil
	}

	checkInvalid := func(a, b any) error {
		callCount++
		return errors.New(errors.CodeUnknown, "", "")
	}

	mockT := &MockT{}
	if _, err := testhelpers.MustApplyFunc(mockT, ruleSet, 10, 10, checkValid); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0, got: %d", mockT.errorCount)
	}
	if callCount != 1 {
		t.Errorf("Expected check function call count to be 1, got: %d", callCount)
	}

	callCount = 0
	mockT = &MockT{}

	if _, err := testhelpers.MustApplyFunc(mockT, ruleSet, 10, 10, checkInvalid); err == nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}
	if callCount != 1 {
		t.Errorf("Expected check function call count to be 1, got: %d", callCount)
	}
}

func TestMustNotApply(t *testing.T) {
	ruleSet := rules.Any().WithRule(testhelpers.NewMockRuleWithErrors[any](1))

	mockT := &MockT{}
	if err := testhelpers.MustNotApply(mockT, ruleSet, 10, errors.CodeUnknown); err == nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0, got: %d", mockT.errorCount)
	}

	mockT = &MockT{}
	// Wrong code
	if err := testhelpers.MustNotApply(mockT, ruleSet, 10, errors.CodeMin); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}

	ruleSet = rules.Any()

	mockT = &MockT{}
	// Is actually valid
	if err := testhelpers.MustNotApply(mockT, ruleSet, 10, errors.CodeUnknown); err != nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}
}

func TestMustApplyMutation(t *testing.T) {
	out := 10

	mockRuleSet := &testhelpers.MockRuleSet[int]{
		OutputValue: &out,
	}
	mockT := &MockT{}

	if _, err := testhelpers.MustApplyMutation(mockT, mockRuleSet.Any(), 5, 10); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0, got: %d", mockT.errorCount)
	}

	mockRuleSet.Reset()

	if _, err := testhelpers.MustApplyMutation(mockT, mockRuleSet.Any(), 5, 7); err == nil {
		t.Errorf("Expected error to not be nil")
	}
}

// MockNilOk is a mock rule set that incorrectly succeeds when applying nil
type MockNilOk struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilOk) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if output == nil {
		return nil
	}
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	return mockRuleSet.Apply(ctx, input, output)
}

// MockNilOk is a mock rule set that incorrectly succeeds when applying a pointer to nil
type MockNilPtrOk struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilPtrOk) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if outputPtr, ok := output.(*int); ok && outputPtr == nil {
		return nil
	}
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	return mockRuleSet.Apply(ctx, input, output)
}

// MockNilOk is a mock rule set that incorrectly succeeds when applying a non-pointer that matches the type
type MockNonPtrOk struct{ testhelpers.MockRuleSet[int] }

func (m *MockNonPtrOk) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if _, ok := output.(int); ok {
		return nil
	}
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	return mockRuleSet.Apply(ctx, input, output)
}

// MockWrongTypeOk is a mock rule set that incorrectly succeeds when applying a pointer with the wrong type
type MockWrongTypeOk struct{ testhelpers.MockRuleSet[int] }

func (m *MockWrongTypeOk) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	// Always succeed on non-nil pointer regardless of type
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
		return nil
	}
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	return mockRuleSet.Apply(ctx, input, output)
}

// MockWrongErrorCode is a mock rule set that fails with the incorrect error code
// errors.CodeInternal should be used and is replaced with errors.CodeUknown
type MockWrongErrorCode struct{ testhelpers.MockRuleSet[int] }

func (m *MockWrongErrorCode) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	errs := mockRuleSet.Apply(ctx, input, output)

	if errs == nil {
		return nil
	}

	// Replace all CodeInternal errors
	for idx := range errs {
		if errs[idx].Code() == errors.CodeInternal {
			errs[idx] = errors.Errorf(errors.CodeUnknown, ctx, "")
		}
	}

	return errs
}

// MockAlwaysError is a mock rule set that always fails
type MockAlwaysError struct{ testhelpers.MockRuleSet[int] }

func (m *MockAlwaysError) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	errs := mockRuleSet.Apply(ctx, input, output)

	if errs != nil {
		return errs
	}

	return errors.Collection(errors.Errorf(errors.CodeUnknown, ctx, ""))
}

func TestMustApplyTypes(t *testing.T) {

	// MockRuleSet should pass all type tests by default
	var mockRuleSet rules.RuleSet[int] = &testhelpers.MockRuleSet[int]{}
	testhelpers.MustApplyTypes[int](t, mockRuleSet, 123)

	// Incorrectly succeeds on nil output only
	mockT := &MockT{}
	mockRuleSet = &MockNilOk{}
	testhelpers.MustApplyTypes[int](mockT, mockRuleSet, 123)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on mock incorrectly succeeding on nil, got: %d", mockT.errorCount)
	}

	// Incorrectly succeeds on nil pointer output only
	mockT = &MockT{}
	mockRuleSet = &MockNilPtrOk{}
	testhelpers.MustApplyTypes[int](mockT, mockRuleSet, 123)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on mock incorrectly succeeding on nil pointer, got: %d", mockT.errorCount)
	}

	// Incorrectly succeeds on non-pointer with correct type output only
	mockT = &MockT{}
	mockRuleSet = &MockNonPtrOk{}
	testhelpers.MustApplyTypes[int](mockT, mockRuleSet, 123)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on mock incorrectly succeeding on non-pointer with correct type, got: %d", mockT.errorCount)
	}

	// Incorrectly succeeds on incorrect output type
	mockT = &MockT{}
	mockRuleSet = &MockWrongTypeOk{}
	testhelpers.MustApplyTypes[int](mockT, mockRuleSet, 123)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on mock incorrectly succeeding on pointer of incompatible type, got: %d", mockT.errorCount)
	}

	// Fails with incorrect error code
	mockT = &MockT{}
	mockRuleSet = &MockWrongErrorCode{}
	testhelpers.MustApplyTypes[int](mockT, mockRuleSet, 123)
	if mockT.errorCount != 4 {
		t.Errorf("Expected 4 errors on mock incorrectly error code, got: %d", mockT.errorCount)
	}

	// Fail success cases
	mockT = &MockT{}
	mockRuleSet = &MockAlwaysError{}
	testhelpers.MustApplyTypes[int](mockT, mockRuleSet, 123)
	if mockT.errorCount != 2 {
		t.Errorf("Expected 2 errors on mock failed success cases, got: %d", mockT.errorCount)
	}
}
