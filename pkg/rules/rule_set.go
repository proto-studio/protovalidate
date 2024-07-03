package rules

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

// RuleSet interface is used to define a collection of Rules that logically apply to the same value.
type RuleSet[T any] interface {
	Rule[T]
	Run(ctx context.Context, value any) (T, errors.ValidationErrorCollection) // Run coerces the value into the correct type then evaluates all the rules in the set
	Any() RuleSet[any]                                                        // Any returns an implementation of rule sets for the "any" type that wraps a typed RuleSet so that the set can be used in nested objects and arrays.
	Required() bool                                                           // Returns true if the value is not allowed to be omitted when nested under other rule sets.
	String() string                                                           // Converts the rule set to a string for printing and debugging.
}
