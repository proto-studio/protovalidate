package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/rules"
)

// Test that WithMin and WithMinExclusive conflict with each other
func TestStringRuleSet_WithMin_WithMinExclusiveConflict(t *testing.T) {
	ruleSet := rules.String().WithMin("b")

	// Adding WithMinExclusive should conflict and replace WithMin
	ruleSet2 := ruleSet.WithMinExclusive("c")

	var output string

	// Original rule set should still have WithMin
	err := ruleSet.Apply(context.TODO(), "b", &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMin at threshold, got %s", err)
	}

	// New rule set should have WithMinExclusive (exclusive, so "c" should fail)
	err = ruleSet2.Apply(context.TODO(), "c", &output)
	if err == nil {
		t.Errorf("Expected error for WithMinExclusive at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := "StringRuleSet.WithMinExclusive(\"c\")"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMax and WithMaxExclusive conflict with each other
func TestStringRuleSet_WithMax_WithMaxExclusiveConflict(t *testing.T) {
	ruleSet := rules.String().WithMax("y")

	// Adding WithMaxExclusive should conflict and replace WithMax
	ruleSet2 := ruleSet.WithMaxExclusive("x")

	var output string

	// Original rule set should still have WithMax
	err := ruleSet.Apply(context.TODO(), "y", &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// New rule set should have WithMaxExclusive (exclusive, so "x" should fail)
	err = ruleSet2.Apply(context.TODO(), "x", &output)
	if err == nil {
		t.Errorf("Expected error for WithMaxExclusive at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := "StringRuleSet.WithMaxExclusive(\"x\")"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMinExclusive and WithMin conflict with each other (reverse order)
func TestStringRuleSet_WithMinExclusive_WithMinConflict(t *testing.T) {
	ruleSet := rules.String().WithMinExclusive("b")

	// Adding WithMin should conflict and replace WithMinExclusive
	ruleSet2 := ruleSet.WithMin("c")

	var output string

	// Original rule set should still have WithMinExclusive
	err := ruleSet.Apply(context.TODO(), "b", &output)
	if err == nil {
		t.Errorf("Expected error for WithMinExclusive at threshold (exclusive)")
	}

	// New rule set should have WithMin (inclusive, so "c" should pass)
	err = ruleSet2.Apply(context.TODO(), "c", &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMin at threshold, got %s", err)
	}

	// Verify serialization shows the conflict resolution
	expected := "StringRuleSet.WithMin(\"c\")"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMaxExclusive and WithMax conflict with each other (reverse order)
func TestStringRuleSet_WithMaxExclusive_WithMaxConflict(t *testing.T) {
	ruleSet := rules.String().WithMaxExclusive("y")

	// Adding WithMax should conflict and replace WithMaxExclusive
	ruleSet2 := ruleSet.WithMax("x")

	var output string

	// Original rule set should still have WithMaxExclusive
	err := ruleSet.Apply(context.TODO(), "y", &output)
	if err == nil {
		t.Errorf("Expected error for WithMaxExclusive at threshold (exclusive)")
	}

	// New rule set should have WithMax (inclusive, so "x" should pass)
	err = ruleSet2.Apply(context.TODO(), "x", &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// Verify serialization shows the conflict resolution
	expected := "StringRuleSet.WithMax(\"x\")"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
