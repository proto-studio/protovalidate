package net

import (
	"context"
	"errors"
	"net/url"
	"testing"

	validateerrors "proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// TestQueryPercentEncodingRule_InvalidEncoding hits the error branch of queryPercentEncodingRule.
// The rule is only invoked with values.Encode() in production, which always produces valid encoding,
// so the error path is unreachable from public API. Same-package test calls the rule directly.
func TestQueryPercentEncodingRule_InvalidEncoding(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name  string
		value string
		want  bool // true = expect error
	}{
		{"empty", "", false},
		{"no percent", "a=b&c=d", false},
		{"valid percent", "a%20=b", false},
		{"valid two hex", "a%2f=b", false},
		{"percent at end", "%", true},
		{"percent one char", "%z", true},
		{"percent one hex", "%1", true},
		{"percent two non-hex", "%zz", true},
		{"percent hex then non-hex", "%1g", true},
		{"mid string invalid", "a%b", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queryPercentEncodingRule(ctx, tt.value)
			if tt.want {
				if err == nil {
					t.Error("expected encoding error, got nil")
					return
				}
				if err.Code() != validateerrors.CodeEncoding {
					t.Errorf("expected CodeEncoding, got %s", err.Code())
				}
			} else {
				if err != nil {
					t.Errorf("expected nil error, got %v", err)
				}
			}
		})
	}
}

// TestQueryRuleSet_Evaluate_NilSpec covers the "if spec == nil { continue }" branch.
// Production code always sets paramRules[name] = &paramSpec{}, so nil spec only possible via whitebox.
func TestQueryRuleSet_Evaluate_NilSpec(t *testing.T) {
	ctx := context.Background()
	q := &QueryRuleSet{
		paramRules: map[string]*paramSpec{
			"x": nil, // nil spec to hit continue
			"y": {ruleSet: rules.String().Any()},
		},
	}
	values := url.Values{"x": {"a"}, "y": {"b"}}
	err := q.Evaluate(ctx, values)
	if err != nil {
		t.Fatalf("expected nil (y is valid), got %v", err)
	}
}

// TestQueryRuleSet_Evaluate_NilRuleSet covers the branch where spec.ruleSet == nil (param ignored).
func TestQueryRuleSet_Evaluate_NilRuleSet(t *testing.T) {
	ctx := context.Background()
	q := Query().WithParam("ignored", nil) // ruleSet is nil
	values := url.Values{"ignored": {"any"}}
	err := q.Evaluate(ctx, values)
	if err != nil {
		t.Fatalf("expected nil (nil ruleSet skips validation), got %v", err)
	}
}

// TestQueryRuleSet_Apply_ParseError covers the parse-error branch in Apply by injecting a failing queryParser.
func TestQueryRuleSet_Apply_ParseError(t *testing.T) {
	ctx := context.Background()
	rs := Query()
	var out string
	saved := queryParser
	defer func() { queryParser = saved }()
	queryParser = func(string) (url.Values, error) {
		return nil, errors.New("injected parse error")
	}
	err := rs.Apply(ctx, "a=b", &out)
	if err == nil {
		t.Fatal("expected error when queryParser fails")
	}
	if err.Code() != validateerrors.CodeEncoding {
		t.Errorf("expected CodeEncoding, got %s", err.Code())
	}
}

// TestQueryRuleSet_Clone_WithOptions ensures clone(options...) runs the option loop (for coverage).
func TestQueryRuleSet_Clone_WithOptions(t *testing.T) {
	base := Query()
	// WithErrorMessage uses clone with a function option
	rs := base.WithErrorMessage("short", "long")
	if rs.String() == base.String() {
		t.Error("expected label to change")
	}
	// Chained clone with option
	rs2 := base.WithParam("q", rules.String().Any()).WithDocsURI("https://example.com")
	ctx := context.Background()
	var out string
	if err := rs2.Apply(ctx, "q=1", &out); err != nil {
		t.Fatalf("Apply: %v", err)
	}
}

// TestQueryRuleSet_Evaluate_EncodingCheckReturnsError covers the branch where defaultQueryStringRuleSet.Evaluate returns an error.
func TestQueryRuleSet_Evaluate_EncodingCheckReturnsError(t *testing.T) {
	ctx := context.Background()
	saved := defaultQueryStringRuleSet
	defer func() { defaultQueryStringRuleSet = saved }()
	defaultQueryStringRuleSet = rules.String().WithRuleFunc(func(context.Context, string) validateerrors.ValidationError {
		return validateerrors.Errorf(validateerrors.CodeEncoding, ctx, "bad", "injected")
	})
	q := Query()
	err := q.Evaluate(ctx, url.Values{"x": {"y"}})
	if err == nil {
		t.Fatal("expected error from injected encoding rule")
	}
	if err.Code() != validateerrors.CodeEncoding {
		t.Errorf("expected CodeEncoding, got %s", err.Code())
	}
}

// TestQueryRuleSet_Evaluate_ParamRuleSetReturnsError covers the branch where spec.ruleSet.Evaluate returns errors (append to allErrors).
func TestQueryRuleSet_Evaluate_ParamRuleSetReturnsError(t *testing.T) {
	ctx := context.Background()
	failRule := rules.String().WithRuleFunc(func(context.Context, string) validateerrors.ValidationError {
		return validateerrors.Errorf(validateerrors.CodePattern, ctx, "bad", "injected")
	}).Any()
	q := Query().WithParam("x", failRule)
	err := q.Evaluate(ctx, url.Values{"x": {"y"}})
	if err == nil {
		t.Fatal("expected error from param rule")
	}
	if err.Code() != validateerrors.CodePattern {
		t.Errorf("expected CodePattern, got %s", err.Code())
	}
}

// TestQueryRuleSet_Apply_WithURLValuesInput covers the url.Values branch in Apply (input type switch).
func TestQueryRuleSet_Apply_WithURLValuesInput(t *testing.T) {
	ctx := context.Background()
	rs := Query()
	var out string
	vals := url.Values{"a": {"b"}}
	err := rs.Apply(ctx, vals, &out)
	if err != nil {
		t.Fatalf("Apply with url.Values: %v", err)
	}
	if out != "a=b" {
		t.Errorf("expected a=b, got %q", out)
	}
}
