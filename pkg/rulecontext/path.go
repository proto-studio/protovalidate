package rulecontext

import (
	"context"
	"fmt"
)

// PathSegment represents a segment in a validation path.
// PathSegment can be either a string segment or an index segment.
type PathSegment interface {
	Parent() PathSegment
	String() string
	FullString() string
}

type pathSegmentString struct {
	parent  PathSegment
	segment string
}

type pathSegmentIndex struct {
	parent  PathSegment
	segment int
}

// Parent returns the previous path segment.
func (s *pathSegmentString) Parent() PathSegment {
	return s.parent
}

// String returns the segment as a string.
func (s *pathSegmentString) String() string {
	return s.segment
}

// FullString returns the full path until there are no more parent segments.
func (s *pathSegmentString) FullString() string {
	if s.parent != nil {
		return s.parent.FullString() + "/" + s.String()
	}
	return "/" + s.String()
}

// Parent returns the previous path segment.
func (s *pathSegmentIndex) Parent() PathSegment {
	return s.parent
}

// String returns the index as a string using brackets.
//
// Example: [0] or [3]
func (s *pathSegmentIndex) String() string {
	return fmt.Sprintf("%d", s.segment)
}

// FullString returns the full path until there are no more parent segments.
func (s *pathSegmentIndex) FullString() string {
	if s.parent != nil {
		return s.parent.FullString() + "/" + s.String()
	}
	return s.String()
}

// WithPathString returns a new context with the path segment added.
func WithPathString(parent context.Context, value string) context.Context {
	newPath := &pathSegmentString{
		segment: value,
	}

	if previousPath := Path(parent); previousPath != nil {
		newPath.parent = previousPath
	}

	return context.WithValue(parent, &pathContextKey, newPath)
}

// WithPathIndex returns a new context with the path segment index added.
func WithPathIndex(parent context.Context, value int) context.Context {
	newPath := &pathSegmentIndex{
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
