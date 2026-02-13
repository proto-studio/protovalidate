package util

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
)

func TestTryNilIfAllowed(t *testing.T) {
	ctx := context.TODO()

	// Input is not nil - should return false, nil
	handled, err := TryNilIfAllowed(ctx, false, "not nil")
	if handled {
		t.Error("Expected handled to be false when input is not nil")
	}
	if err != nil {
		t.Errorf("Expected no error when input is not nil, got: %s", err)
	}

	// Input is nil, withNil is false - should return true with CodeNull error
	handled, err = TryNilIfAllowed(ctx, false, nil)
	if !handled {
		t.Error("Expected handled to be true when nil is not allowed")
	}
	if err == nil {
		t.Error("Expected error when nil is not allowed")
	} else if err.Code() != errors.CodeNull {
		t.Errorf("Expected error code to be CodeNull, got: %s", err.Code())
	}

	// Input is nil, withNil is true - should return true, nil
	handled, err = TryNilIfAllowed(ctx, true, nil)
	if !handled {
		t.Error("Expected handled to be true when nil is allowed")
	}
	if err != nil {
		t.Errorf("Expected no error when nil is allowed, got: %s", err)
	}
}
