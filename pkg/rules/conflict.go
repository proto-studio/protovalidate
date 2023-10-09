package rules

type NoConflict[T any] struct {
}

func (_ NoConflict[T]) Conflict(_ Rule[T]) bool {
	return false
}
