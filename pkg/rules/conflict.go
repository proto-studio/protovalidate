package rules

// NoConflict is a helper struct that implements the Replaces method to always return false.
// NoConflict can be embedded in rule sets to indicate that they never replace other rules.
type NoConflict[T any] struct {
}

func (_ NoConflict[T]) Replaces(_ Rule[T]) bool {
	return false
}
