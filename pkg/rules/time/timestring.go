package time

import (
	"context"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// TimeStringRuleSet is identical to TimeStringRuleSet except that validation returns a string instead of a
// time.Time struct. This is useful when dealing with APIs that treat date times as opaque strings and
// don't care about the actual date.Time object.
type TimeStringRuleSet struct {
	rules.NoConflict[string]
	inner  *TimeRuleSet
	layout string
}

// backgroundTimeRuleSet is the base time rule set. Since rule sets are immutable.
var backgroundTimeStringRuleSet TimeRuleSet = TimeRuleSet{
	label: "TimeStringRuleSet",
}

// NewTime creates a new time.Time RuleSet.
// Layout contains the target string date time format.
//
// You may allow for different input layouts than the target using WithLayouts() but the output string
// will always be in the target layout.
func NewTimeString(layout string) *TimeStringRuleSet {
	return &TimeStringRuleSet{
		layout: layout,
		inner:  backgroundTimeStringRuleSet.WithLayouts(layout),
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *TimeStringRuleSet) Required() bool {
	return ruleSet.inner.required
}

// WithRequired returns a new rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (ruleSet *TimeStringRuleSet) WithRequired() *TimeStringRuleSet {
	return &TimeStringRuleSet{
		inner:  ruleSet.inner.WithRequired(),
		layout: ruleSet.layout,
	}
}

// WithLayouts returns the a new rule set with the specified string layouts allowed for string coercion.
// The validation function will attempt each format in the order they are provided so it is recommended
// to list more specific layouts first.
//
// Unlike TimeRuleSet, TimeStringRuleSet has a default layout identical to the specified output format.
//
// If you call WithLayouts, take care to also list the output format. If you sue WithLayout without the output
// format then it will not be allowed as input, which is probably not what you want.
func (ruleSet *TimeStringRuleSet) WithLayouts(first string, rest ...string) *TimeStringRuleSet {
	return &TimeStringRuleSet{
		inner:  ruleSet.inner.WithLayouts(first, rest...),
		layout: ruleSet.layout,
	}
}

// Validate performs a validation of a RuleSet against a value and returns a string value or
// a ValidationErrorCollection.
//
// Deprecated: Validate is deprecated and will be removed in v1.0.0. Use Run instead.
func (ruleSet *TimeStringRuleSet) Validate(value any) (string, errors.ValidationErrorCollection) {
	return ruleSet.Run(context.Background(), value)
}

// Validate performs a validation of a RuleSet against a value and returns a string value or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by rules and error formatting.
//
// Deprecated: ValidateWithContext is deprecated and will be removed in v1.0.0. Use Run instead.
func (ruleSet *TimeStringRuleSet) ValidateWithContext(value any, ctx context.Context) (string, errors.ValidationErrorCollection) {
	return ruleSet.Run(ctx, value)
}

// Run performs a validation of a RuleSet against a value and returns a string value or
// a ValidationErrorCollection.
func (ruleSet *TimeStringRuleSet) Run(ctx context.Context, value any) (string, errors.ValidationErrorCollection) {
	t, err := ruleSet.inner.ValidateWithContext(value, ctx)
	if err != nil {
		return "", err
	}

	return t.Format(ruleSet.layout), nil
}

// Evaluate performs a validation of a RuleSet against a string value and returns a string value of the
// same type or a ValidationErrorCollection.
func (ruleSet *TimeStringRuleSet) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	// We have to cast from string to time no matter what in this case so we call ValidateWithContext
	// and ignore the potentially mutated value.
	_, errs := ruleSet.ValidateWithContext(value, ctx)
	return errs
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the time.Time type.
//
// This is different than other implementations since the WithRule method takes a different
// data type (time.Time) than the output type (string). This is because string validation on
// dates and times is rarely meaningful.
//
// Use this when implementing custom rules.
func (ruleSet *TimeStringRuleSet) WithRule(rule rules.Rule[time.Time]) *TimeStringRuleSet {
	return &TimeStringRuleSet{
		inner:  ruleSet.inner.WithRule(rule),
		layout: ruleSet.layout,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule interface
// for the time.Time type.
//
// This is different than other implementations since the WithRuleFunc method takes a different
// data type (time.Time) than the output type (string). This is because string validation on
// dates and times is rarely meaningful.
//
// Use this when implementing custom rules.
func (v *TimeStringRuleSet) WithRuleFunc(rule rules.RuleFunc[time.Time]) *TimeStringRuleSet {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the domain RuleSet in any Any rule set
// which can then be used in nested validation.
func (ruleSet *TimeStringRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[string](ruleSet)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *TimeStringRuleSet) String() string {
	return ruleSet.inner.String()
}
