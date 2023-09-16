package errors

import (
	"context"
)

// NewCoercionError creates a new ValidationError with the CodeType code given an
// expected and received type name.
//
// Use when you expected one type and received another.
func NewCoercionError(ctx context.Context, expected, received string) ValidationError {
	return Errorf(CodeType, ctx, "error converting %s to %s", received, expected)
}
