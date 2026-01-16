package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestDefaultPathSerializer_StringSegments tests:
// - Default serializer with string segments
func TestDefaultPathSerializer_StringSegments(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DefaultPathSerializer{}
	path := err.PathAs(serializer)

	expected := "/a/b"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestDefaultPathSerializer_IndexSegments tests:
// - Default serializer with index segments
func TestDefaultPathSerializer_IndexSegments(t *testing.T) {
	ctx := rulecontext.WithPathIndex(context.Background(), 0)
	ctx = rulecontext.WithPathIndex(ctx, 1)
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DefaultPathSerializer{}
	path := err.PathAs(serializer)

	expected := "0/1"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestDefaultPathSerializer_MixedSegments tests:
// - Default serializer with mixed string and index segments
func TestDefaultPathSerializer_MixedSegments(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	ctx = rulecontext.WithPathIndex(ctx, 0)
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DefaultPathSerializer{}
	path := err.PathAs(serializer)

	expected := "/a/b/0"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestDefaultPathSerializer_SingleIndex tests:
// - Default serializer with single index segment (no leading slash)
func TestDefaultPathSerializer_SingleIndex(t *testing.T) {
	ctx := rulecontext.WithPathIndex(context.Background(), 5)
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DefaultPathSerializer{}
	path := err.PathAs(serializer)

	expected := "5"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestDefaultPathSerializer_SingleString tests:
// - Default serializer with single string segment
func TestDefaultPathSerializer_SingleString(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "field")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DefaultPathSerializer{}
	path := err.PathAs(serializer)

	expected := "/field"
	if path != expected {
		t.Errorf("Expected path to be '%s', got: '%s'", expected, path)
	}
}

// TestDefaultPathSerializer_EmptyPath tests:
// - Default serializer with empty path
func TestDefaultPathSerializer_EmptyPath(t *testing.T) {
	err := errors.Errorf(errors.CodeMin, context.Background(), "short", "message")

	serializer := errors.DefaultPathSerializer{}
	path := err.PathAs(serializer)

	if path != "" {
		t.Errorf("Expected empty path, got: '%s'", path)
	}
}

// TestDefaultPathSerializer_MatchesPath tests:
// - Default serializer should match the default Path() behavior
func TestDefaultPathSerializer_MatchesPath(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	serializer := errors.DefaultPathSerializer{}
	pathAs := err.PathAs(serializer)
	path := err.Path()

	if pathAs != path {
		t.Errorf("Expected PathAs to match Path, got PathAs: '%s', Path: '%s'", pathAs, path)
	}
}
