package util

import (
	"fmt"
	"strings"

	"proto.zip/studio/validate/pkg/errors"
)

const maxStringDisplayLength = 50

// TruncateString truncates a string to a maximum length, adding ellipsis if truncated.
// This is used for displaying string arguments in rule labels to keep them readable.
func TruncateString(s string) string {
	if len(s) <= maxStringDisplayLength {
		return s
	}
	return s[:maxStringDisplayLength] + "..."
}

// StringsToRuleOutput formats a rule name and a slice of values into a string representation.
// All values are converted to strings, with string values being quoted and any internal quotes escaped.
// This generic version works with slices of any type.
func StringsToRuleOutput[T any](ruleName string, values []T) string {
	l := len(values)

	var sb strings.Builder
	sb.WriteString(ruleName)
	sb.WriteRune('(')

	// Append up to the first 3 values or the total number of values if less than 3
	for i := 0; i < 3 && i < l; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		v := values[i]
		str, ok := any(v).(string)
		if ok {
			// Truncate and escape any internal double quotes if v is a string
			truncated := TruncateString(str)
			escapedValue := strings.ReplaceAll(truncated, "\"", "\\\"")
			sb.WriteString(fmt.Sprintf("\"%s\"", escapedValue))
		} else {
			// Convert the value to a string and quote it if v is not a string
			sb.WriteString(fmt.Sprintf("%v", v))
		}
	}

	// If there are more than 3 values, append the "... and X more" message
	if l > 3 {
		sb.WriteString(fmt.Sprintf(" ... and %d more", l-3))
	}

	sb.WriteRune(')')

	return sb.String()
}

// FormatErrorMessageLabel formats a label for WithErrorMessage with truncated string values.
func FormatErrorMessageLabel(short, long string) string {
	shortTruncated := TruncateString(short)
	longTruncated := TruncateString(long)
	return fmt.Sprintf("WithErrorMessage(%q, %q)", shortTruncated, longTruncated)
}

// FormatStringArgLabel formats a label for methods that take a single string argument.
func FormatStringArgLabel(methodName, value string) string {
	truncated := TruncateString(value)
	return fmt.Sprintf("%s(%q)", methodName, truncated)
}

// FormatErrorCodeLabel formats a label for WithErrorCode with the error code value.
func FormatErrorCodeLabel(code errors.ErrorCode) string {
	return fmt.Sprintf("WithErrorCode(%v)", code)
}

// FormatErrorMetaLabel formats a label for WithErrorMeta with truncated key and value.
func FormatErrorMetaLabel(key string, value any) string {
	keyTruncated := TruncateString(key)
	var valueStr string
	if str, ok := value.(string); ok {
		valueStr = TruncateString(str)
		return fmt.Sprintf("WithErrorMeta(%q, %q)", keyTruncated, valueStr)
	}
	return fmt.Sprintf("WithErrorMeta(%q, %v)", keyTruncated, value)
}

// FormatErrorCallbackLabel formats a label for WithErrorCallback.
// Since callbacks are functions, we show a generic representation.
func FormatErrorCallbackLabel() string {
	return "WithErrorCallback(<func>)"
}
