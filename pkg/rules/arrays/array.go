package arrays

import (
	"context"
	"reflect"
	"strconv"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// Implementation of RuleSet for arrays of a given type.
type ArrayRuleSet[T any] struct {
	itemRules rules.RuleSet[T]
	rule      rules.Rule[[]T]
	required  bool
	parent    *ArrayRuleSet[T]
}

// NewInt creates a new array RuleSet.
func New[T any]() *ArrayRuleSet[T] {
	return &ArrayRuleSet[T]{}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *ArrayRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
// Required has no effect on integer if the RuleSet is strict since nil is not a valid number.
func (v *ArrayRuleSet[T]) WithRequired() *ArrayRuleSet[T] {
	return &ArrayRuleSet[T]{
		parent:   v,
		required: true,
	}
}

// WithItemRuleSet takes a new rule set to use to validate array items and returns a new child rule set.
//
// If this function is called more than once, only the most recent one will be used to validate the items.
// If you don't set an item rule set then the validator will attempt to cast the items to the correct type
// and perform no additional validation.
func (v *ArrayRuleSet[T]) WithItemRuleSet(itemRules rules.RuleSet[T]) *ArrayRuleSet[T] {
	return &ArrayRuleSet[T]{
		itemRules: itemRules,
		parent:    v,
		required:  v.required,
	}
}

// Validate performs a validation of a RuleSet against a value and returns an array of the correct type or
// a ValidationErrorCollection.
func (v *ArrayRuleSet[T]) Validate(value any) ([]T, errors.ValidationErrorCollection) {
	return v.ValidateWithContext(value, context.Background())
}

// ValidateWithContext performs a validation of a RuleSet against a value and returns an array of the correct type or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by rules and error formatting.
func (v *ArrayRuleSet[T]) ValidateWithContext(value any, ctx context.Context) ([]T, errors.ValidationErrorCollection) {

	valueOf := reflect.ValueOf(value)
	typeOf := valueOf.Type()
	kind := typeOf.Kind()

	if kind != reflect.Slice && kind != reflect.Array {
		return nil, errors.Collection(errors.NewCoercionError(ctx, "array", kind.String()))
	}

	l := valueOf.Len()

	output := make([]T, l)

	var allErrors = errors.Collection()

	// Check for a RuleSet first
	var itemRuleSet rules.RuleSet[T]

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.itemRules != nil {
			itemRuleSet = currentRuleSet.itemRules
			break
		}
	}

	// Default to a plain type cast if the rule set is nil
	if itemRuleSet == nil {
		var ok bool
		expected := ""

		for i := 0; i < l; i++ {
			output[i], ok = valueOf.Index(i).Interface().(T)
			if !ok {
				subContext := rulecontext.WithPathString(ctx, strconv.Itoa(i))
				if expected == "" {
					expected = reflect.TypeOf(new(T)).Name()
				}
				actual := valueOf.Index(i).Kind().String()
				allErrors = append(allErrors, errors.NewCoercionError(subContext, expected, actual))
			}
		}
	} else {
		var itemErrors errors.ValidationErrorCollection
		for i := 0; i < l; i++ {
			subContext := rulecontext.WithPathIndex(ctx, i)
			output[i], itemErrors = itemRuleSet.ValidateWithContext(valueOf.Index(i).Interface(), subContext)
			if itemErrors != nil {
				allErrors = append(allErrors, itemErrors...)
			}
		}
	}

	// Next apply array level rules
	// This must be done after the item rules because we want to make sure all values are cast first.
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			newOutput, err := currentRuleSet.rule.Evaluate(ctx, output)
			if err != nil {
				allErrors = append(allErrors, err...)
			} else {
				output = newOutput
			}
		}
	}

	if len(allErrors) != 0 {
		return output, allErrors
	} else {
		return output, nil
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the given array and item type.
//
// Use this when implementing custom rules.
func (v *ArrayRuleSet[T]) WithRule(rule rules.Rule[[]T]) *ArrayRuleSet[T] {
	return &ArrayRuleSet[T]{
		rule:     rule,
		parent:   v,
		required: v.required,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// for the given array and item type.
//
// Use this when implementing custom rules.
func (v *ArrayRuleSet[T]) WithRuleFunc(rule rules.RuleFunc[[]T]) *ArrayRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the array RuleSet in any Any rule set
// which can then be used in nested validation.
func (v *ArrayRuleSet[T]) Any() rules.RuleSet[any] {
	return rules.WrapAny[[]T](v)
}
