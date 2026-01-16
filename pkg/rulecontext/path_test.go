package rulecontext_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

func fullPathHelper(t testing.TB, ctx context.Context, expected string) {
	t.Helper()

	if ctx == nil {
		t.Fatal("Expected path to not be nil")
	}

	path := rulecontext.Path(ctx)
	if path == nil {
		t.Fatal("Expected path to not be nil")
	}

	// Use the default serializer to get the full path
	serializer := errors.DefaultPathSerializer{}
	segments := extractPathSegmentsForTest(path)
	fullpath := serializer.Serialize(segments)
	if fullpath != expected {
		t.Errorf("Expected full path to be '%s', got: '%s'", expected, fullpath)
	}
}

// extractPathSegmentsForTest extracts all segments from a PathSegment into an array,
// ordered from root to leaf (top to bottom). This is a test helper that mirrors
// the private function in the errors package.
func extractPathSegmentsForTest(segment rulecontext.PathSegment) []rulecontext.PathSegment {
	if segment == nil {
		return nil
	}

	// First, collect all segments by traversing up to the root
	var segments []rulecontext.PathSegment
	current := segment
	for current != nil {
		segments = append(segments, current)
		current = current.Parent()
	}

	// Reverse to get root-to-leaf order
	for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
		segments[i], segments[j] = segments[j], segments[i]
	}

	return segments
}

// TestPathNil tests:
// - Returns nil when context is nil
// - Returns nil when no path is set in context
func TestPathNil(t *testing.T) {
	//lint:ignore SA1012 Testing nil context handling
	if path := rulecontext.Path(nil); path != nil {
		t.Errorf("Expected path to be nil, got: %v", path)
	}

	ctx := context.Background()

	if path := rulecontext.Path(ctx); path != nil {
		t.Errorf("Expected path to be nil, got: %v", path)
	}
}

// TestParentString tests:
// - Parent path segments are correctly retrieved for string paths
func TestParentString(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "patha")
	ctx = rulecontext.WithPathString(ctx, "pathb")

	firstPath := rulecontext.Path(ctx)
	if firstPath == nil {
		t.Fatalf("Expected first path to not be nil")
	}
	if str := firstPath.String(); str != "pathb" {
		t.Errorf("Expected first path be 'pathb', got: '%s'", str)
	}

	secondPath := firstPath.Parent()
	if secondPath == nil {
		t.Fatalf("Expected second path to not be nil")
	}
	if str := secondPath.String(); str != "patha" {
		t.Errorf("Expected second path be 'patha', got: '%s'", str)
	}

	thirdPath := firstPath.Parent()
	if thirdPath == nil {
		t.Fatalf("Expected third path to be nil, got: %v", thirdPath)
	}
}

// TestParentIndex tests:
// - Parent path segments are correctly retrieved for index paths
func TestParentIndex(t *testing.T) {
	ctx := rulecontext.WithPathIndex(context.Background(), 1)
	ctx = rulecontext.WithPathIndex(ctx, 2)

	firstPath := rulecontext.Path(ctx)
	if firstPath == nil {
		t.Fatalf("Expected first path to not be nil")
	}
	if str := firstPath.String(); str != "2" {
		t.Errorf("Expected first path be '2', got: '%s'", str)
	}

	secondPath := firstPath.Parent()
	if secondPath == nil {
		t.Fatalf("Expected second path to not be nil")
	}
	if str := secondPath.String(); str != "1" {
		t.Errorf("Expected second path be '1', got: '%s'", str)
	}

	thirdPath := firstPath.Parent()
	if thirdPath == nil {
		t.Fatalf("Expected third path to be nil, got: %v", thirdPath)
	}
}

// TestWithPathCombined tests:
// - Combined string and index paths work correctly
func TestWithPathCombined(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "patha")
	fullPathHelper(t, ctx, "/patha")

	ctx = rulecontext.WithPathString(ctx, "pathb")
	fullPathHelper(t, ctx, "/patha/pathb")

	ctx = rulecontext.WithPathIndex(ctx, 1)
	fullPathHelper(t, ctx, "/patha/pathb/1")

	ctx = rulecontext.WithPathIndex(ctx, 2)
	fullPathHelper(t, ctx, "/patha/pathb/1/2")

	ctx = rulecontext.WithPathString(ctx, "pathc")
	fullPathHelper(t, ctx, "/patha/pathb/1/2/pathc")
}

// TestPathIndexFullStringNilParent tests:
// - PathSegmentIndex with nil parent serializes to just the index (no leading slash)
func TestPathIndexFullStringNilParent(t *testing.T) {
	// Create a context with only an index path (no parent string path)
	ctx := rulecontext.WithPathIndex(context.Background(), 5)
	fullPathHelper(t, ctx, "5")
}
