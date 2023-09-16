// Package string provides a RuleSet implementation that can be used to validate bstring values.
//
// It implements standard rules and allows the developer to set a rule set used to validate items.
package strings

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// Implementation of RuleSet for strings.
type StringRuleSet struct {
	strict   bool
	rule     rules.Rule[string]
	required bool
	parent   *StringRuleSet
}

// New creates a new string RuleSet.
func New() *StringRuleSet {
	return &StringRuleSet{}
}

// WithStrict returns a new child RuleSet with the strict flag applied.
// A strict rule will only validate if the value is already a string.
func (v *StringRuleSet) WithStrict() *StringRuleSet {
	return &StringRuleSet{
		strict:   true,
		parent:   v,
		required: v.required,
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
	}
}

// Validate performs a validation of a RuleSet against a value and returns a stringvalue or
// a ValidationErrorCollection.
func (v *StringRuleSet) Validate(value any) (string, errors.ValidationErrorCollection) {
	return v.ValidateWithContext(value, context.Background())
}

// Validate performs a validation of a RuleSet against a value and returns a stringvalue or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by validaton rules and error formatting.
func (v *StringRuleSet) ValidateWithContext(value interface{}, ctx context.Context) (string, errors.ValidationErrorCollection) {
	allErrors := errors.Collection()

	str, validationErr := v.coerce(value, ctx)

	if validationErr != nil {
		allErrors.Add(validationErr)
		return "", allErrors
	}

	currentRuleSet := v
	ctx = rulecontext.WithRuleSet(ctx, v)

	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			newStr, errs := currentRuleSet.rule.Evaluate(ctx, str)
			if errs != nil {
				allErrors.Add(errs.All()...)
			} else {
				str = newStr
			}
		}

		currentRuleSet = currentRuleSet.parent
	}

	if allErrors.Size() > 0 {
		return str, allErrors
	} else {
		return str, nil
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the string type.
//
// Use this when implementing custom rules.
func (v *StringRuleSet) WithRule(rule rules.Rule[string]) *StringRuleSet {
	return &StringRuleSet{
		strict:   v.strict,
		rule:     rule,
		parent:   v,
		required: v.required,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// for the string type.
//
// Use this when implementing custom rules.
func (v *StringRuleSet) WithRuleFunc(rule rules.RuleFunc[string]) *StringRuleSet {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the string RuleSet in any Any rule set
// which can then be used in nested validation.
func (v *StringRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[string](v)
}
