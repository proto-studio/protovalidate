package rules

// Define a custom constraint that includes types that can be passed to len
// Used by minLenRule and maxLenRule
type lengthy[T any] interface {
	~string | ~[]T | ~chan T
}
