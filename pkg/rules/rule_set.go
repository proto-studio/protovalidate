package rules

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

// RuleSet interface is used to define a collection of Rules that logically apply to the same value.
type RuleSet[T any] interface {
	Rule[T]
	Validate(value any) (T, errors.ValidationErrorCollection)                                 // Validate is used to take in a value (or any type) and return a value of the correct type for the RuleSet or a collection of validation errors.
	ValidateWithContext(value any, ctx context.Context) (T, errors.ValidationErrorCollection) // ValidateWithContext does the same as Validate but takes a Context object that can be accessed by the rules.
	Any() RuleSet[any]                                                                        // Any returns an implementation of rule sets for the "any" type that wraps a typed RuleSet so that the set can be used in nested objects and arrays.
	Required() bool                                                                           // Returns true if the value is not allowed to be omitted when nested under other rule sets.
	String() string                                                                           // Converts the rule set to a string for printing and debugging.
}
