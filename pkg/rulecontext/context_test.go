package rulecontext_test

import (
	"context"
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// testContextKey is a custom type for test context keys to avoid key collisions.
type testContextKey string

const (
	testContextKeyA testContextKey = "keyA"
)

// TestReturnsPrinter tests:
// - Default printer is returned when context is nil
// - Default printer is returned when no printer is set in context
// - Custom printer is returned when set in context
// - Context values are preserved when setting printer
func TestReturnsPrinter(t *testing.T) {
	//lint:ignore SA1012 Testing nil context handling
	defaultPrinter := rulecontext.Printer(nil)
	if defaultPrinter == nil {
		t.Error("Expected default printer to not be nil")
		return
	}

	ctx := context.Background()
	p := rulecontext.Printer(ctx)
	if p == nil {
		t.Error("Expected printer to not be nil")
	} else if p != defaultPrinter {
		t.Error("Expected default printer")
	}

	ctx = context.WithValue(ctx, testContextKeyA, "valA")
	p = rulecontext.Printer(ctx)
	if p == nil {
		t.Error("Expected printer to not be nil")
	} else if p != defaultPrinter {
		t.Error("Expected default printer")
	}

	ctx = rulecontext.WithPrinter(ctx, message.NewPrinter(language.Spanish))
	p = rulecontext.Printer(ctx)
	if p == nil {
		t.Error("Expected printer to not be nil")
	} else if p == defaultPrinter {
		t.Error("Expected non-default printer")
	}

	v := ctx.Value(testContextKeyA)
	if v == nil {
		t.Error("Expected keyA to not be nil")
	} else if v != "valA" {
		t.Errorf("Expected valB for keyA but got %s", v)
	}
}

// TestWithPrinterNil tests:
// - Panics when nil printer is provided
func TestWithPrinterNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	rulecontext.WithPrinter(context.Background(), nil)
}

// TestReturnsRuleSet tests:
// - Returns nil when context is nil
// - Returns nil when no rule set is set in context
// - Returns rule set when set in context
func TestReturnsRuleSet(t *testing.T) {
	//lint:ignore SA1012 Testing nil context handling
	v := rulecontext.RuleSet(nil)
	if v != nil {
		t.Error("Expected rule set to be nil")
	}

	ctx := context.Background()
	v = rulecontext.RuleSet(ctx)
	if v != nil {
		t.Error("Expected rule set to be nil")
	}

	ctx = context.WithValue(ctx, testContextKeyA, "valA")
	v = rulecontext.RuleSet(ctx)
	if v != nil {
		t.Error("Expected rule set to be nil")
	}

	ctx = rulecontext.WithRuleSet(ctx, 123)
	v = rulecontext.RuleSet(ctx)
	if v == nil {
		t.Error("Expected rule set to not be nil")
	} else if v != 123 {
		t.Error("Expected rule set")
	}
}

// TestWithRuleSetil tests:
// - Panics when nil rule set is provided
func TestWithRuleSetil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	rulecontext.WithRuleSet(context.Background(), nil)
}

// TestPathRuleSet tests:
// - Path is correctly set in context
func TestPathRuleSet(t *testing.T) {
	const segmentA = "test"
	const segmentB = "123"

	ctx := context.Background()
	p := rulecontext.Path(ctx)
	if p != nil {
		t.Error("Expected path segment to be nil")
	}

	ctx = rulecontext.WithPathString(ctx, segmentA)
	p = rulecontext.Path(ctx)
	if p == nil {
		t.Error("Expected path segment to not be nil")
	} else if p.String() != segmentA {
		t.Errorf("Expected path segment to be `%s` got `%s`", segmentA, p.String())
	}

	ctx = context.WithValue(ctx, testContextKeyA, "valA")
	p = rulecontext.Path(ctx)
	if p.String() != segmentA {
		t.Errorf("Expected path segment to be `%s` got `%s`", segmentA, p.String())
	}

	ctx = rulecontext.WithPathString(ctx, segmentB)
	p = rulecontext.Path(ctx)
	expectedFullPath := "/" + segmentA + "/" + segmentB

	if p == nil {
		t.Error("Expected path segment to not be nil")
	} else if p.String() != segmentB {
		t.Errorf("Expected  path segment to be `%s` got `%s`", segmentB, p.String())
	} else if p.FullString() != expectedFullPath {
		t.Errorf("Expected full path to be `%s` got `%s`", expectedFullPath, p.FullString())
	}
}
