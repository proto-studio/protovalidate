package net

import (
	"context"
	"reflect"
	"regexp"
	"strings"

	"golang.org/x/net/idna"
	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// baseDomainRuleSet is the base domain rule set. Since rule sets are immutable.
var baseDomainRuleSet DomainRuleSet = DomainRuleSet{
	label: "DomainRuleSet",
}

// domainLabelPattern matches valid domains after they have been converted to punycode
var domainLabelPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]$`)

// conflictType identifies the type of method that was called on a ruleset.
// Used for fast conflict checking instead of slow string prefix matching.
type domainConflictType int

const (
	domainConflictTypeNone domainConflictType = iota
	domainConflictTypeRequired
	domainConflictTypeNil
)

// Conflict returns true if this conflict type conflicts with the other conflict type.
// Two conflict types conflict if they are the same (non-zero) value.
func (ct domainConflictType) Conflict(other domainConflictType) bool {
	return ct != domainConflictTypeNone && ct == other
}

// Replaces returns true if this conflict type replaces the given rule.
// It attempts to cast the rule to *DomainRuleSet and checks if the conflictType conflicts.
func (ct domainConflictType) Replaces(r rules.Rule[string]) bool {
	rs, ok := r.(*DomainRuleSet)
	if !ok {
		return false
	}
	return ct.Conflict(rs.conflictType)
}

// DomainRuleSet implements the RuleSet interface for the domain names.
type DomainRuleSet struct {
	rules.NoConflict[string]
	required     bool
	withNil      bool
	parent       *DomainRuleSet
	rule         rules.Rule[string]
	label        string
	conflictType domainConflictType
	errorConfig  *errors.ErrorConfig
}

// Domain returns the base domain RuleSet.
func Domain() *DomainRuleSet {
	return &baseDomainRuleSet
}

// clone returns a shallow copy of the rule set with parent set to the current instance.
// domainCloneOption is a functional option for cloning DomainRuleSet.
type domainCloneOption func(*DomainRuleSet)

func (ruleSet *DomainRuleSet) clone(options ...domainCloneOption) *DomainRuleSet {
	newRuleSet := &DomainRuleSet{
		required:    ruleSet.required,
		withNil:     ruleSet.withNil,
		parent:      ruleSet,
		errorConfig: ruleSet.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func domainWithLabel(label string) domainCloneOption {
	return func(rs *DomainRuleSet) { rs.label = label }
}

func domainWithErrorConfig(config *errors.ErrorConfig) domainCloneOption {
	return func(rs *DomainRuleSet) { rs.errorConfig = config }
}

func domainWithConflictType(ct domainConflictType) domainCloneOption {
	return func(rs *DomainRuleSet) {
		// Check for conflicts and update parent if needed
		if rs.parent != nil {
			rs.parent = rs.parent.noConflict(ct)
		}
		rs.conflictType = ct
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *DomainRuleSet) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (ruleSet *DomainRuleSet) WithRequired() *DomainRuleSet {
	newRuleSet := ruleSet.clone(domainWithLabel("WithRequired()"), domainWithConflictType(domainConflictTypeRequired))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (ruleSet *DomainRuleSet) WithNil() *DomainRuleSet {
	newRuleSet := ruleSet.clone(domainWithLabel("WithNil()"), domainWithConflictType(domainConflictTypeNil))
	newRuleSet.withNil = true
	return newRuleSet
}

// Apply performs a validation of a RuleSet against a value and assigns the result to the output parameter.
// It returns a ValidationErrorCollection if any validation errors occur.
func (ruleSet *DomainRuleSet) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
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

// validateBasicDomain performs general domain validation that is valid for any and all domains.
// This function always returns a collection even if it is empty.
func validateBasicDomain(ctx context.Context, value string) errors.ValidationErrorCollection {
	allErrors := errors.Collection()

	// Convert to punycode
	punycode, err := idna.ToASCII(value)

	if err != nil {
		allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "invalid format", "domain contains invalid characters"))
		return allErrors
	}

	// Check total length
	if len(punycode) >= 256 {
		allErrors = append(allErrors, errors.Errorf(errors.CodeMaxLen, ctx, "too long", "domain exceeds maximum length"))
		return allErrors
	}

	// Each labels should contain only valid characters
	parts := strings.Split(punycode, ".")

	for _, part := range parts {
		if !domainLabelPattern.MatchString(part) {
			allErrors = append(allErrors, errors.Errorf(errors.CodePattern, ctx, "invalid format", "domain segment is invalid"))
			break
		}
	}

	return allErrors
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
func (ruleSet *DomainRuleSet) noConflict(checker rules.Replaces[string]) *DomainRuleSet {
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
func (ruleSet *DomainRuleSet) WithRule(rule rules.Rule[string]) *DomainRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = rule
	newRuleSet.parent = ruleSet.noConflict(rule)
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
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

// WithErrorMessage returns a new RuleSet with custom short and long error messages.
func (ruleSet *DomainRuleSet) WithErrorMessage(short, long string) *DomainRuleSet {
	return ruleSet.clone(domainWithLabel(util.FormatErrorMessageLabel(short, long)), domainWithErrorConfig(ruleSet.errorConfig.WithMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (ruleSet *DomainRuleSet) WithDocsURI(uri string) *DomainRuleSet {
	return ruleSet.clone(domainWithLabel(util.FormatStringArgLabel("WithDocsURI", uri)), domainWithErrorConfig(ruleSet.errorConfig.WithDocs(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (ruleSet *DomainRuleSet) WithTraceURI(uri string) *DomainRuleSet {
	return ruleSet.clone(domainWithLabel(util.FormatStringArgLabel("WithTraceURI", uri)), domainWithErrorConfig(ruleSet.errorConfig.WithTrace(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (ruleSet *DomainRuleSet) WithErrorCode(code errors.ErrorCode) *DomainRuleSet {
	return ruleSet.clone(domainWithLabel(util.FormatErrorCodeLabel(code)), domainWithErrorConfig(ruleSet.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (ruleSet *DomainRuleSet) WithErrorMeta(key string, value any) *DomainRuleSet {
	return ruleSet.clone(domainWithLabel(util.FormatErrorMetaLabel(key, value)), domainWithErrorConfig(ruleSet.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (ruleSet *DomainRuleSet) WithErrorCallback(fn errors.ErrorCallback) *DomainRuleSet {
	return ruleSet.clone(domainWithLabel(util.FormatErrorCallbackLabel()), domainWithErrorConfig(ruleSet.errorConfig.WithCallback(fn)))
}
