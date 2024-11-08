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
	NoConflict[any]
	required bool
	inner    RuleSet[T]
	rule     Rule[any]
	parent   *WrapAnyRuleSet[T]
	label    string
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
		label:    "WithRequired()",
	}
}

// evaluateRules runs all the rules and returns any errors.
// Returns a collection regardless of if there are any errors.
func (v *WrapAnyRuleSet[T]) evaluateRules(ctx context.Context, value any) errors.ValidationErrorCollection {
	allErrors := errors.Collection()

	currentRuleSet := v
	ctx = rulecontext.WithRuleSet(ctx, v)

	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			if errs := currentRuleSet.rule.Evaluate(ctx, value); errs != nil {
				allErrors = append(allErrors, errs...)
			}
		}

		currentRuleSet = currentRuleSet.parent
	}

	return allErrors
}

// Run performs a validation of a RuleSet against a value and returns a value of the same type
// as the wrapped RuleSet or a ValidationErrorCollection. The wrapped rules are called before any rules
// added directly to the WrapAnyRuleSet.
func (v *WrapAnyRuleSet[T]) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	innerErrors := v.inner.Apply(ctx, input, output)
	allErrors := v.evaluateRules(ctx, output)

	if innerErrors != nil {
		allErrors = append(allErrors, innerErrors...)
	}

	if len(allErrors) > 0 {
		return allErrors
	} else {
		return nil
	}
}

// Evaluate performs a validation of a RuleSet against a value of any type.
// If the input value implements the same type as the wrapped RuleSet then Evaluate is called directly, otherwise
// Apply is called. This approach is usually more efficient since it does not need to allocate an output variable.
func (ruleSet *WrapAnyRuleSet[T]) Evaluate(ctx context.Context, value any) errors.ValidationErrorCollection {
	if v, ok := value.(T); ok {
		innerErrors := ruleSet.inner.Evaluate(ctx, v)
		allErrors := ruleSet.evaluateRules(ctx, value)

		if innerErrors != nil {
			allErrors = append(allErrors, innerErrors...)
		}

		if len(allErrors) != 0 {
			return allErrors
		} else {
			return nil
		}
	} else {
		var out T
		errs := ruleSet.Apply(ctx, value, &out)
		return errs
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

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *WrapAnyRuleSet[T]) String() string {
	if ruleSet.parent != nil {
		label := ruleSet.label

		if label == "" {
			if ruleSet.rule != nil {
				label = ruleSet.rule.String()
			}
		}

		return ruleSet.parent.String() + "." + label
	}

	return ruleSet.inner.String() + ".Any()"
}
