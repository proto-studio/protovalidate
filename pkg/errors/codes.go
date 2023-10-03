package errors

// Error codes allow identifying of the error without having to do string comparison.
//
// All user defined and module errors should have a number greater than 1000.
type ErrorCode string

const (
	CodeUnknown    ErrorCode = "UNK" // The cause of the validation error was not specified.
	CodeType                 = "TYP" // Unable to coerce a value to the correct type.
	CodeRange                = "RNG" // The data falls outside the range allowed by the type.
	CodeRequired             = "REQ" // Value is required to not be nil.
	CodeUnexpected           = "UNE" // Value was not expected to be defined.
	CodeMin                  = "MIN" // Value does not satisfy minimum constraints.
	CodeMax                  = "MAX" // Value does not satisfy maximum constraints.
	CodePattern              = "PAT" // Value does not match an expected pattern or expression.
	CodeExpired              = "EXP" // Value has expired
)
