package errors

import "fmt"

// ValidationErrorCollection interface is used to return 1 or more validation errors
// from a function while preserving all the details.
type ValidationErrorCollection interface {
	Add(err ...ValidationError)
	First() ValidationError
	All() []ValidationError
	For(path string) ValidationErrorCollection
	Error() string
	Size() int
}

// ValidationErrorCollection implements a standard Error interface and also ValidationErrorCollection interface
// while preserving the validation data.
type validationErrorCollection struct {
	errors []ValidationError
}

// Collection takes one or more ValidationError pointers and creates a new instance of a collection.
func Collection(errs ...ValidationError) ValidationErrorCollection {
	if errs == nil {
		return &validationErrorCollection{
			errors: make([]ValidationError, 0),
		}
	}

	return &validationErrorCollection{
		errors: errs[:],
	}
}

// Add appends a new error onto the end of the collection.
// This method is not thread safe. RuleSets must provide their own locking.
func (collection *validationErrorCollection) Add(errs ...ValidationError) {
	collection.errors = append(collection.errors, errs...)
}

// Size returns the number of errors in the collection.
func (collection *validationErrorCollection) Size() int {
	return len(collection.errors)
}

// All returns an array of all the errors in the collection.
// If there is more than one error, the order they are returned is not guaranteed to be deturministic.
func (collection *validationErrorCollection) All() []ValidationError {
	return collection.errors
}

// Error implements the standard Error interface to rerurn a string.
//
// If there is more than one error, only the first will be returned and the total count
// will also be returned with the string.
//
// When possible you should use the ValidationError object since this method loses contextual data.
//
// As with the First() method, if there is more than one error, which error is displayed is not guaranteed to be deturministic.
//
// An empty collection should never be returned from a function. Return nil instead. This method panics if called on an empty collection.
func (collection *validationErrorCollection) Error() string {
	if len(collection.errors) > 1 {
		return fmt.Sprintf("%s (and %d more)", collection.errors[0].Error(), len(collection.errors)-1)
	}

	if len(collection.errors) > 0 {
		return collection.errors[0].Error()
	}

	panic("Empty collection")
}

// First returns only the first error.
// If there is more than one error, the error returned is not guaranteed to be deturministic.
func (collection *validationErrorCollection) First() ValidationError {
	if len(collection.errors) == 0 {
		return nil
	}

	return collection.errors[0]
}

// For returns a new collection containing only errors for a specific path.
func (collection *validationErrorCollection) For(path string) ValidationErrorCollection {
	if len(collection.errors) == 0 {
		return nil
	}

	var filteredErrors []ValidationError
	for _, err := range collection.errors {
		if err.Path() == path {
			filteredErrors = append(filteredErrors, err)
		}
	}

	if len(filteredErrors) == 0 {
		return nil
	}

	return Collection(filteredErrors...)
}
