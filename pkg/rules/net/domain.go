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

// domainLabelPattern matches valid domains after they have been converted to punycode
var domainLabelPatter = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]$`)

// DomainRuleSet implements the RuleSet interface for the domain names.
type DomainRuleSet struct {
	required bool
	parent   *DomainRuleSet
	rule     rules.Rule[string]
}

// NewDomain creates a new domain RuleSet
func NewDomain() *DomainRuleSet {
	return &DomainRuleSet{}
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
		if !domainLabelPatter.MatchString(part) {
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

	allErrors := validateBasicDomain(ctx, valueStr)

	if len(allErrors) > 0 {
		return valueStr, allErrors
	}

	currentRuleSet := ruleSet
	ctx = rulecontext.WithRuleSet(ctx, ruleSet)

	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			newStr, errs := currentRuleSet.rule.Evaluate(ctx, valueStr)
			if errs != nil {
				allErrors = append(allErrors, errs...)
			} else {
				valueStr = newStr
			}
		}

		currentRuleSet = currentRuleSet.parent
	}

	if len(allErrors) > 0 {
		return valueStr, allErrors
	} else {
		return valueStr, nil
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
		parent:   ruleSet,
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

func (ruleSet *DomainRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[string](ruleSet)
}
