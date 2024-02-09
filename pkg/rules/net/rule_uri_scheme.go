package net

import (
	"proto.zip/studio/validate/internal/util"
)

// WithAllowedSchemes returns a new child RuleSet that is checked against the provided list of allowed values.
//
// This method can be called more than once and the allowed values are cumulative.
// Allowed values must still pass all other rules.
func (ruleSet *URIRuleSet) WithAllowedSchemes(value string, rest ...string) *URIRuleSet {
	newRuleSet := ruleSet.copyWithParent(ruleSet)
	newRuleSet.schemeRuleSet = newRuleSet.schemeRuleSet.WithAllowedValues(value, rest...)

	list := append([]string{value}, rest...)

	newRuleSet.label = util.StringsToRuleOutput[string]("WithAllowedSchemes", list)
	return newRuleSet
}
