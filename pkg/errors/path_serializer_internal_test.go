package errors

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestExtractPathSegments_SingleSegment tests:
// - extractPathSegments with a single segment (tests reversal loop with one element)
func TestExtractPathSegments_SingleSegment(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "field")
	segment := rulecontext.Path(ctx)
	
	segments := extractPathSegments(segment)
	
	if len(segments) != 1 {
		t.Errorf("Expected 1 segment, got: %d", len(segments))
	}
	if segments[0].String() != "field" {
		t.Errorf("Expected segment to be 'field', got: '%s'", segments[0].String())
	}
}

// TestExtractPathSegments_MultipleSegments tests:
// - extractPathSegments with multiple segments (tests reversal)
func TestExtractPathSegments_MultipleSegments(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	ctx = rulecontext.WithPathString(ctx, "c")
	segment := rulecontext.Path(ctx)
	
	segments := extractPathSegments(segment)
	
	if len(segments) != 3 {
		t.Errorf("Expected 3 segments, got: %d", len(segments))
	}
	if segments[0].String() != "a" {
		t.Errorf("Expected first segment to be 'a', got: '%s'", segments[0].String())
	}
	if segments[1].String() != "b" {
		t.Errorf("Expected second segment to be 'b', got: '%s'", segments[1].String())
	}
	if segments[2].String() != "c" {
		t.Errorf("Expected third segment to be 'c', got: '%s'", segments[2].String())
	}
}

// TestExtractPathSegments_MixedSegments tests:
// - extractPathSegments with mixed string and index segments
func TestExtractPathSegments_MixedSegments(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathIndex(ctx, 0)
	ctx = rulecontext.WithPathString(ctx, "b")
	segment := rulecontext.Path(ctx)
	
	segments := extractPathSegments(segment)
	
	if len(segments) != 3 {
		t.Errorf("Expected 3 segments, got: %d", len(segments))
	}
	if segments[0].String() != "a" {
		t.Errorf("Expected first segment to be 'a', got: '%s'", segments[0].String())
	}
	if segments[1].String() != "0" {
		t.Errorf("Expected second segment to be '0', got: '%s'", segments[1].String())
	}
	if segments[2].String() != "b" {
		t.Errorf("Expected third segment to be 'b', got: '%s'", segments[2].String())
	}
}
