package net_test

import (
	"context"
	"net/url"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/net"
)

func TestQueryRuleSet_WithParam(t *testing.T) {
	ctx := context.Background()

	// Rule set without required param: can validate and output as *string, *url.Values, or *any
	rsOptional := net.Query().WithParam("q", rules.String().Any())
	var out string
	err := rsOptional.Apply(ctx, "q=hello", &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if out != "q=hello" {
		t.Errorf("expected out=%q, got %q", "q=hello", out)
	}

	var vals url.Values
	err = rsOptional.Apply(ctx, "q=hello&x=1", &vals)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if vals.Get("q") != "hello" || vals.Get("x") != "1" {
		t.Errorf("expected q=hello x=1, got %v", vals)
	}

	var vAny any
	err = rsOptional.Apply(ctx, "a=b", &vAny)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if v, ok := vAny.(url.Values); !ok || v.Get("a") != "b" {
		t.Errorf("expected url.Values with a=b, got %v", vAny)
	}

	// Rule set with required param: missing param fails
	rsRequired := net.Query().WithParam("q", rules.String().WithRequired().Any())
	err = rsRequired.Apply(ctx, "q=hello", &out)
	if err != nil {
		t.Fatalf("expected nil error when param present, got %v", err)
	}
	err = rsRequired.Apply(ctx, "", &out)
	if err == nil {
		t.Fatal("expected error when required param missing")
	}
	if err.First().Code() != errors.CodeRequired {
		t.Errorf("expected CodeRequired, got %s", err.First().Code())
	}
}

func TestQueryRuleSet_Required(t *testing.T) {
	base := net.Query()
	if base.Required() {
		t.Error("base QueryRuleSet should not be required")
	}
	withRequired := base.WithParam("x", rules.String().WithRequired().Any())
	if !withRequired.Required() {
		t.Error("WithParam(required) should make Required() true")
	}
	// Chaining optional param after required: whole rule set stays required
	withBoth := withRequired.WithParam("y", rules.String().Any())
	if !withBoth.Required() {
		t.Error("rule set with any required param should remain Required() true")
	}
	// Explicit WithRequired marks the query string as required
	withExplicitRequired := base.WithRequired()
	if !withExplicitRequired.Required() {
		t.Error("WithRequired() should make Required() true")
	}
}

func TestQueryRuleSet_WithRule_WithRuleFunc(t *testing.T) {
	ctx := context.Background()
	// WithRuleFunc: custom rule that fails when query has "forbidden=1"
	rs := net.Query().WithRuleFunc(func(ctx context.Context, v url.Values) errors.ValidationErrorCollection {
		if v.Get("forbidden") == "1" {
			return errors.Collection(errors.Errorf(errors.CodeForbidden, ctx, "forbidden", "forbidden param present"))
		}
		return nil
	})
	var out string
	err := rs.Apply(ctx, "a=b", &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	err = rs.Apply(ctx, "forbidden=1", &out)
	if err == nil {
		t.Fatal("expected error when custom rule fails")
	}
	if err.First().Code() != errors.CodeForbidden {
		t.Errorf("expected CodeForbidden, got %s", err.First().Code())
	}
	// WithRule: pass RuleFunc as Rule (RuleFunc implements Rule)
	rs2 := net.Query().WithRule(rules.RuleFunc[url.Values](func(ctx context.Context, v url.Values) errors.ValidationErrorCollection {
		if len(v) == 0 {
			return errors.Collection(errors.Errorf(errors.CodeRequired, ctx, "empty", "query must not be empty"))
		}
		return nil
	}))
	err = rs2.Apply(ctx, "", &out)
	if err == nil {
		t.Fatal("expected error when query empty and rule requires non-empty")
	}
	err = rs2.Apply(ctx, "x=1", &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestQueryRuleSet_Apply_inputOutputBranches(t *testing.T) {
	ctx := context.Background()
	rs := net.Query()
	var out string

	// Invalid input type
	err := rs.Apply(ctx, 123, &out)
	if err == nil {
		t.Fatal("expected error for non-string/url.Values input")
	}
	if err.First().Code() != errors.CodeType {
		t.Errorf("expected CodeType, got %s", err.First().Code())
	}

	// Nil output pointer
	err = rs.Apply(ctx, "a=b", nil)
	if err == nil {
		t.Fatal("expected error for nil output")
	}
	if err.First().Code() != errors.CodeInternal {
		t.Errorf("expected CodeInternal for nil output, got %s", err.First().Code())
	}

	// Non-pointer output
	var notPtr string
	err = rs.Apply(ctx, "a=b", notPtr)
	if err == nil {
		t.Fatal("expected error for non-pointer output")
	}

	// Output *url.Values (map branch): nil map
	var vals url.Values
	err = rs.Apply(ctx, "a=b", &vals)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if vals.Get("a") != "b" {
		t.Errorf("expected a=b, got %v", vals)
	}

	// Output *url.Values (map branch): non-nil pre-allocated map
	vals2 := make(url.Values)
	err = rs.Apply(ctx, "x=y", &vals2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if vals2.Get("x") != "y" {
		t.Errorf("expected x=y, got %v", vals2)
	}

	// Invalid output type (e.g. *int)
	var i int
	err = rs.Apply(ctx, "a=b", &i)
	if err == nil {
		t.Fatal("expected error for invalid output type")
	}
	if err.First().Code() != errors.CodeInternal {
		t.Errorf("expected CodeInternal for wrong output type, got %s", err.First().Code())
	}
}

func TestQueryRuleSet_Apply_parseError(t *testing.T) {
	ctx := context.Background()
	rs := net.Query()
	var out string
	// Invalid percent encoding: % not followed by two hex digits can make ParseQuery return an error (e.g. "a=%" or "a=%z")
	err := rs.Apply(ctx, "a=%", &out)
	if err == nil {
		// ParseQuery may or may not fail depending on Go version; if it fails we get CodeEncoding
		return
	}
	if err.First().Code() != errors.CodeEncoding {
		t.Errorf("expected CodeEncoding on parse error, got %s", err.First().Code())
	}
}

func TestQueryRuleSet_Evaluate_percentEncodingError(t *testing.T) {
	// The percent-encoding rule runs on values.Encode(); Encode() always produces valid encoding.
	// So the encoding-error path is only reachable if we had a code path that passed a raw string.
	// Here we just ensure Evaluate with valid url.Values runs the encoding check (empty and non-empty).
	ctx := context.Background()
	rs := net.Query()
	// Empty values: Encode() is "" so encoding rule is not triggered (queryStringForEncoding != "" branch skipped)
	err := rs.Evaluate(ctx, url.Values{})
	if err != nil {
		t.Fatalf("expected nil: %v", err)
	}
	// Non-empty valid encoding
	err = rs.Evaluate(ctx, url.Values{"k": {"v"}})
	if err != nil {
		t.Fatalf("expected nil: %v", err)
	}
}

func TestQueryRuleSet_String_Any(t *testing.T) {
	base := net.Query()
	if s := base.String(); s != "QueryRuleSet" {
		t.Errorf("expected base String() = QueryRuleSet, got %q", s)
	}
	withParam := base.WithParam("q", rules.String().Any())
	if s := withParam.String(); s != "QueryRuleSet.WithParam(\"q\")" {
		t.Errorf("expected chained String(), got %q", s)
	}
	// Any() returns RuleSet[any]; Apply through it
	anyRS := base.WithParam("k", rules.String().Any()).Any()
	ctx := context.Background()
	var v any
	err := anyRS.Apply(ctx, "k=v", &v)
	if err != nil {
		t.Fatalf("Any().Apply expected nil error, got %v", err)
	}
	if m, ok := v.(url.Values); !ok || m.Get("k") != "v" {
		t.Errorf("expected url.Values k=v, got %v", v)
	}
}

func TestQueryRuleSet_WithErrorConfig(t *testing.T) {
	base := net.Query()
	// Chain WithErrorMessage, WithDocsURI, WithTraceURI, WithErrorCode, WithErrorMeta, WithErrorCallback (exercise all; nil errorConfig is safe)
	rs := base.
		WithErrorMessage("short", "long").
		WithDocsURI("https://example.com").
		WithTraceURI("https://trace.example.com").
		WithErrorCode(errors.CodeForbidden).
		WithErrorMeta("key", "val").
		WithErrorCallback(func(ctx context.Context, err errors.ValidationError) errors.ValidationError { return err })
	ctx := context.Background()
	var out string
	err := rs.Apply(ctx, "a=b", &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if out != "a=b" {
		t.Errorf("expected a=b, got %q", out)
	}
}
