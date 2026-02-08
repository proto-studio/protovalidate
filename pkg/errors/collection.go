package errors

import "fmt"

// multiError holds multiple ValidationErrors and implements ValidationError by
// delegating Code(), Path(), etc. to the first error. Unwrap() returns the list.
type multiError struct {
	errs []ValidationError
}

var _ ValidationError = (*multiError)(nil)

// Unwrap returns the list of wrapped errors for use with errors.Is and errors.As. Nil receiver returns nil.
func (e *multiError) Unwrap() []error {
	if e == nil {
		return nil
	}
	out := make([]error, len(e.errs))
	for i := range e.errs {
		out[i] = e.errs[i]
	}
	return out
}

// Error returns the long-form message; for multiple errors, returns the first message plus a count.
func (e *multiError) Error() string {
	if len(e.errs) == 0 {
		return "(no validation errors)"
	}
	if len(e.errs) > 1 {
		return fmt.Sprintf("%s (and %d more)", e.errs[0].Error(), len(e.errs)-1)
	}
	return e.errs[0].Error()
}

// Code returns the first error's code, or empty if there are no errors.
func (e *multiError) Code() ErrorCode {
	if len(e.errs) == 0 {
		return ""
	}
	return e.errs[0].Code()
}

// Path returns the first error's path, or empty if there are no errors.
func (e *multiError) Path() string {
	if len(e.errs) == 0 {
		return ""
	}
	return e.errs[0].Path()
}

// PathAs returns the first error's path using the given serializer.
func (e *multiError) PathAs(serializer PathSerializer) string {
	if len(e.errs) == 0 {
		return ""
	}
	return e.errs[0].PathAs(serializer)
}

// ShortError returns the first error's short description.
func (e *multiError) ShortError() string {
	if len(e.errs) == 0 {
		return ""
	}
	return e.errs[0].ShortError()
}

// DocsURI returns the first error's documentation URI.
func (e *multiError) DocsURI() string {
	if len(e.errs) == 0 {
		return ""
	}
	return e.errs[0].DocsURI()
}

// TraceURI returns the first error's trace URI.
func (e *multiError) TraceURI() string {
	if len(e.errs) == 0 {
		return ""
	}
	return e.errs[0].TraceURI()
}

// Meta returns the first error's metadata.
func (e *multiError) Meta() map[string]any {
	if len(e.errs) == 0 {
		return nil
	}
	return e.errs[0].Meta()
}

// Params returns the first error's format params.
func (e *multiError) Params() []any {
	if len(e.errs) == 0 {
		return nil
	}
	return e.errs[0].Params()
}

// Internal returns true if any wrapped error is internal.
func (e *multiError) Internal() bool {
	for _, err := range e.errs {
		if err.Internal() {
			return true
		}
	}
	return false
}

// Permission returns true if any wrapped error is a permission error and none are internal.
func (e *multiError) Permission() bool {
	if e.Internal() {
		return false
	}
	for _, err := range e.errs {
		if err.Permission() {
			return true
		}
	}
	return false
}

// Validation returns true if there is at least one error and none are internal or permission.
func (e *multiError) Validation() bool {
	if len(e.errs) == 0 {
		return false
	}
	return !e.Internal() && !e.Permission()
}

// Join combines zero or more errors into a single ValidationError.
// Nil entries are skipped. Non-ValidationError entries are skipped.
// If an argument is a ValidationError that wraps multiple errors (Unwrap() non-empty), it is flattened so those errors are merged in. Single errors are added as-is.
// Returns nil for zero ValidationErrors, the single error unchanged for one, or a multiError for two or more.
func Join(errs ...error) ValidationError {
	var verrs []ValidationError
	for _, e := range errs {
		if e == nil {
			continue
		}
		ve, ok := e.(ValidationError)
		if !ok {
			continue
		}
		u := ve.Unwrap()
		if len(u) == 0 {
			verrs = append(verrs, ve)
		} else {
			for _, sub := range u {
				if v, ok := sub.(ValidationError); ok {
					verrs = append(verrs, v)
				}
			}
		}
	}
	switch len(verrs) {
	case 0:
		return nil
	case 1:
		return verrs[0]
	default:
		return &multiError{errs: verrs}
	}
}

// Unwrap returns the list of errors from err for iteration or len. Nil err returns nil.
// For a single error (err.Unwrap() is nil), returns []error{err}. Otherwise returns err.Unwrap().
func Unwrap(err ValidationError) []error {
	if err == nil {
		return nil
	}
	u := err.Unwrap()
	if len(u) == 0 {
		return []error{err}
	}
	return u
}

// For returns a ValidationError containing only the wrapped errors whose Path() equals path.
// If err is nil or no errors match, returns nil. If exactly one matches, returns that error;
// if multiple match, returns Join of the matches.
func For(err ValidationError, path string) ValidationError {
	unwrapped := Unwrap(err)
	if len(unwrapped) == 0 {
		return nil
	}
	var matched []error
	for _, e := range unwrapped {
		if ve, ok := e.(ValidationError); ok && ve.Path() == path {
			matched = append(matched, ve)
		}
	}
	return Join(matched...)
}

// ForPathAs is like For but compares paths using the given serializer (e.g. to filter by JSON Pointer or JSONPath).
// Use when the path string is in a different format than the default.
func ForPathAs(err ValidationError, path string, serializer PathSerializer) ValidationError {
	unwrapped := Unwrap(err)
	if len(unwrapped) == 0 {
		return nil
	}
	var matched []error
	for _, e := range unwrapped {
		if ve, ok := e.(ValidationError); ok && ve.PathAs(serializer) == path {
			matched = append(matched, ve)
		}
	}
	return Join(matched...)
}
