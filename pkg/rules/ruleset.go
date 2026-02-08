package rules

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

// RuleSet interface is used to define a collection of Rules that logically apply to the same value.
//
// RuleSet implementations should also implement these error customization methods (not on interface for chaining):
//   - WithErrorMessage(short, long string) <ConcreteType>
//   - WithDocsURI(uri string) <ConcreteType>
//   - WithTraceURI(uri string) <ConcreteType>
//   - WithErrorCode(code errors.ErrorCode) <ConcreteType>
//   - WithErrorMeta(key string, value any) <ConcreteType>
//   - WithErrorCallback(fn errors.ErrorCallback) <ConcreteType>
type RuleSet[T any] interface {
	Rule[T]

	// Apply coerces value into the correct type, evaluates all rules in the rule set, and assigns the result to out.
	// Returns a ValidationError if coercion or validation fails. out must be a non-nil pointer to the output type.
	Apply(ctx context.Context, value any, out any) errors.ValidationError

	// Any returns a RuleSet[any] that wraps this rule set for use in nested objects and arrays.
	Any() RuleSet[any]

	// Required returns true if the value must be present when nested under other rule sets (e.g. required field).
	Required() bool

	// String returns a string representation of the rule set for debugging and serialization.
	String() string
}
