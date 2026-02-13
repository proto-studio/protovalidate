package testhelpers

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules"
)

// TestNeverAssignableImpl_priv tests that the priv() method works.
// This is needed for coverage even though it's an internal implementation detail.
func TestNeverAssignableImpl_priv(t *testing.T) {
	na := &neverAssignableImpl{}
	// Just call it - it's a no-op but needed for coverage
	na.priv()
}

// TestCheckRuleSetInterface ensures CheckRuleSetInterface is covered when testing this package in isolation.
func TestCheckRuleSetInterface(t *testing.T) {
	if !CheckRuleSetInterface[string](rules.String()) {
		t.Error("CheckRuleSetInterface: expected rules.String() to implement RuleSet[string]")
	}
	if CheckRuleSetInterface[string](123) {
		t.Error("CheckRuleSetInterface: expected 123 not to implement RuleSet[string]")
	}
}

// TestCheckRuleInterface ensures CheckRuleInterface is covered when testing this package in isolation.
func TestCheckRuleInterface(t *testing.T) {
	if !CheckRuleInterface[string](rules.String()) {
		t.Error("CheckRuleInterface: expected rules.String() to implement Rule[string]")
	}
	if CheckRuleInterface[string](123) {
		t.Error("CheckRuleInterface: expected 123 not to implement Rule[string]")
	}
}

// TestMustApplyAny ensures MustApplyAny (and checkAlways) are covered when testing this package in isolation.
func TestMustApplyAny(t *testing.T) {
	out, err := MustApplyAny(t, rules.Any(), 42)
	if err != nil {
		t.Errorf("MustApplyAny: %v", err)
	}
	if out != 42 {
		t.Errorf("MustApplyAny: expected 42, got %v", out)
	}
}

// TestNewMockRuleSetWithErrors_ApplyCallCount ensures NewMockRuleSetWithErrors and ApplyCallCount are covered.
func TestNewMockRuleSetWithErrors_ApplyCallCount(t *testing.T) {
	mock := NewMockRuleSetWithErrors[int](2)
	if mock == nil {
		t.Fatal("NewMockRuleSetWithErrors returned nil")
	}
	if c := mock.ApplyCallCount(); c != 0 {
		t.Errorf("ApplyCallCount: expected 0 before Apply, got %d", c)
	}
	_, err := mock.Apply(nil, 10)
	if err == nil {
		t.Error("expected error from MockRuleSetWithErrors")
	}
	if c := mock.ApplyCallCount(); c != 1 {
		t.Errorf("ApplyCallCount: expected 1 after Apply, got %d", c)
	}
}

