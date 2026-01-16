package rulecontext_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestPathJSONPointerEscaping tests that path segments with special characters
// are properly escaped according to RFC 6901 when serialized as JSON Pointer.
//
// RFC 6901 requires:
// - `~` must be escaped as `~0`
// - `/` must be escaped as `~1`
func TestPathJSONPointerEscaping(t *testing.T) {
	tests := []struct {
		name     string
		segments []string
		expected string
	}{
		{
			name:     "segment with tilde",
			segments: []string{"a~b"},
			expected: "/a~0b",
		},
		{
			name:     "segment with slash",
			segments: []string{"a/b"},
			expected: "/a~1b",
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
			name:     "multiple segments with special chars",
			segments: []string{"a~b", "c/d", "e~f"},
			expected: "/a~0b/c~1d/e~0f",
		},
		{
			name:     "normal segment",
			segments: []string{"normal"},
			expected: "/normal",
		},
		{
			name:     "empty segment",
			segments: []string{""},
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			for _, seg := range tt.segments {
				ctx = rulecontext.WithPathString(ctx, seg)
			}

			path := rulecontext.Path(ctx)
			if path == nil {
				t.Fatal("Expected path to not be nil")
			}

			// Use JSON Pointer serializer to get the properly escaped path
			serializer := errors.JSONPointerSerializer{}
			segments := extractPathSegmentsForTest(path)
			actual := serializer.Serialize(segments)

			if actual != tt.expected {
				t.Errorf("Expected JSON Pointer path '%s', got '%s'", tt.expected, actual)
			}
		})
	}
}
