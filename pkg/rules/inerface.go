package rules

import (
	"context"
	"fmt"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// InterfaceRuleSet implements RuleSet for the a generic interface.
type InterfaceRuleSet[T any] struct {
	NoConflict[T]
	required    bool
	withNil     bool
	rule        Rule[T]
	parent      *InterfaceRuleSet[T]
	label       string
	cast        func(ctx context.Context, value any) (T, errors.ValidationError)
	errorConfig *errors.ErrorConfig
}

// Interface creates a new Interface rule set.
func Interface[T any]() *InterfaceRuleSet[T] {
	return &InterfaceRuleSet[T]{
		label: fmt.Sprintf("InterfaceRuleSet[%s]", reflect.TypeOf(new(T)).Elem().Name()),
	}
}

// interfaceCloneOption is a functional option for cloning InterfaceRuleSet.
type interfaceCloneOption[T any] func(*InterfaceRuleSet[T])

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (v *InterfaceRuleSet[T]) clone(options ...interfaceCloneOption[T]) *InterfaceRuleSet[T] {
	newRuleSet := &InterfaceRuleSet[T]{
		required:    v.required,
		withNil:     v.withNil,
		cast:        v.cast,
		parent:      v,
		errorConfig: v.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func interfaceWithLabel[T any](label string) interfaceCloneOption[T] {
	return func(rs *InterfaceRuleSet[T]) { rs.label = label }
}

func interfaceWithErrorConfig[T any](config *errors.ErrorConfig) interfaceCloneOption[T] {
	return func(rs *InterfaceRuleSet[T]) { rs.errorConfig = config }
}

// WithCast creates a new Interface rule set that has the set cast function.
// The cast function should take "any" and return a value of the appropriate interface type.
// Run will always try to directly cast the value. Adding a function is useful for when the
// value may need to be wrapped in another type in order to satisfy the interface.
//
// Cast functions are stacking, You may call this function as many times as you need in order
// to cast from different type. Newly defined cast functions take priority. Execution will stop
// at the first function to return a non-nil value or an error collection.
//
// A third boolean return value is added to differentiate between a successful cast to a nil value
// and
func (v *InterfaceRuleSet[T]) WithCast(fn func(ctx context.Context, value any) (T, errors.ValidationError)) *InterfaceRuleSet[T] {
	newRuleSet := v.clone(interfaceWithLabel[T]("WithCast(<function>)"))
	newRuleSet.cast = fn
	return newRuleSet
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *InterfaceRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (v *InterfaceRuleSet[T]) WithRequired() *InterfaceRuleSet[T] {
	if v.required {
		return v
	}

	newRuleSet := v.clone(interfaceWithLabel[T]("WithRequired()"))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *InterfaceRuleSet[T]) WithNil() *InterfaceRuleSet[T] {
	newRuleSet := v.clone(interfaceWithLabel[T]("WithNil()"))
	newRuleSet.withNil = true
	return newRuleSet
}

// Apply coerces input to T, evaluates all rules, and returns the result.
func (ruleSet *InterfaceRuleSet[T]) Apply(ctx context.Context, input any) (T, errors.ValidationError) {
	var zero T
	ctx = errors.WithErrorConfig(ctx, ruleSet.errorConfig)

	if handled, err := util.TryNilIfAllowed(ctx, ruleSet.withNil, input); handled {
		return zero, err
	}

	if v, ok := input.(T); ok {
		if errs := ruleSet.Evaluate(ctx, v); errs != nil {
			return zero, errs
		}
		return v, nil
	}

	for curRuleSet := ruleSet; curRuleSet != nil; curRuleSet = curRuleSet.parent {
		if curRuleSet.cast != nil {
			if v, errs := curRuleSet.cast(ctx, input); any(v) != nil || errs != nil {
				if errs != nil {
					return zero, errs
				}
				if evalErrs := ruleSet.Evaluate(ctx, v); evalErrs != nil {
					return zero, evalErrs
				}
				return v, nil
			}
		}
	}

	return zero, errors.Error(errors.CodeType,
		ctx,
		reflect.TypeOf(new(T)).Elem().Name(),
		reflect.ValueOf(input).Kind().String(),
	)
}

// Evaluate performs a validation of a RuleSet against all the defined rules.
func (v *InterfaceRuleSet[T]) Evaluate(ctx context.Context, value T) errors.ValidationError {
	var errs errors.ValidationError
	currentRuleSet := v
	ctx = rulecontext.WithRuleSet(ctx, v)
	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			if err := currentRuleSet.rule.Evaluate(ctx, value); err != nil {
				errs = errors.Join(errs, err)
			}
		}
		currentRuleSet = currentRuleSet.parent
	}
	return errs
}

// WithRule returns a new child rule set that applies a custom validation rule.
// The custom rule is evaluated during validation and any errors it returns are included in the validation result.
//
// Use this when implementing custom rules.
func (v *InterfaceRuleSet[T]) WithRule(rule Rule[T]) *InterfaceRuleSet[T] {
	newRuleSet := v.clone()
	newRuleSet.rule = rule
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
//
// Use this when implementing custom rules.
func (v *InterfaceRuleSet[T]) WithRuleFunc(rule RuleFunc[T]) *InterfaceRuleSet[T] {
	return v.WithRule(rule)
}

// Interface is an identity function for this implementation and returns the current rule set.
func (v *InterfaceRuleSet[T]) Any() RuleSet[any] {
	return WrapAny(v)
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

// WithErrorMessage returns a new RuleSet with custom short and long error messages.
func (v *InterfaceRuleSet[T]) WithErrorMessage(short, long string) *InterfaceRuleSet[T] {
	return v.clone(interfaceWithLabel[T](util.FormatErrorMessageLabel(short, long)), interfaceWithErrorConfig[T](v.errorConfig.WithErrorMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (v *InterfaceRuleSet[T]) WithDocsURI(uri string) *InterfaceRuleSet[T] {
	return v.clone(interfaceWithLabel[T](util.FormatStringArgLabel("WithDocsURI", uri)), interfaceWithErrorConfig[T](v.errorConfig.WithDocsURI(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (v *InterfaceRuleSet[T]) WithTraceURI(uri string) *InterfaceRuleSet[T] {
	return v.clone(interfaceWithLabel[T](util.FormatStringArgLabel("WithTraceURI", uri)), interfaceWithErrorConfig[T](v.errorConfig.WithTraceURI(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (v *InterfaceRuleSet[T]) WithErrorCode(code errors.ErrorCode) *InterfaceRuleSet[T] {
	return v.clone(interfaceWithLabel[T](util.FormatErrorCodeLabel(code)), interfaceWithErrorConfig[T](v.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (v *InterfaceRuleSet[T]) WithErrorMeta(key string, value any) *InterfaceRuleSet[T] {
	return v.clone(interfaceWithLabel[T](util.FormatErrorMetaLabel(key, value)), interfaceWithErrorConfig[T](v.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (v *InterfaceRuleSet[T]) WithErrorCallback(fn errors.ErrorCallback) *InterfaceRuleSet[T] {
	return v.clone(interfaceWithLabel[T](util.FormatErrorCallbackLabel()), interfaceWithErrorConfig[T](v.errorConfig.WithCallback(fn)))
}
