package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

func TestNew(t *testing.T) {
	err := errors.New(errors.CodeMin, "a.b.c", "testmessage")

	if err.Code() != errors.CodeMin {
		t.Errorf("Expected code to be %d, got: %d", errors.CodeMin, err.Code())
	}

	if err.Path() != "a.b.c" {
		t.Errorf("Expected path to be %s, got: %s", "a.b.c", err.Path())
	}

	if err.Error() != "testmessage" {
		t.Errorf("Expected path to be %s, got: %s", "testmessage", err)
	}
}

func TestErrorfContainsFullPath(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")
	err := errors.Errorf(errors.CodeUnknown, ctx, "error")

	if err.Path() != "a.b" {
		t.Errorf("Expected full path to be set, got: %s", err.Path())
	}
}

func TestErrorfContainsCode(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "a")
	ctx = rulecontext.WithPathString(ctx, "b")

	err := errors.Errorf(errors.CodeUnknown, ctx, "error")

	if err.Code() != errors.CodeUnknown {
		t.Errorf("Expected code to be %d, got: %d", errors.CodeUnknown, err.Code())
	}

	err = errors.Errorf(errors.CodeMin, ctx, "error")

	if err.Code() != errors.CodeMin {
		t.Errorf("Expected code to be %d, got: %d", errors.CodeMin, err.Code())
	}
}

func TestErrorMessage(t *testing.T) {
	err := errors.Errorf(errors.CodeUnknown, context.Background(), "error123")

	if msg := err.Error(); msg != "error123" {
		t.Errorf("Expected error message to be %s, got: %s", "error123", msg)
	}
}
