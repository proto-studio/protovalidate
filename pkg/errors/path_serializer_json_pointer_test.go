package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestJSONPointerSerializer_StringSegments tests:
// - JSON Pointer serializer with string segments
func TestJSONPointerSerializer_StringSegments(t *testing.T) {
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

// TestJSONPointerSerializer_IndexSegments tests:
// - JSON Pointer serializer with index segments
func TestJSONPointerSerializer_IndexSegments(t *testing.T) {
	ctx := rulecontext.WithPathIndex(context.Background(), 0)
	ctx = rulecontext.WithPathIndex(ctx, 1)
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPointerSerializer{}
	path := err.PathAs(serializer)

	expected := "/0/1"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestJSONPointerSerializer_MixedSegments tests:
// - JSON Pointer serializer with mixed string and index segments
func TestJSONPointerSerializer_MixedSegments(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	ctx = rulecontext.WithPathIndex(ctx, 0)
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPointerSerializer{}
	path := err.PathAs(serializer)

	expected := "/a/b/0"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestJSONPointerSerializer_Escaping tests:
// - JSON Pointer serializer escapes special characters
func TestJSONPointerSerializer_Escaping(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a/b")
	ctx = rulecontext.WithPathString(ctx, "c~d")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPointerSerializer{}
	path := err.PathAs(serializer)

	expected := "/a~1b/c~0d"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestJSONPointerSerializer_EmptyPath tests:
// - JSON Pointer serializer with empty path
func TestJSONPointerSerializer_EmptyPath(t *testing.T) {
	err := errors.Errorf(errors.CodeMin, context.Background(), "short", "message")

	serializer := errors.JSONPointerSerializer{}
	path := err.PathAs(serializer)

	if path != "" {
		t.Errorf("Expected empty path, got: '%s'", path)
	}
}

// TestJSONPointerSerializer_SingleSegment tests:
// - JSON Pointer serializer with single segment
func TestJSONPointerSerializer_SingleSegment(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "field")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPointerSerializer{}
	path := err.PathAs(serializer)

	expected := "/field"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}
