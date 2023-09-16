package rulecontext_test

import (
	"context"
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"proto.zip/studio/validate/pkg/rulecontext"
)

func TestReturnsPrinter(t *testing.T) {
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

	ctx = context.WithValue(ctx, "keyA", "valA")
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

	v := ctx.Value("keyA")
	if v == nil {
		t.Error("Expected keyA to not be nil")
	} else if v != "valA" {
		t.Errorf("Expected valB for keyA but got %s", v)
	}
}

func TestWithPrinterNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	rulecontext.WithPrinter(context.Background(), nil)
}

func TestReturnsRuleSet(t *testing.T) {
	v := rulecontext.RuleSet(nil)
	if v != nil {
		t.Error("Expected rule set to be nil")
	}

	ctx := context.Background()
	v = rulecontext.RuleSet(ctx)
	if v != nil {
		t.Error("Expected rule set to be nil")
	}

	ctx = context.WithValue(ctx, "keyA", "valA")
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

func TestWithRuleSetil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	rulecontext.WithRuleSet(context.Background(), nil)
}

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

	ctx = context.WithValue(ctx, "keyA", "valA")
	p = rulecontext.Path(ctx)
	if p.String() != segmentA {
		t.Errorf("Expected path segment to be `%s` got `%s`", segmentA, p.String())
	}

	ctx = rulecontext.WithPathString(ctx, segmentB)
	p = rulecontext.Path(ctx)
	expectedFullPath := segmentA + "." + segmentB

	if p == nil {
		t.Error("Expected path segment to not be nil")
	} else if p.String() != segmentB {
		t.Errorf("Expected  path segment to be `%s` got `%s`", segmentB, p.String())
	} else if p.FullString() != expectedFullPath {
		t.Errorf("Expected full path to be `%s` got `%s`", expectedFullPath, p.FullString())
	}
}
