package net

import (
	"context"
	"reflect"
	"strings"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// conflictType identifies the type of method that was called on a ruleset.
// Used for fast conflict checking instead of slow string prefix matching.
type emailConflictType int

const (
	emailConflictTypeNone emailConflictType = iota
	emailConflictTypeRequired
	emailConflictTypeNil
)

// Conflict returns true if this conflict type conflicts with the other conflict type.
// Two conflict types conflict if they are the same (non-zero) value.
func (ct emailConflictType) Conflict(other emailConflictType) bool {
	return ct != emailConflictTypeNone && ct == other
}

// Replaces returns true if this conflict type replaces the given rule.
// It attempts to cast the rule to *EmailRuleSet and checks if the conflictType conflicts.
func (ct emailConflictType) Replaces(r rules.Rule[string]) bool {
	rs, ok := r.(*EmailRuleSet)
	if !ok {
		return false
	}
	return ct.Conflict(rs.conflictType)
}

// EmailRuleSet implements the RuleSet interface for the domain names.
type EmailRuleSet struct {
	rules.NoConflict[string]
	required      bool
	withNil       bool
	parent        *EmailRuleSet
	rule          rules.Rule[string]
	domainRuleSet rules.RuleSet[string]
	label         string
	conflictType  emailConflictType
	errorConfig   *errors.ErrorConfig
}

// baseEmailRuleSet is the base email rule set. Since rule sets are immutable.
var baseEmailRuleSet EmailRuleSet = EmailRuleSet{
	label: "EmailRuleSet",
}

// Email returns the base email RuleSet.
func Email() *EmailRuleSet {
	return &baseEmailRuleSet
}

// emailCloneOption is a functional option for cloning EmailRuleSet.
type emailCloneOption func(*EmailRuleSet)

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (ruleSet *EmailRuleSet) clone(options ...emailCloneOption) *EmailRuleSet {
	newRuleSet := &EmailRuleSet{
		required:      ruleSet.required,
		withNil:       ruleSet.withNil,
		domainRuleSet: ruleSet.domainRuleSet,
		parent:        ruleSet,
		errorConfig:   ruleSet.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func emailWithLabel(label string) emailCloneOption {
	return func(rs *EmailRuleSet) { rs.label = label }
}

func emailWithErrorConfig(config *errors.ErrorConfig) emailCloneOption {
	return func(rs *EmailRuleSet) { rs.errorConfig = config }
}

func emailWithConflictType(ct emailConflictType) emailCloneOption {
	return func(rs *EmailRuleSet) {
		// Check for conflicts and update parent if needed
		if rs.parent != nil {
			rs.parent = rs.parent.noConflict(ct)
		}
		rs.conflictType = ct
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *EmailRuleSet) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (ruleSet *EmailRuleSet) WithRequired() *EmailRuleSet {
	newRuleSet := ruleSet.clone(emailWithLabel("WithRequired()"), emailWithConflictType(emailConflictTypeRequired))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (ruleSet *EmailRuleSet) WithNil() *EmailRuleSet {
	newRuleSet := ruleSet.clone(emailWithLabel("WithNil()"), emailWithConflictType(emailConflictTypeNil))
	newRuleSet.withNil = true
	return newRuleSet
}

// Apply performs a validation of a RuleSet against a value and assigns the result to the output parameter.
// It returns a ValidationErrorCollection if any validation errors occur.
func (ruleSet *EmailRuleSet) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	// Add error config to context for error customization
	ctx = errors.WithErrorConfig(ctx, ruleSet.errorConfig)

	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, ruleSet.withNil, input, output); handled {
		return err
	}

	// Attempt to cast the input to a string
	valueStr, ok := input.(string)
	if !ok {
		return errors.Collection(errors.Error(errors.CodeType, ctx, "string", reflect.ValueOf(input).Kind().String()))
	}

	// Perform the validation
	if err := ruleSet.Evaluate(ctx, valueStr); err != nil {
		return err
	}

	outputVal := reflect.ValueOf(output)

	// Check if the output is a non-nil pointer
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "internal error", "output must be a non-nil pointer",
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
			errors.CodeInternal, ctx, "internal error", "cannot assign string to %T", output,
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
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "invalid format", "missing @ symbol"))
		return allErrors
	}
	if len(parts) > 2 {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "invalid format", "too many @ symbols"))
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
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "invalid format", "local part is empty"))
		return allErrors
	}

	if strings.HasPrefix(local, ".") {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "invalid format", "cannot start with a dot"))
	}

	if strings.HasSuffix(local, ".") {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "invalid format", "cannot end with a dot"))
	}

	if strings.Contains(local, "..") {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "invalid format", "cannot contain consecutive dots"))
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

// WithDomain returns a new child rule set that uses a custom domain validator
// instead of the default domain rule set for validating the domain portion of email addresses.
//
// The default domain rule set for email validation is the equivalent of:
//
//	NewDomain().WithTLD()
func (ruleSet *EmailRuleSet) WithDomain(domainRuleSet rules.RuleSet[string]) *EmailRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.domainRuleSet = domainRuleSet
	return newRuleSet
}

// noConflict returns the new rule set with all conflicting rules removed.
func (ruleSet *EmailRuleSet) noConflict(checker rules.Replaces[string]) *EmailRuleSet {
	// Check if current node conflicts (either via rule or conflictType)
	conflicts := false
	if ruleSet.rule != nil && checker.Replaces(ruleSet.rule) {
		conflicts = true
	} else if checker.Replaces(ruleSet) {
		conflicts = true
	}
	if conflicts {
		// Skip this node, continue up the parent chain
		if ruleSet.parent == nil {
			return nil
		}
		return ruleSet.parent.noConflict(checker)
	}

	// Current node doesn't conflict, process parent
	if ruleSet.parent == nil {
		return ruleSet
	}

	newParent := ruleSet.parent.noConflict(checker)

	// If parent didn't change, return current node unchanged
	if newParent == ruleSet.parent {
		return ruleSet
	}

	// Parent changed, clone current node with new parent
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = ruleSet.rule
	newRuleSet.parent = newParent
	newRuleSet.label = ruleSet.label
	newRuleSet.conflictType = ruleSet.conflictType
	return newRuleSet
}

// WithRule returns a new child rule set that applies a custom validation rule.
// The custom rule is evaluated during validation and any errors it returns are included in the validation result.
//
// Use this when implementing custom rules.
func (ruleSet *EmailRuleSet) WithRule(rule rules.Rule[string]) *EmailRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = rule
	if newParent := ruleSet.noConflict(rule); newParent != nil {
		newRuleSet.parent = newParent
	}
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
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

// WithErrorMessage returns a new RuleSet with custom short and long error messages.
func (ruleSet *EmailRuleSet) WithErrorMessage(short, long string) *EmailRuleSet {
	return ruleSet.clone(emailWithLabel("WithErrorMessage(...)"), emailWithErrorConfig(ruleSet.errorConfig.WithMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (ruleSet *EmailRuleSet) WithDocsURI(uri string) *EmailRuleSet {
	return ruleSet.clone(emailWithLabel("WithDocsURI(...)"), emailWithErrorConfig(ruleSet.errorConfig.WithDocs(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (ruleSet *EmailRuleSet) WithTraceURI(uri string) *EmailRuleSet {
	return ruleSet.clone(emailWithLabel("WithTraceURI(...)"), emailWithErrorConfig(ruleSet.errorConfig.WithTrace(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (ruleSet *EmailRuleSet) WithErrorCode(code errors.ErrorCode) *EmailRuleSet {
	return ruleSet.clone(emailWithLabel("WithErrorCode(...)"), emailWithErrorConfig(ruleSet.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (ruleSet *EmailRuleSet) WithErrorMeta(key string, value any) *EmailRuleSet {
	return ruleSet.clone(emailWithLabel("WithErrorMeta(...)"), emailWithErrorConfig(ruleSet.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (ruleSet *EmailRuleSet) WithErrorCallback(fn errors.ErrorCallback) *EmailRuleSet {
	return ruleSet.clone(emailWithLabel("WithErrorCallback(...)"), emailWithErrorConfig(ruleSet.errorConfig.WithCallback(fn)))
}
