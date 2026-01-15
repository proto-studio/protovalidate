package rules

import (
	"context"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// AnyRuleSet implements RuleSet for the "any" interface.
// AnyRuleSet is used when you don't care about the data type passed in and want to return it unaltered from the Validate method.
//
// See also: WrapAny which also implements the "any" interface and wraps another RuleSet.
type AnyRuleSet struct {
	NoConflict[any]
	required    bool
	forbidden   bool
	withNil     bool
	rule        Rule[any]
	parent      *AnyRuleSet
	label       string
	errorConfig *errors.ErrorConfig
}

// backgroundAnyRUleSet is the main AnyRuleSet.
// Any returns this since rule sets are immutable and AnyRuleSet does not contain generics.
var backgroundAnyRuleSet AnyRuleSet = AnyRuleSet{
	label: "AnyRuleSet",
}

// Any creates a new Any rule set.
func Any() *AnyRuleSet {
	return &backgroundAnyRuleSet
}

// anyCloneOption is a functional option for cloning AnyRuleSet.
type anyCloneOption func(*AnyRuleSet)

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (v *AnyRuleSet) clone(options ...anyCloneOption) *AnyRuleSet {
	newRuleSet := &AnyRuleSet{
		required:    v.required,
		forbidden:   v.forbidden,
		withNil:     v.withNil,
		parent:      v,
		errorConfig: v.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func anyWithLabel(label string) anyCloneOption {
	return func(rs *AnyRuleSet) { rs.label = label }
}

func anyWithErrorConfig(config *errors.ErrorConfig) anyCloneOption {
	return func(rs *AnyRuleSet) { rs.errorConfig = config }
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *AnyRuleSet) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (v *AnyRuleSet) WithRequired() *AnyRuleSet {
	newRuleSet := v.clone(anyWithLabel("WithRequired()"))
	newRuleSet.required = true
	return newRuleSet
}

// WithForbidden returns a new child rule set that requires values to be nil or omitted.
// When a value is present, validation fails with an error.
func (v *AnyRuleSet) WithForbidden() *AnyRuleSet {
	newRuleSet := v.clone(anyWithLabel("WithForbidden()"))
	newRuleSet.forbidden = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *AnyRuleSet) WithNil() *AnyRuleSet {
	newRuleSet := v.clone(anyWithLabel("WithNil()"))
	newRuleSet.withNil = true
	return newRuleSet
}

// Apply performs validation of a RuleSet against a value and assigns the value to the output.
// Apply returns a ValidationErrorCollection.
func (v *AnyRuleSet) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	// Add error config to context for error customization
	ctx = errors.WithErrorConfig(ctx, v.errorConfig)

	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, v.withNil, input, output); handled {
		return err
	}

	err := v.Evaluate(ctx, input)
	if err != nil {
		return err
	}

	// Ensure output is a pointer
	rv := reflect.ValueOf(output)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.Collection(
			errors.Errorf(errors.CodeInternal, ctx, "internal error", "Output must be a non-nil pointer"),
		)
	}

	// Get the element the pointer points to
	elem := rv.Elem()

	// Convert input to reflect.Value
	inputValue := reflect.ValueOf(input)

	// Check if the input can be assigned to the output
	if inputValue.Type().AssignableTo(elem.Type()) {
		elem.Set(inputValue)
		return nil
	}

	return errors.Collection(
		errors.Errorf(errors.CodeInternal, ctx, "internal error", "Cannot assign %T to %T", input, output),
	)
}

// Evaluate performs validation of a RuleSet against a value and returns a ValidationErrorCollection.
// Evaluate calls wrapped rules before any rules added directly to the AnyRuleSet.
func (v *AnyRuleSet) Evaluate(ctx context.Context, value any) errors.ValidationErrorCollection {
	if v.forbidden {
		return errors.Collection(errors.Error(errors.CodeForbidden, ctx))
	}

	allErrors := errors.Collection()

	currentRuleSet := v
	ctx = rulecontext.WithRuleSet(ctx, v)

	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			err := currentRuleSet.rule.Evaluate(ctx, value)
			if err != nil {
				allErrors = append(allErrors, err...)
			}
		}

		currentRuleSet = currentRuleSet.parent
	}

	if len(allErrors) != 0 {
		return allErrors
	} else {
		return nil
	}
}

// WithRule returns a new child rule set that applies a custom validation rule.
// The custom rule is evaluated during validation and any errors it returns are included in the validation result.
func (v *AnyRuleSet) WithRule(rule Rule[any]) *AnyRuleSet {
	newRuleSet := v.clone()
	newRuleSet.rule = rule
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
func (v *AnyRuleSet) WithRuleFunc(rule RuleFunc[any]) *AnyRuleSet {
	return v.WithRule(rule)
}

// Any returns the current rule set.
func (v *AnyRuleSet) Any() RuleSet[any] {
	return v
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *AnyRuleSet) String() string {
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
func (v *AnyRuleSet) WithErrorMessage(short, long string) *AnyRuleSet {
	return v.clone(anyWithLabel(util.FormatErrorMessageLabel(short, long)), anyWithErrorConfig(v.errorConfig.WithMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (v *AnyRuleSet) WithDocsURI(uri string) *AnyRuleSet {
	return v.clone(anyWithLabel(util.FormatStringArgLabel("WithDocsURI", uri)), anyWithErrorConfig(v.errorConfig.WithDocs(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (v *AnyRuleSet) WithTraceURI(uri string) *AnyRuleSet {
	return v.clone(anyWithLabel(util.FormatStringArgLabel("WithTraceURI", uri)), anyWithErrorConfig(v.errorConfig.WithTrace(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (v *AnyRuleSet) WithErrorCode(code errors.ErrorCode) *AnyRuleSet {
	return v.clone(anyWithLabel(util.FormatErrorCodeLabel(code)), anyWithErrorConfig(v.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (v *AnyRuleSet) WithErrorMeta(key string, value any) *AnyRuleSet {
	return v.clone(anyWithLabel(util.FormatErrorMetaLabel(key, value)), anyWithErrorConfig(v.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (v *AnyRuleSet) WithErrorCallback(fn errors.ErrorCallback) *AnyRuleSet {
	return v.clone(anyWithLabel(util.FormatErrorCallbackLabel()), anyWithErrorConfig(v.errorConfig.WithCallback(fn)))
}
