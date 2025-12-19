package rules

// NoConflict is a helper struct that implements the Conflict method to always return false.
// NoConflict can be embedded in rule sets to indicate that they never conflict with other rules.
type NoConflict[T any] struct {
}

func (_ NoConflict[T]) Conflict(_ Rule[T]) bool {
	return false
}
