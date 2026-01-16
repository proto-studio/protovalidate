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
// - JSON Pointer serializer escapes special characters according to RFC 6901
func TestJSONPointerSerializer_Escaping(t *testing.T) {
	tests := []struct {
		name     string
		segments []string
		expected string
	}{
		{
			name:     "segment with slash",
			segments: []string{"a/b"},
			expected: "/a~1b",
		},
		{
			name:     "segment with tilde",
			segments: []string{"c~d"},
			expected: "/c~0d",
		},
		{
			name:     "segment with both tilde and slash",
			segments: []string{"a~/b"},
			expected: "/a~0~1b",
		},
		{
			name:     "segment starting with tilde",
			segments: []string{"~something"},
			expected: "/~0something",
		},
		{
			name:     "segment starting with slash",
			segments: []string{"/leading"},
			expected: "/~1leading",
		},
		{
			name:     "segment ending with tilde",
			segments: []string{"ending~"},
			expected: "/ending~0",
		},
		{
			name:     "segment ending with slash",
			segments: []string{"ending/"},
			expected: "/ending~1",
		},
		{
			name:     "segment with tilde followed by 0 (literal ~0)",
			segments: []string{"a~0b"},
			expected: "/a~00b",
		},
		{
			name:     "segment with tilde followed by 1 (literal ~1)",
			segments: []string{"a~1b"},
			expected: "/a~01b",
		},
		{
			name:     "multiple segments with special chars",
			segments: []string{"a~b", "c/d", "e~f"},
			expected: "/a~0b/c~1d/e~0f",
		},
		{
			name:     "empty segment",
			segments: []string{""},
			expected: "/",
		},
		{
			name:     "segment with only tilde",
			segments: []string{"~"},
			expected: "/~0",
		},
		{
			name:     "segment with only slash",
			segments: []string{"/"},
			expected: "/~1",
		},
		{
			name:     "segment with multiple tildes",
			segments: []string{"a~~b"},
			expected: "/a~0~0b",
		},
		{
			name:     "segment with multiple slashes",
			segments: []string{"a//b"},
			expected: "/a~1~1b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			for _, seg := range tt.segments {
				ctx = rulecontext.WithPathString(ctx, seg)
			}
			err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

			serializer := errors.JSONPointerSerializer{}
			path := err.PathAs(serializer)

			if path != tt.expected {
				t.Errorf("Expected path to be '%s', got: '%s'", tt.expected, path)
			}
		})
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
