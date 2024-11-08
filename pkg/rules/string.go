package rules

import (
	"context"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// Implementation of RuleSet for strings.
type StringRuleSet struct {
	NoConflict[string]
	strict   bool
	rule     Rule[string]
	required bool
	parent   *StringRuleSet
	label    string
}

// baseStringRuleSet is the main RuleSet.
// New returns this since rule sets are immutable and StringRuleSet does not contain generics.
var baseStringRuleSet StringRuleSet = StringRuleSet{
	label: "StringRuleSet",
}

// String returns the base StringRuleSet.
func String() *StringRuleSet {
	return &baseStringRuleSet
}

// WithStrict returns a new child RuleSet with the strict flag applied.
// A strict rule will only validate if the value is already a string.
func (v *StringRuleSet) WithStrict() *StringRuleSet {
	return &StringRuleSet{
		strict:   true,
		parent:   v,
		required: v.required,
		label:    "WithStrict()",
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *StringRuleSet) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (v *StringRuleSet) WithRequired() *StringRuleSet {
	return &StringRuleSet{
		strict:   v.strict,
		parent:   v,
		required: true,
		label:    "WithRequired()",
	}
}

// Apply performs a validation of a RuleSet against a value and assigns the resulting string to the output pointer
// a ValidationErrorCollection.
func (v *StringRuleSet) Apply(ctx context.Context, value, output any) errors.ValidationErrorCollection {
	// Ensure output is a pointer that can be set
	rv := reflect.ValueOf(output)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.Collection(
			errors.Errorf(errors.CodeInternal, ctx, "Output must be a non-nil pointer"),
		)
	}

	// Attempt to coerce the input to a string
	str, validationErr := v.coerce(value, ctx)

	if validationErr != nil {
		return errors.Collection(validationErr)
	}

	verrs := v.Evaluate(ctx, str)
	if verrs != nil {
		return verrs
	}

	// Set the string result in the output parameter
	elem := rv.Elem()

	// Check if the output is an interface
	if elem.Kind() == reflect.Interface {
		// Create a new string value and set the interface to point to it
		elem.Set(reflect.ValueOf(str))
		return nil
	}

	// If the element is a string, replace it with the new string value
	if elem.Kind() == reflect.String {
		elem.SetString(str)
		return nil
	}

	return errors.Collection(
		errors.Errorf(errors.CodeInternal, ctx, "Cannot assign string to %T", output),
	)
}

// Evaluate performs a validation of a RuleSet against a string value and returns a string value or
// a ValidationErrorCollection.
func (v *StringRuleSet) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
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

	if len(allErrors) > 0 {
		return allErrors
	} else {
		return nil
	}
}

// noConflict returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *StringRuleSet) noConflict(rule Rule[string]) *StringRuleSet {
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

	return &StringRuleSet{
		rule:     ruleSet.rule,
		parent:   newParent,
		required: ruleSet.required,
		strict:   ruleSet.strict,
		label:    ruleSet.label,
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the string type.
//
// Use this when implementing custom rules.
func (ruleSet *StringRuleSet) WithRule(rule Rule[string]) *StringRuleSet {
	return &StringRuleSet{
		strict:   ruleSet.strict,
		rule:     rule,
		parent:   ruleSet.noConflict(rule),
		required: ruleSet.required,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// for the string type.
//
// Use this when implementing custom rules.
func (v *StringRuleSet) WithRuleFunc(rule RuleFunc[string]) *StringRuleSet {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the string RuleSet in any Any rule set
// which can then be used in nested validation.
func (v *StringRuleSet) Any() RuleSet[any] {
	return WrapAny[string](v)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *StringRuleSet) String() string {
	label := ruleSet.label

	if label == "" && ruleSet.rule != nil {
		label = ruleSet.rule.String()
	}

	if ruleSet.parent != nil {
		return ruleSet.parent.String() + "." + label
	}
	return label
}
