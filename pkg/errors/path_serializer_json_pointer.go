package errors

import (
	"strings"

	"proto.zip/studio/validate/pkg/rulecontext"
)

// JSONPointerSerializer implements JSON Pointer serialization as defined in RFC 6901.
// JSON Pointer uses "/" to separate segments and escapes "/" as "~1" and "~" as "~0".
type JSONPointerSerializer struct{}

// Serialize serializes path segments using JSON Pointer format (RFC 6901).
// The format is: /segment1/segment2/0
// Special characters are escaped: "/" becomes "~1" and "~" becomes "~0".
func (s JSONPointerSerializer) Serialize(segments []rulecontext.PathSegment) string {
	if len(segments) == 0 {
		return ""
	}

	var parts []string
	for _, seg := range segments {
		var part string
		switch v := seg.(type) {
		case *rulecontext.PathSegmentIndex:
			part = v.String()
		case *rulecontext.PathSegmentString:
			// Escape JSON Pointer special characters
			part = escapeJSONPointer(v.Segment())
		}
		parts = append(parts, part)
	}
	
	return "/" + strings.Join(parts, "/")
}

// escapeJSONPointer escapes special characters in JSON Pointer format.
// According to RFC 6901:
// - "~" must be encoded as "~0"
// - "/" must be encoded as "~1"
func escapeJSONPointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}
