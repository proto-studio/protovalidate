package errors

import (
	"context"
	"testing"
)

// Tests for unexported mergeErrorConfig function

func TestMergeErrorConfig_BothNil(t *testing.T) {
	result := mergeErrorConfig(nil, nil)
	if result != nil {
		t.Error("Expected nil when both parent and child are nil")
	}
}

func TestMergeErrorConfig_ChildNil(t *testing.T) {
	parent := &ErrorConfig{Short: "parent short"}
	result := mergeErrorConfig(parent, nil)
	if result != parent {
		t.Error("Expected parent to be returned when child is nil")
	}
	if result.Short != "parent short" {
		t.Errorf("Expected Short = %q, got %q", "parent short", result.Short)
	}
}

func TestMergeErrorConfig_ParentNil(t *testing.T) {
	child := &ErrorConfig{Short: "child short"}
	result := mergeErrorConfig(nil, child)
	if result != child {
		t.Error("Expected child to be returned when parent is nil")
	}
	if result.Short != "child short" {
		t.Errorf("Expected Short = %q, got %q", "child short", result.Short)
	}
}

func TestMergeErrorConfig_ChildLongTakesPrecedence(t *testing.T) {
	parent := &ErrorConfig{Long: "parent long"}
	child := &ErrorConfig{Long: "child long"}
	result := mergeErrorConfig(parent, child)
	if result.Long != "child long" {
		t.Errorf("Expected Long = %q, got %q", "child long", result.Long)
	}
}

func TestMergeErrorConfig_ParentLongUsedWhenChildEmpty(t *testing.T) {
	parent := &ErrorConfig{Long: "parent long"}
	child := &ErrorConfig{Short: "child short"} // Long is empty
	result := mergeErrorConfig(parent, child)
	if result.Long != "parent long" {
		t.Errorf("Expected Long = %q, got %q", "parent long", result.Long)
	}
}

func TestMergeErrorConfig_ChildCallbackTakesPrecedence(t *testing.T) {
	parentCalled := false
	childCalled := false

	parent := &ErrorConfig{
		Callback: func(ctx context.Context, err ValidationError) ValidationError {
			parentCalled = true
			return err
		},
	}
	child := &ErrorConfig{
		Callback: func(ctx context.Context, err ValidationError) ValidationError {
			childCalled = true
			return err
		},
	}

	result := mergeErrorConfig(parent, child)
	result.Callback(context.Background(), nil)

	if !childCalled {
		t.Error("Expected child callback to be called")
	}
	if parentCalled {
		t.Error("Expected parent callback NOT to be called when child has callback")
	}
}

func TestMergeErrorConfig_ParentCallbackUsedWhenChildNil(t *testing.T) {
	parentCalled := false

	parent := &ErrorConfig{
		Callback: func(ctx context.Context, err ValidationError) ValidationError {
			parentCalled = true
			return err
		},
	}
	child := &ErrorConfig{Short: "child short"} // Callback is nil

	result := mergeErrorConfig(parent, child)
	result.Callback(context.Background(), nil)

	if !parentCalled {
		t.Error("Expected parent callback to be called when child callback is nil")
	}
}
