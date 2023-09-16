package rules

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// AnyRuleSet implements RuleSet for the "any" interface.
// Use when you don't care about the date type passed in and want to return it unaltered from the Validate method.
//
// See also: WrapAny which also implements the "any" interface and wraps another RuleSet.
type AnyRuleSet struct {
	required  bool
	forbidden bool
	rule      Rule[any]
	parent    *AnyRuleSet
}

// Any creates a new Any rule set.
func Any() *AnyRuleSet {
	return &AnyRuleSet{}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *AnyRuleSet) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (v *AnyRuleSet) WithRequired() *AnyRuleSet {
	return &AnyRuleSet{
		required:  true,
		forbidden: v.forbidden,
		parent:    v,
	}
}

// WithForbidden returns a new child rule set with the forbidden flag set.
// Use WithForbidden when a value is expected to always be nil or omitted.
func (v *AnyRuleSet) WithForbidden() *AnyRuleSet {
	return &AnyRuleSet{
		required:  v.required,
		forbidden: true,
		parent:    v,
	}
}

// Validate performs a validation of a RuleSet against a value and returns the unaltered supplied value
// or a ValidationErrorCollection.
func (v *AnyRuleSet) Validate(value interface{}) (any, errors.ValidationErrorCollection) {
	return v.ValidateWithContext(value, context.Background())
}

// ValidateWithContext performs a validation of a RuleSet against a value and returns the unaltered supplied value
// or a ValidationErrorCollection.
//
// Also, takes a Context which can be used by validaton rules and error formatting.
func (v *AnyRuleSet) ValidateWithContext(value interface{}, ctx context.Context) (any, errors.ValidationErrorCollection) {
	if v.forbidden {
		return nil, errors.Collection(errors.Errorf(errors.CodeUnexpected, ctx, "value is not allowed"))
	}

	allErrors := errors.Collection()
	retValue := value

	currentRuleSet := v
	ctx = rulecontext.WithRuleSet(ctx, v)

	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			newStr, err := currentRuleSet.rule.Evaluate(ctx, retValue)
			if err != nil {
				allErrors.Add(err.All()...)
			} else {
				retValue = newStr
			}
		}

		currentRuleSet = currentRuleSet.parent
	}

	if allErrors.Size() != 0 {
		return retValue, allErrors
	} else {
		return retValue, nil
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// explicitly for the "any" interface.
//
// Use this when implementing custom rules.
func (v *AnyRuleSet) WithRule(rule Rule[any]) *AnyRuleSet {
	return &AnyRuleSet{
		required:  v.required,
		forbidden: v.forbidden,
		rule:      rule,
		parent:    v,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// explicitly for the "any" interface.
//
// Use this when implementing custom rules.
func (v *AnyRuleSet) WithRuleFunc(rule RuleFunc[any]) *AnyRuleSet {
	return v.WithRule(rule)
}

// Any is an identity function for this implementation and returns the current rule set.
func (v *AnyRuleSet) Any() RuleSet[any] {
	return v
}
