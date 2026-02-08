package testhelpers_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestErrorConfigTestRule_Replaces tests:
// - ErrorConfigTestRule.Replaces returns true for same type
// - ErrorConfigTestRule.Replaces returns false for different types
func TestErrorConfigTestRule_Replaces(t *testing.T) {
	rule1 := &testhelpers.ErrorConfigTestRule[string]{}
	rule2 := &testhelpers.ErrorConfigTestRule[string]{}
	mockRule := testhelpers.NewMockRule[string]()

	// Same type should replace
	if !rule1.Replaces(rule2) {
		t.Error("Expected ErrorConfigTestRule to replace another ErrorConfigTestRule")
	}

	// Different type should not replace
	if rule1.Replaces(mockRule) {
		t.Error("Expected ErrorConfigTestRule to not replace MockRule")
	}
}

// TestErrorConfigTestRule_String tests:
// - ErrorConfigTestRule.String returns correct representation
func TestErrorConfigTestRule_String(t *testing.T) {
	rule := &testhelpers.ErrorConfigTestRule[string]{}
	expected := "ErrorConfigTestRule"

	if s := rule.String(); s != expected {
		t.Errorf("Expected String() to be %q, got %q", expected, s)
	}
}

// TestErrorConfigTestRuleFunc tests:
// - ErrorConfigTestRuleFunc returns a function that produces errors
func TestErrorConfigTestRuleFunc(t *testing.T) {
	fn := testhelpers.ErrorConfigTestRuleFunc[string]()
	if fn == nil {
		t.Fatal("Expected ErrorConfigTestRuleFunc to return non-nil function")
	}

	// The function should return errors (since it's for testing error config)
	errs := fn(context.Background(), "test")
	if len(errors.Unwrap(errs)) == 0 {
		t.Error("Expected ErrorConfigTestRuleFunc to produce errors")
	}
}

// =============================================================================
// Mock implementations with broken error config methods for testing failures
// =============================================================================

// MockBrokenMessage has a WithErrorMessage that doesn't actually set the message
type MockBrokenMessage struct{ testhelpers.MockRuleSet[int] }

func (m *MockBrokenMessage) WithErrorMessage(short, long string) *MockBrokenMessage {
	return m // Returns itself unchanged - broken implementation
}

func (m *MockBrokenMessage) WithDocsURI(uri string) *MockBrokenMessage {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithDocsURI(uri)
	return &newRuleSet
}

func (m *MockBrokenMessage) WithTraceURI(uri string) *MockBrokenMessage {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithTraceURI(uri)
	return &newRuleSet
}

func (m *MockBrokenMessage) WithErrorCode(code errors.ErrorCode) *MockBrokenMessage {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorCode(code)
	return &newRuleSet
}

func (m *MockBrokenMessage) WithErrorMeta(key string, value any) *MockBrokenMessage {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorMeta(key, value)
	return &newRuleSet
}

func (m *MockBrokenMessage) WithErrorCallback(fn errors.ErrorCallback) *MockBrokenMessage {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorCallback(fn)
	return &newRuleSet
}

// MockBrokenDocsURI has a WithDocsURI that doesn't actually set the URI
type MockBrokenDocsURI struct{ testhelpers.MockRuleSet[int] }

func (m *MockBrokenDocsURI) WithErrorMessage(short, long string) *MockBrokenDocsURI {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorMessage(short, long)
	return &newRuleSet
}

func (m *MockBrokenDocsURI) WithDocsURI(uri string) *MockBrokenDocsURI {
	return m // Returns itself unchanged - broken implementation
}

func (m *MockBrokenDocsURI) WithTraceURI(uri string) *MockBrokenDocsURI {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithTraceURI(uri)
	return &newRuleSet
}

func (m *MockBrokenDocsURI) WithErrorCode(code errors.ErrorCode) *MockBrokenDocsURI {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorCode(code)
	return &newRuleSet
}

func (m *MockBrokenDocsURI) WithErrorMeta(key string, value any) *MockBrokenDocsURI {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorMeta(key, value)
	return &newRuleSet
}

func (m *MockBrokenDocsURI) WithErrorCallback(fn errors.ErrorCallback) *MockBrokenDocsURI {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorCallback(fn)
	return &newRuleSet
}

// MockBrokenTraceURI has a WithTraceURI that doesn't actually set the URI
type MockBrokenTraceURI struct{ testhelpers.MockRuleSet[int] }

func (m *MockBrokenTraceURI) WithErrorMessage(short, long string) *MockBrokenTraceURI {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorMessage(short, long)
	return &newRuleSet
}

func (m *MockBrokenTraceURI) WithDocsURI(uri string) *MockBrokenTraceURI {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithDocsURI(uri)
	return &newRuleSet
}

func (m *MockBrokenTraceURI) WithTraceURI(uri string) *MockBrokenTraceURI {
	return m // Returns itself unchanged - broken implementation
}

func (m *MockBrokenTraceURI) WithErrorCode(code errors.ErrorCode) *MockBrokenTraceURI {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorCode(code)
	return &newRuleSet
}

func (m *MockBrokenTraceURI) WithErrorMeta(key string, value any) *MockBrokenTraceURI {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorMeta(key, value)
	return &newRuleSet
}

func (m *MockBrokenTraceURI) WithErrorCallback(fn errors.ErrorCallback) *MockBrokenTraceURI {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorCallback(fn)
	return &newRuleSet
}

// MockBrokenCode has a WithErrorCode that doesn't actually set the code
type MockBrokenCode struct{ testhelpers.MockRuleSet[int] }

func (m *MockBrokenCode) WithErrorMessage(short, long string) *MockBrokenCode {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorMessage(short, long)
	return &newRuleSet
}

func (m *MockBrokenCode) WithDocsURI(uri string) *MockBrokenCode {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithDocsURI(uri)
	return &newRuleSet
}

func (m *MockBrokenCode) WithTraceURI(uri string) *MockBrokenCode {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithTraceURI(uri)
	return &newRuleSet
}

func (m *MockBrokenCode) WithErrorCode(code errors.ErrorCode) *MockBrokenCode {
	return m // Returns itself unchanged - broken implementation
}

func (m *MockBrokenCode) WithErrorMeta(key string, value any) *MockBrokenCode {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorMeta(key, value)
	return &newRuleSet
}

func (m *MockBrokenCode) WithErrorCallback(fn errors.ErrorCallback) *MockBrokenCode {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorCallback(fn)
	return &newRuleSet
}

// MockBrokenMeta has a WithErrorMeta that doesn't actually set metadata
type MockBrokenMeta struct{ testhelpers.MockRuleSet[int] }

func (m *MockBrokenMeta) WithErrorMessage(short, long string) *MockBrokenMeta {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorMessage(short, long)
	return &newRuleSet
}

func (m *MockBrokenMeta) WithDocsURI(uri string) *MockBrokenMeta {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithDocsURI(uri)
	return &newRuleSet
}

func (m *MockBrokenMeta) WithTraceURI(uri string) *MockBrokenMeta {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithTraceURI(uri)
	return &newRuleSet
}

func (m *MockBrokenMeta) WithErrorCode(code errors.ErrorCode) *MockBrokenMeta {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorCode(code)
	return &newRuleSet
}

func (m *MockBrokenMeta) WithErrorMeta(key string, value any) *MockBrokenMeta {
	return m // Returns itself unchanged - broken implementation
}

func (m *MockBrokenMeta) WithErrorCallback(fn errors.ErrorCallback) *MockBrokenMeta {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorCallback(fn)
	return &newRuleSet
}

// MockBrokenCallback has a WithErrorCallback that doesn't actually call the callback
type MockBrokenCallback struct{ testhelpers.MockRuleSet[int] }

func (m *MockBrokenCallback) WithErrorMessage(short, long string) *MockBrokenCallback {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorMessage(short, long)
	return &newRuleSet
}

func (m *MockBrokenCallback) WithDocsURI(uri string) *MockBrokenCallback {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithDocsURI(uri)
	return &newRuleSet
}

func (m *MockBrokenCallback) WithTraceURI(uri string) *MockBrokenCallback {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithTraceURI(uri)
	return &newRuleSet
}

func (m *MockBrokenCallback) WithErrorCode(code errors.ErrorCode) *MockBrokenCallback {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorCode(code)
	return &newRuleSet
}

func (m *MockBrokenCallback) WithErrorMeta(key string, value any) *MockBrokenCallback {
	newRuleSet := *m
	newRuleSet.MockRuleSet = *m.MockRuleSet.WithErrorMeta(key, value)
	return &newRuleSet
}

func (m *MockBrokenCallback) WithErrorCallback(fn errors.ErrorCallback) *MockBrokenCallback {
	return m // Returns itself unchanged - callback won't be stored or called
}

// brokenValidationError is a custom ValidationError that returns empty strings for testing
type brokenValidationError struct {
	code     errors.ErrorCode
	shortErr string
	longErr  string
}

func (e *brokenValidationError) Code() errors.ErrorCode { return e.code }
func (e *brokenValidationError) Path() string           { return "" }
func (e *brokenValidationError) PathAs(serializer errors.PathSerializer) string { return "" }
func (e *brokenValidationError) ShortError() string     { return e.shortErr }
func (e *brokenValidationError) Error() string          { return e.longErr }
func (e *brokenValidationError) DocsURI() string        { return "" }
func (e *brokenValidationError) TraceURI() string       { return "" }
func (e *brokenValidationError) Meta() map[string]any   { return nil }
func (e *brokenValidationError) Params() []any          { return nil }
func (e *brokenValidationError) Internal() bool         { return false }
func (e *brokenValidationError) Validation() bool       { return true }
func (e *brokenValidationError) Permission() bool       { return false }
func (e *brokenValidationError) Unwrap() []error        { return nil }

// MockCallbackBrokenError produces errors with empty fields that the callback will capture
type MockCallbackBrokenError struct {
	testhelpers.MockRuleSet[int]
	brokenErr *brokenValidationError
	callback  errors.ErrorCallback
}

func (m *MockCallbackBrokenError) Apply(ctx context.Context, input, output any) errors.ValidationError {
	if input == nil {
		if m.callback != nil {
			return m.callback(ctx, m.brokenErr)
		}
		return m.brokenErr
	}
	return m.MockRuleSet.Apply(ctx, input, output)
}

func (m *MockCallbackBrokenError) WithErrorMessage(short, long string) *MockCallbackBrokenError {
	return m
}

func (m *MockCallbackBrokenError) WithDocsURI(uri string) *MockCallbackBrokenError {
	return m
}

func (m *MockCallbackBrokenError) WithTraceURI(uri string) *MockCallbackBrokenError {
	return m
}

func (m *MockCallbackBrokenError) WithErrorCode(code errors.ErrorCode) *MockCallbackBrokenError {
	return m
}

func (m *MockCallbackBrokenError) WithErrorMeta(key string, value any) *MockCallbackBrokenError {
	return m
}

func (m *MockCallbackBrokenError) WithErrorCallback(fn errors.ErrorCallback) *MockCallbackBrokenError {
	newRuleSet := *m
	newRuleSet.callback = fn
	return &newRuleSet
}

// MockNoErrors has Apply that succeeds even when it should fail (doesn't return errors for nil)
type MockNoErrors struct{ testhelpers.MockRuleSet[int] }

func (m *MockNoErrors) Apply(ctx context.Context, input, output any) errors.ValidationError {
	return nil // Always succeeds - broken for testing nil error handling
}

func (m *MockNoErrors) WithErrorMessage(short, long string) *MockNoErrors {
	return m
}

func (m *MockNoErrors) WithDocsURI(uri string) *MockNoErrors {
	return m
}

func (m *MockNoErrors) WithTraceURI(uri string) *MockNoErrors {
	return m
}

func (m *MockNoErrors) WithErrorCode(code errors.ErrorCode) *MockNoErrors {
	return m
}

func (m *MockNoErrors) WithErrorMeta(key string, value any) *MockNoErrors {
	return m
}

func (m *MockNoErrors) WithErrorCallback(fn errors.ErrorCallback) *MockNoErrors {
	return m
}

// =============================================================================
// Tests for MustImplementErrorConfig catching broken implementations
// =============================================================================

// TestMustImplementErrorConfig_ValidRuleSet tests:
// - MustImplementErrorConfig passes for valid implementations
func TestMustImplementErrorConfig_ValidRuleSet(t *testing.T) {
	mockT := &MockT{}
	testhelpers.MustImplementErrorConfig[int, *testhelpers.MockRuleSet[int]](mockT, testhelpers.NewMockRuleSet[int]())
	if mockT.errorCount != 0 {
		t.Errorf("Expected 0 errors for valid rule set, got: %d", mockT.errorCount)
	}
}

// TestMustImplementErrorConfig_BrokenMessage tests:
// - MustImplementErrorConfig catches broken WithErrorMessage
func TestMustImplementErrorConfig_BrokenMessage(t *testing.T) {
	mockT := &MockT{}
	testhelpers.MustImplementErrorConfig[int, *MockBrokenMessage](mockT, &MockBrokenMessage{})
	if mockT.errorCount == 0 {
		t.Error("Expected errors for broken WithErrorMessage implementation")
	}
}

// TestMustImplementErrorConfig_BrokenDocsURI tests:
// - MustImplementErrorConfig catches broken WithDocsURI
func TestMustImplementErrorConfig_BrokenDocsURI(t *testing.T) {
	mockT := &MockT{}
	testhelpers.MustImplementErrorConfig[int, *MockBrokenDocsURI](mockT, &MockBrokenDocsURI{})
	if mockT.errorCount == 0 {
		t.Error("Expected errors for broken WithDocsURI implementation")
	}
}

// TestMustImplementErrorConfig_BrokenTraceURI tests:
// - MustImplementErrorConfig catches broken WithTraceURI
func TestMustImplementErrorConfig_BrokenTraceURI(t *testing.T) {
	mockT := &MockT{}
	testhelpers.MustImplementErrorConfig[int, *MockBrokenTraceURI](mockT, &MockBrokenTraceURI{})
	if mockT.errorCount == 0 {
		t.Error("Expected errors for broken WithTraceURI implementation")
	}
}

// TestMustImplementErrorConfig_BrokenCode tests:
// - MustImplementErrorConfig catches broken WithErrorCode
func TestMustImplementErrorConfig_BrokenCode(t *testing.T) {
	mockT := &MockT{}
	testhelpers.MustImplementErrorConfig[int, *MockBrokenCode](mockT, &MockBrokenCode{})
	if mockT.errorCount == 0 {
		t.Error("Expected errors for broken WithErrorCode implementation")
	}
}

// TestMustImplementErrorConfig_BrokenMeta tests:
// - MustImplementErrorConfig catches broken WithErrorMeta
func TestMustImplementErrorConfig_BrokenMeta(t *testing.T) {
	mockT := &MockT{}
	testhelpers.MustImplementErrorConfig[int, *MockBrokenMeta](mockT, &MockBrokenMeta{})
	if mockT.errorCount == 0 {
		t.Error("Expected errors for broken WithErrorMeta implementation")
	}
}

// TestMustImplementErrorConfig_BrokenCallback tests:
// - MustImplementErrorConfig catches broken WithErrorCallback
func TestMustImplementErrorConfig_BrokenCallback(t *testing.T) {
	mockT := &MockT{}
	testhelpers.MustImplementErrorConfig[int, *MockBrokenCallback](mockT, &MockBrokenCallback{})
	if mockT.errorCount == 0 {
		t.Error("Expected errors for broken WithErrorCallback implementation")
	}
}

// TestMustImplementErrorConfig_NoErrors tests:
// - MustImplementErrorConfig catches implementations that don't return errors when they should
func TestMustImplementErrorConfig_NoErrors(t *testing.T) {
	mockT := &MockT{}
	testhelpers.MustImplementErrorConfig[int, *MockNoErrors](mockT, &MockNoErrors{})
	if mockT.errorCount == 0 {
		t.Error("Expected errors when Apply doesn't return errors for nil input")
	}
}

// TestMustImplementErrorConfig_CallbackEmptyCode tests:
// - MustImplementErrorConfig catches callback receiving error with empty Code
func TestMustImplementErrorConfig_CallbackEmptyCode(t *testing.T) {
	mockT := &MockT{}
	brokenMock := &MockCallbackBrokenError{
		brokenErr: &brokenValidationError{
			code:     "", // Empty code
			shortErr: "short",
			longErr:  "long",
		},
	}
	testhelpers.MustImplementErrorConfig[int, *MockCallbackBrokenError](mockT, brokenMock)
	if mockT.errorCount == 0 {
		t.Error("Expected errors when callback receives error with empty Code")
	}
}

// TestMustImplementErrorConfig_CallbackEmptyShortError tests:
// - MustImplementErrorConfig catches callback receiving error with empty ShortError
func TestMustImplementErrorConfig_CallbackEmptyShortError(t *testing.T) {
	mockT := &MockT{}
	brokenMock := &MockCallbackBrokenError{
		brokenErr: &brokenValidationError{
			code:     errors.CodeUnknown,
			shortErr: "", // Empty short error
			longErr:  "long",
		},
	}
	testhelpers.MustImplementErrorConfig[int, *MockCallbackBrokenError](mockT, brokenMock)
	if mockT.errorCount == 0 {
		t.Error("Expected errors when callback receives error with empty ShortError")
	}
}

// TestMustImplementErrorConfig_CallbackEmptyLongError tests:
// - MustImplementErrorConfig catches callback receiving error with empty Error (long)
func TestMustImplementErrorConfig_CallbackEmptyLongError(t *testing.T) {
	mockT := &MockT{}
	brokenMock := &MockCallbackBrokenError{
		brokenErr: &brokenValidationError{
			code:     errors.CodeUnknown,
			shortErr: "short",
			longErr:  "", // Empty long error
		},
	}
	testhelpers.MustImplementErrorConfig[int, *MockCallbackBrokenError](mockT, brokenMock)
	if mockT.errorCount == 0 {
		t.Error("Expected errors when callback receives error with empty Error")
	}
}

// =============================================================================
// Tests for MustApplyErrorConfigWithCustomRule catching failures
// =============================================================================

// TestMustApplyErrorConfigWithCustomRule_Success tests:
// - MustApplyErrorConfigWithCustomRule passes for correct implementation
func TestMustApplyErrorConfigWithCustomRule_Success(t *testing.T) {
	mockT := &MockT{}

	ruleSet := rules.String().
		WithRule(&testhelpers.ErrorConfigTestRule[string]{}).
		WithDocsURI("https://example.com/test")

	testhelpers.MustApplyErrorConfigWithCustomRule[string](mockT, ruleSet, "trigger", "https://example.com/test")

	if mockT.errorCount != 0 {
		t.Errorf("Expected 0 errors for valid rule set, got: %d with messages: %v", mockT.errorCount, mockT.errorValues)
	}
}

// TestMustApplyErrorConfigWithCustomRule_WrongURI tests:
// - MustApplyErrorConfigWithCustomRule catches wrong DocsURI
func TestMustApplyErrorConfigWithCustomRule_WrongURI(t *testing.T) {
	mockT := &MockT{}

	ruleSet := rules.String().
		WithRule(&testhelpers.ErrorConfigTestRule[string]{}).
		WithDocsURI("https://example.com/test")

	testhelpers.MustApplyErrorConfigWithCustomRule[string](mockT, ruleSet, "trigger", "https://wrong.com/uri")

	if mockT.errorCount == 0 {
		t.Error("Expected error when DocsURI doesn't match")
	}
}

// TestMustApplyErrorConfigWithCustomRule_NoError tests:
// - MustApplyErrorConfigWithCustomRule catches when no error is triggered
func TestMustApplyErrorConfigWithCustomRule_NoError(t *testing.T) {
	mockT := &MockT{}

	// Rule set without any rules that would trigger errors
	ruleSet := rules.String()

	testhelpers.MustApplyErrorConfigWithCustomRule[string](mockT, ruleSet, "valid-input", "https://example.com/test")

	if mockT.errorCount == 0 {
		t.Error("Expected error when no validation error is triggered")
	}
}

// =============================================================================
// Tests for MustApplyErrorConfigWithMetaOnInput catching failures
// =============================================================================

// TestMustApplyErrorConfigWithMetaOnInput_Success tests:
// - MustApplyErrorConfigWithMetaOnInput passes for correct implementation
func TestMustApplyErrorConfigWithMetaOnInput_Success(t *testing.T) {
	mockT := &MockT{}

	ruleSet := rules.String().
		WithRule(&testhelpers.ErrorConfigTestRule[string]{}).
		WithErrorMeta("key", "value")

	testhelpers.MustApplyErrorConfigWithMetaOnInput[string](mockT, ruleSet, "trigger", "key", "value")

	if mockT.errorCount != 0 {
		t.Errorf("Expected 0 errors for valid rule set, got: %d with messages: %v", mockT.errorCount, mockT.errorValues)
	}
}

// TestMustApplyErrorConfigWithMetaOnInput_WrongValue tests:
// - MustApplyErrorConfigWithMetaOnInput catches wrong meta value
func TestMustApplyErrorConfigWithMetaOnInput_WrongValue(t *testing.T) {
	mockT := &MockT{}

	ruleSet := rules.String().
		WithRule(&testhelpers.ErrorConfigTestRule[string]{}).
		WithErrorMeta("key", "value")

	testhelpers.MustApplyErrorConfigWithMetaOnInput[string](mockT, ruleSet, "trigger", "key", "wrong-value")

	if mockT.errorCount == 0 {
		t.Error("Expected error when meta value doesn't match")
	}
}

// TestMustApplyErrorConfigWithMetaOnInput_MissingKey tests:
// - MustApplyErrorConfigWithMetaOnInput catches missing meta key
func TestMustApplyErrorConfigWithMetaOnInput_MissingKey(t *testing.T) {
	mockT := &MockT{}

	ruleSet := rules.String().
		WithRule(&testhelpers.ErrorConfigTestRule[string]{}).
		WithErrorMeta("different-key", "value")

	testhelpers.MustApplyErrorConfigWithMetaOnInput[string](mockT, ruleSet, "trigger", "expected-key", "value")

	if mockT.errorCount == 0 {
		t.Error("Expected error when meta key doesn't exist")
	}
}

// TestMustApplyErrorConfigWithMetaOnInput_NoError tests:
// - MustApplyErrorConfigWithMetaOnInput catches when no error is triggered
func TestMustApplyErrorConfigWithMetaOnInput_NoError(t *testing.T) {
	mockT := &MockT{}

	// Rule set without any rules that would trigger errors
	ruleSet := rules.String()

	testhelpers.MustApplyErrorConfigWithMetaOnInput[string](mockT, ruleSet, "valid-input", "key", "value")

	if mockT.errorCount == 0 {
		t.Error("Expected error when no validation error is triggered")
	}
}
