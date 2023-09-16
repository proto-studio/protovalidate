package errors

import (
	"context"

	"proto.zip/studio/validate/pkg/rulecontext"
)

// ValidationError stores information necessary to identify where the validation error
// is, as well as implementing the Error interface to work with standard errors.
type ValidationError interface {
	Code() ErrorCode // Code returns the error code.
	Path() string    // Path returns the full path to the error in the data structure.
	Error() string   // Error returns the error message.
}

// validationError implements a standard Error interface and also ValidationError interface
// while preserving the validation data.
type validationError struct {
	code    ErrorCode // Error code helps identify the error without string comparisons.
	path    string    // The full path to the error separated by dots.
	message string    // The error message converted to the context locale.
}

// New instantiates a validator error given a code, path, and message.
func New(code ErrorCode, path, message string) ValidationError {
	return &validationError{
		code:    code,
		path:    path,
		message: message,
	}
}

// Errorf instantiates a new error given context and a format string.
// This uses message.Sprintf to format the message.
func Errorf(code ErrorCode, ctx context.Context, key string, args ...interface{}) ValidationError {
	printer := rulecontext.Printer(ctx)
	segment := rulecontext.Path(ctx)

	if segment == nil {
		return New(code, "", printer.Sprintf(key, args...))
	}

	return New(code, segment.FullString(), printer.Sprintf(key, args...))
}

// Error implements the standard Error interface to return a string for validation errors.
// When possible you should use the ValidationError object since this method loses contextual data.
func (err *validationError) Error() string {
	return err.message
}

// Code returns the error code. It can be used to look up the error without relying on string checks.
func (err *validationError) Code() ErrorCode {
	return err.code
}

// Path returns the full path to the error in the data structure.
func (err *validationError) Path() string {
	return err.path
}
