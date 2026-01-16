package errors

import (
	"fmt"
	"strings"

	"proto.zip/studio/validate/pkg/rulecontext"
)

// DotNotationSerializer implements dot notation serialization format.
// Dot notation uses "field1.field2[0]" format without a prefix.
type DotNotationSerializer struct{}

// Serialize serializes path segments using dot notation format.
// The format is: segment1.segment2[0]
// String segments are separated by "." and index segments use bracket notation.
func (s DotNotationSerializer) Serialize(segments []rulecontext.PathSegment) string {
	if len(segments) == 0 {
		return ""
	}

	var result strings.Builder

	for i, seg := range segments {
		switch v := seg.(type) {
		case *rulecontext.PathSegmentIndex:
			result.WriteString(fmt.Sprintf("[%d]", v.Index()))
		case *rulecontext.PathSegmentString:
			// Escape special characters in dot notation
			escaped := escapeDotNotation(v.Segment())
			// Add "." if not the first segment (whether previous was index or string)
			// But if escaped segment starts with '[', don't add dot (it's bracket notation)
			if i > 0 && !strings.HasPrefix(escaped, "[") {
				result.WriteString(".")
			}
			result.WriteString(escaped)
		}
	}

	return result.String()
}

// escapeDotNotation escapes special characters in dot notation format.
// For segments containing dots or brackets, we use bracket notation with quotes.
func escapeDotNotation(s string) string {
	// If the segment contains dots, brackets, or other special characters,
	// we should use bracket notation with quotes: ['field.name']
	if strings.ContainsAny(s, ".[]") {
		// Escape single quotes in the value
		escaped := strings.ReplaceAll(s, "'", "\\'")
		return fmt.Sprintf("['%s']", escaped)
	}
	return s
}
