package rules

import (
	"context"
	"fmt"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
)

type constCache[T comparable] map[T]*ConstantRuleSet[T]

var constCacheMap map[any]any

// ConstantRuleSet implements RuleSet that returns an error for
// any value that does not match the constant.
//
// This is primarily used for conditional validation. To test a constant of a specific
// type it is usually best to use that type.
type ConstantRuleSet[T comparable] struct {
	required    bool
	withNil     bool
	value       T
	empty       T // Leave this empty
	errorConfig *errors.ErrorConfig
}

// Constant creates a new Constant rule set for the specified value.
// Constant returns the same Rule Set when called multiple times with the same value.
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

// constantCloneOption is a functional option for cloning ConstantRuleSet.
type constantCloneOption[T comparable] func(*ConstantRuleSet[T])

// clone returns a shallow copy of the rule set.
func (ruleSet *ConstantRuleSet[T]) clone(options ...constantCloneOption[T]) *ConstantRuleSet[T] {
	newRuleSet := &ConstantRuleSet[T]{
		value:       ruleSet.value,
		required:    ruleSet.required,
		withNil:     ruleSet.withNil,
		errorConfig: ruleSet.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func constantWithErrorConfig[T comparable](config *errors.ErrorConfig) constantCloneOption[T] {
	return func(rs *ConstantRuleSet[T]) { rs.errorConfig = config }
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (ruleSet *ConstantRuleSet[T]) WithRequired() *ConstantRuleSet[T] {
	if ruleSet.required {
		return ruleSet
	}

	newRuleSet := ruleSet.clone()
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (ruleSet *ConstantRuleSet[T]) WithNil() *ConstantRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.withNil = true
	return newRuleSet
}

// Apply validates a RuleSet against an input value and assigns the validated value to output.
// Apply returns a ValidationErrorCollection.
func (ruleSet *ConstantRuleSet[T]) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	// Add error config to context for error customization
	ctx = errors.WithErrorConfig(ctx, ruleSet.errorConfig)

	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, ruleSet.withNil, input, output); handled {
		return err
	}

	// Attempt to coerce input to type T.
	v, ok := input.(T)
	if !ok {
		// Return a coercion error if input is not of type T.
		return errors.Collection(errors.Error(errors.CodeType, ctx, reflect.TypeOf(ruleSet.empty).String(), reflect.TypeOf(input).String()))
	}

	// Ensure the output is assignable to the coerced value.
	outVal := reflect.ValueOf(output)
	if outVal.Kind() != reflect.Ptr || outVal.IsNil() || !reflect.ValueOf(v).Type().AssignableTo(outVal.Elem().Type()) {
		// Return an error if the output is not assignable.
		return errors.Collection(errors.Errorf(errors.CodeInternal, ctx, "internal error", "Cannot assign %T to %T", input, output))
	}

	// Assign the validated value to the output.
	outVal.Elem().Set(reflect.ValueOf(v))

	// Evaluate the RuleSet and return any validation errors.
	return ruleSet.Evaluate(ctx, v)
}

// Evaluate performs validation of a RuleSet against a value and returns any errors.
func (ruleSet *ConstantRuleSet[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	if value != ruleSet.value {
		return errors.Collection(errors.Errorf(errors.CodePattern, ctx, "value mismatch", "value does not match"))
	}
	return nil
}

// Replaces returns true for all rules since by definition no rule can be a superset of a constant rule.
func (ruleSet *ConstantRuleSet[T]) Replaces(other Rule[T]) bool {
	return true
}

// Any returns the current rule set wrapped as a RuleSet[any].
func (ruleSet *ConstantRuleSet[T]) Any() RuleSet[any] {
	return WrapAny(ruleSet)
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

// WithErrorMessage returns a new RuleSet with custom short and long error messages.
func (ruleSet *ConstantRuleSet[T]) WithErrorMessage(short, long string) *ConstantRuleSet[T] {
	return ruleSet.clone(constantWithErrorConfig[T](ruleSet.errorConfig.WithErrorMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (ruleSet *ConstantRuleSet[T]) WithDocsURI(uri string) *ConstantRuleSet[T] {
	return ruleSet.clone(constantWithErrorConfig[T](ruleSet.errorConfig.WithDocsURI(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (ruleSet *ConstantRuleSet[T]) WithTraceURI(uri string) *ConstantRuleSet[T] {
	return ruleSet.clone(constantWithErrorConfig[T](ruleSet.errorConfig.WithTraceURI(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (ruleSet *ConstantRuleSet[T]) WithErrorCode(code errors.ErrorCode) *ConstantRuleSet[T] {
	return ruleSet.clone(constantWithErrorConfig[T](ruleSet.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (ruleSet *ConstantRuleSet[T]) WithErrorMeta(key string, value any) *ConstantRuleSet[T] {
	return ruleSet.clone(constantWithErrorConfig[T](ruleSet.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (ruleSet *ConstantRuleSet[T]) WithErrorCallback(fn errors.ErrorCallback) *ConstantRuleSet[T] {
	return ruleSet.clone(constantWithErrorConfig[T](ruleSet.errorConfig.WithCallback(fn)))
}
