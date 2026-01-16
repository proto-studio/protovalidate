package errors

import (
	"proto.zip/studio/validate/pkg/rulecontext"
)

// DefaultPathSerializer implements the default path serialization format.
// This matches the current behavior: "/segment1/segment2" for string segments
// and "/segment1/0" for index segments.
// If the first segment is an index with no parent, it returns just the index without a leading "/".
type DefaultPathSerializer struct{}

// Serialize serializes path segments using the default format.
// String segments are separated by "/" with a leading "/".
// Index segments are represented as their numeric value.
// If the path starts with an index segment, no leading "/" is added.
func (s DefaultPathSerializer) Serialize(segments []rulecontext.PathSegment) string {
	if len(segments) == 0 {
		return ""
	}

	// If first segment is an index, don't add leading "/" (matches current behavior)
	if len(segments) > 0 {
		if _, ok := segments[0].(*rulecontext.PathSegmentIndex); ok {
			var result string
			for i, seg := range segments {
				if i > 0 {
					result += "/"
				}
				result += seg.String()
			}
			return result
		}
	}

	// Otherwise, use leading "/" for string segments
	var result string
	for _, seg := range segments {
		result += "/" + seg.String()
	}
	
	return result
}
