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
}

// baseTimeRuleSet is the base time rule set. Since rule sets are immutable.
var baseTimeRuleSet TimeRuleSet = TimeRuleSet{
	label: "TimeRuleSet",
}

// Time returns the base time.Time RuleSet.
func Time() *TimeRuleSet {
	return &baseTimeRuleSet
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *TimeRuleSet) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (ruleSet *TimeRuleSet) WithRequired() *TimeRuleSet {
	return &TimeRuleSet{
		required:     true,
		withNil:      ruleSet.withNil,
		parent:       ruleSet,
		outputLayout: ruleSet.outputLayout,
		label:        "WithRequired()",
	}
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (ruleSet *TimeRuleSet) WithNil() *TimeRuleSet {
	return &TimeRuleSet{
		required:     ruleSet.required,
		withNil:      true,
		parent:       ruleSet,
		outputLayout: ruleSet.outputLayout,
		label:        "WithNil()",
	}
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

	return &TimeRuleSet{
		required:     ruleSet.required,
		withNil:      ruleSet.withNil,
		layouts:      layouts,
		parent:       ruleSet,
		outputLayout: ruleSet.outputLayout,
		label:        util.StringsToRuleOutput("WithLayouts", layouts),
	}
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

	return &TimeRuleSet{
		required:     ruleSet.required,
		withNil:      ruleSet.withNil,
		parent:       ruleSet,
		outputLayout: layout,
		label:        util.StringsToRuleOutput("WithOutputLayout", []string{layout}),
	}
}

// Apply performs validation of a RuleSet against a value and assigns the result to the output parameter.
// Apply returns a ValidationErrorCollection if any validation errors occur.
func (ruleSet *TimeRuleSet) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, ruleSet.withNil, input, output); handled {
		return err
	}

	// Ensure output is a non-nil pointer
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Output must be a non-nil pointer",
		))
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
			return errors.Collection(errors.NewCoercionError(ctx, "date time", "string"))
		}
	default:
		return errors.Collection(errors.NewCoercionError(ctx, "date time", reflect.TypeOf(input).String()))
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
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Cannot assign %T to %T", t, outputElem.Interface(),
		))
	}

	// Evaluate the time value and return any validation errors
	return ruleSet.Evaluate(ctx, t)
}

// Evaluate performs validation of a RuleSet against a time.Time value and returns a ValidationErrorCollection.
func (ruleSet *TimeRuleSet) Evaluate(ctx context.Context, value time.Time) errors.ValidationErrorCollection {
	allErrors := errors.Collection()

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
func (ruleSet *TimeRuleSet) noConflict(rule rules.Rule[time.Time]) *TimeRuleSet {
	if ruleSet.rule != nil {

		// Conflicting rules, skip this and return the parent
		if rule.Conflict(ruleSet.rule) {
			return ruleSet.parent.noConflict(rule)
		}

	}

	if ruleSet.parent == nil {
		return ruleSet
	}

	newParent := ruleSet.parent.noConflict(rule)

	if newParent == ruleSet.parent {
		return ruleSet
	}

	return &TimeRuleSet{
		rule:         ruleSet.rule,
		layouts:      ruleSet.layouts,
		outputLayout: ruleSet.outputLayout,
		parent:       newParent,
		required:     ruleSet.required,
		withNil:      ruleSet.withNil,
		label:        ruleSet.label,
	}
}

// WithRule returns a new child rule set that applies a custom validation rule.
// The custom rule is evaluated during validation and any errors it returns are included in the validation result.
func (ruleSet *TimeRuleSet) WithRule(rule rules.Rule[time.Time]) *TimeRuleSet {
	return &TimeRuleSet{
		rule:     rule,
		parent:   ruleSet.noConflict(rule),
		required: ruleSet.required,
		withNil:  ruleSet.withNil,
	}
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
