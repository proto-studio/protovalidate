package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestPathAs_WithDefaultSerializer tests:
// - PathAs method works with default serializer
func TestPathAs_WithDefaultSerializer(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DefaultPathSerializer{}
	path := err.PathAs(serializer)

	if path == "" {
		t.Error("Expected non-empty path")
	}
}

// TestPathAs_WithJSONPointerSerializer tests:
// - PathAs method works with JSON Pointer serializer
func TestPathAs_WithJSONPointerSerializer(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPointerSerializer{}
	path := err.PathAs(serializer)

	expected := "/a/b"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestPathAs_WithJSONPathSerializer tests:
// - PathAs method works with JSONPath serializer
func TestPathAs_WithJSONPathSerializer(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPathSerializer{}
	path := err.PathAs(serializer)

	expected := "$.a.b"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestPathAs_WithDotNotationSerializer tests:
// - PathAs method works with dot notation serializer
func TestPathAs_WithDotNotationSerializer(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DotNotationSerializer{}
	path := err.PathAs(serializer)

	expected := "a.b"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestPathAs_EmptyPath tests:
// - PathAs returns empty string for errors without path
func TestPathAs_EmptyPath(t *testing.T) {
	err := errors.Errorf(errors.CodeMin, context.Background(), "short", "message")

	serializer := errors.DefaultPathSerializer{}
	path := err.PathAs(serializer)

	if path != "" {
		t.Errorf("Expected empty path, got: '%s'", path)
	}
}
