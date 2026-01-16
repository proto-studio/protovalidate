package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestForPathAs_WithDefaultSerializer tests:
// - ForPathAs method works with default serializer
func TestForPathAs_WithDefaultSerializer(t *testing.T) {
	ctx1 := rulecontext.WithPathString(context.Background(), "a")
	ctx1 = rulecontext.WithPathString(ctx1, "b")
	err1 := errors.Errorf(errors.CodeMin, ctx1, "short", "message1")

	ctx2 := rulecontext.WithPathString(context.Background(), "a")
	ctx2 = rulecontext.WithPathString(ctx2, "c")
	err2 := errors.Errorf(errors.CodeMin, ctx2, "short", "message2")

	collection := errors.Collection(err1, err2)

	serializer := errors.DefaultPathSerializer{}
	filtered := collection.ForPathAs("/a/b", serializer)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 error, got: %d", len(filtered))
	}

	if filtered[0].Error() != "message1" {
		t.Errorf("Expected error message 'message1', got: '%s'", filtered[0].Error())
	}
}

// TestForPathAs_WithJSONPointerSerializer tests:
// - ForPathAs method works with JSON Pointer serializer
func TestForPathAs_WithJSONPointerSerializer(t *testing.T) {
	ctx1 := rulecontext.WithPathString(context.Background(), "a")
	ctx1 = rulecontext.WithPathString(ctx1, "b")
	err1 := errors.Errorf(errors.CodeMin, ctx1, "short", "message1")

	ctx2 := rulecontext.WithPathString(context.Background(), "a")
	ctx2 = rulecontext.WithPathString(ctx2, "c")
	err2 := errors.Errorf(errors.CodeMin, ctx2, "short", "message2")

	collection := errors.Collection(err1, err2)

	serializer := errors.JSONPointerSerializer{}
	filtered := collection.ForPathAs("/a/b", serializer)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 error, got: %d", len(filtered))
	}
}

// TestForPathAs_WithJSONPathSerializer tests:
// - ForPathAs method works with JSONPath serializer
func TestForPathAs_WithJSONPathSerializer(t *testing.T) {
	ctx1 := rulecontext.WithPathString(context.Background(), "a")
	ctx1 = rulecontext.WithPathString(ctx1, "b")
	err1 := errors.Errorf(errors.CodeMin, ctx1, "short", "message1")

	ctx2 := rulecontext.WithPathString(context.Background(), "a")
	ctx2 = rulecontext.WithPathString(ctx2, "c")
	err2 := errors.Errorf(errors.CodeMin, ctx2, "short", "message2")

	collection := errors.Collection(err1, err2)

	serializer := errors.JSONPathSerializer{}
	filtered := collection.ForPathAs("$.a.b", serializer)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 error, got: %d", len(filtered))
	}
}

// TestForPathAs_WithDotNotationSerializer tests:
// - ForPathAs method works with dot notation serializer
func TestForPathAs_WithDotNotationSerializer(t *testing.T) {
	ctx1 := rulecontext.WithPathString(context.Background(), "a")
	ctx1 = rulecontext.WithPathString(ctx1, "b")
	err1 := errors.Errorf(errors.CodeMin, ctx1, "short", "message1")

	ctx2 := rulecontext.WithPathString(context.Background(), "a")
	ctx2 = rulecontext.WithPathString(ctx2, "c")
	err2 := errors.Errorf(errors.CodeMin, ctx2, "short", "message2")

	collection := errors.Collection(err1, err2)

	serializer := errors.DotNotationSerializer{}
	filtered := collection.ForPathAs("a.b", serializer)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 error, got: %d", len(filtered))
	}
}

// TestForPathAs_NoMatches tests:
// - ForPathAs returns empty collection when no matches
func TestForPathAs_NoMatches(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

	collection := errors.Collection(err)

	serializer := errors.DefaultPathSerializer{}
	filtered := collection.ForPathAs("/nonexistent", serializer)

	if filtered != nil && len(filtered) != 0 {
		t.Errorf("Expected empty collection, got: %d errors", len(filtered))
	}
}

// TestForPathAs_EmptyCollection tests:
// - ForPathAs handles empty collection
func TestForPathAs_EmptyCollection(t *testing.T) {
	collection := errors.Collection()

	serializer := errors.DefaultPathSerializer{}
	filtered := collection.ForPathAs("/a/b", serializer)

	if filtered != nil && len(filtered) != 0 {
		t.Errorf("Expected nil or empty collection, got: %d errors", len(filtered))
	}
}
