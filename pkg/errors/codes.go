package errors

// Error codes allow identifying of the error without having to do string comparison.
//
// All user defined and module errors should have a number greater than 1000.
type ErrorCode string

const (
	CodeUnknown    ErrorCode = "UNKNOWN"    // The cause of the validation error was not specified.
	CodeInternal   ErrorCode = "INTERNAL"   // An internal error ocurred. We may know the reason but should not convey that to the user.
	CodeTimeout    ErrorCode = "TIMEOUT"    // The request timed out before validation could be completed.
	CodeCancelled  ErrorCode = "CANCELED"   // The request was cancelled before it could be completed.
	CodeType       ErrorCode = "TYPE"       // Unable to coerce a value to the correct type.
	CodeRange      ErrorCode = "RANGE"      // The data falls outside the range allowed by the type.
	CodeRequired   ErrorCode = "REQUIRED"   // Value is required to not be nil.
	CodeUnexpected ErrorCode = "UNEXPECTED" // Value was not expected to be defined.
	CodeMin        ErrorCode = "MIN"        // Value does not satisfy minimum constraints.
	CodeMax        ErrorCode = "MAX"        // Value does not satisfy maximum constraints.
	CodePattern    ErrorCode = "PATTERN"    // Value does not match an expected pattern or expression.
	CodeExpired    ErrorCode = "EXPIRED"    // Value has expired
	CodeForbidden  ErrorCode = "DENIED"     // Value is in a list of forbidden values.
	CodeNotAllowed ErrorCode = "NOTALLOWED" // Value is not one of the allowed values.
)
