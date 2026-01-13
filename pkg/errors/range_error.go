package errors

import (
	"context"
)

// NewRangeError creates a new ValidationError with the CodeRange code given a
// a target data type.
//
// Use when you understand the provided type and can convert from it but there is too
// much data to be contained in the new type.
//
// For example, converting from int to int8 is ok if the value is less than 128 but anything
// higher cannot be converted and should throw an error.
func NewRangeError(ctx context.Context, target string) ValidationError {
	return Error(CodeRange, ctx, target)
}
