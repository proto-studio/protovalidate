package objects

import (
	"context"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// knownKeys is a utility structure to track which keys are seen during validation.
type knownKeys struct {
	keys map[string]bool
}

// newKnownKeys creates a new instance of knownKeys.
func newKnownKeys(allowUnknown, fromMap bool) *knownKeys {
	if !allowUnknown && fromMap {
		return &knownKeys{keys: make(map[string]bool)}
	}
	return &knownKeys{}
}

// Add registers a known key.
func (k *knownKeys) Add(key string) {
	if k.keys != nil {
		k.keys[key] = true
	}
}

// exists checks if a given key is known.
func (k *knownKeys) exists(key string) bool {
	_, ok := k.keys[key]
	return ok
}

// Check validates if all keys in the provided reflect.Value are known.
// It returns a ValidationErrorCollection with errors for each unexpected key.
//
// If allowUnknown is true when creating the object then this always returns an
// empty error collection.
func (k *knownKeys) Check(inValue reflect.Value) errors.ValidationErrorCollection {
	errs := errors.Collection()

	// If the knownKeys map is not initialized, return an empty error collection.
	if k.keys == nil {
		return errs
	}

	// Loop through each key in the input value and check if it's a known key.
	for _, key := range inValue.MapKeys() {
		keyStr := key.String()
		if !k.exists(keyStr) {
			subContext := rulecontext.WithPathString(context.Background(), keyStr)
			errs = append(errs, errors.Errorf(errors.CodeUnexpected, subContext, "unexpected field"))
		}
	}
	return errs
}
