package objects

import (
	"errors"
)

// refTracker represents a structure to track references and their dependencies.
type refTracker struct {
	edges map[string][]string // edges represent the directed graph of dependencies.
}

// newRefTracker initializes and returns a new refTracker.
func newRefTracker() *refTracker {
	return &refTracker{
		edges: make(map[string][]string),
	}
}

// Add adds a new dependency between key and dependsOnKey.
// It returns an error if adding this dependency results in a circular reference.
func (rt *refTracker) Add(key, dependsOnKey string) error {
	// Initialize the key in the map if it doesn't exist.
	if _, exists := rt.edges[key]; !exists {
		rt.edges[key] = []string{}
	}
	// Add the dependency.
	rt.edges[key] = append(rt.edges[key], dependsOnKey)

	// Check for circular references.
	visited := make(map[string]bool)
	stack := make(map[string]bool)

	if rt.hasCycle(key, visited, stack) {
		return errors.New("circular reference detected")
	}
	return nil
}

// hasCycle recursively checks for cycles in the graph using depth-first search.
// It returns true if a cycle is detected.
func (rt *refTracker) hasCycle(node string, visited, stack map[string]bool) bool {
	// If the node is in the stack, it means we've encountered a cycle.
	if stack[node] {
		return true
	}

	// If we've already visited the node, no need to revisit.
	if visited[node] {
		return false
	}

	// Mark the node as visited and added to the stack.
	visited[node] = true
	stack[node] = true

	// Recursively check all dependencies of the current node.
	for _, child := range rt.edges[node] {
		if rt.hasCycle(child, visited, stack) {
			return true
		}
	}

	// Once we're done processing the current node, remove it from the stack.
	stack[node] = false
	return false
}

func (rt *refTracker) Clone() *refTracker {
	clone := &refTracker{
		edges: make(map[string][]string),
	}

	for key, values := range rt.edges {
		clonedValues := make([]string, len(values))
		copy(clonedValues, values)
		clone.edges[key] = clonedValues
	}

	return clone
}
