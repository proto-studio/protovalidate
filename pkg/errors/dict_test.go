package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
)

// TestErrorTypeString tests:
// - ErrorType.String() returns correct values for all types
func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errType  errors.ErrorType
		expected string
	}{
		{errors.ErrorTypeValidation, "validation"},
		{errors.ErrorTypePermission, "permission"},
		{errors.ErrorTypeInternal, "internal"},
		{errors.ErrorType(99), "unknown"}, // Test unknown value
	}

	for _, tt := range tests {
		if got := tt.errType.String(); got != tt.expected {
			t.Errorf("ErrorType(%d).String() = %q, want %q", tt.errType, got, tt.expected)
		}
	}
}

// TestNewDict tests:
// - NewDict creates a new dictionary
// - New dict inherits from default dict
func TestNewDict(t *testing.T) {
	dict := errors.NewDict()

	// Should inherit default values
	if shortDesc := dict.ShortError(errors.CodeMin); shortDesc == "" {
		t.Error("Expected inherited short description, got empty string")
	}
}

// TestWithCode tests:
// - WithCode creates a new dict with overridden entry
// - Original entry is preserved in parent
func TestWithCode(t *testing.T) {
	dict := errors.DefaultDict()

	customEntry := errors.ErrorEntry{
		Type:         errors.ErrorTypePermission,
		ShortError:   "custom short",
		ErrorPattern: "custom long",
	}

	newDict := dict.WithCode(errors.CodeMin, customEntry)

	// New dict should have custom entry
	if got := newDict.ShortError(errors.CodeMin); got != "custom short" {
		t.Errorf("newDict.ShortError(CodeMin) = %q, want %q", got, "custom short")
	}

	if got := newDict.ErrorPattern(errors.CodeMin); got != "custom long" {
		t.Errorf("newDict.ErrorPattern(CodeMin) = %q, want %q", got, "custom long")
	}

	if got := newDict.ErrorType(errors.CodeMin); got != errors.ErrorTypePermission {
		t.Errorf("newDict.ErrorType(CodeMin) = %v, want %v", got, errors.ErrorTypePermission)
	}

	// Original dict should be unchanged
	if got := dict.ShortError(errors.CodeMin); got == "custom short" {
		t.Error("Original dict should not be modified")
	}
}

// TestWithDict tests:
// - WithDict adds dict to context
// - Dict retrieves dict from context
func TestWithDict(t *testing.T) {
	customDict := errors.NewDict().WithCode(errors.CodeMin, errors.ErrorEntry{
		Type:         errors.ErrorTypePermission,
		ShortError:   "context test",
		ErrorPattern: "context test long",
	})

	ctx := errors.WithDict(context.Background(), customDict)

	retrievedDict := errors.Dict(ctx)
	if retrievedDict == nil {
		t.Fatal("Expected dict from context, got nil")
	}

	if got := retrievedDict.ShortError(errors.CodeMin); got != "context test" {
		t.Errorf("Dict from context ShortError = %q, want %q", got, "context test")
	}
}

// TestDictReturnsDefaultWhenNotSet tests:
// - Dict returns DefaultDict when no dict in context
func TestDictReturnsDefaultWhenNotSet(t *testing.T) {
	ctx := context.Background()
	dict := errors.Dict(ctx)

	if dict != errors.DefaultDict() {
		t.Error("Expected DefaultDict when no dict in context")
	}
}

// TestDictReturnsDefaultWhenContextNil tests:
// - Dict returns DefaultDict when context is nil
func TestDictReturnsDefaultWhenContextNil(t *testing.T) {
	//lint:ignore SA1012 intentionally testing nil context behavior
	dict := errors.Dict(nil)

	if dict != errors.DefaultDict() {
		t.Error("Expected DefaultDict when context is nil")
	}
}

// TestWithDictPanicsOnNil tests:
// - WithDict panics when dict is nil
func TestWithDictPanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when dict is nil")
		}
	}()

	errors.WithDict(context.Background(), nil)
}

// TestDictEntryForUnknownCode tests:
// - Entry returns default unknown entry for unknown codes
func TestDictEntryForUnknownCode(t *testing.T) {
	dict := errors.DefaultDict()

	// Use a code that doesn't exist
	entry := dict.Entry("NONEXISTENT_CODE")

	if entry.Type != errors.ErrorTypeInternal {
		t.Errorf("Expected unknown code to have ErrorTypeInternal, got %v", entry.Type)
	}
}

// TestErrorConfigFluentMethods tests:
// - Fluent methods correctly set values
// - Chaining works correctly
// - Child values take precedence over parent
// - Meta maps are merged
func TestErrorConfigFluentMethods(t *testing.T) {
	code := errors.ErrorCode("CODE1")

	// Start with nil and chain fluent methods
	config := (*errors.ErrorConfig)(nil).
		WithErrorMessage("short", "long").
		WithDocsURI("https://docs.example.com").
		WithTraceURI("https://trace.example.com").
		WithCode(code).
		WithMeta("key1", "value1").
		WithMeta("key2", "value2")

	if config.Short != "short" {
		t.Errorf("Short = %q, want %q", config.Short, "short")
	}
	if config.Long != "long" {
		t.Errorf("Long = %q, want %q", config.Long, "long")
	}
	if config.DocsURI != "https://docs.example.com" {
		t.Errorf("DocsURI = %q, want %q", config.DocsURI, "https://docs.example.com")
	}
	if config.TraceURI != "https://trace.example.com" {
		t.Errorf("TraceURI = %q, want %q", config.TraceURI, "https://trace.example.com")
	}
	if config.Code == nil || *config.Code != code {
		t.Error("Code should be set")
	}
	if config.Meta["key1"] != "value1" || config.Meta["key2"] != "value2" {
		t.Error("Meta should contain both keys")
	}
}

// TestErrorConfigFluentMethodsChaining tests:
// - Values are preserved when chaining
// - Later values override earlier ones for same field
func TestErrorConfigFluentMethodsChaining(t *testing.T) {
	parent := &errors.ErrorConfig{
		Short:    "parent short",
		Long:     "parent long",
		DocsURI:  "https://parent.com/docs",
		TraceURI: "https://parent.com/trace",
	}

	// Child overrides some values
	merged := parent.WithErrorMessage("child short", "").WithDocsURI("https://child.com/docs")

	// Child takes precedence for Short
	if merged.Short != "child short" {
		t.Errorf("Short = %q, want %q", merged.Short, "child short")
	}

	// Parent used when child is empty
	if merged.Long != "parent long" {
		t.Errorf("Long = %q, want %q", merged.Long, "parent long")
	}

	// Child takes precedence for DocsURI
	if merged.DocsURI != "https://child.com/docs" {
		t.Errorf("DocsURI = %q, want %q", merged.DocsURI, "https://child.com/docs")
	}

	// Parent used when child is empty
	if merged.TraceURI != "https://parent.com/trace" {
		t.Errorf("TraceURI = %q, want %q", merged.TraceURI, "https://parent.com/trace")
	}
}

// TestErrorConfigFluentMethodsNilReceiver tests:
// - All fluent methods work on nil receiver
func TestErrorConfigFluentMethodsNilReceiver(t *testing.T) {
	var config *errors.ErrorConfig

	// All methods should work on nil receiver
	if result := config.WithErrorMessage("short", "long"); result.Short != "short" {
		t.Error("WithErrorMessage should work on nil receiver")
	}
	if result := config.WithDocsURI("docs"); result.DocsURI != "docs" {
		t.Error("WithDocsURI should work on nil receiver")
	}
	if result := config.WithTraceURI("trace"); result.TraceURI != "trace" {
		t.Error("WithTraceURI should work on nil receiver")
	}
	code := errors.ErrorCode("CODE")
	if result := config.WithCode(code); result.Code == nil || *result.Code != code {
		t.Error("WithCode should work on nil receiver")
	}
	if result := config.WithMeta("key", "value"); result.Meta["key"] != "value" {
		t.Error("WithMeta should work on nil receiver")
	}
}

// TestErrorConfigWithCallback tests:
// - Callback is set correctly
// - Later callback replaces earlier one
func TestErrorConfigWithCallback(t *testing.T) {
	called := false

	callback := func(ctx context.Context, err errors.ValidationError) errors.ValidationError {
		called = true
		return err
	}

	config := (*errors.ErrorConfig)(nil).WithCallback(callback)

	if config.Callback == nil {
		t.Error("Callback should be set")
	}

	config.Callback(context.Background(), nil)

	if !called {
		t.Error("Callback should have been called")
	}
}
