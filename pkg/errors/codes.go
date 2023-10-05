package errors

// Error codes allow identifying of the error without having to do string comparison.
//
// All user defined and module errors should have a number greater than 1000.
type ErrorCode string

const (
	CodeUnknown    ErrorCode = "UNKNOWN"    // The cause of the validation error was not specified.
	CodeType                 = "TYPE"       // Unable to coerce a value to the correct type.
	CodeRange                = "RANGE"      // The data falls outside the range allowed by the type.
	CodeRequired             = "REQUIRED"   // Value is required to not be nil.
	CodeUnexpected           = "UNEXPECTED" // Value was not expected to be defined.
	CodeMin                  = "MIN"        // Value does not satisfy minimum constraints.
	CodeMax                  = "MAX"        // Value does not satisfy maximum constraints.
	CodePattern              = "PATTERN"    // Value does not match an expected pattern or expression.
	CodeExpired              = "EXPIRED"    // Value has expired
	CodeForbidden            = "DENIED"     // Value is in a list of forbidden values.
	CodeNotAllowed           = "NOTALLOWED" // Value is not one of the allowed values.
)
