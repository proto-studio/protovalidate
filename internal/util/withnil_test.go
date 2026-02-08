package util

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
)

func TestTrySetNilIfAllowed(t *testing.T) {
	ctx := context.TODO()

	// Test case 1: input is not nil - should return false, nil
	handled, err := TrySetNilIfAllowed(ctx, false, "not nil", nil)
	if handled {
		t.Error("Expected handled to be false when input is not nil")
	}
	if err != nil {
		t.Errorf("Expected no error when input is not nil, got: %s", err)
	}

	// Test case 2: input is nil, withNil is false - should return true with CodeNull error
	var output *string
	handled, err = TrySetNilIfAllowed(ctx, false, nil, &output)
	if !handled {
		t.Error("Expected handled to be true when nil is not allowed")
	}
	if err == nil {
		t.Error("Expected error when nil is not allowed")
	} else if err.Code() != errors.CodeNull {
		t.Errorf("Expected error code to be CodeNull, got: %s", err.Code())
	}

	// Test case 3: input is nil, withNil is true, output is not a pointer - should return false with CodeInternal error
	var nonPointer string
	handled, err = TrySetNilIfAllowed(ctx, true, nil, nonPointer)
	if handled {
		t.Error("Expected handled to be false when output is not a pointer")
	}
	if err == nil {
		t.Error("Expected error when output is not a pointer")
	} else if err.Code() != errors.CodeInternal {
		t.Errorf("Expected error code to be CodeInternal, got: %s", err.Code())
	}

	// Test case 4: input is nil, withNil is true, output is nil pointer - should return false with CodeInternal error
	var nilPointer *string
	handled, err = TrySetNilIfAllowed(ctx, true, nil, nilPointer)
	if handled {
		t.Error("Expected handled to be false when output pointer is nil")
	}
	if err == nil {
		t.Error("Expected error when output pointer is nil")
	} else if err.Code() != errors.CodeInternal {
		t.Errorf("Expected error code to be CodeInternal, got: %s", err.Code())
	}

	// Test case 5: input is nil, withNil is true, output points to nil-able type (pointer) - should set to nil
	var outputPtr *string
	handled, err = TrySetNilIfAllowed(ctx, true, nil, &outputPtr)
	if !handled {
		t.Error("Expected handled to be true when nil is set")
	}
	if err != nil {
		t.Errorf("Expected no error when nil is set, got: %s", err)
	}
	if outputPtr != nil {
		t.Error("Expected output to be set to nil")
	}

	// Test case 6: input is nil, withNil is true, output points to interface - should set to nil
	var outputInterface interface{}
	handled, err = TrySetNilIfAllowed(ctx, true, nil, &outputInterface)
	if !handled {
		t.Error("Expected handled to be true when nil is set for interface")
	}
	if err != nil {
		t.Errorf("Expected no error when nil is set for interface, got: %s", err)
	}
	if outputInterface != nil {
		t.Error("Expected output interface to be set to nil")
	}

	// Test case 7: input is nil, withNil is true, output points to slice - should set to nil
	var outputSlice []string
	handled, err = TrySetNilIfAllowed(ctx, true, nil, &outputSlice)
	if !handled {
		t.Error("Expected handled to be true when nil is set for slice")
	}
	if err != nil {
		t.Errorf("Expected no error when nil is set for slice, got: %s", err)
	}
	if outputSlice != nil {
		t.Error("Expected output slice to be set to nil")
	}

	// Test case 8: input is nil, withNil is true, output points to map - should set to nil
	var outputMap map[string]int
	handled, err = TrySetNilIfAllowed(ctx, true, nil, &outputMap)
	if !handled {
		t.Error("Expected handled to be true when nil is set for map")
	}
	if err != nil {
		t.Errorf("Expected no error when nil is set for map, got: %s", err)
	}
	if outputMap != nil {
		t.Error("Expected output map to be set to nil")
	}

	// Test case 9: input is nil, withNil is true, output points to channel - should set to nil
	var outputChan chan int
	handled, err = TrySetNilIfAllowed(ctx, true, nil, &outputChan)
	if !handled {
		t.Error("Expected handled to be true when nil is set for channel")
	}
	if err != nil {
		t.Errorf("Expected no error when nil is set for channel, got: %s", err)
	}
	if outputChan != nil {
		t.Error("Expected output channel to be set to nil")
	}

	// Test case 10: input is nil, withNil is true, output points to function - should set to nil
	var outputFunc func()
	handled, err = TrySetNilIfAllowed(ctx, true, nil, &outputFunc)
	if !handled {
		t.Error("Expected handled to be true when nil is set for function")
	}
	if err != nil {
		t.Errorf("Expected no error when nil is set for function, got: %s", err)
	}
	if outputFunc != nil {
		t.Error("Expected output function to be set to nil")
	}

	// Test case 11: input is nil, withNil is true, output points to non-nil-able type (int) - should return false, nil
	var outputInt int
	handled, err = TrySetNilIfAllowed(ctx, true, nil, &outputInt)
	if handled {
		t.Error("Expected handled to be false when element type doesn't support nil")
	}
	if err != nil {
		t.Errorf("Expected no error when element type doesn't support nil, got: %s", err)
	}
}
