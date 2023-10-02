package rules

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// WrapAnyRuleSet implements RuleSet for the "any" interface and wraps around another type of rule set.
// Use it when you need to use a more specific RuleSet in a nested validator or to pass into a function.
//
// Unless you are implementing a brand new RuleSet you probably want to use the .Any() method on the RuleSet
// itself instead, which usually returns this interface.
type WrapAnyRuleSet[T any] struct {
	required bool
	inner    RuleSet[T]
	rule     Rule[any]
	parent   *WrapAnyRuleSet[T]
}

// WrapAny wraps an existing RuleSet in an "Any" rule set which can then be used to pass into nested validators
// or any function where the type of RuleSet is not known ahead of time.
//
// Unless you are implementing a brand new RuleSet you probably want to use the .Any() method on the RuleSet
// itself instead, which usually calls this function.
func WrapAny[T any](inner RuleSet[T]) *WrapAnyRuleSet[T] {
	return &WrapAnyRuleSet[T]{
		required: inner.Required(),
		inner:    inner,
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *WrapAnyRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
//
// Required defaults to the value of the wrapped RuleSet so if it is already required then there is
// no need to call this again.
func (v *WrapAnyRuleSet[T]) WithRequired() *WrapAnyRuleSet[T] {
	return &WrapAnyRuleSet[T]{
		required: true,
		inner:    v.inner,
		parent:   v,
	}
}

// Validate performs a validation of a RuleSet against a value and returns a value of the same type
// as the wrapped RuleSet or a ValidationErrorCollection. The wrapped rules are called before any rules
// added directly to the WrapAnyRuleSet.
func (v *WrapAnyRuleSet[T]) Validate(value any) (any, errors.ValidationErrorCollection) {
	return v.ValidateWithContext(value, context.Background())
}

// Validate performs a validation of a RuleSet against a value and returns a value of the same type
// as the wrapped RuleSet or a ValidationErrorCollection. The wrapped rules are called before any rules
// added directly to the WrapAnyRuleSet.
//
// Also, takes a Context which can be used by validation rules and error formatting.
func (v *WrapAnyRuleSet[T]) ValidateWithContext(value any, ctx context.Context) (any, errors.ValidationErrorCollection) {
	var retValue any

	retValue, innerErrors := v.inner.ValidateWithContext(value, ctx)

	allErrors := errors.Collection()

	if innerErrors != nil {
		allErrors = append(allErrors, innerErrors...)
	}

	currentRuleSet := v
	ctx = rulecontext.WithRuleSet(ctx, v)

	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			newValue, err := currentRuleSet.rule.Evaluate(ctx, retValue)
			if err != nil {
				allErrors = append(allErrors, err...)
			} else {
				retValue = newValue
			}
		}

		currentRuleSet = currentRuleSet.parent
	}

	if len(allErrors) != 0 {
		return retValue, allErrors
	} else {
		return retValue, nil
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// explicitly for the "any" interface.
//
// If you want to add a rule directly to the wrapped RuleSet you must do it before wrapping it.
//
// Use this when implementing custom rules.
func (v *WrapAnyRuleSet[T]) WithRule(rule Rule[any]) *WrapAnyRuleSet[T] {
	return &WrapAnyRuleSet[T]{
		required: v.required,
		inner:    v.inner,
		rule:     rule,
		parent:   v,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// explicitly for the "any" interface.
//
// If you want to add a rule directly to the wrapped RuleSet you must do it before wrapping it.
//
// Use this when implementing custom rules.
func (v *WrapAnyRuleSet[T]) WithRuleFunc(rule RuleFunc[any]) *WrapAnyRuleSet[T] {
	return v.WithRule(rule)
}

// Any is an identity function for this implementation and returns the current rule set.
func (v *WrapAnyRuleSet[T]) Any() RuleSet[any] {
	return v
}
