package rulecontext

import (
	"context"
	"fmt"
)

// pathSegment represents a segment in a validation path.
// pathSegment can be either a string segment or an index segment.
// This interface is unexported to prevent external implementations.
type pathSegment interface {
	Parent() pathSegment
	String() string
}

// PathSegment represents a segment in a validation path.
// PathSegment is a type alias for the unexported pathSegment interface.
// While external code can reference PathSegment, they cannot implement it
// because the underlying pathSegment interface is unexported and the concrete
// types (PathSegmentString and PathSegmentIndex) are the only implementations.
type PathSegment = pathSegment

// PathSegmentString represents a string segment in a validation path.
type PathSegmentString struct {
	parent  pathSegment
	segment string
}

// PathSegmentIndex represents an index segment in a validation path.
type PathSegmentIndex struct {
	parent  pathSegment
	segment int
}

// Parent returns the previous path segment.
func (s *PathSegmentString) Parent() pathSegment {
	return s.parent
}

// String returns the segment as a string.
func (s *PathSegmentString) String() string {
	return s.segment
}

// Segment returns the string value of this segment.
func (s *PathSegmentString) Segment() string {
	return s.segment
}

// Parent returns the previous path segment.
func (s *PathSegmentIndex) Parent() pathSegment {
	return s.parent
}

// String returns the index as a string using brackets.
//
// Example: [0] or [3]
func (s *PathSegmentIndex) String() string {
	return fmt.Sprintf("%d", s.segment)
}

// Index returns the numeric index value of this segment.
func (s *PathSegmentIndex) Index() int {
	return s.segment
}

// WithPathString returns a new context with the path segment added.
func WithPathString(parent context.Context, value string) context.Context {
	newPath := &PathSegmentString{
		segment: value,
	}

	if previousPath := Path(parent); previousPath != nil {
		newPath.parent = previousPath
	}

	return context.WithValue(parent, &pathContextKey, newPath)
}

// WithPathIndex returns a new context with the path segment index added.
func WithPathIndex(parent context.Context, value int) context.Context {
	newPath := &PathSegmentIndex{
		segment: value,
	}

	if previousPath := Path(parent); previousPath != nil {
		newPath.parent = previousPath
	}

	return context.WithValue(parent, &pathContextKey, newPath)
}

// Path returns the most recently added path segment.
// Path can be used to build out the full path.
func Path(ctx context.Context) PathSegment {
	if ctx == nil {
		return nil
	}

	if segment := ctx.Value(&pathContextKey); segment != nil {
		return segment.(PathSegment)
	}
	return nil
}
