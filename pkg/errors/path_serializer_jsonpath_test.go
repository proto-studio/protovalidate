package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestJSONPathSerializer_StringSegments tests:
// - JSONPath serializer with string segments
func TestJSONPathSerializer_StringSegments(t *testing.T) {
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

// TestJSONPathSerializer_IndexSegments tests:
// - JSONPath serializer with index segments
func TestJSONPathSerializer_IndexSegments(t *testing.T) {
	ctx := rulecontext.WithPathIndex(context.Background(), 0)
	ctx = rulecontext.WithPathIndex(ctx, 1)
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPathSerializer{}
	path := err.PathAs(serializer)

	expected := "$[0][1]"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestJSONPathSerializer_MixedSegments tests:
// - JSONPath serializer with mixed string and index segments
func TestJSONPathSerializer_MixedSegments(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	ctx = rulecontext.WithPathIndex(ctx, 0)
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPathSerializer{}
	path := err.PathAs(serializer)

	expected := "$.a.b[0]"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestJSONPathSerializer_ComplexPath tests:
// - JSONPath serializer with complex path
func TestJSONPathSerializer_ComplexPath(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "users")
	ctx = rulecontext.WithPathIndex(ctx, 0)
	ctx = rulecontext.WithPathString(ctx, "name")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPathSerializer{}
	path := err.PathAs(serializer)

	expected := "$.users[0].name"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestJSONPathSerializer_SpecialCharacters tests:
// - JSONPath serializer handles special characters
func TestJSONPathSerializer_SpecialCharacters(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "field.name")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPathSerializer{}
	path := err.PathAs(serializer)

	// Should use bracket notation for fields with dots
	expected := "$['field.name']"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestJSONPathSerializer_EmptyPath tests:
// - JSONPath serializer with empty path
func TestJSONPathSerializer_EmptyPath(t *testing.T) {
	err := errors.Errorf(errors.CodeMin, context.Background(), "short", "message")

	serializer := errors.JSONPathSerializer{}
	path := err.PathAs(serializer)

	expected := "$"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestJSONPathSerializer_SingleSegment tests:
// - JSONPath serializer with single segment
func TestJSONPathSerializer_SingleSegment(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "field")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.JSONPathSerializer{}
	path := err.PathAs(serializer)

	expected := "$.field"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}
