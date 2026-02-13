package rules

import (
	"context"
	"reflect"
	"strconv"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
)

var baseBool BoolRuleSet = BoolRuleSet{
	label: "BoolRuleSet",
}

// conflictType identifies the type of method that was called on a ruleset.
// Used for fast conflict checking instead of slow string prefix matching.
type boolConflictType int

const (
	boolConflictTypeNone boolConflictType = iota
	boolConflictTypeRequired
	boolConflictTypeNil
	boolConflictTypeStrict
)

// Conflict returns true if this conflict type conflicts with the other conflict type.
// Two conflict types conflict if they are the same (non-zero) value.
func (ct boolConflictType) Conflict(other boolConflictType) bool {
	return ct != boolConflictTypeNone && ct == other
}

// Implementation of RuleSet for booleans.
type BoolRuleSet struct {
	NoConflict[bool]
	strict       bool
	rule         Rule[bool]
	required     bool
	withNil      bool
	parent       *BoolRuleSet
	label        string
	conflictType boolConflictType
	errorConfig  *errors.ErrorConfig
}

// Bool creates a new boolean RuleSet.
func Bool() *BoolRuleSet {
	return &baseBool
}

// boolCloneOption is a functional option for cloning BoolRuleSet.
type boolCloneOption func(*BoolRuleSet)

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (v *BoolRuleSet) clone(options ...boolCloneOption) *BoolRuleSet {
	newRuleSet := &BoolRuleSet{
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

func boolWithLabel(label string) boolCloneOption {
	return func(rs *BoolRuleSet) { rs.label = label }
}

func boolWithErrorConfig(config *errors.ErrorConfig) boolCloneOption {
	return func(rs *BoolRuleSet) { rs.errorConfig = config }
}

func boolWithConflictType(ct boolConflictType) boolCloneOption {
	return func(rs *BoolRuleSet) {
		// Check for conflicts and update parent if needed
		if rs.parent != nil {
			rs.parent = rs.parent.noConflict(boolConflictTypeReplacesWrapper{ct: ct})
		}
		rs.conflictType = ct
	}
}

// getConflictType returns the conflict type of the rule set.
// This is used by the conflict type wrapper to check for conflicts.
func (v *BoolRuleSet) getConflictType() boolConflictType {
	return v.conflictType
}

// boolConflictTypeReplacesWrapper wraps a conflict type to implement Replaces[bool]
type boolConflictTypeReplacesWrapper struct {
	ct boolConflictType
}

// Replaces returns true if the other rule is a BoolRuleSet with a conflicting conflict type.
func (w boolConflictTypeReplacesWrapper) Replaces(r Rule[bool]) bool {
	// Try to cast to BoolRuleSet to access conflictType
	if rs, ok := r.(interface{ getConflictType() boolConflictType }); ok {
		return w.ct.Conflict(rs.getConflictType())
	}
	return false
}

// WithStrict returns a new child RuleSet that disables type coercion.
// When strict mode is enabled, validation only succeeds if the value is already a boolean.
func (v *BoolRuleSet) WithStrict() *BoolRuleSet {
	newRuleSet := v.clone(boolWithLabel("WithStrict()"), boolWithConflictType(boolConflictTypeStrict))
	newRuleSet.strict = true
	return newRuleSet
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *BoolRuleSet) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (v *BoolRuleSet) WithRequired() *BoolRuleSet {
	newRuleSet := v.clone(boolWithLabel("WithRequired()"), boolWithConflictType(boolConflictTypeRequired))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *BoolRuleSet) WithNil() *BoolRuleSet {
	newRuleSet := v.clone(boolWithLabel("WithNil()"), boolWithConflictType(boolConflictTypeNil))
	newRuleSet.withNil = true
	return newRuleSet
}

// Apply coerces input to bool, evaluates all rules, and returns the result.
func (ruleSet *BoolRuleSet) Apply(ctx context.Context, input any) (bool, errors.ValidationError) {
	ctx = errors.WithErrorConfig(ctx, ruleSet.errorConfig)

	if handled, err := util.TryNilIfAllowed(ctx, ruleSet.withNil, input); handled {
		return false, err
	}

	boolval, validationErr := ruleSet.coerceBool(input, ctx)
	if validationErr != nil {
		return false, validationErr
	}

	if errs := ruleSet.Evaluate(ctx, boolval); errs != nil {
		return false, errs
	}
	return boolval, nil
}

// Evaluate performs validation of a RuleSet against a boolean value and returns a ValidationError.
func (v *BoolRuleSet) Evaluate(ctx context.Context, value bool) errors.ValidationError {
	var errs errors.ValidationError
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			if err := currentRuleSet.rule.Evaluate(ctx, value); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	return errs
}

// noConflict returns the new rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *BoolRuleSet) noConflict(checker Replaces[bool]) *BoolRuleSet {
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
func (ruleSet *BoolRuleSet) WithRule(rule Rule[bool]) *BoolRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = rule
	newRuleSet.parent = ruleSet.noConflict(rule)
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
func (v *BoolRuleSet) WithRuleFunc(rule RuleFunc[bool]) *BoolRuleSet {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the boolean RuleSet in an Any rule set
// which can then be used in nested validation.
func (v *BoolRuleSet) Any() RuleSet[any] {
	return WrapAny(v)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *BoolRuleSet) String() string {
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
func (v *BoolRuleSet) WithErrorMessage(short, long string) *BoolRuleSet {
	return v.clone(boolWithLabel("WithErrorMessage(...)"), boolWithErrorConfig(v.errorConfig.WithErrorMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (v *BoolRuleSet) WithDocsURI(uri string) *BoolRuleSet {
	return v.clone(boolWithLabel("WithDocsURI(...)"), boolWithErrorConfig(v.errorConfig.WithDocsURI(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (v *BoolRuleSet) WithTraceURI(uri string) *BoolRuleSet {
	return v.clone(boolWithLabel("WithTraceURI(...)"), boolWithErrorConfig(v.errorConfig.WithTraceURI(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (v *BoolRuleSet) WithErrorCode(code errors.ErrorCode) *BoolRuleSet {
	return v.clone(boolWithLabel("WithErrorCode(...)"), boolWithErrorConfig(v.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (v *BoolRuleSet) WithErrorMeta(key string, value any) *BoolRuleSet {
	return v.clone(boolWithLabel("WithErrorMeta(...)"), boolWithErrorConfig(v.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (v *BoolRuleSet) WithErrorCallback(fn errors.ErrorCallback) *BoolRuleSet {
	return v.clone(boolWithLabel("WithErrorCallback(...)"), boolWithErrorConfig(v.errorConfig.WithCallback(fn)))
}

// coerceBool attempts to convert the value to a boolean and returns a validation error if it can't.
func (ruleSet *BoolRuleSet) coerceBool(value any, ctx context.Context) (bool, errors.ValidationError) {
	switch x := value.(type) {
	case bool:
		return x, nil
	case *bool:
		if x == nil {
			return false, errors.Error(errors.CodeNull, ctx, "bool", "nil pointer")
		}
		return *x, nil
	default:
		if ruleSet.strict {
			return false, errors.Error(errors.CodeType, ctx, "bool", reflect.ValueOf(value).Kind().String())
		}

		// Try to coerce from string
		if str, ok := value.(string); ok {
			boolval, err := strconv.ParseBool(str)
			if err != nil {
				return false, errors.Error(errors.CodeType, ctx, "bool", "string")
			}
			return boolval, nil
		}

		// Try to coerce from numeric types (0 = false, non-zero = true)
		switch x := value.(type) {
		case int:
			return x != 0, nil
		case int8:
			return x != 0, nil
		case int16:
			return x != 0, nil
		case int32:
			return x != 0, nil
		case int64:
			return x != 0, nil
		case uint:
			return x != 0, nil
		case uint8:
			return x != 0, nil
		case uint16:
			return x != 0, nil
		case uint32:
			return x != 0, nil
		case uint64:
			return x != 0, nil
		case float32:
			return x != 0, nil
		case float64:
			return x != 0, nil
		}

		return false, errors.Error(errors.CodeType, ctx, "bool", reflect.ValueOf(value).Kind().String())
	}
}
