package rules

import (
	"context"
	"fmt"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// InterfaceRuleSet implements RuleSet for the a generic interface.
type InterfaceRuleSet[T any] struct {
	NoConflict[T]
	required bool
	rule     Rule[T]
	parent   *InterfaceRuleSet[T]
	label    string
	cast     func(value any) (T, bool)
	empty    T // leave empty
}

// Interface creates a new Interface rule set.
func Interface[T any]() *InterfaceRuleSet[T] {
	return &InterfaceRuleSet[T]{
		label: fmt.Sprintf("InterfaceRuleSet[%s]", reflect.TypeOf(new(T)).Elem().Name()),
	}
}

// WithCast creates a new Interface rule set that has the set cast function.
// The cast function should take "any" and return a value of the appropriate interface type.
// Run will always try to directly cast the value. Adding a function is useful for when the
// value may need to be wrapped in another type in order to satisfy the interface.
//
// Cast functions are stacking, You may call this function as many times as you need in order
// to cast from different type. Newly defined cast functions take priority. Execution will stop
// at the first successful function.
func (v *InterfaceRuleSet[T]) WithCast(fn func(value any) (T, bool)) *InterfaceRuleSet[T] {
	return &InterfaceRuleSet[T]{
		required: v.required,
		parent:   v,
		cast:     fn,
		label:    "WithCast(...)",
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *InterfaceRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (v *InterfaceRuleSet[T]) WithRequired() *InterfaceRuleSet[T] {
	if v.required {
		return v
	}

	return &InterfaceRuleSet[T]{
		required: true,
		parent:   v,
		label:    "WithRequired()",
	}
}

// Run performs a validation of a RuleSet against a value and returns the unaltered supplied value
// or a ValidationErrorCollection.
func (ruleSet *InterfaceRuleSet[T]) Run(ctx context.Context, value any) (T, errors.ValidationErrorCollection) {
	if v, ok := value.(T); ok {
		return v, ruleSet.Evaluate(ctx, v)
	}

	for curRuleSet := ruleSet; curRuleSet != nil; curRuleSet = curRuleSet.parent {
		if curRuleSet.cast != nil {
			if v, ok := curRuleSet.cast(value); ok {
				return v, ruleSet.Evaluate(ctx, v)
			}
		}
	}

	return ruleSet.empty, errors.Collection(
		errors.NewCoercionError(
			ctx,
			reflect.TypeOf(new(T)).Elem().Name(),
			reflect.ValueOf(value).Kind().String(),
		),
	)
}

// Evaluate performs a validation of a RuleSet against all the defined rules.
func (v *InterfaceRuleSet[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	allErrors := errors.Collection()

	currentRuleSet := v
	ctx = rulecontext.WithRuleSet(ctx, v)

	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			err := currentRuleSet.rule.Evaluate(ctx, value)
			if err != nil {
				allErrors = append(allErrors, err...)
			}
		}

		currentRuleSet = currentRuleSet.parent
	}

	if len(allErrors) != 0 {
		return allErrors
	} else {
		return nil
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// explicitly for the "any" interface.
//
// Use this when implementing custom rules.
func (v *InterfaceRuleSet[T]) WithRule(rule Rule[T]) *InterfaceRuleSet[T] {
	return &InterfaceRuleSet[T]{
		required: v.required,
		cast:     v.cast,
		rule:     rule,
		parent:   v,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// explicitly for the "any" interface.
//
// Use this when implementing custom rules.
func (v *InterfaceRuleSet[T]) WithRuleFunc(rule RuleFunc[T]) *InterfaceRuleSet[T] {
	return v.WithRule(rule)
}

// Interface is an identity function for this implementation and returns the current rule set.
func (v *InterfaceRuleSet[T]) Any() RuleSet[any] {
	return WrapAny[T](v)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *InterfaceRuleSet[T]) String() string {
	label := ruleSet.label

	if label == "" {
		if ruleSet.rule != nil {
			label = ruleSet.rule.String()
		}
	}

	if ruleSet.parent != nil {
		return ruleSet.parent.String() + "." + label
	}
	return label
}
