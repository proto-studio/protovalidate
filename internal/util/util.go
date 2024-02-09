package util

import (
	"fmt"
	"strings"
)

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
			// Escape any internal double quotes if v is a string
			escapedValue := strings.ReplaceAll(str, "\"", "\\\"")
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
