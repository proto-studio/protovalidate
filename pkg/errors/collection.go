package errors

import "fmt"

// ValidationErrorCollection implements a standard Error interface and also ValidationErrorCollection interface
// while preserving the validation data.
type ValidationErrorCollection []ValidationError

// Collection takes one or more ValidationError pointers and creates a new instance of a collection.
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
// If there is more than one error, only the first will be returned and the total count
// will also be returned with the string.
//
// When possible you should use the ValidationError object since this method loses contextual data.
//
// If there is more than one error, which error is displayed is not guaranteed to be deterministic.
//
// An empty collection should never be returned from a function. Return nil instead. This method panics if called on an empty collection.
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
//
// An empty collection should never be returned from a function. Return nil instead. This method panics if called on an empty collection.
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
