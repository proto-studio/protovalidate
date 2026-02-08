package time

import (
	"context"
	"reflect"
	"time"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// conflictType identifies the type of method that was called on a ruleset.
// Used for fast conflict checking instead of slow string prefix matching.
type conflictType int

const (
	conflictTypeNone conflictType = iota
	conflictTypeRequired
	conflictTypeNil
	conflictTypeLayouts
	conflictTypeOutputLayout
)

// Conflict returns true if this conflict type conflicts with the other conflict type.
// Two conflict types conflict if they are the same (non-zero) value.
func (ct conflictType) Conflict(other conflictType) bool {
	return ct != conflictTypeNone && ct == other
}

// Replaces returns true if this conflict type replaces the given rule.
// It attempts to cast the rule to *TimeRuleSet and checks if the conflictType conflicts.
func (ct conflictType) Replaces(r rules.Rule[time.Time]) bool {
	rs, ok := r.(*TimeRuleSet)
	if !ok {
		return false
	}
	return ct.Conflict(rs.conflictType)
}

// TimeRuleSet implements the RuleSet interface for the time.Time struct.
type TimeRuleSet struct {
	rules.NoConflict[time.Time]
	required     bool
	withNil      bool
	layouts      []string
	outputLayout string
	parent       *TimeRuleSet
	rule         rules.Rule[time.Time]
	label        string
	conflictType conflictType
	errorConfig  *errors.ErrorConfig
}

// baseTimeRuleSet is the base time rule set. Since rule sets are immutable.
var baseTimeRuleSet TimeRuleSet = TimeRuleSet{
	label: "TimeRuleSet",
}

// Time returns the base time.Time RuleSet.
func Time() *TimeRuleSet {
	return &baseTimeRuleSet
}

// timeCloneOption is a functional option for cloning TimeRuleSet.
type timeCloneOption func(*TimeRuleSet)

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (ruleSet *TimeRuleSet) clone(options ...timeCloneOption) *TimeRuleSet {
	newRuleSet := &TimeRuleSet{
		required:     ruleSet.required,
		withNil:      ruleSet.withNil,
		layouts:      ruleSet.layouts,
		outputLayout: ruleSet.outputLayout,
		parent:       ruleSet,
		errorConfig:  ruleSet.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func timeWithLabel(label string) timeCloneOption {
	return func(rs *TimeRuleSet) { rs.label = label }
}

func timeWithErrorConfig(config *errors.ErrorConfig) timeCloneOption {
	return func(rs *TimeRuleSet) { rs.errorConfig = config }
}

func timeWithConflictType(ct conflictType) timeCloneOption {
	return func(rs *TimeRuleSet) {
		// Check for conflicts and update parent if needed
		if rs.parent != nil {
			rs.parent = rs.parent.noConflict(ct)
		}
		rs.conflictType = ct
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *TimeRuleSet) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (ruleSet *TimeRuleSet) WithRequired() *TimeRuleSet {
	newRuleSet := ruleSet.clone(timeWithLabel("WithRequired()"), timeWithConflictType(conflictTypeRequired))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (ruleSet *TimeRuleSet) WithNil() *TimeRuleSet {
	newRuleSet := ruleSet.clone(timeWithLabel("WithNil()"), timeWithConflictType(conflictTypeNil))
	newRuleSet.withNil = true
	return newRuleSet
}

// WithLayouts returns a new rule set that allows string-to-time conversion using the specified layouts.
// The rule set attempts each format in the order they are provided and stops when a match
// is found, so it is recommended to list more specific layouts first.
//
// Layouts are cumulative, calling this method multiple times will result in all provided layouts across
// all calls being allowed.
//
// If this method is not called then coercion from strings will not be allowed and providing a string
// will return an error.
//
// By default if both the input and output of Apply are strings, the output value will be formatted to be
// the same format as the input and non-string inputs will always be formatted as time.RFC3339. To change
// this behavior, use WithOutputLayout.
func (ruleSet *TimeRuleSet) WithLayouts(first string, rest ...string) *TimeRuleSet {
	layouts := make([]string, 0, 1+len(rest))
	layouts = append(layouts, first)
	layouts = append(layouts, rest...)

	newRuleSet := ruleSet.clone(
		timeWithLabel(util.StringsToRuleOutput("WithLayouts", layouts)),
		timeWithConflictType(conflictTypeLayouts),
	)
	newRuleSet.layouts = layouts
	return newRuleSet
}

// WithOutputLayout returns a new rule set that formats time values as strings using the specified layout.
// When the output value of Apply is a string pointer, the time will be formatted using this layout
// regardless of the type or format of the input.
//
// This method has no effect on input layouts. Use WithLayouts to set which layouts are allowed on input.
// The default output format is time.RFC3339 unless the input is also a string.
func (ruleSet *TimeRuleSet) WithOutputLayout(layout string) *TimeRuleSet {
	if ruleSet.outputLayout == layout {
		return ruleSet
	}

	newRuleSet := ruleSet.clone(
		timeWithLabel(util.StringsToRuleOutput("WithOutputLayout", []string{layout})),
		timeWithConflictType(conflictTypeOutputLayout),
	)
	newRuleSet.outputLayout = layout
	return newRuleSet
}

// Apply performs validation of a RuleSet against a value and assigns the result to the output parameter.
// Apply returns a ValidationError if any validation errors occur.
func (ruleSet *TimeRuleSet) Apply(ctx context.Context, input any, output any) errors.ValidationError {
	// Add error config to context for error customization
	ctx = errors.WithErrorConfig(ctx, ruleSet.errorConfig)

	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, ruleSet.withNil, input, output); handled {
		return err
	}

	// Ensure output is a non-nil pointer
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Errorf(errors.CodeInternal, ctx, "internal error", "Output must be a non-nil pointer")
	}

	var t time.Time
	ok := false

	// Set the default layout
	layout := time.RFC3339

	// Handle different types of input
	switch x := input.(type) {
	case time.Time:
		t = x
		ok = true
	case *time.Time:
		if x != nil {
			t = *x
			ok = true
		}
	case string:
		for currentRuleSet := ruleSet; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
			if currentRuleSet.layouts != nil {
				for _, l := range currentRuleSet.layouts {
					var err error
					t, err = time.Parse(l, x)
					if err == nil {
						layout = l // Overwrite layout with the one used for parsing
						ok = true
						break
					}
				}
				if ok {
					break
				}
			}
		}
		if !ok {
			return errors.Error(errors.CodeType, ctx, "date time", "string")
		}
	default:
		return errors.Error(errors.CodeType, ctx, "date time", reflect.TypeOf(input).String())
	}

	// Overwrite layout if outputLayout is set
	if ruleSet.outputLayout != "" {
		layout = ruleSet.outputLayout
	}

	// Handle setting the value in output
	outputElem := outputVal.Elem()

	// If output is assignable from time.Time, set it directly to the new time value
	if outputElem.Kind() == reflect.Interface && outputElem.IsNil() {
		outputElem.Set(reflect.ValueOf(t))
	} else if outputElem.Type().AssignableTo(reflect.TypeOf(t)) {
		outputElem.Set(reflect.ValueOf(t))
	} else if outputElem.Type().AssignableTo(reflect.TypeOf("")) { // Check if output is assignable from string
		// Use the determined layout to format time as a string
		formattedTime := t.Format(layout)
		outputElem.Set(reflect.ValueOf(formattedTime))
	} else {
		return errors.Errorf(errors.CodeInternal, ctx, "internal error", "Cannot assign %T to %T", t, outputElem.Interface())
	}

	// Evaluate the time value and return any validation errors
	return ruleSet.Evaluate(ctx, t)
}

// Evaluate performs validation of a RuleSet against a time.Time value and returns a ValidationError.
func (ruleSet *TimeRuleSet) Evaluate(ctx context.Context, value time.Time) errors.ValidationError {
	var errs errors.ValidationError
	currentRuleSet := ruleSet
	ctx = rulecontext.WithRuleSet(ctx, ruleSet)
	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			if e := currentRuleSet.rule.Evaluate(ctx, value); e != nil {
				errs = errors.Join(errs, e)
			}
		}
		currentRuleSet = currentRuleSet.parent
	}
	return errs
}

// noConflict returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *TimeRuleSet) noConflict(checker rules.Replaces[time.Time]) *TimeRuleSet {
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
func (ruleSet *TimeRuleSet) WithRule(rule rules.Rule[time.Time]) *TimeRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = rule
	newRuleSet.parent = ruleSet.noConflict(rule)
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
//
// Use this when implementing custom rules.
func (v *TimeRuleSet) WithRuleFunc(rule rules.RuleFunc[time.Time]) *TimeRuleSet {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the time RuleSet in an Any rule set
// which can then be used in nested validation.
func (ruleSet *TimeRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[time.Time](ruleSet)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *TimeRuleSet) String() string {
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
func (ruleSet *TimeRuleSet) WithErrorMessage(short, long string) *TimeRuleSet {
	return ruleSet.clone(timeWithLabel(util.FormatErrorMessageLabel(short, long)), timeWithErrorConfig(ruleSet.errorConfig.WithErrorMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (ruleSet *TimeRuleSet) WithDocsURI(uri string) *TimeRuleSet {
	return ruleSet.clone(timeWithLabel(util.FormatStringArgLabel("WithDocsURI", uri)), timeWithErrorConfig(ruleSet.errorConfig.WithDocsURI(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (ruleSet *TimeRuleSet) WithTraceURI(uri string) *TimeRuleSet {
	return ruleSet.clone(timeWithLabel(util.FormatStringArgLabel("WithTraceURI", uri)), timeWithErrorConfig(ruleSet.errorConfig.WithTraceURI(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (ruleSet *TimeRuleSet) WithErrorCode(code errors.ErrorCode) *TimeRuleSet {
	return ruleSet.clone(timeWithLabel(util.FormatErrorCodeLabel(code)), timeWithErrorConfig(ruleSet.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (ruleSet *TimeRuleSet) WithErrorMeta(key string, value any) *TimeRuleSet {
	return ruleSet.clone(timeWithLabel(util.FormatErrorMetaLabel(key, value)), timeWithErrorConfig(ruleSet.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (ruleSet *TimeRuleSet) WithErrorCallback(fn errors.ErrorCallback) *TimeRuleSet {
	return ruleSet.clone(timeWithLabel(util.FormatErrorCallbackLabel()), timeWithErrorConfig(ruleSet.errorConfig.WithCallback(fn)))
}
