package net

import (
	"context"
	"reflect"
	"strings"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// EmailRuleSet implements the RuleSet interface for the domain names.
type EmailRuleSet struct {
	rules.NoConflict[string]
	required      bool
	parent        *EmailRuleSet
	rule          rules.Rule[string]
	domainRuleSet rules.RuleSet[any]
	label         string
}

// backgroundEmailRuleSet is the base email rule set. Since rule sets are immutable.
var backgroundEmailRuleSet EmailRuleSet = EmailRuleSet{
	label: "EmailRuleSet",
}

// NewEmail creates a new domain RuleSet
func NewEmail() *EmailRuleSet {
	return &backgroundEmailRuleSet
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *EmailRuleSet) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (ruleSet *EmailRuleSet) WithRequired() *EmailRuleSet {
	return &EmailRuleSet{
		required:      true,
		parent:        ruleSet,
		domainRuleSet: ruleSet.domainRuleSet,
		label:         "WithRequired()",
	}
}

// Validate performs a validation of a RuleSet against a value and returns a string value or
// a ValidationErrorCollection.
func (ruleSet *EmailRuleSet) Validate(value any) (string, errors.ValidationErrorCollection) {
	return ruleSet.ValidateWithContext(value, context.Background())
}

// validateBasicEmail performs general domain validation that is valid for any and all domains.
// This function always returns a collection even if it is empty.
func (ruleSet *EmailRuleSet) validateBasicEmail(ctx context.Context, value string) errors.ValidationErrorCollection {
	allErrors := errors.Collection()

	parts := strings.Split(value, "@")

	if len(parts) < 2 {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "Missing @ symbol"))
		return allErrors
	}
	if len(parts) > 2 {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "Too many @ symbols"))
		return allErrors
	}

	local := parts[0]
	domain := parts[1]

	domainRuleSet := ruleSet.domainRuleSet
	if domainRuleSet == nil {
		domainRuleSet = NewDomain().WithTLD().Any()
	}

	_, domainErrs := domainRuleSet.ValidateWithContext(domain, ctx)

	if len(domainErrs) > 0 {
		allErrors = append(allErrors, domainErrs...)
	}

	if len(local) == 0 {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "Local address is empty"))
		return allErrors
	}

	if strings.HasPrefix(local, ".") {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "Address cannot start with a dot"))
	}

	if strings.HasSuffix(local, ".") {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "Address cannot end with a dot"))
	}

	if strings.Contains(local, "..") {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "Address cannot contain \"..\""))
	}

	return allErrors
}

// Validate performs a validation of a RuleSet against a value and returns a string value or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by rules and error formatting.
func (ruleSet *EmailRuleSet) ValidateWithContext(value any, ctx context.Context) (string, errors.ValidationErrorCollection) {

	valueStr, ok := value.(string)

	if !ok {
		return "", errors.Collection(errors.NewCoercionError(ctx, "string", reflect.ValueOf(value).Kind().String()))
	}

	return valueStr, ruleSet.Evaluate(ctx, valueStr)
}

// Evaluate performs a validation of a RuleSet against a string and returns an object value of the
// same type or a ValidationErrorCollection.
func (ruleSet *EmailRuleSet) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {

	allErrors := ruleSet.validateBasicEmail(ctx, value)

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

// WithDomain returns a new child rule set with the domain validator assigned to
// the provided RuleSet instead of the default domain rule set.
//
// The default domain rule set for email validation is the equivalent of:
//
//	NewDomain().WithTLD()
func (ruleSet *EmailRuleSet) WithDomain(domainRuleSet rules.RuleSet[any]) *EmailRuleSet {
	return &EmailRuleSet{
		parent:        ruleSet,
		required:      ruleSet.required,
		domainRuleSet: domainRuleSet,
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the string type.
//
// Use this when implementing custom rules.
func (ruleSet *EmailRuleSet) WithRule(rule rules.Rule[string]) *EmailRuleSet {
	return &EmailRuleSet{
		rule:          rule,
		parent:        ruleSet,
		required:      ruleSet.required,
		domainRuleSet: ruleSet.domainRuleSet,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule interface
// for the string type.
//
// Use this when implementing custom rules.
func (v *EmailRuleSet) WithRuleFunc(rule rules.RuleFunc[string]) *EmailRuleSet {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the domain RuleSet in any Any rule set
// which can then be used in nested validation.
func (ruleSet *EmailRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[string](ruleSet)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *EmailRuleSet) String() string {
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
