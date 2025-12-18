package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/rules"
)

// Test that WithMin and WithMore conflict with each other
func TestString_MinMoreConflict(t *testing.T) {
	ruleSet := rules.String().WithMin("b")

	// Adding WithMore should conflict and replace WithMin
	ruleSet2 := ruleSet.WithMore("c")

	var output string

	// Original rule set should still have WithMin
	err := ruleSet.Apply(context.TODO(), "b", &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMin at threshold, got %s", err)
	}

	// New rule set should have WithMore (exclusive, so "c" should fail)
	err = ruleSet2.Apply(context.TODO(), "c", &output)
	if err == nil {
		t.Errorf("Expected error for WithMore at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := "StringRuleSet.WithMore(\"c\")"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMax and WithLess conflict with each other
func TestString_MaxLessConflict(t *testing.T) {
	ruleSet := rules.String().WithMax("y")

	// Adding WithLess should conflict and replace WithMax
	ruleSet2 := ruleSet.WithLess("x")

	var output string

	// Original rule set should still have WithMax
	err := ruleSet.Apply(context.TODO(), "y", &output)
	if err != nil {
		t.Errorf("Expected error to be nil for WithMax at threshold, got %s", err)
	}

	// New rule set should have WithLess (exclusive, so "x" should fail)
	err = ruleSet2.Apply(context.TODO(), "x", &output)
	if err == nil {
		t.Errorf("Expected error for WithLess at threshold (exclusive)")
	}

	// Verify serialization shows the conflict resolution
	expected := "StringRuleSet.WithLess(\"x\")"
	if s := ruleSet2.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Test that WithMore and WithMin conflict with each other (reverse order)
func TestString_MoreMinConflict(t *testing.T) {
	ruleSet := rules.String().WithMore("b")

	// Adding WithMin should conflict and replace WithMore
	ruleSet2 := ruleSet.WithMin("c")

	var output string

	// Original rule set should still have WithMore
	err := ruleSet.Apply(context.TODO(), "b", &output)
	if err == nil {
		t.Errorf("Expected error for WithMore at threshold (exclusive)")
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

// Test that WithLess and WithMax conflict with each other (reverse order)
func TestString_LessMaxConflict(t *testing.T) {
	ruleSet := rules.String().WithLess("y")

	// Adding WithMax should conflict and replace WithLess
	ruleSet2 := ruleSet.WithMax("x")

	var output string

	// Original rule set should still have WithLess
	err := ruleSet.Apply(context.TODO(), "y", &output)
	if err == nil {
		t.Errorf("Expected error for WithLess at threshold (exclusive)")
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
