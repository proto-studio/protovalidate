package rules

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// Implementation of RuleSet for arrays of a given type.
type SliceRuleSet[T any] struct {
	NoConflict[[]T]
	itemRules RuleSet[T]
	rule      Rule[[]T]
	required  bool
	withNil   bool
	parent    *SliceRuleSet[T]
	label     string
}

// Slice creates a new slice RuleSet.
func Slice[T any]() *SliceRuleSet[T] {
	var empty [0]T

	return &SliceRuleSet[T]{
		label: fmt.Sprintf("SliceRuleSet[%s]", reflect.TypeOf(empty).Elem().Kind()),
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *SliceRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
// WithRequired has no effect on slices if the RuleSet is strict since nil is not a valid slice.
func (v *SliceRuleSet[T]) WithRequired() *SliceRuleSet[T] {
	return &SliceRuleSet[T]{
		parent:   v,
		required: true,
		withNil:  v.withNil,
		label:    "WithRequired()",
	}
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *SliceRuleSet[T]) WithNil() *SliceRuleSet[T] {
	return &SliceRuleSet[T]{
		parent:   v,
		required: v.required,
		withNil:  true,
		label:    "WithNil()",
	}
}

// WithItemRuleSet takes a new rule set to use to validate array items and returns a new child rule set.
//
// If this function is called more than once, only the most recent one will be used to validate the items.
// If you don't set an item rule set then the validator will attempt to cast the items to the correct type
// and perform no additional validation.
func (v *SliceRuleSet[T]) WithItemRuleSet(itemRules RuleSet[T]) *SliceRuleSet[T] {
	return &SliceRuleSet[T]{
		itemRules: itemRules,
		parent:    v,
		required:  v.required,
		withNil:   v.withNil,
	}
}

// Apply performs validation of a RuleSet against a value and assigns the result to the output parameter.
// Apply returns a ValidationErrorCollection if any validation errors occur.
func (v *SliceRuleSet[T]) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, v.withNil, input, output); handled {
		return err
	}

	// Ensure output is a non-nil pointer
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Output must be a non-nil pointer",
		))
	}

	valueOf := reflect.ValueOf(input)
	typeOf := valueOf.Type()
	kind := typeOf.Kind()

	if kind != reflect.Slice && kind != reflect.Array {
		return errors.Collection(errors.NewCoercionError(ctx, "array", kind.String()))
	}

	l := valueOf.Len()

	outputSlice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf((*T)(nil)).Elem()), l, l)

	var allErrors = errors.Collection()

	// Check for an item RuleSet
	var itemRuleSet RuleSet[T]

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.itemRules != nil {
			itemRuleSet = currentRuleSet.itemRules
			break
		}
	}

	// Default to a plain type cast if the rule set is nil
	if itemRuleSet == nil {
		expected := ""

		for i := 0; i < l; i++ {
			item := valueOf.Index(i).Interface()
			castItem, castOk := item.(T)
			outputSlice.Index(i).Set(reflect.ValueOf(castItem))
			if !castOk {
				subContext := rulecontext.WithPathString(ctx, strconv.Itoa(i))
				if expected == "" {
					expected = reflect.TypeOf(new(T)).Elem().Name()
				}
				actual := valueOf.Index(i).Kind().String()
				allErrors = append(allErrors, errors.NewCoercionError(subContext, expected, actual))
			}
		}
	} else {
		for i := 0; i < l; i++ {
			subContext := rulecontext.WithPathIndex(ctx, i)
			item := valueOf.Index(i).Interface()

			// Prepare the output location for the item
			var itemOutput T
			itemErr := itemRuleSet.Apply(subContext, item, &itemOutput)
			outputSlice.Index(i).Set(reflect.ValueOf(itemOutput))

			if itemErr != nil {
				allErrors = append(allErrors, itemErr...)
			}
		}
	}

	// Apply array-level rules after all items are validated and cast
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			if err := currentRuleSet.rule.Evaluate(ctx, outputSlice.Interface().([]T)); err != nil {
				allErrors = append(allErrors, err...)
			}
		}
	}

	// Assign the result to the output
	outputElem := outputVal.Elem()
	if outputElem.Kind() == reflect.Interface && outputElem.IsNil() {
		outputElem.Set(outputSlice)
	} else if outputSlice.Type().AssignableTo(outputElem.Type()) {
		outputElem.Set(outputSlice)
	} else {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Cannot assign %T to %T", outputSlice.Interface(), outputElem.Interface(),
		))
	}

	// Return any accumulated errors
	if len(allErrors) != 0 {
		return allErrors
	}

	return nil
}

// Evaluate performs validation of a RuleSet against a slice type and returns a ValidationErrorCollection.
func (ruleSet *SliceRuleSet[T]) Evaluate(ctx context.Context, value []T) errors.ValidationErrorCollection {
	var out any
	return ruleSet.Apply(ctx, value, &out)
}

// noConflict returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *SliceRuleSet[T]) noConflict(rule Rule[[]T]) *SliceRuleSet[T] {

	if ruleSet.rule != nil {

		// Conflicting rules, skip this and return the parent
		if rule.Conflict(ruleSet.rule) {
			return ruleSet.parent.noConflict(rule)
		}

	}

	if ruleSet.parent == nil {
		return ruleSet
	}

	newParent := ruleSet.parent.noConflict(rule)

	if newParent == ruleSet.parent {
		return ruleSet
	}

	return &SliceRuleSet[T]{
		rule:      ruleSet.rule,
		parent:    newParent,
		required:  ruleSet.required,
		withNil:   ruleSet.withNil,
		itemRules: ruleSet.itemRules,
		label:     ruleSet.label,
	}
}

// WithRule returns a new child rule set that applies a custom validation rule.
// The custom rule is evaluated during validation and any errors it returns are included in the validation result.
func (v *SliceRuleSet[T]) WithRule(rule Rule[[]T]) *SliceRuleSet[T] {
	return &SliceRuleSet[T]{
		rule:     rule,
		parent:   v.noConflict(rule),
		required: v.required,
		withNil:  v.withNil,
	}
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
func (v *SliceRuleSet[T]) WithRuleFunc(rule RuleFunc[[]T]) *SliceRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the slice RuleSet in an Any rule set
// which can then be used in nested validation.
func (v *SliceRuleSet[T]) Any() RuleSet[any] {
	return WrapAny[[]T](v)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *SliceRuleSet[T]) String() string {
	label := ruleSet.label

	if label == "" {
		if ruleSet.rule != nil {
			label = ruleSet.rule.String()
		} else if ruleSet.itemRules != nil {
			label = fmt.Sprintf("WithItemRuleSet(%s)", ruleSet.itemRules)
		}
	}

	if ruleSet.parent != nil {
		return ruleSet.parent.String() + "." + label
	}
	return label
}
