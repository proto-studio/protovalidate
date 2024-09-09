package rules

import (
	"errors"
)

// refTracker[T] represents a structure to track references and their dependencies.
type refTracker[T comparable] struct {
	edges map[T][]T // edges represent the directed graph of dependencies.
}

// newRefTracker initializes and returns a new refTracker[T].
func newRefTracker[T comparable]() *refTracker[T] {
	return &refTracker[T]{
		edges: make(map[T][]T),
	}
}

// Add adds a new dependency between key and dependsOnKey.
// It returns an error if adding this dependency results in a circular reference.
func (rt *refTracker[T]) Add(keyRule, dependsOnKeyRule Rule[T]) error {

	// For now both key and depends on must be constants
	constKeyRule, keyIsConstant := keyRule.(*ConstantRuleSet[T])
	constDependsOnKeyRule, dependsOnKeyIsConstant := dependsOnKeyRule.(*ConstantRuleSet[T])

	if !keyIsConstant || !dependsOnKeyIsConstant {
		return errors.New("conditional rules do not support dynamic keys at this time")
	}

	key := constKeyRule.Value()
	dependsOnKey := constDependsOnKeyRule.Value()

	// Initialize the key in the map if it doesn't exist.
	if _, exists := rt.edges[key]; !exists {
		rt.edges[key] = []T{}
	}
	// Add the dependency.
	rt.edges[key] = append(rt.edges[key], dependsOnKey)

	// Check for circular references.
	visited := make(map[T]bool)
	stack := make(map[T]bool)

	if rt.hasCycle(key, visited, stack) {
		return errors.New("circular reference detected")
	}
	return nil
}

// hasCycle recursively checks for cycles in the graph using depth-first search.
// It returns true if a cycle is detected.
func (rt *refTracker[T]) hasCycle(node T, visited, stack map[T]bool) bool {
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

func (rt *refTracker[T]) Clone() *refTracker[T] {
	clone := &refTracker[T]{
		edges: make(map[T][]T),
	}

	for key, values := range rt.edges {
		clonedValues := make([]T, len(values))
		copy(clonedValues, values)
		clone.edges[key] = clonedValues
	}

	return clone
}
