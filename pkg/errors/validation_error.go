package errors

import (
	"context"

	"proto.zip/studio/validate/pkg/rulecontext"
)

// ValidationError is the interface for validation errors. It extends the standard error
// interface with Unwrap() []error for multiple errors and rich metadata (Code, Path, etc.).
// When an error wraps multiple errors (Unwrap() returns non-empty), Code(), Path(), and
// similar methods return the first error's information.
type ValidationError interface {
	error

	// Unwrap returns the list of wrapped errors for use with errors.Is and errors.As.
	// Returns nil for a single error (no wrapped errors). Nil receiver returns nil.
	Unwrap() []error

	// Code returns the error code. For multi-errors, returns the first error's code.
	Code() ErrorCode

	// Path returns the full path to the error (e.g. "/field/subfield"). For multi-errors, returns the first error's path.
	Path() string

	// PathAs returns the path serialized with the given serializer (e.g. JSON Pointer, JSONPath).
	PathAs(serializer PathSerializer) string

	// ShortError returns a brief, constant description suitable for API responses.
	ShortError() string

	// DocsURI returns an optional documentation URL for this error code.
	DocsURI() string

	// TraceURI returns an optional trace or debug URL.
	TraceURI() string

	// Meta returns arbitrary key-value metadata attached to the error.
	Meta() map[string]any

	// Params returns the format arguments used to build the long error message.
	Params() []any

	// Internal returns true if any wrapped error is classified as internal (e.g. CodeInternal, CodeTimeout).
	Internal() bool

	// Validation returns true if all wrapped errors are validation errors (user input issues).
	Validation() bool

	// Permission returns true if any wrapped error is a permission/authorization error and none are internal.
	Permission() bool
}

// singleError is the concrete type for a single validation error.
type singleError struct {
	code        ErrorCode
	pathSegment rulecontext.PathSegment
	message     string
	docsURI     string
	traceURI    string
	shortMsg    string
	meta        map[string]any
	params      []any
}

// Ensure singleError implements ValidationError.
var _ ValidationError = (*singleError)(nil)

// Unwrap returns nil for a single error (no wrapped errors). Nil receiver returns nil.
func (err *singleError) Unwrap() []error {
	if err == nil {
		return nil
	}
	return nil
}

// Errorf creates a new ValidationError with explicit short and long messages.
// The long message is formatted with the provided args using the printer from context.
// Short messages should be constant strings (no formatting), long messages can contain format verbs.
//
// If an ErrorConfig is present in the context, its values override the provided short/long messages,
// code, URIs, and metadata. The callback is also applied if present.
func Errorf(code ErrorCode, ctx context.Context, short, long string, args ...interface{}) ValidationError {
	printer := rulecontext.Printer(ctx)
	segment := rulecontext.Path(ctx)
	config := ErrorConfigFromContext(ctx)

	// Apply config overrides if present
	actualCode := code
	actualShort := short
	actualLong := long
	var docsURI string
	var traceURI string
	var meta map[string]any

	if config != nil {
		if config.Code != nil {
			actualCode = *config.Code
		}
		if config.Short != "" {
			actualShort = config.Short
		}
		if config.Long != "" {
			actualLong = config.Long
		}
		docsURI = config.DocsURI
		traceURI = config.TraceURI
		meta = config.Meta
	}

	err := ValidationError(&singleError{
		code:        actualCode,
		pathSegment: segment,
		message:     printer.Sprintf(actualLong, args...),
		shortMsg:    actualShort,
		docsURI:     docsURI,
		traceURI:    traceURI,
		meta:        meta,
		params:      args,
	})

	// Apply callback if present
	if config != nil && config.Callback != nil {
		err = config.Callback(ctx, err)
	}

	return err
}

// Error creates a new ValidationError by looking up messages from the dictionary.
// It uses the error code to find the short and long descriptions in the dictionary,
// then formats the long description with the provided args.
func Error(code ErrorCode, ctx context.Context, args ...interface{}) ValidationError {
	dict := Dict(ctx)
	return Errorf(code, ctx, dict.ShortError(code), dict.ErrorPattern(code), args...)
}

// Error implements the error interface and returns the long-form message.
func (err *singleError) Error() string {
	return err.message
}

// Code returns the error code.
func (err *singleError) Code() ErrorCode {
	return err.code
}

// Path returns the full path using the default serializer.
func (err *singleError) Path() string {
	return err.PathAs(DefaultPathSerializer{})
}

// PathAs returns the path serialized with the given serializer.
func (err *singleError) PathAs(serializer PathSerializer) string {
	var segments []rulecontext.PathSegment
	if err.pathSegment != nil {
		segments = extractPathSegments(err.pathSegment)
	}
	return serializer.Serialize(segments)
}

// DocsURI returns the optional documentation URI.
func (err *singleError) DocsURI() string {
	return err.docsURI
}

// TraceURI returns the optional trace/debug URI.
func (err *singleError) TraceURI() string {
	return err.traceURI
}

// ShortError returns the brief, constant description.
func (err *singleError) ShortError() string {
	return err.shortMsg
}

// Meta returns the metadata map, if any.
func (err *singleError) Meta() map[string]any {
	return err.meta
}

// Params returns the format arguments for the long message.
func (err *singleError) Params() []any {
	return err.params
}

// Internal returns true if the error code is classified as internal.
func (err *singleError) Internal() bool {
	return DefaultDict().ErrorType(err.code) == ErrorTypeInternal
}

// Validation returns true if the error code is classified as a validation error.
func (err *singleError) Validation() bool {
	return DefaultDict().ErrorType(err.code) == ErrorTypeValidation
}

// Permission returns true if the error code is classified as a permission error.
func (err *singleError) Permission() bool {
	return DefaultDict().ErrorType(err.code) == ErrorTypePermission
}
