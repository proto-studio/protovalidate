package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestErrorf_WithPath tests:
// - Creates validation error with correct code, path, and message
func TestErrorf_WithPath(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a.b.c")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "testmessage")

	if err.Code() != errors.CodeMin {
		t.Errorf("Expected code to be %s, got: %s", errors.CodeMin, err.Code())
	}

	if err.Path() != "/a.b.c" {
		t.Errorf("Expected path to be %s, got: %s", "/a.b.c", err.Path())
	}

	if err.Error() != "testmessage" {
		t.Errorf("Expected message to be %s, got: %s", "testmessage", err)
	}
}

// TestErrorf_EmptyPath tests:
// - Errorf with no path in context returns empty path
func TestErrorf_EmptyPath(t *testing.T) {
	err := errors.Errorf(errors.CodeMin, context.Background(), "short", "testmessage")

	if err.Path() != "" {
		t.Errorf("Expected empty path, got: %s", err.Path())
	}
}

// TestErrorfContainsFullPath tests:
// - Errorf includes full path from context
func TestErrorfContainsFullPath(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	err := errors.Errorf(errors.CodeUnknown, ctx, "unknown error", "error")

	if err.Path() != "/a/b" {
		t.Errorf("Expected full path to be /a/b, got: %s", err.Path())
	}
}

// TestErrorfContainsCode tests:
// - Errorf includes correct error code
func TestErrorfContainsCode(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")

	err := errors.Errorf(errors.CodeUnknown, ctx, "unknown error", "error")

	if err.Code() != errors.CodeUnknown {
		t.Errorf("Expected code to be %s, got: %s", errors.CodeUnknown, err.Code())
	}

	err = errors.Errorf(errors.CodeMin, ctx, "below minimum", "error")

	if err.Code() != errors.CodeMin {
		t.Errorf("Expected code to be %s, got: %s", errors.CodeMin, err.Code())
	}
}

// TestErrorMessage tests:
// - Error message is correctly formatted
func TestErrorMessage(t *testing.T) {
	err := errors.Errorf(errors.CodeUnknown, context.Background(), "unknown error", "error123")

	if msg := err.Error(); msg != "error123" {
		t.Errorf("Expected error message to be %s, got: %s", "error123", msg)
	}
}

// TestErrorf_ShortAndLong tests:
// - Errorf correctly sets short and long messages
func TestErrorf_ShortAndLong(t *testing.T) {
	err := errors.Errorf(errors.CodeMin, context.Background(), "short msg", "long message with %d", 42)

	if err.ShortError() != "short msg" {
		t.Errorf("Expected short error to be 'short msg', got: %s", err.ShortError())
	}

	if err.Error() != "long message with 42" {
		t.Errorf("Expected long error to be 'long message with 42', got: %s", err.Error())
	}
}

// TestErrorf_Params tests:
// - Errorf stores the format params for callback access
func TestErrorf_Params(t *testing.T) {
	err := errors.Errorf(errors.CodeMin, context.Background(), "short", "must be at least %d and at most %d", 10, 100)

	params := err.Params()
	if len(params) != 2 {
		t.Fatalf("Expected 2 params, got: %d", len(params))
	}

	if params[0] != 10 {
		t.Errorf("Expected params[0] to be 10, got: %v", params[0])
	}

	if params[1] != 100 {
		t.Errorf("Expected params[1] to be 100, got: %v", params[1])
	}
}

// TestError_DictionaryLookup tests:
// - Error looks up short and long from dictionary
func TestError_DictionaryLookup(t *testing.T) {
	err := errors.Error(errors.CodeMin, context.Background(), 10)

	// Should use dictionary values
	if err.ShortError() != "below minimum" {
		t.Errorf("Expected short error from dict 'below minimum', got: %s", err.ShortError())
	}

	if err.Error() != "must be at least 10" {
		t.Errorf("Expected long error from dict 'must be at least 10', got: %s", err.Error())
	}
}

// TestErrorf_WithErrorConfig_OverridesCode tests:
// - ErrorConfig in context overrides error code
func TestErrorf_WithErrorConfig_OverridesCode(t *testing.T) {
	customCode := errors.CodeForbidden
	config := &errors.ErrorConfig{
		Code: &customCode,
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	err := errors.Errorf(errors.CodeMin, ctx, "short", "long")

	if err.Code() != errors.CodeForbidden {
		t.Errorf("Expected code to be overridden to %s, got: %s", errors.CodeForbidden, err.Code())
	}
}

// TestErrorf_WithErrorConfig_OverridesShort tests:
// - ErrorConfig in context overrides short message
func TestErrorf_WithErrorConfig_OverridesShort(t *testing.T) {
	config := &errors.ErrorConfig{
		Short: "custom short",
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	err := errors.Errorf(errors.CodeMin, ctx, "original short", "long")

	if err.ShortError() != "custom short" {
		t.Errorf("Expected short to be overridden to 'custom short', got: %s", err.ShortError())
	}
}

// TestErrorf_WithErrorConfig_OverridesLong tests:
// - ErrorConfig in context overrides long message
func TestErrorf_WithErrorConfig_OverridesLong(t *testing.T) {
	config := &errors.ErrorConfig{
		Long: "custom long with %d",
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	err := errors.Errorf(errors.CodeMin, ctx, "short", "original long with %d", 42)

	if err.Error() != "custom long with 42" {
		t.Errorf("Expected long to be overridden to 'custom long with 42', got: %s", err.Error())
	}
}

// TestErrorf_WithErrorConfig_SetsDocsURI tests:
// - ErrorConfig in context sets DocsURI
func TestErrorf_WithErrorConfig_SetsDocsURI(t *testing.T) {
	config := &errors.ErrorConfig{
		DocsURI: "https://example.com/docs/error",
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	err := errors.Errorf(errors.CodeMin, ctx, "short", "long")

	if err.DocsURI() != "https://example.com/docs/error" {
		t.Errorf("Expected DocsURI to be 'https://example.com/docs/error', got: %s", err.DocsURI())
	}
}

// TestErrorf_WithErrorConfig_SetsTraceURI tests:
// - ErrorConfig in context sets TraceURI
func TestErrorf_WithErrorConfig_SetsTraceURI(t *testing.T) {
	config := &errors.ErrorConfig{
		TraceURI: "https://example.com/trace/12345",
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	err := errors.Errorf(errors.CodeMin, ctx, "short", "long")

	if err.TraceURI() != "https://example.com/trace/12345" {
		t.Errorf("Expected TraceURI to be 'https://example.com/trace/12345', got: %s", err.TraceURI())
	}
}

// TestErrorf_WithErrorConfig_SetsMeta tests:
// - ErrorConfig in context sets metadata
func TestErrorf_WithErrorConfig_SetsMeta(t *testing.T) {
	config := &errors.ErrorConfig{
		Meta: map[string]any{
			"field": "username",
			"limit": 100,
		},
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	err := errors.Errorf(errors.CodeMin, ctx, "short", "long")

	meta := err.Meta()
	if meta == nil {
		t.Fatal("Expected meta to be set")
	}

	if meta["field"] != "username" {
		t.Errorf("Expected meta['field'] to be 'username', got: %v", meta["field"])
	}

	if meta["limit"] != 100 {
		t.Errorf("Expected meta['limit'] to be 100, got: %v", meta["limit"])
	}
}

// TestErrorf_WithErrorConfig_CallsCallback tests:
// - ErrorConfig callback is called and can modify the error
func TestErrorf_WithErrorConfig_CallsCallback(t *testing.T) {
	callbackCalled := false
	config := &errors.ErrorConfig{
		Callback: func(ctx context.Context, err errors.ValidationError) errors.ValidationError {
			callbackCalled = true
			// Return a new error by calling Errorf again with modified long message
			return errors.Errorf(err.Code(), ctx, err.ShortError(), "modified by callback")
		},
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	// Clear the callback to avoid infinite recursion when creating the modified error
	innerConfig := &errors.ErrorConfig{
		Callback: nil,
	}
	config.Callback = func(outerCtx context.Context, err errors.ValidationError) errors.ValidationError {
		callbackCalled = true
		innerCtx := errors.WithErrorConfig(outerCtx, innerConfig)
		return errors.Errorf(err.Code(), innerCtx, err.ShortError(), "modified by callback")
	}

	err := errors.Errorf(errors.CodeMin, ctx, "short", "original long")

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}

	if err.Error() != "modified by callback" {
		t.Errorf("Expected error to be modified by callback, got: %s", err.Error())
	}
}

// TestErrorf_WithErrorConfig_CallbackHasAccessToParams tests:
// - ErrorConfig callback has access to Params()
func TestErrorf_WithErrorConfig_CallbackHasAccessToParams(t *testing.T) {
	var capturedParams []any
	config := &errors.ErrorConfig{
		Callback: func(ctx context.Context, err errors.ValidationError) errors.ValidationError {
			capturedParams = err.Params()
			return err
		},
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	errors.Errorf(errors.CodeMin, ctx, "short", "must be at least %d", 42)

	if len(capturedParams) != 1 {
		t.Fatalf("Expected callback to capture 1 param, got: %d", len(capturedParams))
	}

	if capturedParams[0] != 42 {
		t.Errorf("Expected captured param to be 42, got: %v", capturedParams[0])
	}
}

// TestErrorf_WithErrorConfig_CallbackHasAccessToAllFields tests:
// - ErrorConfig callback has access to all error fields
func TestErrorf_WithErrorConfig_CallbackHasAccessToAllFields(t *testing.T) {
	customCode := errors.CodeForbidden
	config := &errors.ErrorConfig{
		Code:     &customCode,
		Short:    "custom short",
		Long:     "custom long %d",
		DocsURI:  "https://docs.example.com",
		TraceURI: "https://trace.example.com/123",
		Meta:     map[string]any{"key": "value"},
		Callback: func(ctx context.Context, err errors.ValidationError) errors.ValidationError {
			// Verify all fields are accessible
			if err.Code() != errors.CodeForbidden {
				t.Errorf("Callback: Expected code %s, got: %s", errors.CodeForbidden, err.Code())
			}
			if err.ShortError() != "custom short" {
				t.Errorf("Callback: Expected short 'custom short', got: %s", err.ShortError())
			}
			if err.Error() != "custom long 42" {
				t.Errorf("Callback: Expected long 'custom long 42', got: %s", err.Error())
			}
			if err.DocsURI() != "https://docs.example.com" {
				t.Errorf("Callback: Expected DocsURI 'https://docs.example.com', got: %s", err.DocsURI())
			}
			if err.TraceURI() != "https://trace.example.com/123" {
				t.Errorf("Callback: Expected TraceURI 'https://trace.example.com/123', got: %s", err.TraceURI())
			}
			if err.Meta()["key"] != "value" {
				t.Errorf("Callback: Expected meta['key'] 'value', got: %v", err.Meta()["key"])
			}
			if len(err.Params()) != 1 || err.Params()[0] != 42 {
				t.Errorf("Callback: Expected params [42], got: %v", err.Params())
			}
			return err
		},
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	errors.Errorf(errors.CodeMin, ctx, "original short", "original long %d", 42)
}

// TestErrorf_WithErrorConfig_AllFieldsOverridden tests:
// - All ErrorConfig fields work together
func TestErrorf_WithErrorConfig_AllFieldsOverridden(t *testing.T) {
	customCode := errors.CodeForbidden
	config := &errors.ErrorConfig{
		Code:     &customCode,
		Short:    "custom short",
		Long:     "custom long",
		DocsURI:  "https://example.com/docs",
		TraceURI: "https://example.com/trace/abc",
		Meta: map[string]any{
			"key": "value",
		},
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	err := errors.Errorf(errors.CodeMin, ctx, "original short", "original long")

	if err.Code() != errors.CodeForbidden {
		t.Errorf("Expected code %s, got: %s", errors.CodeForbidden, err.Code())
	}

	if err.ShortError() != "custom short" {
		t.Errorf("Expected short 'custom short', got: %s", err.ShortError())
	}

	if err.Error() != "custom long" {
		t.Errorf("Expected long 'custom long', got: %s", err.Error())
	}

	if err.DocsURI() != "https://example.com/docs" {
		t.Errorf("Expected DocsURI 'https://example.com/docs', got: %s", err.DocsURI())
	}

	if err.TraceURI() != "https://example.com/trace/abc" {
		t.Errorf("Expected TraceURI 'https://example.com/trace/abc', got: %s", err.TraceURI())
	}

	if err.Meta()["key"] != "value" {
		t.Errorf("Expected meta['key'] 'value', got: %v", err.Meta()["key"])
	}
}

// TestError_WithErrorConfig tests:
// - Error function also respects ErrorConfig
func TestError_WithErrorConfig(t *testing.T) {
	config := &errors.ErrorConfig{
		Short: "custom short from Error()",
	}
	ctx := errors.WithErrorConfig(context.Background(), config)

	err := errors.Error(errors.CodeMin, ctx, 10)

	if err.ShortError() != "custom short from Error()" {
		t.Errorf("Expected short to be overridden, got: %s", err.ShortError())
	}
}

// TestErrorConfigFromContext_NilContext tests:
// - ErrorConfigFromContext handles nil context
func TestErrorConfigFromContext_NilContext(t *testing.T) {
	//lint:ignore SA1012 Testing nil context handling
	config := errors.ErrorConfigFromContext(nil)
	if config != nil {
		t.Errorf("Expected nil config from nil context, got: %v", config)
	}
}

// TestErrorConfigFromContext_NoConfig tests:
// - ErrorConfigFromContext returns nil when no config set
func TestErrorConfigFromContext_NoConfig(t *testing.T) {
	config := errors.ErrorConfigFromContext(context.Background())
	if config != nil {
		t.Errorf("Expected nil config from context without config, got: %v", config)
	}
}

// TestWithErrorConfig_NilConfig tests:
// - WithErrorConfig with nil config returns same context
func TestWithErrorConfig_NilConfig(t *testing.T) {
	ctx := context.Background()
	newCtx := errors.WithErrorConfig(ctx, nil)

	if newCtx != ctx {
		t.Error("Expected same context when config is nil")
	}
}

// TestErrorf_AllFieldsViaConfig tests:
// - Errorf with ErrorConfig correctly sets all fields
func TestErrorf_AllFieldsViaConfig(t *testing.T) {
	// Create path via context
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")

	customCode := errors.CodeForbidden
	config := &errors.ErrorConfig{
		Code:     &customCode,
		Short:    "short",
		Long:     "long message with %d %d %d",
		DocsURI:  "https://docs.example.com",
		TraceURI: "https://trace.example.com",
		Meta:     map[string]any{"key": "value"},
	}
	ctx = errors.WithErrorConfig(ctx, config)

	err := errors.Errorf(errors.CodeMin, ctx, "orig short", "orig long", 1, 2, 3)

	if err.Code() != errors.CodeForbidden {
		t.Errorf("Expected code %s, got: %s", errors.CodeForbidden, err.Code())
	}
	if err.Path() != "/a/b" {
		t.Errorf("Expected path '/a/b', got: %s", err.Path())
	}
	if err.Error() != "long message with 1 2 3" {
		t.Errorf("Expected message 'long message with 1 2 3', got: %s", err.Error())
	}
	if err.ShortError() != "short" {
		t.Errorf("Expected short 'short', got: %s", err.ShortError())
	}
	if err.DocsURI() != "https://docs.example.com" {
		t.Errorf("Expected DocsURI 'https://docs.example.com', got: %s", err.DocsURI())
	}
	if err.TraceURI() != "https://trace.example.com" {
		t.Errorf("Expected TraceURI 'https://trace.example.com', got: %s", err.TraceURI())
	}
	if err.Meta()["key"] != "value" {
		t.Errorf("Expected meta['key'] 'value', got: %v", err.Meta()["key"])
	}
	if len(err.Params()) != 3 {
		t.Errorf("Expected 3 params, got: %d", len(err.Params()))
	}
}
