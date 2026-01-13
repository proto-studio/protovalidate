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
	Apply(ctx context.Context, value any, out any) errors.ValidationErrorCollection // Apply attempts to coerce the value into the correct type and evaluates all rules in the rule set, then assigns the results to an interface.
	Any() RuleSet[any]                                                              // Any returns an implementation of rule sets for the "any" type that wraps a typed RuleSet so that the set can be used in nested objects and arrays.
	Required() bool                                                                 // Returns true if the value is not allowed to be omitted when nested under other rule sets.
	String() string                                                                 // Converts the rule set to a string for printing and debugging.
}
