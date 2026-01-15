package errors

import "fmt"

// ValidationErrorCollection implements a standard Error interface and also ValidationErrorCollection interface
// while preserving the validation data.
type ValidationErrorCollection []ValidationError

// Collection creates a new ValidationErrorCollection from one or more ValidationError values.
func Collection(errs ...ValidationError) ValidationErrorCollection {
	var arr []ValidationError

	if errs == nil {
		arr = make([]ValidationError, 0)
	} else {
		arr = errs[:]
	}

	return ValidationErrorCollection(arr)
}

// Size returns the number of errors in the collection.
//
// Deprecated: Size is deprecated and will be removed in v1.0.0. Use len(collection) instead.
func (collection ValidationErrorCollection) Size() int {
	return len(collection)
}

// All returns an array of all the errors in the collection.
// If there is more than one error, the order they are returned is not guaranteed to be deterministic.
//
// Deprecated: All is deprecated and will be removed in v1.0.0. Use as you would a normal slice or call Unwrap instead.
func (collection ValidationErrorCollection) All() []ValidationError {
	return collection
}

// Error implements the standard Error interface to return a string.
//
// Error returns only the first error if there is more than one, along with the total count.
// Error loses contextual data, so use the ValidationError object when possible.
//
// If there is more than one error, which error is displayed is not guaranteed to be deterministic.
//
// An empty collection should never be returned from a function. Return nil instead. Error panics if called on an empty collection.
func (collection ValidationErrorCollection) Error() string {
	if len(collection) > 1 {
		return fmt.Sprintf("%s (and %d more)", []ValidationError(collection)[0].Error(), len(collection)-1)
	}

	if len(collection) > 0 {
		return []ValidationError(collection)[0].Error()
	}

	panic("Empty collection")
}

// Unwrap implements the wrapped Error interface to return an array of errors.
// This enables support for errors.Is and errors.As from the standard library.
//
// Returns an empty slice for empty collections. An empty collection should never be returned from a function. Return nil instead.
func (collection ValidationErrorCollection) Unwrap() []error {
	errs := make([]error, len(collection))
	for i := range collection {
		errs[i] = collection[i]
	}
	return errs
}

// First returns only the first error.
// If there is more than one error, the error returned is not guaranteed to be deterministic.
func (collection ValidationErrorCollection) First() ValidationError {
	if len(collection) == 0 {
		return nil
	}

	return collection[0]
}

// For returns a new collection containing only errors for a specific path.
func (collection ValidationErrorCollection) For(path string) ValidationErrorCollection {
	if len(collection) == 0 {
		return nil
	}

	var filteredErrors []ValidationError
	for _, err := range collection {
		if err.Path() == path {
			filteredErrors = append(filteredErrors, err)
		}
	}

	if len(filteredErrors) == 0 {
		return nil
	}

	return Collection(filteredErrors...)
}

// Internal returns true if any error in the collection is an internal error.
// Internal errors are the most general classification and take precedence.
// Returns false for empty collections.
func (collection ValidationErrorCollection) Internal() bool {
	for _, err := range collection {
		if err.Internal() {
			return true
		}
	}
	return false
}

// Permission returns true if the most general error classification is permission.
// Permission errors are more general than validation errors but less general than internal errors.
// Returns true if any error is a permission error and no errors are internal.
// Returns false for empty collections.
func (collection ValidationErrorCollection) Permission() bool {
	if collection.Internal() {
		return false
	}
	for _, err := range collection {
		if err.Permission() {
			return true
		}
	}
	return false
}

// Validation returns true if all errors are validation errors.
// Validation errors are the most specific classification.
// Returns true only if no errors are internal or permission errors.
// Returns false for empty collections.
func (collection ValidationErrorCollection) Validation() bool {
	if len(collection) == 0 {
		return false
	}
	return !collection.Internal() && !collection.Permission()
}
