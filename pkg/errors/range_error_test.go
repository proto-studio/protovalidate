package errors_test

import (
	"context"
	"testing"

	pkgerrors "proto.zip/studio/validate/pkg/errors"
)

func TestNewRangeError(t *testing.T) {
	ctx := context.Background()
	err := pkgerrors.NewRangeError(ctx, "int8")
	if err == nil {
		t.Fatal("NewRangeError should not return nil")
	}
	if err.Code() != pkgerrors.CodeRange {
		t.Errorf("Code() = %s, want %s", err.Code(), pkgerrors.CodeRange)
	}
	if err.Error() == "" {
		t.Error("Error() should not be empty")
	}
}
