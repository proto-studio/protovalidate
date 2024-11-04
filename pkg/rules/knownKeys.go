package rules

import (
	"context"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

type knownKeyType byte

const (
	knownConstKey knownKeyType = iota
	knownDynamicKey
)

// knownKeys is a utility structure to track which keys are seen during validation.
type knownKeys[TK comparable] struct {
	keys map[TK]knownKeyType
}

// newKnownKeys creates a new instance of knownKeys.
func newKnownKeys[TK comparable](track bool) *knownKeys[TK] {
	if track {
		return &knownKeys[TK]{keys: make(map[TK]knownKeyType)}
	}
	return &knownKeys[TK]{}
}

// Add registers a known key.
func (k *knownKeys[TK]) Add(key TK) {
	if k.keys != nil {
		k.keys[key] = knownConstKey
	}
}

// exists checks if a given key is known.
func (k *knownKeys[TK]) exists(key TK) bool {
	_, ok := k.keys[key]
	return ok
}

// Check validates if all keys in the provided reflect.Value are known.
// It returns a ValidationErrorCollection with errors for each unexpected key.
//
// If allowUnknown is true when creating the object then this always returns an
// empty error collection.
func (k *knownKeys[TK]) Check(ctx context.Context, inValue reflect.Value) errors.ValidationErrorCollection {
	errs := errors.Collection()

	// If the knownKeys map is not initialized, return an empty error collection.
	if k.keys == nil {
		return errs
	}

	unk := k.Unknown(inValue)
	for _, key := range unk {
		subContext := rulecontext.WithPathString(ctx, toPath(key))
		errs = append(errs, errors.Errorf(errors.CodeUnexpected, subContext, "unexpected field"))
	}
	return errs
}

// Unknown returns all the unexpected keys.
func (k *knownKeys[TK]) Unknown(inValue reflect.Value) []TK {
	var out []TK

	// Loop through each key in the input value and check if it's a known key.
	for _, key := range inValue.MapKeys() {
		keyVal := key.Interface().(TK)
		if !k.exists(keyVal) {
			out = append(out, keyVal)
		}
	}

	return out
}
