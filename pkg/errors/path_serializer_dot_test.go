package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestDotNotationSerializer_StringSegments tests:
// - Dot notation serializer with string segments
func TestDotNotationSerializer_StringSegments(t *testing.T) {
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

// TestDotNotationSerializer_IndexSegments tests:
// - Dot notation serializer with index segments
func TestDotNotationSerializer_IndexSegments(t *testing.T) {
	ctx := rulecontext.WithPathIndex(context.Background(), 0)
	ctx = rulecontext.WithPathIndex(ctx, 1)
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DotNotationSerializer{}
	path := err.PathAs(serializer)

	expected := "[0][1]"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestDotNotationSerializer_MixedSegments tests:
// - Dot notation serializer with mixed string and index segments
func TestDotNotationSerializer_MixedSegments(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	ctx = rulecontext.WithPathIndex(ctx, 0)
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DotNotationSerializer{}
	path := err.PathAs(serializer)

	expected := "a.b[0]"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestDotNotationSerializer_ComplexPath tests:
// - Dot notation serializer with complex path
func TestDotNotationSerializer_ComplexPath(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "users")
	ctx = rulecontext.WithPathIndex(ctx, 0)
	ctx = rulecontext.WithPathString(ctx, "name")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DotNotationSerializer{}
	path := err.PathAs(serializer)

	expected := "users[0].name"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestDotNotationSerializer_SpecialCharacters tests:
// - Dot notation serializer handles special characters
func TestDotNotationSerializer_SpecialCharacters(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "field.name")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DotNotationSerializer{}
	path := err.PathAs(serializer)

	// Should use bracket notation for fields with dots
	expected := "['field.name']"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestDotNotationSerializer_EmptyPath tests:
// - Dot notation serializer with empty path
func TestDotNotationSerializer_EmptyPath(t *testing.T) {
	err := errors.Errorf(errors.CodeMin, context.Background(), "short", "message")

	serializer := errors.DotNotationSerializer{}
	path := err.PathAs(serializer)

	if path != "" {
		t.Errorf("Expected empty path, got: '%s'", path)
	}
}

// TestDotNotationSerializer_SingleSegment tests:
// - Dot notation serializer with single segment
func TestDotNotationSerializer_SingleSegment(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "field")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DotNotationSerializer{}
	path := err.PathAs(serializer)

	expected := "field"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}
