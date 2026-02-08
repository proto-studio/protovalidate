package time

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// conflictTypeDuration identifies the type of method that was called on a duration ruleset.
// Used for fast conflict checking instead of slow string prefix matching.
type conflictTypeDuration int

const (
	conflictTypeDurationNone conflictTypeDuration = iota
	conflictTypeDurationRequired
	conflictTypeDurationNil
	conflictTypeDurationUnit
	conflictTypeDurationRounding
)

// Conflict returns true if this conflict type conflicts with the other conflict type.
// Two conflict types conflict if they are the same (non-zero) value.
func (ct conflictTypeDuration) Conflict(other conflictTypeDuration) bool {
	return ct != conflictTypeDurationNone && ct == other
}

// Replaces returns true if this conflict type replaces the given rule.
// It attempts to cast the rule to *DurationRuleSet and checks if the conflictType conflicts.
func (ct conflictTypeDuration) Replaces(r rules.Rule[time.Duration]) bool {
	rs, ok := r.(*DurationRuleSet)
	if !ok {
		return false
	}
	return ct.Conflict(rs.conflictType)
}

// DurationRuleSet implements the RuleSet interface for the time.Duration type.
type DurationRuleSet struct {
	rules.NoConflict[time.Duration]
	required     bool
	withNil      bool
	unit         time.Duration  // Unit multiplier for numeric inputs (defaults to nanoseconds)
	rounding     rules.Rounding // Rounding method for unit conversion
	parent       *DurationRuleSet
	rule         rules.Rule[time.Duration]
	label        string
	conflictType conflictTypeDuration
	errorConfig  *errors.ErrorConfig
}

// baseDurationRuleSet is the base duration rule set. Since rule sets are immutable.
var baseDurationRuleSet DurationRuleSet = DurationRuleSet{
	label: "DurationRuleSet",
	unit:  time.Nanosecond, // Default unit is nanoseconds
}

// Duration returns the base time.Duration RuleSet.
func Duration() *DurationRuleSet {
	return &baseDurationRuleSet
}

// durationCloneOption is a functional option for cloning DurationRuleSet.
type durationCloneOption func(*DurationRuleSet)

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (ruleSet *DurationRuleSet) clone(options ...durationCloneOption) *DurationRuleSet {
	newRuleSet := &DurationRuleSet{
		required:    ruleSet.required,
		withNil:     ruleSet.withNil,
		unit:        ruleSet.unit,
		rounding:    ruleSet.rounding,
		parent:      ruleSet,
		errorConfig: ruleSet.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func durationWithLabel(label string) durationCloneOption {
	return func(rs *DurationRuleSet) { rs.label = label }
}

func durationWithErrorConfig(config *errors.ErrorConfig) durationCloneOption {
	return func(rs *DurationRuleSet) { rs.errorConfig = config }
}

func durationWithConflictType(ct conflictTypeDuration) durationCloneOption {
	return func(rs *DurationRuleSet) {
		// Check for conflicts and update parent if needed
		if rs.parent != nil {
			rs.parent = rs.parent.noConflict(ct)
		}
		rs.conflictType = ct
	}
}

func durationWithUnit(unit time.Duration) durationCloneOption {
	return func(rs *DurationRuleSet) {
		rs.unit = unit
	}
}

func durationWithRounding(rounding rules.Rounding) durationCloneOption {
	return func(rs *DurationRuleSet) {
		rs.rounding = rounding
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *DurationRuleSet) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (ruleSet *DurationRuleSet) WithRequired() *DurationRuleSet {
	newRuleSet := ruleSet.clone(durationWithLabel("WithRequired()"), durationWithConflictType(conflictTypeDurationRequired))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (ruleSet *DurationRuleSet) WithNil() *DurationRuleSet {
	newRuleSet := ruleSet.clone(durationWithLabel("WithNil()"), durationWithConflictType(conflictTypeDurationNil))
	newRuleSet.withNil = true
	return newRuleSet
}

// WithUnit returns a new child rule set that specifies the unit for numeric duration inputs.
// When an int or int64 is provided as input, it will be multiplied by the specified unit to convert it to a time.Duration.
// For example, WithUnit(time.Second) means that the number 5 represents 5 seconds.
// Common units include: time.Nanosecond, time.Microsecond, time.Millisecond, time.Second, time.Minute, time.Hour, 24*time.Hour (days), 7*24*time.Hour (weeks).
// By default, numeric inputs are treated as nanoseconds (time.Duration's native unit).
// This method conflicts with previous calls to WithUnit - only the most recent unit setting is used.
func (ruleSet *DurationRuleSet) WithUnit(unit time.Duration) *DurationRuleSet {
	if unit <= 0 {
		// Invalid unit, return unchanged
		return ruleSet
	}
	newRuleSet := ruleSet.clone(
		durationWithLabel(util.StringsToRuleOutput("WithUnit", []string{unit.String()})),
		durationWithConflictType(conflictTypeDurationUnit),
		durationWithUnit(unit),
	)
	return newRuleSet
}

// WithRounding returns a new child rule set that applies the specified rounding method when converting durations to numeric output.
// When a duration is not evenly divisible by the unit, rounding determines how the remainder is handled.
// This method conflicts with previous calls to WithRounding - only the most recent rounding setting is used.
func (ruleSet *DurationRuleSet) WithRounding(rounding rules.Rounding) *DurationRuleSet {
	newRuleSet := ruleSet.clone(
		durationWithLabel(fmt.Sprintf("WithRounding(%s)", rounding.String())),
		durationWithConflictType(conflictTypeDurationRounding),
		durationWithRounding(rounding),
	)
	return newRuleSet
}

// Apply performs validation of a RuleSet against a value and assigns the result to the output parameter.
// Apply returns a ValidationError if any validation errors occur.
func (ruleSet *DurationRuleSet) Apply(ctx context.Context, input any, output any) errors.ValidationError {
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

	var d time.Duration
	ok := false

	// Get the unit from the current node (tail node)
	// The unit is copied to each clone, so it's always on the current node
	unit := ruleSet.unit

	// Handle different types of input
	switch x := input.(type) {
	case time.Duration:
		// Input is explicitly time.Duration - use it directly without unit conversion
		d = x
		ok = true
	case *time.Duration:
		// Input is *time.Duration - use it directly without unit conversion
		if x != nil {
			d = *x
			ok = true
		}
	case int64:
		// int64 (not time.Duration) - multiply by unit to convert from specified unit to nanoseconds
		d = time.Duration(x) * unit
		ok = true
	case int:
		// int - multiply by unit to convert from specified unit to nanoseconds
		d = time.Duration(x) * unit
		ok = true
	case string:
		var err error
		d, err = time.ParseDuration(x)
		if err == nil {
			ok = true
		} else {
			// String parsing failed - this is a format/pattern error, not a type error
			return errors.Errorf(errors.CodePattern, ctx, "invalid format", "invalid duration format: %v", err)
		}
	default:
		return errors.Error(errors.CodeType, ctx, "duration", reflect.TypeOf(input).String())
	}

	if !ok {
		return errors.Error(errors.CodeType, ctx, "duration", reflect.TypeOf(input).String())
	}

	// Handle setting the value in output
	outputElem := outputVal.Elem()

	var assignable bool

	// If output is an interface, get the underlying element to check its type
	// This allows `var output any = int64(0)` to be treated as int64
	actualOutputElem := outputElem
	if outputElem.Kind() == reflect.Interface && !outputElem.IsNil() {
		actualOutputElem = outputElem.Elem()
	}

	// Apply rounding to d if rounding is set
	remainder := d % unit
	if remainder != 0 {
		if ruleSet.rounding == rules.RoundingNone {
			// Only error if output is numeric (not duration)
			// Duration output can accept any value
			if actualOutputElem.Type() != reflect.TypeOf(d) {
				return errors.Errorf(errors.CodeRange, ctx, "duration", "Duration %s is not evenly divisible by unit %s", d, unit)
			}
		} else {
			// Apply rounding based on the remainder
			quotient := int64(d / unit)
			halfUnit := unit / 2
			switch ruleSet.rounding {
			case rules.RoundingDown:
				// Floor - use quotient as-is
			case rules.RoundingUp:
				if remainder > 0 {
					quotient++
				}
			case rules.RoundingHalfUp:
				if remainder >= halfUnit {
					quotient++
				}
			case rules.RoundingHalfEven:
				if remainder > halfUnit {
					quotient++
				} else if remainder == halfUnit && quotient%2 != 0 {
					quotient++
				}
			}
			d = time.Duration(quotient) * unit
		}
	}

	// Now set d to output based on output type
	// Check if output is time.Duration first (before numeric check because Duration has Kind() == Int64)
	if actualOutputElem.Type() == reflect.TypeOf(d) {
		if outputElem.Kind() == reflect.Interface {
			outputElem.Set(reflect.ValueOf(d))
		} else {
			actualOutputElem.Set(reflect.ValueOf(d))
		}
		assignable = true
	} else if outputKind := actualOutputElem.Kind(); outputKind >= reflect.Int && outputKind <= reflect.Uintptr {
		// Numeric output - convert duration to numeric by dividing by unit
		quotient := int64(d / unit)

		// Use reflection to set the value - check bounds first
		if outputKind >= reflect.Uint && outputKind <= reflect.Uintptr {
			// For unsigned types, check if value is non-negative and fits
			if quotient < 0 {
				return errors.NewRangeError(ctx, "duration")
			}
			// Check if the value fits in the target type by converting and checking if we lose information
			targetType := actualOutputElem.Type()
			testValue := reflect.New(targetType).Elem()
			testValue.SetUint(uint64(quotient))
			// Convert back to see if we lost information
			convertedBack := int64(testValue.Uint())
			if convertedBack != quotient {
				return errors.NewRangeError(ctx, "duration")
			}
			// Set the value - if output was an interface, set the new value into it
			if outputElem.Kind() == reflect.Interface {
				outputElem.Set(testValue)
			} else {
				outputElem.SetUint(uint64(quotient))
			}
		} else {
			// Signed integer types - check if the value fits by converting and checking
			targetType := actualOutputElem.Type()
			testValue := reflect.New(targetType).Elem()
			testValue.SetInt(quotient)
			// Convert back to see if we lost information (same pattern as number_coerce.go)
			convertedBack := int64(testValue.Int())
			if convertedBack != quotient {
				return errors.NewRangeError(ctx, "duration")
			}
			// Set the value - if output was an interface, set the new value into it
			if outputElem.Kind() == reflect.Interface {
				outputElem.Set(testValue)
			} else {
				outputElem.SetInt(quotient)
			}
		}
		assignable = true
	} else if outputElem.Kind() == reflect.Interface && outputElem.IsNil() {
		// If output is a nil interface, set it to the duration value
		outputElem.Set(reflect.ValueOf(d))
		assignable = true
	}

	// If the types are incompatible, return an error
	if !assignable {
		return errors.Errorf(errors.CodeInternal, ctx, "internal error", "Cannot assign %T to %T", d, outputElem.Interface())
	}

	// Evaluate the duration value and return any validation errors
	return ruleSet.Evaluate(ctx, d)
}

// Evaluate performs validation of a RuleSet against a time.Duration value and returns a ValidationError.
func (ruleSet *DurationRuleSet) Evaluate(ctx context.Context, value time.Duration) errors.ValidationError {
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

// noConflict returns the new duration rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *DurationRuleSet) noConflict(checker rules.Replaces[time.Duration]) *DurationRuleSet {
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
	// Note: clone() already preserves unit, required, withNil, and errorConfig
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = ruleSet.rule
	newRuleSet.parent = newParent
	newRuleSet.label = ruleSet.label
	newRuleSet.conflictType = ruleSet.conflictType
	return newRuleSet
}

// WithRule returns a new child rule set that applies a custom validation rule.
// The custom rule is evaluated during validation and any errors it returns are included in the validation result.
func (ruleSet *DurationRuleSet) WithRule(rule rules.Rule[time.Duration]) *DurationRuleSet {
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = rule
	newRuleSet.parent = ruleSet.noConflict(rule)
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
//
// Use this when implementing custom rules.
func (v *DurationRuleSet) WithRuleFunc(rule rules.RuleFunc[time.Duration]) *DurationRuleSet {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the duration RuleSet in an Any rule set
// which can then be used in nested validation.
func (ruleSet *DurationRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[time.Duration](ruleSet)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *DurationRuleSet) String() string {
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
func (ruleSet *DurationRuleSet) WithErrorMessage(short, long string) *DurationRuleSet {
	return ruleSet.clone(durationWithLabel(util.FormatErrorMessageLabel(short, long)), durationWithErrorConfig(ruleSet.errorConfig.WithErrorMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (ruleSet *DurationRuleSet) WithDocsURI(uri string) *DurationRuleSet {
	return ruleSet.clone(durationWithLabel(util.FormatStringArgLabel("WithDocsURI", uri)), durationWithErrorConfig(ruleSet.errorConfig.WithDocsURI(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (ruleSet *DurationRuleSet) WithTraceURI(uri string) *DurationRuleSet {
	return ruleSet.clone(durationWithLabel(util.FormatStringArgLabel("WithTraceURI", uri)), durationWithErrorConfig(ruleSet.errorConfig.WithTraceURI(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (ruleSet *DurationRuleSet) WithErrorCode(code errors.ErrorCode) *DurationRuleSet {
	return ruleSet.clone(durationWithLabel(util.FormatErrorCodeLabel(code)), durationWithErrorConfig(ruleSet.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (ruleSet *DurationRuleSet) WithErrorMeta(key string, value any) *DurationRuleSet {
	return ruleSet.clone(durationWithLabel(util.FormatErrorMetaLabel(key, value)), durationWithErrorConfig(ruleSet.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (ruleSet *DurationRuleSet) WithErrorCallback(fn errors.ErrorCallback) *DurationRuleSet {
	return ruleSet.clone(durationWithLabel(util.FormatErrorCallbackLabel()), durationWithErrorConfig(ruleSet.errorConfig.WithCallback(fn)))
}
