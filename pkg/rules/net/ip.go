package net

import (
	"context"
	"net"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// baseIPRuleSet is the base IP rule set. Since rule sets are immutable.
var baseIPRuleSet IPRuleSet = IPRuleSet{
	label: "IPRuleSet",
}

// conflictType identifies the type of method that was called on a ruleset.
// Used for fast conflict checking instead of slow string prefix matching.
type ipConflictType int

const (
	ipConflictTypeNone ipConflictType = iota
	ipConflictTypeRequired
	ipConflictTypeNil
)

// Conflict returns true if this conflict type conflicts with the other conflict type.
// Two conflict types conflict if they are the same (non-zero) value.
func (ct ipConflictType) Conflict(other ipConflictType) bool {
	return ct != ipConflictTypeNone && ct == other
}

// Replaces returns true if this conflict type replaces the given rule.
// It attempts to cast the rule to *IPRuleSet and checks if the conflictType conflicts.
func (ct ipConflictType) Replaces(r rules.Rule[net.IP]) bool {
	rs, ok := r.(*IPRuleSet)
	if !ok {
		return false
	}
	return ct.Conflict(rs.conflictType)
}

// IPRuleSet implements the RuleSet interface for IP addresses.
type IPRuleSet struct {
	rules.NoConflict[net.IP]
	required     bool
	withNil      bool
	parent       *IPRuleSet
	rule         rules.Rule[net.IP]
	label        string
	conflictType ipConflictType
	errorConfig  *errors.ErrorConfig
}

// IP returns the base IP RuleSet.
func IP() *IPRuleSet {
	return &baseIPRuleSet
}

// ipCloneOption is a functional option for cloning IPRuleSet.
type ipCloneOption func(*IPRuleSet)

func (ruleSet *IPRuleSet) clone(options ...ipCloneOption) *IPRuleSet {
	newRuleSet := &IPRuleSet{
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

func ipWithLabel(label string) ipCloneOption {
	return func(rs *IPRuleSet) { rs.label = label }
}

func ipWithErrorConfig(config *errors.ErrorConfig) ipCloneOption {
	return func(rs *IPRuleSet) { rs.errorConfig = config }
}

func ipWithConflictType(ct ipConflictType) ipCloneOption {
	return func(rs *IPRuleSet) {
		// Check for conflicts and update parent if needed
		if rs.parent != nil {
			rs.parent = rs.parent.noConflict(ct)
		}
		rs.conflictType = ct
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *IPRuleSet) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (ruleSet *IPRuleSet) WithRequired() *IPRuleSet {
	newRuleSet := ruleSet.clone(ipWithLabel("WithRequired()"), ipWithConflictType(ipConflictTypeRequired))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (ruleSet *IPRuleSet) WithNil() *IPRuleSet {
	newRuleSet := ruleSet.clone(ipWithLabel("WithNil()"), ipWithConflictType(ipConflictTypeNil))
	newRuleSet.withNil = true
	return newRuleSet
}

// parseIP attempts to parse the input as either a string or net.IP and returns a net.IP.
func parseIP(ctx context.Context, input any) (net.IP, errors.ValidationError) {
	if ip, ok := input.(net.IP); ok {
		if ip == nil {
			return nil, errors.Error(errors.CodeNull, ctx)
		}
		return ip, nil
	}
	if str, ok := input.(string); ok {
		ip := net.ParseIP(str)
		if ip == nil {
			return nil, errors.Errorf(errors.CodePattern, ctx, "invalid format", "invalid IP address format")
		}
		return ip, nil
	}
	if strPtr, ok := input.(*string); ok && strPtr != nil {
		ip := net.ParseIP(*strPtr)
		if ip == nil {
			return nil, errors.Errorf(errors.CodePattern, ctx, "invalid format", "invalid IP address format")
		}
		return ip, nil
	}
	return nil, errors.Error(errors.CodeType, ctx, "string or net.IP", reflect.ValueOf(input).Kind().String())
}

// setOutput sets the output value to the given IP address.
func setOutput(ctx context.Context, output any, ip net.IP) errors.ValidationError {
	outputVal := reflect.ValueOf(output)

	// Check if the output is a non-nil pointer
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Errorf(errors.CodeInternal, ctx, "internal error", "output must be a non-nil pointer")
	}

	// Dereference the pointer to get the actual value that needs to be set
	outputElem := outputVal.Elem()
	outputType := outputElem.Type()

	// Check if it's net.IP type (net.IP is []byte, so we check for slice of uint8)
	if outputType == reflect.TypeOf(net.IP{}) {
		outputElem.Set(reflect.ValueOf(ip))
		return nil
	}

	switch outputElem.Kind() {
	case reflect.String:
		outputElem.SetString(ip.String())
	case reflect.Interface:
		// Set as net.IP for interface types
		outputElem.Set(reflect.ValueOf(ip))
	default:
		return errors.Errorf(errors.CodeInternal, ctx, "internal error", "cannot assign IP to %T", output)
	}

	return nil
}

// Apply performs a validation of a RuleSet against a value and assigns the result to the output parameter.
// It returns a ValidationError if any validation errors occur.
// Input can be either a string or net.IP, and output can be either *string or *net.IP.
func (ruleSet *IPRuleSet) Apply(ctx context.Context, input any, output any) errors.ValidationError {
	// Add error config to context for error customization
	ctx = errors.WithErrorConfig(ctx, ruleSet.errorConfig)

	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, ruleSet.withNil, input, output); handled {
		return err
	}

	// Parse the input to net.IP
	ip, err := parseIP(ctx, input)
	if err != nil {
		return err
	}

	// Perform the validation
	if err := ruleSet.Evaluate(ctx, ip); err != nil {
		return err
	}

	// Set the output
	return setOutput(ctx, output, ip)
}

// validateBasicIP performs general IP validation that is valid for any and all IP addresses.
// This function always returns a collection even if it is empty.
func validateBasicIP(ctx context.Context, ip net.IP) errors.ValidationError {
	if ip == nil {
		return errors.Error(errors.CodeNull, ctx)
	}
	return nil
}

// Evaluate performs a validation of a RuleSet against a net.IP and returns a ValidationError.
func (ruleSet *IPRuleSet) Evaluate(ctx context.Context, ip net.IP) errors.ValidationError {
	var errs errors.ValidationError
	if err := validateBasicIP(ctx, ip); err != nil {
		errs = errors.Join(errs, err)
	}
	currentRuleSet := ruleSet
	ctx = rulecontext.WithRuleSet(ctx, ruleSet)
	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			if e := currentRuleSet.rule.Evaluate(ctx, ip); e != nil {
				errs = errors.Join(errs, e)
			}
		}
		currentRuleSet = currentRuleSet.parent
	}
	return errs
}

// noConflict returns the new rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *IPRuleSet) noConflict(checker rules.Replaces[net.IP]) *IPRuleSet {
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
func (ruleSet *IPRuleSet) WithRule(rule rules.Rule[net.IP]) *IPRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = rule
	newRuleSet.parent = ruleSet.noConflict(rule)
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
//
// Use this when implementing custom rules.
func (v *IPRuleSet) WithRuleFunc(rule rules.RuleFunc[net.IP]) *IPRuleSet {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the IP RuleSet in any Any rule set
// which can then be used in nested validation.
func (ruleSet *IPRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[net.IP](ruleSet)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *IPRuleSet) String() string {
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
func (ruleSet *IPRuleSet) WithErrorMessage(short, long string) *IPRuleSet {
	return ruleSet.clone(ipWithLabel(util.FormatErrorMessageLabel(short, long)), ipWithErrorConfig(ruleSet.errorConfig.WithErrorMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (ruleSet *IPRuleSet) WithDocsURI(uri string) *IPRuleSet {
	return ruleSet.clone(ipWithLabel(util.FormatStringArgLabel("WithDocsURI", uri)), ipWithErrorConfig(ruleSet.errorConfig.WithDocsURI(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (ruleSet *IPRuleSet) WithTraceURI(uri string) *IPRuleSet {
	return ruleSet.clone(ipWithLabel(util.FormatStringArgLabel("WithTraceURI", uri)), ipWithErrorConfig(ruleSet.errorConfig.WithTraceURI(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (ruleSet *IPRuleSet) WithErrorCode(code errors.ErrorCode) *IPRuleSet {
	return ruleSet.clone(ipWithLabel(util.FormatErrorCodeLabel(code)), ipWithErrorConfig(ruleSet.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (ruleSet *IPRuleSet) WithErrorMeta(key string, value any) *IPRuleSet {
	return ruleSet.clone(ipWithLabel(util.FormatErrorMetaLabel(key, value)), ipWithErrorConfig(ruleSet.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (ruleSet *IPRuleSet) WithErrorCallback(fn errors.ErrorCallback) *IPRuleSet {
	return ruleSet.clone(ipWithLabel(util.FormatErrorCallbackLabel()), ipWithErrorConfig(ruleSet.errorConfig.WithCallback(fn)))
}
