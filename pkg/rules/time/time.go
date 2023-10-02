package time

import (
	"context"
	"reflect"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// TimeRuleSet implements the RuleSet interface for the time.Time struct.
type TimeRuleSet struct {
	required bool
	layouts  []string
	parent   *TimeRuleSet
	rule     rules.Rule[time.Time]
}

// NewTime creates a new time.Time RuleSet
func NewTime() *TimeRuleSet {
	return &TimeRuleSet{}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *TimeRuleSet) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (ruleSet *TimeRuleSet) WithRequired() *TimeRuleSet {
	return &TimeRuleSet{
		required: true,
		parent:   ruleSet,
	}
}

// WithLayouts returns the a new rule set with the specified string layouts allowed for string coercion.
// The validation function will attempt each format in the order they are provided so it is recommended
// to list more specific layouts first.
//
// If this method is not called then coercion from strings will not be allowed and providing a string
// will return an error.
func (ruleSet *TimeRuleSet) WithLayouts(first string, rest ...string) *TimeRuleSet {
	layouts := make([]string, 0, 1+len(rest))
	layouts = append(layouts, first)
	layouts = append(layouts, rest...)

	return &TimeRuleSet{
		required: ruleSet.required,
		layouts:  layouts,
		parent:   ruleSet,
	}
}

// Validate performs a validation of a RuleSet against a value and returns a time.Time value or
// a ValidationErrorCollection.
func (ruleSet *TimeRuleSet) Validate(value any) (time.Time, errors.ValidationErrorCollection) {
	return ruleSet.ValidateWithContext(value, context.Background())
}

// Validate performs a validation of a RuleSet against a value and returns a time.Time value or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by rules and error formatting.
func (ruleSet *TimeRuleSet) ValidateWithContext(value any, ctx context.Context) (time.Time, errors.ValidationErrorCollection) {
	var t time.Time
	allErrors := errors.Collection()

	switch x := value.(type) {
	case time.Time:
		t = x
	case *time.Time:
		t = *x
	case string:
		ok := false

		for currentRuleSet := ruleSet; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
			if ruleSet.layouts != nil {
				for _, layout := range ruleSet.layouts {
					var err error
					t, err = time.Parse(layout, x)
					if err == nil {
						ok = true
						break
					}
				}
			}
		}

		if !ok {
			allErrors.Add(errors.NewCoercionError(ctx, "date time", "string"))
			return t, allErrors
		}
	default:
		return t, errors.Collection(errors.NewCoercionError(ctx, "date time", reflect.ValueOf(value).Kind().String()))
	}

	currentRuleSet := ruleSet
	ctx = rulecontext.WithRuleSet(ctx, ruleSet)

	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			newTime, errs := currentRuleSet.rule.Evaluate(ctx, t)
			if errs != nil {
				allErrors.Add(errs.All()...)
			} else {
				t = newTime
			}
		}

		currentRuleSet = currentRuleSet.parent
	}

	if allErrors.Size() > 0 {
		return t, allErrors
	} else {
		return t, nil
	}
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the time.Time type.
//
// Use this when implementing custom rules.
func (ruleSet *TimeRuleSet) WithRule(rule rules.Rule[time.Time]) *TimeRuleSet {
	return &TimeRuleSet{
		rule:     rule,
		parent:   ruleSet,
		required: ruleSet.required,
	}
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule interface
// for the time.Time type.
//
// Use this when implementing custom rules.
func (v *TimeRuleSet) WithRuleFunc(rule rules.RuleFunc[time.Time]) *TimeRuleSet {
	return v.WithRule(rule)
}

func (ruleSet *TimeRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[time.Time](ruleSet)
}
