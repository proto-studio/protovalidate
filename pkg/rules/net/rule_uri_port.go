package net

import (
	"fmt"

	"proto.zip/studio/validate/internal/util"
)

// WithMinPort returns a new rule set that validates the port number is at least the specified minimum.
func (ruleSet *URIRuleSet) WithMinPort(min int) *URIRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.portRuleSet = newRuleSet.portRuleSet.WithMin(min)
	newRuleSet.label = fmt.Sprintf("WithMinPort(%d)", min)
	return newRuleSet
}

// WithMaxPort returns a new rule set that validates the port number is at most the specified maximum.
func (ruleSet *URIRuleSet) WithMaxPort(max int) *URIRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.portRuleSet = newRuleSet.portRuleSet.WithMax(max)
	newRuleSet.label = fmt.Sprintf("WithMaxPort(%d)", max)
	return newRuleSet
}

// WithAllowedPorts returns a new child RuleSet that is checked against the provided list of allowed values.
//
// This method can be called more than once and the allowed values are cumulative.
// Allowed values must still pass all other rules.
func (ruleSet *URIRuleSet) WithAllowedPorts(value int, rest ...int) *URIRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.portRuleSet = newRuleSet.portRuleSet.WithAllowedValues(value, rest...)

	list := append([]int{value}, rest...)

	newRuleSet.label = util.StringsToRuleOutput[int]("WithAllowedPorts", list)
	return newRuleSet
}
