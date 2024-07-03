package net

import (
	"context"
	"reflect"
	"regexp"
	"strings"

	"golang.org/x/net/idna"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// backgroundDomainRuleSet is the base domain rule set. Since rule sets are immutable.
var backgroundDomainRuleSet DomainRuleSet = DomainRuleSet{
	label: "DomainRuleSet",
}

// domainLabelPattern matches valid domains after they have been converted to punycode
var domainLabelPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]$`)

// DomainRuleSet implements the RuleSet interface for the domain names.
type DomainRuleSet struct {
	rules.NoConflict[string]
	required bool
	parent   *DomainRuleSet
	rule     rules.Rule[string]
	label    string
}

// NewDomain creates a new domain RuleSet
func NewDomain() *DomainRuleSet {
	return &backgroundDomainRuleSet
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *DomainRuleSet) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (ruleSet *DomainRuleSet) WithRequired() *DomainRuleSet {
	return &DomainRuleSet{
		required: true,
		parent:   ruleSet,
		label:    "WithRequired()",
	}
}

// Validate performs a validation of a RuleSet against a value and returns a string value or
// a ValidationErrorCollection.
func (ruleSet *DomainRuleSet) Validate(value any) (string, errors.ValidationErrorCollection) {
	return ruleSet.ValidateWithContext(value, context.Background())
}

// validateBasicDomain performs general domain validation that is valid for any and all domains.
// This function always returns a collection even if it is empty.
func validateBasicDomain(ctx context.Context, value string) errors.ValidationErrorCollection {
	allErrors := errors.Collection()

	// Convert to punycode
	punycode, err := idna.ToASCII(value)

	if err != nil {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "domain contains invalid unicode"))
		return allErrors
	}

	// Check total length
	if len(punycode) >= 256 {
		allErrors = append(allErrors, errors.Errorf(errors.CodeMax, ctx, "domain exceeds maximum length"))
		return allErrors
	}

	// Each labels should contain only valid characters
	parts := strings.Split(punycode, ".")

	for _, part := range parts {
		if !domainLabelPattern.MatchString(part) {
			allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "domain segment is invalid"))
			break
		}
	}

	return allErrors
}

// Validate performs a validation of a RuleSet against a value and returns a string value or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by rules and error formatting.
func (ruleSet *DomainRuleSet) ValidateWithContext(value any, ctx context.Context) (string, errors.ValidationErrorCollection) {

	valueStr, ok := value.(string)

	if !ok {
		return "", errors.Collection(errors.NewCoercionError(ctx, "string", reflect.ValueOf(value).Kind().String()))
	}

	return valueStr, ruleSet.Evaluate(ctx, valueStr)
}

// Evaluate performs a validation of a RuleSet against a string and returns an object value of the
// same type or a ValidationErrorCollection.
func (ruleSet *DomainRuleSet) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	allErrors := validateBasicDomain(ctx, value)

	if len(allErrors) > 0 {
		return allErrors
	}

	currentRuleSet := ruleSet
	ctx = rulecontext.WithRuleSet(ctx, ruleSet)

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
func (ruleSet *DomainRuleSet) noConflict(rule rules.Rule[string]) *DomainRuleSet {
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

	return &DomainRuleSet{
		rule:     ruleSet.rule,
		parent:   newParent,
		required: ruleSet.required,
		label:    ruleSet.label,
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the string type.
//
// Use this when implementing custom rules.
func (ruleSet *DomainRuleSet) WithRule(rule rules.Rule[string]) *DomainRuleSet {
	return &DomainRuleSet{
		rule:     rule,
		parent:   ruleSet.noConflict(rule),
		required: ruleSet.required,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule interface
// for the string type.
//
// Use this when implementing custom rules.
func (v *DomainRuleSet) WithRuleFunc(rule rules.RuleFunc[string]) *DomainRuleSet {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the domain RuleSet in any Any rule set
// which can then be used in nested validation.
func (ruleSet *DomainRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[string](ruleSet)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *DomainRuleSet) String() string {
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
