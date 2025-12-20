package rules

import (
	"context"

	"proto.zip/studio/validate/internal/util"
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
	withNil  bool
	inner    RuleSet[T]
	rule     Rule[any]
	parent   *WrapAnyRuleSet[T]
	label    string
}

// WrapAny wraps an existing RuleSet in an "Any" rule set which can then be used to pass into nested validators
// or any function where the type of RuleSet is not known ahead of time.
//
// WrapAny is usually called by the .Any() method on RuleSet implementations.
// Unless you are implementing a brand new RuleSet you probably want to use the .Any() method instead.
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

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
//
// Required defaults to the value of the wrapped RuleSet so if it is already required then there is
// no need to call this again.
func (v *WrapAnyRuleSet[T]) WithRequired() *WrapAnyRuleSet[T] {
	return &WrapAnyRuleSet[T]{
		required: true,
		withNil:  v.withNil,
		inner:    v.inner,
		parent:   v,
		label:    "WithRequired()",
	}
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *WrapAnyRuleSet[T]) WithNil() *WrapAnyRuleSet[T] {
	return &WrapAnyRuleSet[T]{
		required: v.required,
		withNil:  true,
		inner:    v.inner,
		parent:   v,
		label:    "WithNil()",
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

// Apply performs validation of a RuleSet against a value and assigns the result to the output parameter.
// Apply calls wrapped rules before any rules added directly to the WrapAnyRuleSet.
// Apply returns a ValidationErrorCollection if any validation errors occur.
func (v *WrapAnyRuleSet[T]) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, v.withNil, input, output); handled {
		return err
	}

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

// Evaluate performs validation of a RuleSet against a value of any type and returns a ValidationErrorCollection.
// Evaluate calls the wrapped RuleSet's Evaluate method directly if the input value implements the same type,
// otherwise it calls Apply. This approach is usually more efficient since it does not need to allocate an output variable.
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

// WithRule returns a new child rule set that applies a custom validation rule.
// The custom rule is evaluated during validation and any errors it returns are included in the validation result.
//
// If you want to add a rule directly to the wrapped RuleSet you must do it before wrapping it.
func (v *WrapAnyRuleSet[T]) WithRule(rule Rule[any]) *WrapAnyRuleSet[T] {
	return &WrapAnyRuleSet[T]{
		required: v.required,
		withNil:  v.withNil,
		inner:    v.inner,
		rule:     rule,
		parent:   v,
	}
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
//
// If you want to add a rule directly to the wrapped RuleSet you must do it before wrapping it.
func (v *WrapAnyRuleSet[T]) WithRuleFunc(rule RuleFunc[any]) *WrapAnyRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns the current rule set.
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
