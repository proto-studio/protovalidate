package errors

import (
	"proto.zip/studio/validate/pkg/rulecontext"
)

// PathSerializer serializes an array of path segments into a string representation.
type PathSerializer interface {
	Serialize(segments []rulecontext.PathSegment) string
}

// extractPathSegments extracts all segments from a PathSegment into an array,
// ordered from root to leaf (top to bottom).
func extractPathSegments(segment rulecontext.PathSegment) []rulecontext.PathSegment {
	if segment == nil {
		return nil
	}

	// First, collect all segments by traversing up to the root
	var segments []rulecontext.PathSegment
	current := segment
	for current != nil {
		segments = append(segments, current)
		current = current.Parent()
	}

	// Reverse to get root-to-leaf order
	for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
		segments[i], segments[j] = segments[j], segments[i]
	}

	return segments
}
