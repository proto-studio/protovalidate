package util

import (
	"fmt"
	"strings"
)

func StringsToRuleOutput(ruleName string, values []string) string {
	l := len(values)

	var sb strings.Builder
	sb.WriteString(ruleName)
	sb.WriteRune('(')

	// Append up to the first 3 strings or the total number of strings if less than 3
	for i := 0; i < 3 && i < l; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteRune('"')
		sb.WriteString(values[i])
		sb.WriteRune('"')
	}

	// If there are more than 3 strings, append the "... and X more" message
	if l > 3 {
		sb.WriteString(fmt.Sprintf(" ... and %d more", l-3))
	}

	sb.WriteRune(')')

	return sb.String()
}
