package rules

import (
	"context"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// AnyRuleSet implements RuleSet for the "any" interface.
// AnyRuleSet is used when you don't care about the data type passed in and want to return it unaltered from the Validate method.
//
// See also: WrapAny which also implements the "any" interface and wraps another RuleSet.
type AnyRuleSet struct {
	NoConflict[any]
	required  bool
	forbidden bool
	withNil   bool
	rule      Rule[any]
	parent    *AnyRuleSet
	label     string
}

// backgroundAnyRUleSet is the main AnyRuleSet.
// Any returns this since rule sets are immutable and AnyRuleSet does not contain generics.
var backgroundAnyRuleSet AnyRuleSet = AnyRuleSet{
	label: "AnyRuleSet",
}

// Any creates a new Any rule set.
func Any() *AnyRuleSet {
	return &backgroundAnyRuleSet
}

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (v *AnyRuleSet) clone() *AnyRuleSet {
	return &AnyRuleSet{
		required:  v.required,
		forbidden: v.forbidden,
		withNil:   v.withNil,
		parent:    v,
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *AnyRuleSet) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (v *AnyRuleSet) WithRequired() *AnyRuleSet {
	newRuleSet := v.clone()
	newRuleSet.required = true
	newRuleSet.label = "WithRequired()"
	return newRuleSet
}

// WithForbidden returns a new child rule set that requires values to be nil or omitted.
// When a value is present, validation fails with an error.
func (v *AnyRuleSet) WithForbidden() *AnyRuleSet {
	newRuleSet := v.clone()
	newRuleSet.forbidden = true
	newRuleSet.label = "WithForbidden()"
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *AnyRuleSet) WithNil() *AnyRuleSet {
	newRuleSet := v.clone()
	newRuleSet.withNil = true
	newRuleSet.label = "WithNil()"
	return newRuleSet
}

// Apply performs validation of a RuleSet against a value and assigns the value to the output.
// Apply returns a ValidationErrorCollection.
func (v *AnyRuleSet) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, v.withNil, input, output); handled {
		return err
	}

	err := v.Evaluate(ctx, input)
	if err != nil {
		return err
	}

	// Ensure output is a pointer
	rv := reflect.ValueOf(output)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.Collection(
			errors.Errorf(errors.CodeInternal, ctx, "Output must be a non-nil pointer"),
		)
	}

	// Get the element the pointer points to
	elem := rv.Elem()

	// Convert input to reflect.Value
	inputValue := reflect.ValueOf(input)

	// Check if the input can be assigned to the output
	if inputValue.Type().AssignableTo(elem.Type()) {
		elem.Set(inputValue)
		return nil
	}

	return errors.Collection(
		errors.Errorf(errors.CodeInternal, ctx, "Cannot assign %T to %T", input, output),
	)
}

// Evaluate performs validation of a RuleSet against a value and returns a ValidationErrorCollection.
// Evaluate calls wrapped rules before any rules added directly to the AnyRuleSet.
func (v *AnyRuleSet) Evaluate(ctx context.Context, value any) errors.ValidationErrorCollection {
	if v.forbidden {
		return errors.Collection(errors.Errorf(errors.CodeForbidden, ctx, "value is not allowed"))
	}

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

// WithRule returns a new child rule set that applies a custom validation rule.
// The custom rule is evaluated during validation and any errors it returns are included in the validation result.
func (v *AnyRuleSet) WithRule(rule Rule[any]) *AnyRuleSet {
	newRuleSet := v.clone()
	newRuleSet.rule = rule
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
func (v *AnyRuleSet) WithRuleFunc(rule RuleFunc[any]) *AnyRuleSet {
	return v.WithRule(rule)
}

// Any returns the current rule set.
func (v *AnyRuleSet) Any() RuleSet[any] {
	return v
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *AnyRuleSet) String() string {
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
