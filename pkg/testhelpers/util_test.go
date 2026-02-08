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

// MockT is a mock testing.T implementation used for testing test helpers.
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

// TestMustApply tests:
// - MustApply correctly validates rule sets
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

// ruleSetWithPlainUnwrapError is a RuleSet whose Apply returns an error whose Unwrap() contains a plain error.
// Used to cover the else branch in MustApplyFunc's error-formatting loop.
type ruleSetWithPlainUnwrapError struct {
	testhelpers.MockRuleSet[any]
}

func (r *ruleSetWithPlainUnwrapError) Apply(_ context.Context, _ any, _ any) errors.ValidationError {
	return &errorWithPlainUnwrap{msg: "apply err"}
}

// TestMustApplyFunc tests:
// - MustApplyFunc correctly validates rule sets with custom check function
// - MustApplyFunc formats unwrapped errors that are not ValidationError (else branch)
func TestMustApplyFunc(t *testing.T) {
	ruleSet := rules.Any()
	callCount := 0

	checkValid := func(a, b any) error {
		callCount++
		return nil
	}

	checkInvalid := func(a, b any) error {
		callCount++
		return errors.Errorf(errors.CodeUnknown, context.Background(), "unknown", "check failed")
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

	// RuleSet that returns an error whose Unwrap() contains a plain error (covers else branch in MustApplyFunc)
	ruleSetPlainUnwrap := &ruleSetWithPlainUnwrapError{
		MockRuleSet: *testhelpers.NewMockRuleSet[any](),
	}
	mockT = &MockT{}
	if _, err := testhelpers.MustApplyFunc(mockT, ruleSetPlainUnwrap, 10, 10, checkValid); err == nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}
}

// TestMustNotApply tests:
// - MustNotApply correctly validates that rule sets return errors
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

// TestMustApplyMutation tests:
// - MustApplyMutation correctly validates rule sets with expected output mutations
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

func (m *MockNilOk) Apply(ctx context.Context, input, output any) errors.ValidationError {
	if output == nil {
		return nil
	}
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	return mockRuleSet.Apply(ctx, input, output)
}

// MockNilOk is a mock rule set that incorrectly succeeds when applying a pointer to nil
type MockNilPtrOk struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilPtrOk) Apply(ctx context.Context, input, output any) errors.ValidationError {
	if outputPtr, ok := output.(*int); ok && outputPtr == nil {
		return nil
	}
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	return mockRuleSet.Apply(ctx, input, output)
}

// MockNilOk is a mock rule set that incorrectly succeeds when applying a non-pointer that matches the type
type MockNonPtrOk struct{ testhelpers.MockRuleSet[int] }

func (m *MockNonPtrOk) Apply(ctx context.Context, input, output any) errors.ValidationError {
	if _, ok := output.(int); ok {
		return nil
	}
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	return mockRuleSet.Apply(ctx, input, output)
}

// MockWrongTypeOk is a mock rule set that incorrectly succeeds when applying a pointer with the wrong type
type MockWrongTypeOk struct{ testhelpers.MockRuleSet[int] }

func (m *MockWrongTypeOk) Apply(ctx context.Context, input, output any) errors.ValidationError {
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

func (m *MockWrongErrorCode) Apply(ctx context.Context, input, output any) errors.ValidationError {
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	errs := mockRuleSet.Apply(ctx, input, output)

	if errs == nil {
		return nil
	}
	coll := errors.Unwrap(errs)
	var out []error
	for _, e := range coll {
		ve, ok := e.(errors.ValidationError)
		if !ok {
			out = append(out, e)
			continue
		}
		if ve.Code() == errors.CodeInternal {
			out = append(out, errors.Errorf(errors.CodeUnknown, ctx, "unknown error", ""))
		} else {
			out = append(out, ve)
		}
	}
	return errors.Join(out...)
}

// MockAlwaysError is a mock rule set that always fails
type MockAlwaysError struct{ testhelpers.MockRuleSet[int] }

func (m *MockAlwaysError) Apply(ctx context.Context, input, output any) errors.ValidationError {
	mockRuleSet := &testhelpers.MockRuleSet[int]{}
	errs := mockRuleSet.Apply(ctx, input, output)

	if errs != nil {
		return errs
	}

	return errors.Join(errors.Errorf(errors.CodeUnknown, ctx, "unknown error", ""))
}

// TestMustApplyTypes tests:
// - MustApplyTypes correctly validates rule sets with type checking
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

// errorWithPlainUnwrap is a ValidationError whose Unwrap() returns a plain (non-ValidationError) error.
// Used to cover the branch in MustEvaluate/MustApplyFunc that formats non-VE unwrapped errors.
type errorWithPlainUnwrap struct {
	msg string
}

func (e *errorWithPlainUnwrap) Error() string                                      { return e.msg }
func (e *errorWithPlainUnwrap) Unwrap() []error                                   { return []error{fmt.Errorf("inner plain error")} }
func (e *errorWithPlainUnwrap) Code() errors.ErrorCode                             { return errors.CodeUnknown }
func (e *errorWithPlainUnwrap) Path() string                                      { return "" }
func (e *errorWithPlainUnwrap) PathAs(_ errors.PathSerializer) string              { return "" }
func (e *errorWithPlainUnwrap) ShortError() string                                 { return "short" }
func (e *errorWithPlainUnwrap) DocsURI() string                                   { return "" }
func (e *errorWithPlainUnwrap) TraceURI() string                                   { return "" }
func (e *errorWithPlainUnwrap) Meta() map[string]any                               { return nil }
func (e *errorWithPlainUnwrap) Params() []any                                      { return nil }
func (e *errorWithPlainUnwrap) Internal() bool                                     { return false }
func (e *errorWithPlainUnwrap) Validation() bool                                  { return true }
func (e *errorWithPlainUnwrap) Permission() bool                                  { return false }

// TestMustEvaluate tests:
// - MustEvaluate correctly validates rules
// - MustEvaluate formats unwrapped errors that are not ValidationError (else branch)
func TestMustEvaluate(t *testing.T) {
	rule := testhelpers.NewMockRuleWithErrors[any](1)

	mockT := &MockT{}
	if err := testhelpers.MustEvaluate[any](mockT, rule, 10); err == nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}

	rule = testhelpers.NewMockRule[any]()
	mockT = &MockT{}
	if err := testhelpers.MustEvaluate[any](mockT, rule, 10); err != nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0, got: %d", mockT.errorCount)
	}

	// Rule that returns an error whose Unwrap() contains a plain error (covers else branch in MustEvaluate)
	ruleWithPlainUnwrap := rules.RuleFunc[any](func(_ context.Context, _ any) errors.ValidationError {
		return &errorWithPlainUnwrap{msg: "wrapper"}
	})
	mockT = &MockT{}
	if err := testhelpers.MustEvaluate[any](mockT, ruleWithPlainUnwrap, 10); err == nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}
}

// TestMustNotEvaluate tests:
// - MustNotEvaluate correctly validates that rules return errors
func TestMustNotEvaluate(t *testing.T) {
	rule := testhelpers.NewMockRuleWithErrors[any](1)

	mockT := &MockT{}
	if err := testhelpers.MustNotEvaluate[any](mockT, rule, 10, errors.CodeUnknown); err == nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0, got: %d", mockT.errorCount)
	}

	mockT = &MockT{}
	// Wrong code
	if err := testhelpers.MustNotEvaluate[any](mockT, rule, 10, errors.CodeMin); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}

	rule = testhelpers.NewMockRule[any]()
	mockT = &MockT{}
	// Is actually valid
	if err := testhelpers.MustNotEvaluate[any](mockT, rule, 10, errors.CodeUnknown); err != nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}
}
