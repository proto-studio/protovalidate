package errors_test

import (
	"context"
	"strings"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestCollectionWrapper tests:
// - Collection correctly wraps multiple errors
func TestCollectionWrapper(t *testing.T) {
	ctx := context.Background()

	err := errors.Collection(
		errors.Errorf(errors.CodeMax, ctx, "error1"),
		errors.Errorf(errors.CodeMax, ctx, "error2"),
	)

	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if s := len(err); s != 2 {
		t.Errorf("Expected error to have size %d, got %d", 2, s)
	}
}

// Legacy method. Will be removed in v1.0.0
func TestCollectionAll(t *testing.T) {
	ctx := context.Background()

	err1 := errors.Errorf(errors.CodeMax, ctx, "error1")
	err2 := errors.Errorf(errors.CodeMax, ctx, "error2")

	colErr := errors.Collection(
		err1,
		err2,
	)

	if colErr == nil {
		t.Fatal("Expected error to not be nil")
	} else if s := colErr.Size(); s != 2 {
		t.Fatalf("Expected error to have size %d, got %d", 2, s)
	}

	all := colErr.All()

	if l := len(all); l != 2 {
		t.Fatalf("Expected error to have length %d, got %d", 2, l)
	}

	if !((all[0] == err1 && all[1] == err2) || (all[0] == err2 && all[1] == err1)) {
		t.Errorf("Expected both errors to be returned")
	}
}

// TestCollectionUnwrap tests:
// - Unwrap should return an array of errors.
func TestCollectionUnwrap(t *testing.T) {
	ctx := context.Background()

	err1 := errors.Errorf(errors.CodeMax, ctx, "error1")
	err2 := errors.Errorf(errors.CodeMax, ctx, "error2")

	colErr := errors.Collection(
		err1,
		err2,
	)

	if colErr == nil {
		t.Fatal("Expected error to not be nil")
	} else if s := colErr.Size(); s != 2 {
		t.Fatalf("Expected error to have size %d, got %d", 2, s)
	}

	all := colErr.Unwrap()

	if l := len(all); l != 2 {
		t.Fatalf("Expected error to have length %d, got %d", 2, l)
	}

	if !((all[0] == err1 && all[1] == err2) || (all[0] == err2 && all[1] == err1)) {
		t.Errorf("Expected both errors to be returned")
	}
}

// Legacy method. Will be removed in v1.0.0
func TestCollectionSize(t *testing.T) {
	ctx := context.Background()

	err1 := errors.Errorf(errors.CodeMax, ctx, "error1")
	err2 := errors.Errorf(errors.CodeMax, ctx, "error2")

	colErr := errors.Collection(err1)

	if s := colErr.Size(); s != 1 {
		t.Errorf("Expected size to be 1, got: %d", s)
	}

	colErr = append(colErr, err2)

	if s := colErr.Size(); s != 2 {
		t.Errorf("Expected size to be 2, got: %d", s)
	}
}

// TestCollectionFirst tests:
// - Returns the first error from a collection
// - Returns one of the errors when multiple errors exist
func TestCollectionFirst(t *testing.T) {
	ctx := context.Background()
	err1 := errors.NewCoercionError(ctx, "int", "float32")
	err2 := errors.NewCoercionError(ctx, "int", "float32")

	colErr := errors.Collection(
		err1,
		err2,
	)

	if colErr == nil {
		t.Fatal("Expected error to not be nil")
	} else if s := len(colErr); s != 2 {
		t.Fatalf("Expected error to have size %d, got %d", 2, s)
	}

	first := colErr.First()

	if first == nil {
		t.Errorf("Expected first to not be nil")
	} else if first != err1 && first != err2 {
		t.Errorf("Expected one of two errors to be returned")
	}
}

// TestCollectionFirstEmpty tests:
// - Returns nil when collection is empty
func TestCollectionFirstEmpty(t *testing.T) {
	col := errors.Collection()
	if first := col.First(); first != nil {
		t.Errorf("Expected first to be nil, got: %s", first)
	}
}

// TestCollectionFor tests:
// - Returns errors matching a specific path
// - Returns nil when no errors match the path
// - Correctly filters errors by path
func TestCollectionFor(t *testing.T) {
	ctx1 := rulecontext.WithPathString(context.Background(), "path1")
	err1 := errors.Errorf(errors.CodeMax, ctx1, "error1")

	ctx2 := rulecontext.WithPathString(context.Background(), "path2a")
	ctx2 = rulecontext.WithPathString(ctx2, "b")
	err2 := errors.Errorf(errors.CodeMax, ctx2, "error2")

	colErr := errors.Collection(
		err1,
		err2,
	)

	if colErr == nil {
		t.Fatal("Expected error to not be nil")
	} else if s := len(colErr); s != 2 {
		t.Fatalf("Expected error to have size %d, got %d", 2, s)
	}

	path1err := colErr.For("/path1")

	if path1err == nil {
		t.Errorf("Expected path1 error to not be nil")
	} else if s := len(path1err); s != 1 {
		t.Errorf("Expected a collection with 1 error, got: '%d'", s)
	} else if first := path1err.First(); first != err1 {
		t.Errorf("Expected '%s' to be returned, got: '%s'", err1, first)
	}

	path1err = colErr.For("/path1/b")

	if path1err != nil {
		t.Errorf("Expected error to be nil, got: %s", path1err)
	}

	path2err := colErr.For("/path2a/b")

	if path2err == nil {
		t.Errorf("Expected path2 error to not be nil")
	} else if s := len(path2err); s != 1 {
		t.Errorf("Expected a collection with 1 error, got: '%d'", s)
	} else if first := path2err.First(); first != err2 {
		t.Errorf("Expected '%s' to be returned, got: '%s'", err2, first)
	}
}

// TestCollectionForEmpty tests:
// - Returns nil when collection is empty
func TestCollectionForEmpty(t *testing.T) {
	col := errors.Collection()
	if first := col.For("a"); first != nil {
		t.Errorf("Expected first to be nil, got: %s", first)
	}
}

// TestCollectionMessage tests:
// - Error message is correctly formatted for single error
// - Error message includes count for multiple errors
func TestCollectionMessage(t *testing.T) {
	err := errors.Errorf(errors.CodeUnknown, context.Background(), "error123")

	col := errors.Collection(err)

	if msg := col.Error(); msg != "error123" {
		t.Errorf("Expected error message to be %s, got: %s", "error123", msg)
	}

	col = append(col, errors.Errorf(errors.CodeUnknown, context.Background(), "error123"))

	if msg := col.Error(); !strings.Contains(msg, "(and 1 more)") {
		t.Errorf("Expected error message to contain the string '(and 1 more)', got: %s", msg)
	}
}

// TestPanicCollectionMessageEmpty tests:
// - Panics when Error is called on an empty collection
func TestPanicCollectionMessageEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	_ = errors.Collection().Error()
}
