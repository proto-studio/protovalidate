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

	err := pkgerrors.Join(
		pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error1"),
		pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error2"),
	)

	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if s := len(pkgerrors.Unwrap(err)); s != 2 {
		t.Errorf("Expected error to have size %d, got %d", 2, s)
	}
}

// TestCollectionAsSlice tests that the collection can be used as a slice (len, range, index).
func TestCollectionAsSlice(t *testing.T) {
	ctx := context.Background()

	err1 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error1")
	err2 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error2")

	colErr := pkgerrors.Join(err1, err2)

	if colErr == nil {
		t.Fatal("Expected error to not be nil")
	}
	all := pkgerrors.Unwrap(colErr)
	if l := len(all); l != 2 {
		t.Fatalf("Expected error to have length %d, got %d", 2, l)
	}
	if !((all[0] == err1 && all[1] == err2) || (all[0] == err2 && all[1] == err1)) {
		t.Errorf("Expected both errors to be in collection")
	}
}

// TestCollectionUnwrap tests:
// - Unwrap should return an array of errors.
func TestCollectionUnwrap(t *testing.T) {
	ctx := context.Background()

	err1 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error1")
	err2 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error2")

	colErr := pkgerrors.Join(err1, err2)

	if colErr == nil {
		t.Fatal("Expected error to not be nil")
	}
	all := pkgerrors.Unwrap(colErr)
	if l := len(all); l != 2 {
		t.Fatalf("Expected error to have length %d, got %d", 2, l)
	}

	if !((all[0] == err1 && all[1] == err2) || (all[0] == err2 && all[1] == err1)) {
		t.Errorf("Expected both errors to be returned")
	}
}

// TestJoinFlattensMultiError tests that Join flattens a multiError (Unwrap() non-empty) into the result.
func TestJoinFlattensMultiError(t *testing.T) {
	ctx := context.Background()
	err1 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error1")
	err2 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error2")
	err3 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error3")

	inner := pkgerrors.Join(err1, err2)
	outer := pkgerrors.Join(inner, err3)
	all := pkgerrors.Unwrap(outer)
	if len(all) != 3 {
		t.Errorf("Expected Join to flatten to 3 errors, got %d", len(all))
	}
}

// TestJoinSkipsNilsAndNonValidationErrors tests that Join skips nil entries and non-ValidationError values.
func TestJoinSkipsNilsAndNonValidationErrors(t *testing.T) {
	ctx := context.Background()
	err1 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error1")
	err2 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error2")

	col := pkgerrors.Join(nil, err1, nil, err2)
	if col == nil {
		t.Fatal("Join(nil, err1, nil, err2) should not be nil")
	}
	if n := len(pkgerrors.Unwrap(col)); n != 2 {
		t.Errorf("Expected 2 errors, got %d", n)
	}

	plainErr := errors.New("plain")
	col = pkgerrors.Join(err1, plainErr, err2)
	if col == nil {
		t.Fatal("Join with plain error should not be nil")
	}
	if n := len(pkgerrors.Unwrap(col)); n != 2 {
		t.Errorf("Expected 2 ValidationErrors (plain skipped), got %d", n)
	}
}

// TestCollectionLen tests that Unwrap() returns the correct number of errors.
func TestCollectionLen(t *testing.T) {
	ctx := context.Background()

	err1 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error1")
	err2 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "error2")

	colErr := pkgerrors.Join(err1)
	if s := len(pkgerrors.Unwrap(colErr)); s != 1 {
		t.Errorf("Expected size to be 1, got: %d", s)
	}

	colErr = pkgerrors.Join(err1, err2)
	if s := len(pkgerrors.Unwrap(colErr)); s != 2 {
		t.Errorf("Expected size to be 2, got: %d", s)
	}
}

// TestCollectionCodeWhenMultiple tests that Code() returns the first error's code when multiple errors exist.
func TestCollectionCodeWhenMultiple(t *testing.T) {
	ctx := context.Background()
	err1 := pkgerrors.Error(pkgerrors.CodeType, ctx, "int", "float32")
	err2 := pkgerrors.Error(pkgerrors.CodeType, ctx, "int", "float32")

	colErr := pkgerrors.Join(err1, err2)
	if colErr == nil {
		t.Fatal("Expected error to not be nil")
	}
	if colErr.Code() != pkgerrors.CodeType {
		t.Errorf("Expected Code() to return CodeType, got: %s", colErr.Code())
	}
}

// TestCollectionDelegation tests that Path, PathAs, ShortError, DocsURI, TraceURI, Meta, and Params
// on a joined (multi) error delegate to the first error.
func TestCollectionDelegation(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "field")
	err1 := pkgerrors.Errorf(pkgerrors.CodeMin, ctx, "above minimum", "value at least %d", 10)
	err2 := pkgerrors.Errorf(pkgerrors.CodeMax, ctx, "above maximum", "value at most %d", 100)

	col := pkgerrors.Join(err1, err2)
	if col == nil {
		t.Fatal("Expected joined error to not be nil")
	}

	if p := col.Path(); p != "/field" {
		t.Errorf("Expected Path() to return first error path /field, got: %s", p)
	}
	if p := col.PathAs(pkgerrors.DefaultPathSerializer{}); p != "/field" {
		t.Errorf("Expected PathAs() to return first error path /field, got: %s", p)
	}
	if s := col.ShortError(); s != "above minimum" {
		t.Errorf("Expected ShortError() to return first error short, got: %s", s)
	}
	if u := col.DocsURI(); u != "" {
		t.Errorf("Expected DocsURI() to return first error docs URI, got: %s", u)
	}
	if u := col.TraceURI(); u != "" {
		t.Errorf("Expected TraceURI() to return first error trace URI, got: %s", u)
	}
	if m := col.Meta(); m != nil {
		t.Errorf("Expected Meta() to return first error meta, got: %v", m)
	}
	if params := col.Params(); len(params) != 1 || params[0] != 10 {
		t.Errorf("Expected Params() to return first error params [10], got: %v", params)
	}
}

// TestUnwrapNil tests that pkgerrors.Unwrap(nil) returns nil (Unwrap() on nil is not called).
func TestUnwrapNil(t *testing.T) {
	all := pkgerrors.Unwrap(nil)
	if all != nil {
		t.Errorf("Expected nil, got length %d", len(all))
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

	colErr := pkgerrors.Join(err1, err2)

	if colErr == nil {
		t.Fatal("Expected error to not be nil")
	}
	if s := len(pkgerrors.Unwrap(colErr)); s != 2 {
		t.Fatalf("Expected error to have size %d, got %d", 2, s)
	}

	path1err := pkgerrors.For(colErr, "/path1")
	if path1err == nil {
		t.Errorf("Expected path1 error to not be nil")
	} else if s := len(pkgerrors.Unwrap(path1err)); s != 1 {
		t.Errorf("Expected a collection with 1 error, got: '%d'", s)
	} else if first := pkgerrors.Unwrap(path1err)[0]; first != err1 {
		t.Errorf("Expected '%s' to be returned, got: '%s'", err1, first)
	}

	path1err = pkgerrors.For(colErr, "/path1/b")
	if path1err != nil {
		t.Errorf("Expected error to be nil, got: %s", path1err)
	}

	path2err := pkgerrors.For(colErr, "/path2a/b")
	if path2err == nil {
		t.Errorf("Expected path2 error to not be nil")
	} else if s := len(pkgerrors.Unwrap(path2err)); s != 1 {
		t.Errorf("Expected a collection with 1 error, got: '%d'", s)
	} else if first := pkgerrors.Unwrap(path2err)[0]; first != err2 {
		t.Errorf("Expected '%s' to be returned, got: '%s'", err2, first)
	}
}

// TestCollectionForNil tests that For(nil, path) returns nil.
func TestCollectionForNil(t *testing.T) {
	if result := pkgerrors.For(nil, "a"); result != nil {
		t.Errorf("Expected For(nil, path) to be nil, got: %s", result)
	}
}

// TestCollectionMessage tests:
// - Error message is correctly formatted for single error
// - Error message includes count for multiple errors
func TestCollectionMessage(t *testing.T) {
	err := pkgerrors.Errorf(pkgerrors.CodeUnknown, context.Background(), "unknown error", "error123")

	col := pkgerrors.Join(err)
	if msg := col.Error(); msg != "error123" {
		t.Errorf("Expected error message to be %s, got: %s", "error123", msg)
	}

	err2 := pkgerrors.Errorf(pkgerrors.CodeUnknown, context.Background(), "unknown error", "error123")
	col = pkgerrors.Join(err, err2)
	if msg := col.Error(); !strings.Contains(msg, "(and 1 more)") {
		t.Errorf("Expected error message to contain the string '(and 1 more)', got: %s", msg)
	}
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
func TestCollectionInternal(t *testing.T) {
	ctx := context.Background()

	// Collection with only validation errors
	validationCol := pkgerrors.Join(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeMax, ctx, 100),
	)
	if validationCol.Internal() {
		t.Error("Expected validation-only collection Internal() to return false")
	}

	// Collection with one internal error
	mixedCol := pkgerrors.Join(
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
func TestCollectionPermission(t *testing.T) {
	ctx := context.Background()

	// Collection with only validation errors
	validationCol := pkgerrors.Join(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeMax, ctx, 100),
	)
	if validationCol.Permission() {
		t.Error("Expected validation-only collection Permission() to return false")
	}

	// Collection with permission error
	permissionCol := pkgerrors.Join(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeForbidden, ctx),
	)
	if !permissionCol.Permission() {
		t.Error("Expected collection with permission error Permission() to return true")
	}

	// Collection with internal and permission errors - internal takes precedence
	internalAndPermissionCol := pkgerrors.Join(
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
func TestCollectionValidation(t *testing.T) {
	ctx := context.Background()

	// Collection with only validation errors
	validationCol := pkgerrors.Join(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeMax, ctx, 100),
	)
	if !validationCol.Validation() {
		t.Error("Expected validation-only collection Validation() to return true")
	}

	// Collection with internal error
	internalCol := pkgerrors.Join(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeInternal, ctx),
	)
	if internalCol.Validation() {
		t.Error("Expected collection with internal error Validation() to return false")
	}

	// Collection with permission error
	permissionCol := pkgerrors.Join(
		pkgerrors.Error(pkgerrors.CodeMin, ctx, 10),
		pkgerrors.Error(pkgerrors.CodeForbidden, ctx),
	)
	if permissionCol.Validation() {
		t.Error("Expected collection with permission error Validation() to return false")
	}
}

// TestUnwrapSingle tests that a single error's Unwrap() returns nil and pkgerrors.Unwrap returns []error{err}.
func TestUnwrapSingle(t *testing.T) {
	ctx := context.Background()
	err := pkgerrors.Error(pkgerrors.CodeMin, ctx, 10)
	all := pkgerrors.Unwrap(err)
	if len(all) != 1 || all[0] != err {
		t.Errorf("Expected one error, got: %d", len(all))
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

	// Single error
	if !errors.Is(err1, err1) {
		t.Error("Expected errors.Is to return true for error in single-error collection")
	}
	if errors.Is(err1, err2) {
		t.Error("Expected errors.Is to return false for different error")
	}
	// Multiple error collection
	col2 := pkgerrors.Join(err1, err2)
	if !errors.Is(col2, err1) {
		t.Error("Expected errors.Is to return true for first error in collection")
	}
	if !errors.Is(col2, err2) {
		t.Error("Expected errors.Is to return true for second error in collection")
	}
	if errors.Is(col2, err3) {
		t.Error("Expected errors.Is to return false for error not in collection")
	}

	// Nil
	if errors.Is(nil, err1) {
		t.Error("Expected errors.Is(nil, err) to return false")
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

	// Single error
	var extractedErr pkgerrors.ValidationError
	if !errors.As(err1, &extractedErr) {
		t.Error("Expected errors.As to return true and extract ValidationError from single error")
	}
	if extractedErr != err1 {
		t.Error("Expected extracted error to match")
	}

	// Multiple error collection - should extract first matching error
	col2 := pkgerrors.Join(err1, err2)
	extractedErr = nil
	if !errors.As(col2, &extractedErr) {
		t.Error("Expected errors.As to return true and extract ValidationError from multi-error collection")
	}
	if extractedErr == nil {
		t.Error("Expected extracted error to not be nil")
	}
	// The extracted error should have a code from the collection (As may return the multiError or first wrapped error)
	if extractedErr.Code() != pkgerrors.CodeMin && extractedErr.Code() != pkgerrors.CodeMax {
		t.Errorf("Expected extracted error code to be CodeMin or CodeMax, got %v", extractedErr.Code())
	}

	// Test with a different error code to verify it extracts correctly
	err3 := pkgerrors.Error(pkgerrors.CodeType, ctx, "int", "string")
	extractedErr = nil
	if !errors.As(err3, &extractedErr) {
		t.Error("Expected errors.As to return true for different error code")
	}
	if extractedErr.Code() != pkgerrors.CodeType {
		t.Errorf("Expected extracted error to have CodeType code, got %v", extractedErr.Code())
	}

	// Nil
	extractedErr = nil
	if errors.As(nil, &extractedErr) {
		t.Error("Expected errors.As(nil, &extractedErr) to return false")
	}
	if extractedErr != nil {
		t.Error("Expected extracted error to remain nil")
	}
}

// TestCollectionErrorsAsWithNestedErrors tests:
// - errors.As works correctly when errors are nested in multiple collections
func TestCollectionErrorsAsWithNestedErrors(t *testing.T) {
	ctx := context.Background()

	err1 := pkgerrors.Error(pkgerrors.CodeMin, ctx, 10)
	err2 := pkgerrors.Error(pkgerrors.CodeMax, ctx, 100)

	// Create a collection and then wrap it (simulating nested error scenarios)
	innerCol := pkgerrors.Join(err1, err2)
	first := pkgerrors.Unwrap(innerCol)[0]
	outerCol := pkgerrors.Join(first, pkgerrors.Error(pkgerrors.CodeType, ctx, "int", "string"))

	var extractedErr pkgerrors.ValidationError
	if !errors.As(outerCol, &extractedErr) {
		t.Error("Expected errors.As to work with nested error scenarios")
	}
	if extractedErr == nil {
		t.Error("Expected extracted error to not be nil")
	}
}
