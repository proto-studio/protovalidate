package testhelpers

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// ErrorConfigTestRule is a rule that creates errors via errors.Error (which respects context config).
// This can be used with WithRule to test error config propagation.
type ErrorConfigTestRule[T any] struct{}

func (r *ErrorConfigTestRule[T]) Evaluate(ctx context.Context, value T) errors.ValidationError {
	return errors.Error(errors.CodePattern, ctx)
}

func (r *ErrorConfigTestRule[T]) Replaces(other rules.Rule[T]) bool {
	_, ok := other.(*ErrorConfigTestRule[T])
	return ok
}

func (r *ErrorConfigTestRule[T]) String() string {
	return "ErrorConfigTestRule"
}

// ErrorConfigTestRuleFunc returns a rule function that creates errors via errors.Error.
// This can be used with WithRuleFunc to test error config propagation.
func ErrorConfigTestRuleFunc[T any]() rules.RuleFunc[T] {
	return func(ctx context.Context, value T) errors.ValidationError {
		return errors.Error(errors.CodePattern, ctx)
	}
}

// ruleSetWithErrorConfig combines RuleSet with ErrorConfigurable for testing.
// This allows type-safe testing of error config methods using F-bounded polymorphism.
type ruleSetWithErrorConfig[T any, Self any] interface {
	rules.RuleSet[T]
	errors.ErrorConfigurable[T, Self]
}

// MustImplementErrorConfig tests that a rule set properly implements error customization methods
// and that they work correctly. It verifies:
// - WithErrorMessage(short, long string) works
// - WithDocsURI(uri string) works
// - WithTraceURI(uri string) works
// - WithErrorCode(code errors.ErrorCode) works
// - WithErrorMeta(key string, value any) works
// - WithErrorCallback(fn errors.ErrorCallback) works and has access to all error fields
//
// The function tests by triggering a null error (nil input without WithNil).
func MustImplementErrorConfig[T any, RS ruleSetWithErrorConfig[T, RS]](t testing.TB, ruleSet RS) {
	t.Helper()

	mustApplyErrorConfigWithMessage(t, ruleSet)
	mustApplyErrorConfigWithDocsURI(t, ruleSet)
	mustApplyErrorConfigWithTraceURI(t, ruleSet)
	mustApplyErrorConfigWithCode(t, ruleSet)
	mustApplyErrorConfigWithMeta(t, ruleSet)
	mustApplyErrorConfigWithCallback(t, ruleSet)
}

// mustApplyErrorConfigWithMessage tests that WithErrorMessage overrides error messages
func mustApplyErrorConfigWithMessage[T any, RS ruleSetWithErrorConfig[T, RS]](t testing.TB, ruleSet RS) {
	t.Helper()

	ruleSetWithConfig := ruleSet.WithErrorMessage("custom short", "custom long")

	// Trigger an error by passing nil without WithNil
	var output T
	errs := errors.Unwrap(ruleSetWithConfig.Apply(context.Background(), nil, &output))
	if len(errs) == 0 {
		t.Error("WithErrorMessage: Expected validation error for nil input")
		return
	}
	ve := errs[0].(errors.ValidationError)
	if ve.ShortError() != "custom short" {
		t.Errorf("WithErrorMessage: Expected short error 'custom short', got: %s", ve.ShortError())
	}
}

// mustApplyErrorConfigWithDocsURI tests that WithDocsURI sets the documentation URI on errors
func mustApplyErrorConfigWithDocsURI[T any, RS ruleSetWithErrorConfig[T, RS]](t testing.TB, ruleSet RS) {
	t.Helper()

	ruleSetWithConfig := ruleSet.WithDocsURI("https://example.com/docs")

	// Trigger an error by passing nil without WithNil
	var output T
	errs := errors.Unwrap(ruleSetWithConfig.Apply(context.Background(), nil, &output))
	if len(errs) == 0 {
		t.Error("WithDocsURI: Expected validation error for nil input")
		return
	}
	ve := errs[0].(errors.ValidationError)
	if ve.DocsURI() != "https://example.com/docs" {
		t.Errorf("WithDocsURI: Expected DocsURI 'https://example.com/docs', got: %s", ve.DocsURI())
	}
}

// mustApplyErrorConfigWithTraceURI tests that WithTraceURI sets the trace URI on errors
func mustApplyErrorConfigWithTraceURI[T any, RS ruleSetWithErrorConfig[T, RS]](t testing.TB, ruleSet RS) {
	t.Helper()

	ruleSetWithConfig := ruleSet.WithTraceURI("https://example.com/trace/123")

	// Trigger an error by passing nil without WithNil
	var output T
	errs := errors.Unwrap(ruleSetWithConfig.Apply(context.Background(), nil, &output))
	if len(errs) == 0 {
		t.Error("WithTraceURI: Expected validation error for nil input")
		return
	}
	ve := errs[0].(errors.ValidationError)
	if ve.TraceURI() != "https://example.com/trace/123" {
		t.Errorf("WithTraceURI: Expected TraceURI 'https://example.com/trace/123', got: %s", ve.TraceURI())
	}
}

// mustApplyErrorConfigWithCode tests that WithErrorCode overrides the error code
func mustApplyErrorConfigWithCode[T any, RS ruleSetWithErrorConfig[T, RS]](t testing.TB, ruleSet RS) {
	t.Helper()

	ruleSetWithConfig := ruleSet.WithErrorCode(errors.CodeForbidden)

	// Trigger an error by passing nil without WithNil
	var output T
	errs := errors.Unwrap(ruleSetWithConfig.Apply(context.Background(), nil, &output))
	if len(errs) == 0 {
		t.Error("WithErrorCode: Expected validation error for nil input")
		return
	}
	ve := errs[0].(errors.ValidationError)
	if ve.Code() != errors.CodeForbidden {
		t.Errorf("WithErrorCode: Expected code %s, got: %s", errors.CodeForbidden, ve.Code())
	}
}

// mustApplyErrorConfigWithMeta tests that WithErrorMeta adds metadata to errors
func mustApplyErrorConfigWithMeta[T any, RS ruleSetWithErrorConfig[T, RS]](t testing.TB, ruleSet RS) {
	t.Helper()

	ruleSetWithConfig := ruleSet.WithErrorMeta("field", "testvalue")

	// Trigger an error by passing nil without WithNil
	var output T
	errs := errors.Unwrap(ruleSetWithConfig.Apply(context.Background(), nil, &output))
	if len(errs) == 0 {
		t.Error("WithErrorMeta: Expected validation error for nil input")
		return
	}
	ve := errs[0].(errors.ValidationError)
	meta := ve.Meta()
	if meta == nil || meta["field"] != "testvalue" {
		t.Errorf("WithErrorMeta: Expected meta['field'] to be 'testvalue', got: %v", meta)
	}
}

// mustApplyErrorConfigWithCallback tests that WithErrorCallback is invoked and has access to error fields
func mustApplyErrorConfigWithCallback[T any, RS ruleSetWithErrorConfig[T, RS]](t testing.TB, ruleSet RS) {
	t.Helper()

	callbackCalled := false
	var capturedErr errors.ValidationError

	callback := func(ctx context.Context, err errors.ValidationError) errors.ValidationError {
		callbackCalled = true
		capturedErr = err
		return err
	}

	ruleSetWithConfig := ruleSet.WithErrorCallback(callback)

	// Trigger an error by passing nil without WithNil
	var output T
	errs := errors.Unwrap(ruleSetWithConfig.Apply(context.Background(), nil, &output))
	if len(errs) == 0 {
		t.Error("WithErrorCallback: Expected validation error for nil input")
		return
	}

	if !callbackCalled {
		t.Error("WithErrorCallback: Expected callback to be called")
		return
	}

	// Verify the callback had access to all error fields
	if capturedErr.Code() == "" {
		t.Error("WithErrorCallback: Expected callback to have access to Code()")
	}
	if capturedErr.ShortError() == "" {
		t.Error("WithErrorCallback: Expected callback to have access to ShortError()")
	}
	if capturedErr.Error() == "" {
		t.Error("WithErrorCallback: Expected callback to have access to Error()")
	}
	// Params should be available (may be nil for some errors but shouldn't panic)
	_ = capturedErr.Params()
	// Meta should be accessible (may be nil)
	_ = capturedErr.Meta()
}

// MustApplyErrorConfigWithCustomRule tests that error config is applied to errors from custom rules.
// Pass a pre-configured rule set with the custom rule and error config already applied.
// The triggerInput should be a valid value that will pass coercion but trigger the custom rule.
func MustApplyErrorConfigWithCustomRule[T any](t testing.TB, ruleSet rules.RuleSet[T], triggerInput any, expectedDocsURI string) {
	t.Helper()

	var output T
	errs := errors.Unwrap(ruleSet.Apply(context.Background(), triggerInput, &output))
	if len(errs) == 0 {
		t.Error("Expected validation error from custom rule")
		return
	}
	ve := errs[0].(errors.ValidationError)
	if ve.DocsURI() != expectedDocsURI {
		t.Errorf("Expected DocsURI '%s', got: %s", expectedDocsURI, ve.DocsURI())
	}
}

// MustApplyErrorConfigWithMetaOnInput tests that error config meta is applied to a pre-configured rule set.
// The triggerInput should be a value that triggers an error from the rule.
func MustApplyErrorConfigWithMetaOnInput[T any](t testing.TB, ruleSet rules.RuleSet[T], triggerInput any, expectedKey string, expectedValue any) {
	t.Helper()

	var output T
	errs := errors.Unwrap(ruleSet.Apply(context.Background(), triggerInput, &output))
	if len(errs) == 0 {
		t.Error("Expected validation error")
		return
	}
	ve := errs[0].(errors.ValidationError)
	meta := ve.Meta()
	if meta == nil || meta[expectedKey] != expectedValue {
		t.Errorf("Expected meta['%s'] to be '%v', got: %v", expectedKey, expectedValue, meta)
	}
}
