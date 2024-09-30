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
	domainRuleSet rules.RuleSet[string]
	label         string
}

// baseEmailRuleSet is the base email rule set. Since rule sets are immutable.
var baseEmailRuleSet EmailRuleSet = EmailRuleSet{
	label: "EmailRuleSet",
}

// Email returns the base email RuleSet.
func Email() *EmailRuleSet {
	return &baseEmailRuleSet
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

// Apply performs a validation of a RuleSet against a value and assigns the result to the output parameter.
// It returns a ValidationErrorCollection if any validation errors occur.
func (ruleSet *EmailRuleSet) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	// Attempt to cast the input to a string
	valueStr, ok := input.(string)
	if !ok {
		return errors.Collection(errors.NewCoercionError(ctx, "string", reflect.ValueOf(input).Kind().String()))
	}

	// Perform the validation
	if err := ruleSet.Evaluate(ctx, valueStr); err != nil {
		return err
	}

	outputVal := reflect.ValueOf(output)

	// Check if the output is a non-nil pointer
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Output must be a non-nil pointer",
		))
	}

	// Dereference the pointer to get the actual value that needs to be set
	outputElem := outputVal.Elem()

	switch outputElem.Kind() {
	case reflect.String:
		outputElem.SetString(valueStr)
	case reflect.Interface:
		outputElem.Set(reflect.ValueOf(valueStr))
	default:
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Cannot assign string to %T", output,
		))
	}

	return nil
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
		domainRuleSet = Domain().WithTLD()
	}

	domainErrs := domainRuleSet.Evaluate(ctx, domain)

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
func (ruleSet *EmailRuleSet) WithDomain(domainRuleSet rules.RuleSet[string]) *EmailRuleSet {
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
