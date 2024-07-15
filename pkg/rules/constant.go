package rules

import (
	"context"
	"fmt"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
)

type constCache[T comparable] map[T]*ConstantRuleSet[T]

var constCacheMap map[any]any

// ConstantRuleSet implements RuleSet that returns an error for
// any value that does not match the constant.
//
// This is primary used for conditional rules. To test a constant of a specific
// type it is usually best to use that type.
type ConstantRuleSet[T comparable] struct {
	required bool
	value    T
	empty    T // Leave this empty
}

// Constant creates a new Constant rule set for the specified value.
func Constant[T comparable](value T) *ConstantRuleSet[T] {
	var empty T
	var typedCache constCache[T]

	if constCacheMap == nil {
		constCacheMap = make(map[any]any)
		typedCache = make(map[T]*ConstantRuleSet[T])
		constCacheMap[empty] = typedCache
	} else if tmp, ok := constCacheMap[empty]; ok {
		typedCache = tmp.(constCache[T])
	} else {
		typedCache = make(map[T]*ConstantRuleSet[T])
		constCacheMap[empty] = typedCache
	}

	if val, ok := typedCache[value]; ok {
		return val
	}

	typedCache[value] = &ConstantRuleSet[T]{
		value: value,
	}
	return typedCache[value]
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *ConstantRuleSet[T]) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (ruleSet *ConstantRuleSet[T]) WithRequired() *ConstantRuleSet[T] {
	if ruleSet.required {
		return ruleSet
	}

	return &ConstantRuleSet[T]{
		value:    ruleSet.value,
		required: true,
	}
}

// Run performs a validation of a RuleSet against a value and returns the unaltered supplied value
// or a ValidationErrorCollection.
func (ruleSet *ConstantRuleSet[T]) Run(ctx context.Context, value any) (T, errors.ValidationErrorCollection) {
	v, ok := value.(T)
	if !ok {
		return ruleSet.empty, errors.Collection(errors.NewCoercionError(ctx, reflect.TypeOf(ruleSet.empty).String(), reflect.TypeOf(value).String()))
	}
	return v, ruleSet.Evaluate(ctx, v)
}

// Evaluate performs a validation of a RuleSet against a value and returns any errors.
func (ruleSet *ConstantRuleSet[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	if value != ruleSet.value {
		return errors.Collection(errors.Errorf(errors.CodePattern, ctx, "value does not match"))
	}
	return nil
}

// Conflict returns true for all rules since by definition no rule can be a superset of it.
func (ruleSet *ConstantRuleSet[T]) Conflict(other Rule[T]) bool {
	return true
}

// Any is an identity function for this implementation and returns the current rule set.
func (ruleSet *ConstantRuleSet[T]) Any() RuleSet[any] {
	return WrapAny[T](ruleSet)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *ConstantRuleSet[T]) String() string {
	str := fmt.Sprintf(`ConstantRuleSet(%v)`, ruleSet.value)
	if ruleSet.required {
		return str + ".WithRequired()"
	}
	return str
}

// Value returns the constant value in the correct type.
func (ruleSet *ConstantRuleSet[T]) Value() T {
	return ruleSet.value
}
