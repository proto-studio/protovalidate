package errors

import (
	"fmt"
	"strings"

	"proto.zip/studio/validate/pkg/rulecontext"
)

// JSONPathSerializer implements JSONPath serialization format.
// JSONPath uses "$.field1.field2[0]" format with "$" prefix and brackets for indices.
type JSONPathSerializer struct{}

// Serialize serializes path segments using JSONPath format.
// The format is: $.segment1.segment2[0]
// String segments are separated by "." and index segments use bracket notation.
func (s JSONPathSerializer) Serialize(segments []rulecontext.PathSegment) string {
	if len(segments) == 0 {
		return "$"
	}

	var result strings.Builder
	result.WriteString("$")
	firstSegment := true
	
	for _, seg := range segments {
		switch v := seg.(type) {
		case *rulecontext.PathSegmentIndex:
			result.WriteString(fmt.Sprintf("[%d]", v.Index()))
			firstSegment = false
		case *rulecontext.PathSegmentString:
			// Escape special characters in JSONPath
			escaped := escapeJSONPath(v.Segment())
			// Add "." if not the first segment (whether previous was index or string)
			// But if escaped segment starts with '[', don't add dot (it's bracket notation)
			if !firstSegment && !strings.HasPrefix(escaped, "[") {
				result.WriteString(".")
			} else if firstSegment && !strings.HasPrefix(escaped, "[") {
				// First string segment after $ needs a dot (unless it's bracket notation)
				result.WriteString(".")
			}
			result.WriteString(escaped)
			firstSegment = false
		}
	}
	
	return result.String()
}

// escapeJSONPath escapes special characters in JSONPath format.
// In JSONPath, dots and brackets need to be escaped or quoted.
// For simplicity, we'll use bracket notation for segments containing special characters.
func escapeJSONPath(s string) string {
	// If the segment contains dots, brackets, or other special characters,
	// we should use bracket notation with quotes: ['field.name']
	if strings.ContainsAny(s, ".[]") {
		// Escape single quotes in the value
		escaped := strings.ReplaceAll(s, "'", "\\'")
		return fmt.Sprintf("['%s']", escaped)
	}
	return s
}
