package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
)

func TestNewCoercionError(t *testing.T) {
	ctx := context.Background()

	err := errors.NewCoercionError(ctx, "int", "float32")

	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if err.Code() != errors.CodeType {
		t.Errorf("Expected error to have code %d, got %d", errors.CodeType, err.Code())
	}
}
