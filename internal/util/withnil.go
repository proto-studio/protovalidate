package util

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

// TryNilIfAllowed reports whether input is nil and how to handle it.
// When input is nil and withNil is true, returns (true, nil)—caller should return (zero value, nil).
// When input is nil and withNil is false, returns (true, CodeNull error)—caller should return (zero, err).
// When input is not nil, returns (false, nil)—caller should continue with normal processing.
func TryNilIfAllowed(ctx context.Context, withNil bool, input any) (handled bool, err errors.ValidationError) {
	if input != nil {
		return false, nil
	}
	if !withNil {
		return true, errors.Error(errors.CodeNull, ctx)
	}
	return true, nil
}
