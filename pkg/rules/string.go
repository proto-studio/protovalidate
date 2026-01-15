package rules

import (
	"context"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// conflictType identifies the type of method that was called on a ruleset.
// Used for fast conflict checking instead of slow string prefix matching.
type stringConflictType int

const (
	stringConflictTypeNone stringConflictType = iota
	stringConflictTypeRequired
	stringConflictTypeNil
	stringConflictTypeStrict
)

// Conflict returns true if this conflict type conflicts with the other conflict type.
// Two conflict types conflict if they are the same (non-zero) value.
func (ct stringConflictType) Conflict(other stringConflictType) bool {
	return ct != stringConflictTypeNone && ct == other
}

// Replaces returns true if this conflict type replaces the given rule.
// It attempts to cast the rule to *StringRuleSet and checks if the conflictType conflicts.
func (ct stringConflictType) Replaces(r Rule[string]) bool {
	rs, ok := r.(*StringRuleSet)
	if !ok {
		return false
	}
	return ct.Conflict(rs.conflictType)
}

// Implementation of RuleSet for strings.
type StringRuleSet struct {
	NoConflict[string]
	strict       bool
	rule         Rule[string]
	required     bool
	withNil      bool
	parent       *StringRuleSet
	label        string
	conflictType stringConflictType
	errorConfig  *errors.ErrorConfig
}

// baseStringRuleSet is the main RuleSet.
// New returns this since rule sets are immutable and StringRuleSet does not contain generics.
var baseStringRuleSet StringRuleSet = StringRuleSet{
	label: "StringRuleSet",
}

// String returns the base StringRuleSet.
func String() *StringRuleSet {
	return &baseStringRuleSet
}

// stringCloneOption is a functional option for cloning StringRuleSet.
type stringCloneOption func(*StringRuleSet)

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (v *StringRuleSet) clone(options ...stringCloneOption) *StringRuleSet {
	newRuleSet := &StringRuleSet{
		strict:       v.strict,
		required:     v.required,
		withNil:      v.withNil,
		parent:      v,
		errorConfig: v.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func stringWithLabel(label string) stringCloneOption {
	return func(rs *StringRuleSet) { rs.label = label }
}

func stringWithErrorConfig(config *errors.ErrorConfig) stringCloneOption {
	return func(rs *StringRuleSet) { rs.errorConfig = config }
}

func stringWithConflictType(ct stringConflictType) stringCloneOption {
	return func(rs *StringRuleSet) {
		// Check for conflicts and update parent if needed
		if rs.parent != nil {
			rs.parent = rs.parent.noConflict(ct)
		}
		rs.conflictType = ct
	}
}

// WithStrict returns a new child RuleSet that disables type coercion.
// When strict mode is enabled, validation only succeeds if the value is already a string.
func (v *StringRuleSet) WithStrict() *StringRuleSet {
	newRuleSet := v.clone(stringWithLabel("WithStrict()"), stringWithConflictType(stringConflictTypeStrict))
	newRuleSet.strict = true
	return newRuleSet
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *StringRuleSet) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (v *StringRuleSet) WithRequired() *StringRuleSet {
	newRuleSet := v.clone(stringWithLabel("WithRequired()"), stringWithConflictType(stringConflictTypeRequired))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *StringRuleSet) WithNil() *StringRuleSet {
	newRuleSet := v.clone(stringWithLabel("WithNil()"), stringWithConflictType(stringConflictTypeNil))
	newRuleSet.withNil = true
	return newRuleSet
}

// Apply performs validation of a RuleSet against a value and assigns the resulting string to the output pointer.
// Apply returns a ValidationErrorCollection.
func (v *StringRuleSet) Apply(ctx context.Context, value, output any) errors.ValidationErrorCollection {
	// Add error config to context for error customization
	ctx = errors.WithErrorConfig(ctx, v.errorConfig)

	// Check if withNil is enabled and value is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, v.withNil, value, output); handled {
		return err
	}

	// Ensure output is a pointer that can be set
	rv := reflect.ValueOf(output)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.Collection(
			errors.Errorf(errors.CodeInternal, ctx, "internal error", "Output must be a non-nil pointer"),
		)
	}

	// Attempt to coerce the input to a string
	str, validationErr := v.coerce(value, ctx)

	if validationErr != nil {
		return errors.Collection(validationErr)
	}

	verrs := v.Evaluate(ctx, str)
	if verrs != nil {
		return verrs
	}

	// Set the string result in the output parameter
	elem := rv.Elem()

	// Check if the output is an interface
	if elem.Kind() == reflect.Interface {
		// Create a new string value and set the interface to point to it
		elem.Set(reflect.ValueOf(str))
		return nil
	}

	// If the element is a string, replace it with the new string value
	if elem.Kind() == reflect.String {
		elem.SetString(str)
		return nil
	}

	return errors.Collection(
		errors.Errorf(errors.CodeInternal, ctx, "internal error", "Cannot assign string to %T", output),
	)
}

// Evaluate performs validation of a RuleSet against a string value and returns a ValidationErrorCollection.
func (v *StringRuleSet) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	allErrors := errors.Collection()

	currentRuleSet := v
	ctx = rulecontext.WithRuleSet(ctx, v)

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
func (ruleSet *StringRuleSet) noConflict(checker Replaces[string]) *StringRuleSet {
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
func (ruleSet *StringRuleSet) WithRule(rule Rule[string]) *StringRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = rule
	newRuleSet.parent = ruleSet.noConflict(rule)
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
func (v *StringRuleSet) WithRuleFunc(rule RuleFunc[string]) *StringRuleSet {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the string RuleSet in an Any rule set
// which can then be used in nested validation.
func (v *StringRuleSet) Any() RuleSet[any] {
	return WrapAny(v)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *StringRuleSet) String() string {
	label := ruleSet.label

	if label == "" && ruleSet.rule != nil {
		label = ruleSet.rule.String()
	}

	if ruleSet.parent != nil {
		return ruleSet.parent.String() + "." + label
	}
	return label
}

// WithErrorMessage returns a new RuleSet with custom short and long error messages.
func (v *StringRuleSet) WithErrorMessage(short, long string) *StringRuleSet {
	return v.clone(stringWithLabel("WithErrorMessage(...)"), stringWithErrorConfig(v.errorConfig.WithMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (v *StringRuleSet) WithDocsURI(uri string) *StringRuleSet {
	return v.clone(stringWithLabel("WithDocsURI(...)"), stringWithErrorConfig(v.errorConfig.WithDocs(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (v *StringRuleSet) WithTraceURI(uri string) *StringRuleSet {
	return v.clone(stringWithLabel("WithTraceURI(...)"), stringWithErrorConfig(v.errorConfig.WithTrace(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (v *StringRuleSet) WithErrorCode(code errors.ErrorCode) *StringRuleSet {
	return v.clone(stringWithLabel("WithErrorCode(...)"), stringWithErrorConfig(v.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (v *StringRuleSet) WithErrorMeta(key string, value any) *StringRuleSet {
	return v.clone(stringWithLabel("WithErrorMeta(...)"), stringWithErrorConfig(v.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (v *StringRuleSet) WithErrorCallback(fn errors.ErrorCallback) *StringRuleSet {
	return v.clone(stringWithLabel("WithErrorCallback(...)"), stringWithErrorConfig(v.errorConfig.WithCallback(fn)))
}
