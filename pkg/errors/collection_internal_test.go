package errors

import (
	"context"
	"testing"
)

// TestMultiErrorEmptyErrList covers multiError methods when errs is empty (defensive branches).
// Join() never returns an empty multiError, but the type supports it.
func TestMultiErrorEmptyErrList(t *testing.T) {
	e := &multiError{errs: nil}
	if out := e.Unwrap(); len(out) != 0 {
		t.Errorf("Unwrap() on empty multiError should have length 0, got %v", out)
	}
	if msg := e.Error(); msg != "(no validation errors)" {
		t.Errorf("Error() = %q", msg)
	}
	if c := e.Code(); c != "" {
		t.Errorf("Code() = %q", c)
	}
	if p := e.Path(); p != "" {
		t.Errorf("Path() = %q", p)
	}
	if p := e.PathAs(DefaultPathSerializer{}); p != "" {
		t.Errorf("PathAs() = %q", p)
	}
	if s := e.ShortError(); s != "" {
		t.Errorf("ShortError() = %q", s)
	}
	if u := e.DocsURI(); u != "" {
		t.Errorf("DocsURI() = %q", u)
	}
	if u := e.TraceURI(); u != "" {
		t.Errorf("TraceURI() = %q", u)
	}
	if m := e.Meta(); m != nil {
		t.Errorf("Meta() = %v", m)
	}
	if params := e.Params(); params != nil {
		t.Errorf("Params() = %v", params)
	}
	if e.Validation() {
		t.Error("Validation() on empty multiError should be false")
	}
}

// TestMultiErrorSingleErr covers multiError.Error() when len(errs)==1 (single message, no "and N more").
func TestMultiErrorSingleErr(t *testing.T) {
	ctx := context.Background()
	single := Errorf(CodeMin, ctx, "short", "long %d", 10)
	e := &multiError{errs: []ValidationError{single}}
	if msg := e.Error(); msg != single.Error() {
		t.Errorf("Error() = %q, want %q", msg, single.Error())
	}
}

// TestMultiErrorNilReceiver covers multiError.Unwrap() with nil receiver.
func TestMultiErrorNilReceiver(t *testing.T) {
	var e *multiError
	if out := e.Unwrap(); out != nil {
		t.Errorf("Unwrap() on nil multiError should return nil, got %v", out)
	}
}

// TestSingleErrorNilReceiverUnwrap covers singleError.Unwrap() with nil receiver.
func TestSingleErrorNilReceiverUnwrap(t *testing.T) {
	var ve ValidationError = (*singleError)(nil)
	if out := ve.Unwrap(); out != nil {
		t.Errorf("Unwrap() on nil singleError should return nil, got %v", out)
	}
}
