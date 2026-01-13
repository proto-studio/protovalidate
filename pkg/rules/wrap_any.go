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
	required    bool
	withNil     bool
	inner       RuleSet[T]
	rule        Rule[any]
	parent      *WrapAnyRuleSet[T]
	label       string
	errorConfig *errors.ErrorConfig
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

// wrapAnyCloneOption is a functional option for cloning WrapAnyRuleSet.
type wrapAnyCloneOption[T any] func(*WrapAnyRuleSet[T])

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (v *WrapAnyRuleSet[T]) clone(options ...wrapAnyCloneOption[T]) *WrapAnyRuleSet[T] {
	newRuleSet := &WrapAnyRuleSet[T]{
		required:    v.required,
		withNil:     v.withNil,
		inner:       v.inner,
		parent:      v,
		errorConfig: v.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func wrapAnyWithLabel[T any](label string) wrapAnyCloneOption[T] {
	return func(rs *WrapAnyRuleSet[T]) { rs.label = label }
}

func wrapAnyWithErrorConfig[T any](config *errors.ErrorConfig) wrapAnyCloneOption[T] {
	return func(rs *WrapAnyRuleSet[T]) { rs.errorConfig = config }
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
	newRuleSet := v.clone(wrapAnyWithLabel[T]("WithRequired()"))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *WrapAnyRuleSet[T]) WithNil() *WrapAnyRuleSet[T] {
	newRuleSet := v.clone(wrapAnyWithLabel[T]("WithNil()"))
	newRuleSet.withNil = true
	return newRuleSet
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
	// Add error config to context for error customization
	ctx = errors.WithErrorConfig(ctx, v.errorConfig)

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
	newRuleSet := v.clone()
	newRuleSet.rule = rule
	return newRuleSet
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

// WithErrorMessage returns a new RuleSet with custom short and long error messages.
func (v *WrapAnyRuleSet[T]) WithErrorMessage(short, long string) *WrapAnyRuleSet[T] {
	return v.clone(wrapAnyWithLabel[T]("WithErrorMessage(...)"), wrapAnyWithErrorConfig[T](v.errorConfig.WithMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (v *WrapAnyRuleSet[T]) WithDocsURI(uri string) *WrapAnyRuleSet[T] {
	return v.clone(wrapAnyWithLabel[T]("WithDocsURI(...)"), wrapAnyWithErrorConfig[T](v.errorConfig.WithDocs(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (v *WrapAnyRuleSet[T]) WithTraceURI(uri string) *WrapAnyRuleSet[T] {
	return v.clone(wrapAnyWithLabel[T]("WithTraceURI(...)"), wrapAnyWithErrorConfig[T](v.errorConfig.WithTrace(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (v *WrapAnyRuleSet[T]) WithErrorCode(code errors.ErrorCode) *WrapAnyRuleSet[T] {
	return v.clone(wrapAnyWithLabel[T]("WithErrorCode(...)"), wrapAnyWithErrorConfig[T](v.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (v *WrapAnyRuleSet[T]) WithErrorMeta(key string, value any) *WrapAnyRuleSet[T] {
	return v.clone(wrapAnyWithLabel[T]("WithErrorMeta(...)"), wrapAnyWithErrorConfig[T](v.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (v *WrapAnyRuleSet[T]) WithErrorCallback(fn errors.ErrorCallback) *WrapAnyRuleSet[T] {
	return v.clone(wrapAnyWithLabel[T]("WithErrorCallback(...)"), wrapAnyWithErrorConfig[T](v.errorConfig.WithCallback(fn)))
}
