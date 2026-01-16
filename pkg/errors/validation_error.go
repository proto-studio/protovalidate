package errors

import (
	"context"

	"proto.zip/studio/validate/pkg/rulecontext"
)

// ValidationError stores information necessary to identify where the validation error
// is, as well as implementing the Error interface to work with standard errors.
type ValidationError interface {
	Code() ErrorCode                         // Code returns the error code.
	Path() string                            // Path returns the full path to the error as a string.
	PathAs(serializer PathSerializer) string // PathAs returns the full path using the provided serializer.
	Error() string                           // Error returns the detailed error message (satisfies Go error interface).
	ShortError() string                      // ShortError returns a brief constant error description.
	DocsURI() string                         // DocsURI returns an optional documentation URL for the error.
	TraceURI() string                        // TraceURI returns an optional trace/debug URL for the error.
	Meta() map[string]any                    // Meta returns arbitrary metadata associated with the error.
	Params() []any                           // Params returns the format arguments used to create the error message.

	// Type classification methods
	Internal() bool   // Internal returns true if the error is an internal error.
	Validation() bool // Validation returns true if the error is a validation error.
	Permission() bool // Permission returns true if the error is a permission error.
}

// validationError implements a standard Error interface and also ValidationError interface
// while preserving the validation data.
type validationError struct {
	code        ErrorCode               // Error code helps identify the error without string comparisons.
	pathSegment rulecontext.PathSegment // The leaf path segment (nil if no path).
	message     string                  // The error message (long description) converted to the context locale.
	docsURI     string                  // Optional documentation URL.
	traceURI    string                  // Optional trace/debug URL.
	shortMsg    string                  // Brief constant description.
	meta        map[string]any          // Arbitrary metadata.
	params      []any                   // Format arguments used to create the message.
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

	err := ValidationError(&validationError{
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

// Error implements the standard Error interface to return a string for validation errors.
// Error loses contextual data, so use the ValidationError object when possible.
func (err *validationError) Error() string {
	return err.message
}

// Code returns the error code.
// Code can be used to look up the error without relying on string checks.
func (err *validationError) Code() ErrorCode {
	return err.code
}

// Path returns the full path to the error as a string.
// Path is a wrapper around PathAs using the default serializer.
func (err *validationError) Path() string {
	return err.PathAs(DefaultPathSerializer{})
}

// PathAs returns the full path to the error as a string using the provided serializer.
func (err *validationError) PathAs(serializer PathSerializer) string {
	var segments []rulecontext.PathSegment
	if err.pathSegment != nil {
		segments = extractPathSegments(err.pathSegment)
	}
	return serializer.Serialize(segments)
}

// DocsURI returns an optional documentation URL for the error.
func (err *validationError) DocsURI() string {
	return err.docsURI
}

// TraceURI returns an optional trace/debug URL for the error.
func (err *validationError) TraceURI() string {
	return err.traceURI
}

// ShortError returns a brief constant error description.
func (err *validationError) ShortError() string {
	return err.shortMsg
}

// Meta returns arbitrary metadata associated with the error.
func (err *validationError) Meta() map[string]any {
	return err.meta
}

// Params returns the format arguments used to create the error message.
// This allows callbacks to access the original values for custom formatting.
func (err *validationError) Params() []any {
	return err.params
}

// Internal returns true if the error is classified as an internal error.
func (err *validationError) Internal() bool {
	return DefaultDict().ErrorType(err.code) == ErrorTypeInternal
}

// Validation returns true if the error is classified as a validation error.
func (err *validationError) Validation() bool {
	return DefaultDict().ErrorType(err.code) == ErrorTypeValidation
}

// Permission returns true if the error is classified as a permission error.
func (err *validationError) Permission() bool {
	return DefaultDict().ErrorType(err.code) == ErrorTypePermission
}
