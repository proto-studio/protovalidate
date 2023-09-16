package errors

// Error codes allow identifying of the error without having to do string comparison.
//
// All user defined and module errors should have a number greater than 1000.
type ErrorCode int

const (
	CodeUnknown    ErrorCode = iota // The cause of the validation error was not specified.
	CodeType                        // Unable to coerce a value to the correct type.
	CodeRange                       // The data falls outside the range allowed by the type.
	CodeRequired                    // Value is required to not be nil.
	CodeUnexpected                  // Value was not expected to be defined.
	CodeMin                         // Value does not satisfy minimum constraints.
	CodeMax                         // Value does not satisfy maximum constraints.
	CodePattern                     // Value does not match an expected pattern or expression.
)
