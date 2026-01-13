package errors

import "context"

// dictContextKey is the context key for storing the error dictionary.
var dictContextKey int

// WithDict adds an ErrorDict to the context.
func WithDict(parent context.Context, dict ErrorDict) context.Context {
	if dict == nil {
		panic("expected dict to not be nil")
	}
	return context.WithValue(parent, &dictContextKey, dict)
}

// Dict returns the ErrorDict from the context.
// If no dictionary is found in the context, it returns the DefaultDict.
//
// Dict never returns nil.
func Dict(ctx context.Context) ErrorDict {
	if ctx == nil {
		return defaultDict
	}

	dict := ctx.Value(&dictContextKey)
	if dict != nil {
		return dict.(ErrorDict)
	}

	return defaultDict
}

// ErrorEntry holds all metadata for an error code.
type ErrorEntry struct {
	Type         ErrorType // Classification of the error (validation, permission, internal)
	ShortError   string    // Brief constant description (e.g., "below minimum")
	ErrorPattern string    // Detailed template for i18n (e.g., "must be at least %v")
}

// ErrorDict provides lookup and customization of error metadata.
type ErrorDict interface {
	// ErrorType returns the error type classification for a code.
	ErrorType(code ErrorCode) ErrorType

	// ShortError returns the short description for a code.
	ShortError(code ErrorCode) string

	// ErrorPattern returns the long description template for a code.
	ErrorPattern(code ErrorCode) string

	// Entry returns the full ErrorEntry for a code.
	Entry(code ErrorCode) ErrorEntry

	// WithCode returns a new ErrorDict with the specified code entry overridden.
	WithCode(code ErrorCode, entry ErrorEntry) ErrorDict
}

// errorDict is the default implementation of ErrorDict.
type errorDict struct {
	parent  *errorDict
	entries map[ErrorCode]ErrorEntry
}

// Short description constants (constant strings, no formatting).
const (
	shortUnknown       = "unknown error"
	shortInternal      = "internal error"
	shortTimeout       = "timeout"
	shortCancelled     = "cancelled"
	shortType          = "invalid type"
	shortRange         = "out of range"
	shortRequired      = "required"
	shortNull          = "null not allowed"
	shortUnexpected    = "unexpected"
	shortMin           = "below minimum"
	shortMax           = "above maximum"
	shortMinExclusive  = "below minimum"
	shortMaxExclusive  = "above maximum"
	shortMinLen        = "too short"
	shortMaxLen        = "too long"
	shortPattern       = "invalid format"
	shortEncoding      = "invalid encoding"
	shortExpired       = "expired"
	shortForbidden     = "forbidden"
	shortNotAllowed    = "not allowed"
	shortValueMismatch = "value mismatch"
)

// Long description message templates for i18n (can contain format verbs).
const (
	longUnknown       = "an unknown error occurred"
	longInternal      = "an internal error occurred"
	longTimeout       = "operation timed out"
	longCancelled     = "operation was cancelled"
	longType          = "expected %s but got %s"
	longRange         = "value is out of range for %s"
	longRequired      = "value is required"
	longNull          = "value cannot be null"
	longUnexpected    = "value was not expected"
	longMin           = "must be at least %v"
	longMax           = "must be at most %v"
	longMinExclusive  = "must be greater than %v"
	longMaxExclusive  = "must be less than %v"
	longMinLen        = "length must be at least %d"
	longMaxLen        = "length must be at most %d"
	longPattern       = "value does not match the required format"
	longEncoding      = "value is not properly encoded"
	longExpired       = "value has expired"
	longForbidden     = "value is forbidden"
	longNotAllowed    = "value is not one of the allowed options"
	longValueMismatch = "value does not match"
)

// defaultEntries contains the built-in error metadata.
var defaultEntries = map[ErrorCode]ErrorEntry{
	CodeUnknown: {
		Type:         ErrorTypeInternal,
		ShortError:   shortUnknown,
		ErrorPattern: longUnknown,
	},
	CodeInternal: {
		Type:         ErrorTypeInternal,
		ShortError:   shortInternal,
		ErrorPattern: longInternal,
	},
	CodeTimeout: {
		Type:         ErrorTypeInternal,
		ShortError:   shortTimeout,
		ErrorPattern: longTimeout,
	},
	CodeCancelled: {
		Type:         ErrorTypeInternal,
		ShortError:   shortCancelled,
		ErrorPattern: longCancelled,
	},
	CodeType: {
		Type:         ErrorTypeValidation,
		ShortError:   shortType,
		ErrorPattern: longType,
	},
	CodeRange: {
		Type:         ErrorTypeValidation,
		ShortError:   shortRange,
		ErrorPattern: longRange,
	},
	CodeRequired: {
		Type:         ErrorTypeValidation,
		ShortError:   shortRequired,
		ErrorPattern: longRequired,
	},
	CodeUnexpected: {
		Type:         ErrorTypeValidation,
		ShortError:   shortUnexpected,
		ErrorPattern: longUnexpected,
	},
	CodeMin: {
		Type:         ErrorTypeValidation,
		ShortError:   shortMin,
		ErrorPattern: longMin,
	},
	CodeMax: {
		Type:         ErrorTypeValidation,
		ShortError:   shortMax,
		ErrorPattern: longMax,
	},
	CodeMinExclusive: {
		Type:         ErrorTypeValidation,
		ShortError:   shortMinExclusive,
		ErrorPattern: longMinExclusive,
	},
	CodeMaxExclusive: {
		Type:         ErrorTypeValidation,
		ShortError:   shortMaxExclusive,
		ErrorPattern: longMaxExclusive,
	},
	CodeMinLen: {
		Type:         ErrorTypeValidation,
		ShortError:   shortMinLen,
		ErrorPattern: longMinLen,
	},
	CodeMaxLen: {
		Type:         ErrorTypeValidation,
		ShortError:   shortMaxLen,
		ErrorPattern: longMaxLen,
	},
	CodePattern: {
		Type:         ErrorTypeValidation,
		ShortError:   shortPattern,
		ErrorPattern: longPattern,
	},
	CodeExpired: {
		Type:         ErrorTypeValidation,
		ShortError:   shortExpired,
		ErrorPattern: longExpired,
	},
	CodeForbidden: {
		Type:         ErrorTypePermission,
		ShortError:   shortForbidden,
		ErrorPattern: longForbidden,
	},
	CodeNotAllowed: {
		Type:         ErrorTypePermission,
		ShortError:   shortNotAllowed,
		ErrorPattern: longNotAllowed,
	},
	CodeEncoding: {
		Type:         ErrorTypeValidation,
		ShortError:   shortEncoding,
		ErrorPattern: longEncoding,
	},
	CodeNull: {
		Type:         ErrorTypeValidation,
		ShortError:   shortNull,
		ErrorPattern: longNull,
	},
}

// defaultDict is the singleton default ErrorDict instance.
var defaultDict = &errorDict{
	entries: defaultEntries,
}

// DefaultDict returns the built-in error dictionary.
func DefaultDict() ErrorDict {
	return defaultDict
}

// NewDict creates a new empty dict that inherits from DefaultDict.
func NewDict() ErrorDict {
	return &errorDict{
		parent:  defaultDict,
		entries: make(map[ErrorCode]ErrorEntry),
	}
}

// Entry returns the full ErrorEntry for a code.
func (d *errorDict) Entry(code ErrorCode) ErrorEntry {
	if entry, ok := d.entries[code]; ok {
		return entry
	}
	if d.parent != nil {
		return d.parent.Entry(code)
	}
	// Return a default unknown entry
	return ErrorEntry{
		Type:         ErrorTypeInternal,
		ShortError:   shortUnknown,
		ErrorPattern: longUnknown,
	}
}

// ErrorType returns the error type classification for a code.
func (d *errorDict) ErrorType(code ErrorCode) ErrorType {
	return d.Entry(code).Type
}

// ShortError returns the short description for a code.
func (d *errorDict) ShortError(code ErrorCode) string {
	return d.Entry(code).ShortError
}

// ErrorPattern returns the long description template for a code.
func (d *errorDict) ErrorPattern(code ErrorCode) string {
	return d.Entry(code).ErrorPattern
}

// WithCode returns a new ErrorDict with the specified code entry overridden.
func (d *errorDict) WithCode(code ErrorCode, entry ErrorEntry) ErrorDict {
	newEntries := make(map[ErrorCode]ErrorEntry)
	newEntries[code] = entry
	return &errorDict{
		parent:  d,
		entries: newEntries,
	}
}
