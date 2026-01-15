package errors_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	pkgerrors "proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestCollectionWrapper tests:
// - Collection correctly wraps multiple errors
func TestCollectionWrapper(t *testing.T) {
	ctx := context.Background()

	err := pkgerrors.Collection(
		pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error1"),
		pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error2"),
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

	err1 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error1")
	err2 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error2")

	colErr := pkgerrors.Collection(
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

	err1 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error1")
	err2 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error2")

	colErr := pkgerrors.Collection(
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

	err1 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error1")
	err2 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error2")

	colErr := pkgerrors.Collection(err1)

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
	err1 := pkgerrors.Error(pkgerrors.CodeType, ctx, "int", "float32")
	err2 := pkgerrors.Error(pkgerrors.CodeType, ctx, "int", "float32")

	colErr := pkgerrors.Collection(
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
	col := pkgerrors.Collection()
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
	err1 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx1, "above maximum", "error1")

	ctx2 := rulecontext.WithPathString(context.Background(), "path2a")
	ctx2 = rulecontext.WithPathString(ctx2, "b")
	err2 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx2, "above maximum", "error2")

	colErr := pkgerrors.Collection(
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
	col := pkgerrors.Collection()
	if first := col.For("a"); first != nil {
		t.Errorf("Expected first to be nil, got: %s", first)
	}
}

// TestCollectionMessage tests:
// - Error message is correctly formatted for single error
// - Error message includes count for multiple errors
func TestCollectionMessage(t *testing.T) {
	err := pkgerrors.Errorf(pkgerrors.CodeUnknown, context.Background(), "unknown error", "error123")

	col := pkgerrors.Collection(err)

	if msg := col.Error(); msg != "error123" {
		t.Errorf("Expected error message to be %s, got: %s", "error123", msg)
	}

	col = append(col, pkgerrors.Errorf(pkgerrors.CodeUnknown, context.Background(), "unknown error", "error123"))

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

	_ = pkgerrors.Collection().Error()
}

// TestValidationErrorInternal tests:
// - Internal() returns true for internal error codes
// - Internal() returns false for validation error codes
func TestValidationErrorInternal(t *testing.T) {
	ctx := context.Background()

	// Internal errors
	internalErr := pkgerrors.Error(pkgerrors.CodeInternal, ctx)
	if !internalErr.Internal() {
		t.Error("Expected CodeInternal to be classified as internal")
	}

	timeoutErr := pkgerrors.Error(pkgerrors.CodeTimeout, ctx)
	if !timeoutErr.Internal() {
		t.Error("Expected CodeTimeout to be classified as internal")
	}

	// Non-internal errors
	validationErr := pkgerrors.Error(pkgerrors.CodeMin, ctx, 10)
	if validationErr.Internal() {
		t.Error("Expected CodeMin to not be classified as internal")
	}
}

// TestValidationErrorValidation tests:
// - Validation() returns true for validation error codes
// - Validation() returns false for internal/permission error codes
func TestValidationErrorValidation(t *testing.T) {
	ctx := context.Background()

	// Validation errors
	minErr := pkgerrors.Error(pkgerrors.CodeMin, ctx, 10)
	if !minErr.Validation() {
		t.Error("Expected CodeMin to be classified as validation")
	}

	rangeErr := pkgerrors.Error(pkgerrors.CodeRange, ctx, "int")
	if !rangeErr.Validation() {
		t.Error("Expected CodeRange to be classified as validation")
	}

	// Non-validation errors
	internalErr := pkgerrors.Error(pkgerrors.CodeInternal, ctx)
	if internalErr.Validation() {
		t.Error("Expected CodeInternal to not be classified as validation")
	}
}

// TestValidationErrorPermission tests:
// - Permission() returns true for permission error codes
// - Permission() returns false for validation/internal error codes
func TestValidationErrorPermission(t *testing.T) {
	ctx := context.Background()

	// Permission errors
	forbiddenErr := pkgerrors.Error(pkgerrors.CodeForbidden, ctx)
	if !forbiddenErr.Permission() {
		t.Error("Expected CodeForbidden to be classified as permission")
	}

	notAllowedErr := pkgerrors.Error(pkgerrors.CodeNotAllowed, ctx)
	if !notAllowedErr.Permission() {
		t.Error("Expected CodeNotAllowed to be classified as permission")
	}

	// Non-permission errors
	validationErr := pkgerrors.Error(pkgerrors.CodeMin, ctx, 10)
	if validationErr.Permission() {
		t.Error("Expected CodeMin to not be classified as permission")
	}
}

// TestCollectionInternal tests:
// - Internal() returns true if any error is internal
// - Internal() returns false if no errors are internal
// - Internal() returns false for empty collections
func TestCollectionInternal(t *testing.T) {
	ctx := context.Background()

	// Empty collection
	emptyCol := pkgerrors.Collection()
	if emptyCol.Internal() {
		t.Error("Expected empty collection Internal() to return false")
	}

	// Collection with only validation errors
	validationCol := pkgerrors.Collection(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeMax, ctx, 100),
	)
	if validationCol.Internal() {
		t.Error("Expected validation-only collection Internal() to return false")
	}

	// Collection with one internal error
	mixedCol := pkgerrors.Collection(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeInternal, ctx),
	)
	if !mixedCol.Internal() {
		t.Error("Expected mixed collection with internal error Internal() to return true")
	}
}

// TestCollectionPermission tests:
// - Permission() returns true if any error is permission and none are internal
// - Permission() returns false if any error is internal
// - Permission() returns false if all errors are validation
// - Permission() returns false for empty collections
func TestCollectionPermission(t *testing.T) {
	ctx := context.Background()

	// Empty collection
	emptyCol := pkgerrors.Collection()
	if emptyCol.Permission() {
		t.Error("Expected empty collection Permission() to return false")
	}

	// Collection with only validation errors
	validationCol := pkgerrors.Collection(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeMax, ctx, 100),
	)
	if validationCol.Permission() {
		t.Error("Expected validation-only collection Permission() to return false")
	}

	// Collection with permission error
	permissionCol := pkgerrors.Collection(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeForbidden, ctx),
	)
	if !permissionCol.Permission() {
		t.Error("Expected collection with permission error Permission() to return true")
	}

	// Collection with internal and permission errors - internal takes precedence
	internalAndPermissionCol := pkgerrors.Collection(
		pkgerrors.Error(pkgerrors.CodeInternal, ctx),
		pkgerrors.Error(pkgerrors.CodeForbidden, ctx),
	)
	if internalAndPermissionCol.Permission() {
		t.Error("Expected collection with internal error Permission() to return false (internal takes precedence)")
	}
}

// TestCollectionValidation tests:
// - Validation() returns true if all errors are validation
// - Validation() returns false if any error is internal or permission
// - Validation() returns false for empty collections
func TestCollectionValidation(t *testing.T) {
	ctx := context.Background()

	// Empty collection
	emptyCol := pkgerrors.Collection()
	if emptyCol.Validation() {
		t.Error("Expected empty collection Validation() to return false")
	}

	// Collection with only validation errors
	validationCol := pkgerrors.Collection(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeMax, ctx, 100),
	)
	if !validationCol.Validation() {
		t.Error("Expected validation-only collection Validation() to return true")
	}

	// Collection with internal error
	internalCol := pkgerrors.Collection(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeInternal, ctx),
	)
	if internalCol.Validation() {
		t.Error("Expected collection with internal error Validation() to return false")
	}

	// Collection with permission error
	permissionCol := pkgerrors.Collection(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeForbidden, ctx),
	)
	if permissionCol.Validation() {
		t.Error("Expected collection with permission error Validation() to return false")
	}
}

// TestCollectionUnwrapEmpty tests:
// - Unwrap returns an empty slice for empty collections
func TestCollectionUnwrapEmpty(t *testing.T) {
	col := pkgerrors.Collection()
	unwrapped := col.Unwrap()
	if unwrapped == nil {
		t.Error("Expected Unwrap() to return an empty slice, not nil")
	}
	if len(unwrapped) != 0 {
		t.Errorf("Expected Unwrap() to return empty slice, got length %d", len(unwrapped))
	}
}

// TestCollectionErrorsIs tests:
// - errors.Is correctly identifies errors in the collection
// - errors.Is works with single error collections
// - errors.Is works with multiple error collections
// - errors.Is returns false for errors not in the collection
func TestCollectionErrorsIs(t *testing.T) {
	ctx := context.Background()

	err1 := pkgerrors.Error(pkgerrors.CodeMin, ctx, 10)
	err2 := pkgerrors.Error(pkgerrors.CodeMax, ctx, 100)
	err3 := pkgerrors.Error(pkgerrors.CodeType, ctx, "int", "string")

	// Single error collection
	col1 := pkgerrors.Collection(err1)
	if !errors.Is(col1, err1) {
		t.Error("Expected errors.Is to return true for error in single-error collection")
	}
	if errors.Is(col1, err2) {
		t.Error("Expected errors.Is to return false for error not in collection")
	}

	// Multiple error collection
	col2 := pkgerrors.Collection(err1, err2)
	if !errors.Is(col2, err1) {
		t.Error("Expected errors.Is to return true for first error in collection")
	}
	if !errors.Is(col2, err2) {
		t.Error("Expected errors.Is to return true for second error in collection")
	}
	if errors.Is(col2, err3) {
		t.Error("Expected errors.Is to return false for error not in collection")
	}

	// Empty collection
	emptyCol := pkgerrors.Collection()
	if errors.Is(emptyCol, err1) {
		t.Error("Expected errors.Is to return false for empty collection")
	}
}

// TestCollectionErrorsAs tests:
// - errors.As correctly extracts ValidationError from collection
// - errors.As works with single error collections
// - errors.As works with multiple error collections
// - errors.As returns false for incompatible types
func TestCollectionErrorsAs(t *testing.T) {
	ctx := context.Background()

	err1 := pkgerrors.Error(pkgerrors.CodeMin, ctx, 10)
	err2 := pkgerrors.Error(pkgerrors.CodeMax, ctx, 100)

	// Single error collection
	col1 := pkgerrors.Collection(err1)
	var extractedErr pkgerrors.ValidationError
	if !errors.As(col1, &extractedErr) {
		t.Error("Expected errors.As to return true and extract ValidationError from single-error collection")
	}
	if extractedErr != err1 {
		t.Error("Expected extracted error to match the error in collection")
	}

	// Multiple error collection - should extract first matching error
	col2 := pkgerrors.Collection(err1, err2)
	extractedErr = nil
	if !errors.As(col2, &extractedErr) {
		t.Error("Expected errors.As to return true and extract ValidationError from multi-error collection")
	}
	if extractedErr == nil {
		t.Error("Expected extracted error to not be nil")
	}
	// The extracted error should be one of the errors in the collection
	if extractedErr != err1 && extractedErr != err2 {
		t.Error("Expected extracted error to be one of the errors in the collection")
	}

	// Test with a different error code to verify it extracts correctly
	err3 := pkgerrors.Error(pkgerrors.CodeType, ctx, "int", "string")
	col3 := pkgerrors.Collection(err3)
	extractedErr = nil
	if !errors.As(col3, &extractedErr) {
		t.Error("Expected errors.As to return true for different error code")
	}
	if extractedErr.Code() != pkgerrors.CodeType {
		t.Errorf("Expected extracted error to have CodeType code, got %v", extractedErr.Code())
	}

	// Empty collection
	emptyCol := pkgerrors.Collection()
	extractedErr = nil
	if errors.As(emptyCol, &extractedErr) {
		t.Error("Expected errors.As to return false for empty collection")
	}
	if extractedErr != nil {
		t.Error("Expected extracted error to remain nil for empty collection")
	}
}

// TestCollectionErrorsAsWithNestedErrors tests:
// - errors.As works correctly when errors are nested in multiple collections
func TestCollectionErrorsAsWithNestedErrors(t *testing.T) {
	ctx := context.Background()

	err1 := pkgerrors.Error(pkgerrors.CodeMin, ctx, 10)
	err2 := pkgerrors.Error(pkgerrors.CodeMax, ctx, 100)

	// Create a collection and then wrap it (simulating nested error scenarios)
	innerCol := pkgerrors.Collection(err1, err2)
	outerCol := pkgerrors.Collection(innerCol.First(), pkgerrors.Error(pkgerrors.CodeType, ctx, "int", "string"))

	var extractedErr pkgerrors.ValidationError
	if !errors.As(outerCol, &extractedErr) {
		t.Error("Expected errors.As to work with nested error scenarios")
	}
	if extractedErr == nil {
		t.Error("Expected extracted error to not be nil")
	}
}
