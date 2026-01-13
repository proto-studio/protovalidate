package util

import (
	"context"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
)

// TrySetNilIfAllowed attempts to set the output to nil if withNil is true and input is nil.
// It returns true if nil was successfully set (and the caller should return), false if normal processing should continue,
// and an error if there was a problem setting nil or if nil is not allowed.
func TrySetNilIfAllowed(ctx context.Context, withNil bool, input, output any) (handled bool, err errors.ValidationErrorCollection) {
	// If input is not nil, continue with normal processing
	if input != nil {
		return false, nil
	}

	// Input is nil - check if nil is allowed
	if !withNil {
		// Nil is not allowed, return error
		return true, errors.Collection(errors.Error(
			errors.CodeNull, ctx,
		))
	}

	// Nil is allowed - ensure output is a pointer
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return false, errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "internal error", "output must be a non-nil pointer",
		))
	}

	// Get the element the pointer points to
	elem := outputVal.Elem()

	// Check if the element type supports nil (pointer, interface, slice, map, channel, function)
	elemKind := elem.Kind()
	if elemKind == reflect.Ptr || elemKind == reflect.Interface || elemKind == reflect.Slice ||
		elemKind == reflect.Map || elemKind == reflect.Chan || elemKind == reflect.Func {
		// Set to nil
		elem.Set(reflect.Zero(elem.Type()))
		return true, nil
	}

	// Element type doesn't support nil, continue with normal processing
	return false, nil
}
