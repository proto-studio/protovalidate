package objects

import (
	"context"
	"fmt"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
)

type testStruct struct {
	X int
	Y int
	z int //lint:ignore U1000 Used in reflection testing but not code
}

type alwaysErrorContext struct {
	context.Context
}

func (c *alwaysErrorContext) Err() error {
	return fmt.Errorf("test error")
}

func TestMissingMapping(t *testing.T) {
	ruleSet := New[*testStruct]().withParent()

	// Manually create a mapping that is not on the struct
	ruleSet.key = "A"
	ruleSet.mapping = "A"

	// This should work
	ruleSet = ruleSet.WithKey("X", numbers.NewInt().Any())

	// This should panic

	defer func() {
		err, ok := recover().(error)

		if err == nil || !ok {
			t.Error("Expected panic with error interface")
		} else if err.Error() != "missing destination mapping for field: A" {
			t.Errorf("Expected missing mapping error, got: %s", err)
		}
	}()

	ruleSet.WithKey("A", numbers.NewInt().Any())
}

func TestUnexportedField(t *testing.T) {
	defer func() {
		err, ok := recover().(error)

		if err == nil || !ok {
			t.Error("Expected panic with error interface")
		} else if err.Error() != "field is not exported: z" {
			t.Errorf("Expected field is not exported error, got: %s", err)
		}
	}()

	ruleSet := New[*testStruct]().withParent()

	// Manually create a mapping for the unexported field
	ruleSet.key = "z"
	ruleSet.mapping = "z"

	ruleSet.WithKey("z", numbers.NewInt().Any())
}

// Requirements:
// - Returns nil
// - Returns cancelled
// - Returns timeout
// - Returns other
func TestContextErrorToValidation(t *testing.T) {

	// No error
	ctx := context.Background()
	err := contextErrorToValidation(ctx)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 0)

	err = contextErrorToValidation(ctx)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if c := err.Code(); c != errors.CodeTimeout {
		t.Errorf("Expected code to be %s, got %s", errors.CodeTimeout, c)
	}

	cancel()

	ctx, cancel = context.WithCancel(context.Background())
	cancel()

	err = contextErrorToValidation(ctx)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if c := err.Code(); c != errors.CodeCancelled {
		t.Errorf("Expected code to be %s, got %s", errors.CodeCancelled, c)
	}

	err = contextErrorToValidation(&alwaysErrorContext{
		Context: context.Background(),
	})
	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if c := err.Code(); c != errors.CodeInternal {
		t.Errorf("Expected code to be %s, got %s", errors.CodeInternal, c)
	}
}

// Requirements:
// - counter panics when it goes negative
func TestRefTrackerNegative(t *testing.T) {
	c := newCounter()
	c.Increment()

	c.Lock()
	// Intentionally empty critical section
	c.Unlock()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	c.Lock()
	c.Unlock()
}
