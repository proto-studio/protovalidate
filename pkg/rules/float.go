package rules

import (
	"context"
	"math"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
)

var baseFloat32 FloatRuleSet[float32] = FloatRuleSet[float32]{
	outputPrecision: -1, // -1 means not set
	label:           "FloatRuleSet[float32]",
}

var baseFloat64 FloatRuleSet[float64] = FloatRuleSet[float64]{
	outputPrecision: -1, // -1 means not set
	label:           "FloatRuleSet[float64]",
}

type floating interface {
	float64 | float32
}

// conflictType identifies the type of method that was called on a ruleset.
// Used for fast conflict checking instead of slow string prefix matching.
type floatConflictType int

const (
	floatConflictTypeNone floatConflictType = iota
	floatConflictTypeRequired
	floatConflictTypeNil
	floatConflictTypeStrict
	floatConflictTypeRounding
	floatConflictTypeFixedOutput
)

// Conflict returns true if this conflict type conflicts with the other conflict type.
// Two conflict types conflict if they are the same (non-zero) value.
func (ct floatConflictType) Conflict(other floatConflictType) bool {
	return ct != floatConflictTypeNone && ct == other
}

// Implementation of RuleSet for floats.
type FloatRuleSet[T floating] struct {
	NoConflict[T]
	strict          bool
	rule            Rule[T]
	required        bool
	withNil         bool
	parent          *FloatRuleSet[T]
	rounding        Rounding
	precision       int // Precision for rounding (used with WithRounding)
	outputPrecision int // Precision for string output (-1 means not set, >= 0 means fixed output)
	label           string
	conflictType    floatConflictType
	errorConfig     *errors.ErrorConfig
}

// Float32 creates a new float32 RuleSet.
func Float32() *FloatRuleSet[float32] {
	return &baseFloat32
}

// Float64 creates a new float64 RuleSet.
func Float64() *FloatRuleSet[float64] {
	return &baseFloat64
}

// floatCloneOption is a functional option for cloning FloatRuleSet.
type floatCloneOption[T floating] func(*FloatRuleSet[T])

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (v *FloatRuleSet[T]) clone(options ...floatCloneOption[T]) *FloatRuleSet[T] {
	newRuleSet := &FloatRuleSet[T]{
		strict:          v.strict,
		required:        v.required,
		withNil:         v.withNil,
		rounding:        v.rounding,
		precision:       v.precision,
		outputPrecision: v.outputPrecision,
		parent:      v,
		errorConfig: v.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func floatWithLabel[T floating](label string) floatCloneOption[T] {
	return func(rs *FloatRuleSet[T]) { rs.label = label }
}

func floatWithErrorConfig[T floating](config *errors.ErrorConfig) floatCloneOption[T] {
	return func(rs *FloatRuleSet[T]) { rs.errorConfig = config }
}

func floatWithConflictType[T floating](ct floatConflictType) floatCloneOption[T] {
	return func(rs *FloatRuleSet[T]) {
		// Check for conflicts and update parent if needed
		if rs.parent != nil {
			rs.parent = rs.parent.noConflict(floatConflictTypeReplacesWrapper[T]{ct: ct})
		}
		rs.conflictType = ct
	}
}

// getConflictType returns the conflict type of the rule set.
// This is used by the conflict type wrapper to check for conflicts.
func (v *FloatRuleSet[T]) getConflictType() floatConflictType {
	return v.conflictType
}

// floatConflictTypeReplacesWrapper wraps a conflict type to implement Replaces[T]
type floatConflictTypeReplacesWrapper[T floating] struct {
	ct floatConflictType
}

func (w floatConflictTypeReplacesWrapper[T]) Replaces(r Rule[T]) bool {
	// Try to cast to FloatRuleSet to access conflictType
	if rs, ok := r.(interface{ getConflictType() floatConflictType }); ok {
		return w.ct.Conflict(rs.getConflictType())
	}
	return false
}

// WithStrict returns a new child RuleSet that disables type coercion.
// When strict mode is enabled, validation only succeeds if the value is already the correct type.
//
// With number types, any type will work in strict mode as long as it can be converted
// deterministically and without loss.
func (v *FloatRuleSet[T]) WithStrict() *FloatRuleSet[T] {
	newRuleSet := v.clone(floatWithLabel[T]("WithStrict()"), floatWithConflictType[T](floatConflictTypeStrict))
	newRuleSet.strict = true
	return newRuleSet
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *FloatRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (v *FloatRuleSet[T]) WithRequired() *FloatRuleSet[T] {
	newRuleSet := v.clone(floatWithLabel[T]("WithRequired()"), floatWithConflictType[T](floatConflictTypeRequired))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *FloatRuleSet[T]) WithNil() *FloatRuleSet[T] {
	newRuleSet := v.clone(floatWithLabel[T]("WithNil()"), floatWithConflictType[T](floatConflictTypeNil))
	newRuleSet.withNil = true
	return newRuleSet
}

// Apply performs validation of a RuleSet against a value and assigns the result to the output parameter.
// Apply returns a ValidationErrorCollection if any validation errors occur.
func (v *FloatRuleSet[T]) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	// Add error config to context for error customization
	ctx = errors.WithErrorConfig(ctx, v.errorConfig)

	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, v.withNil, input, output); handled {
		return err
	}

	// Ensure output is a non-nil pointer
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "internal error", "Output must be a non-nil pointer",
		))
	}

	// Attempt to coerce the input value to the correct float type
	floatval, validationErr := v.coerceFloat(input, ctx)
	if validationErr != nil {
		return errors.Collection(validationErr)
	}

	// Apply rounding if specified
	if v.rounding != RoundingNone {
		mul := math.Pow10(v.precision)
		tempFloatval := float64(floatval) * mul

		switch v.rounding {
		case RoundingDown:
			tempFloatval = math.Floor(tempFloatval)
		case RoundingUp:
			tempFloatval = math.Ceil(tempFloatval)
		case RoundingHalfUp:
			tempFloatval = math.Round(tempFloatval)
		case RoundingHalfEven:
			tempFloatval = math.RoundToEven(tempFloatval)
		}

		tempFloatval /= mul
		floatval = T(tempFloatval)
	}

	// Handle setting the value in output
	outputElem := outputVal.Elem()

	var assignable bool

	// Format the float as a string with the appropriate precision
	strVal := formatFloat(v, floatval)

	// Check if output is a string type
	if outputElem.Kind() == reflect.String {
		outputElem.SetString(strVal)
		assignable = true
	} else if outputElem.Kind() == reflect.Ptr && outputElem.Type().Elem().Kind() == reflect.String {
		// Handle pointer to string
		if outputElem.IsNil() {
			newStrPtr := reflect.New(outputElem.Type().Elem())
			newStrPtr.Elem().SetString(strVal)
			outputElem.Set(newStrPtr)
		} else {
			outputElem.Elem().SetString(strVal)
		}
		assignable = true
	} else if outputElem.Kind() == reflect.Bool {
		// Handle bool output: non-zero values are true, zero is false
		outputElem.SetBool(floatval != 0)
		assignable = true
	} else if (outputElem.Kind() == reflect.Interface && outputElem.IsNil()) ||
		(outputElem.Kind() == reflect.Float32 || outputElem.Kind() == reflect.Float64 ||
			outputElem.Type().AssignableTo(reflect.TypeOf(floatval))) {

		// If output is a nil interface, or an assignable type, set it directly to the new float value
		outputElem.Set(reflect.ValueOf(floatval))
		assignable = true
	}

	// If the types are incompatible, return an error
	if !assignable {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "internal error", "Cannot assign %T to %T", floatval, outputElem.Interface(),
		))
	}

	allErrors := errors.Collection()

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			if err := currentRuleSet.rule.Evaluate(ctx, floatval); err != nil {
				allErrors = append(allErrors, err...)
			}
		}
	}

	if len(allErrors) != 0 {
		return allErrors
	}
	return nil
}

// Evaluate performs validation of a RuleSet against a float value and returns a ValidationErrorCollection.
func (v *FloatRuleSet[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	var out T
	return v.Apply(ctx, value, &out)
}

// noConflict returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *FloatRuleSet[T]) noConflict(checker Replaces[T]) *FloatRuleSet[T] {
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
func (ruleSet *FloatRuleSet[T]) WithRule(rule Rule[T]) *FloatRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = rule
	newRuleSet.parent = ruleSet.noConflict(rule)
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
func (v *FloatRuleSet[T]) WithRuleFunc(rule RuleFunc[T]) *FloatRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the number RuleSet in an Any rule set
// which can then be used in nested validation.
func (v *FloatRuleSet[T]) Any() RuleSet[any] {
	return WrapAny(v)
}

// typeName returns the name for the target integer type.
// Used for error message formatting.
func (v *FloatRuleSet[T]) typeName() string {
	return reflect.ValueOf(*new(T)).Kind().String()
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *FloatRuleSet[T]) String() string {
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
func (v *FloatRuleSet[T]) WithErrorMessage(short, long string) *FloatRuleSet[T] {
	return v.clone(floatWithLabel[T](util.FormatErrorMessageLabel(short, long)), floatWithErrorConfig[T](v.errorConfig.WithErrorMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (v *FloatRuleSet[T]) WithDocsURI(uri string) *FloatRuleSet[T] {
	return v.clone(floatWithLabel[T](util.FormatStringArgLabel("WithDocsURI", uri)), floatWithErrorConfig[T](v.errorConfig.WithDocsURI(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (v *FloatRuleSet[T]) WithTraceURI(uri string) *FloatRuleSet[T] {
	return v.clone(floatWithLabel[T](util.FormatStringArgLabel("WithTraceURI", uri)), floatWithErrorConfig[T](v.errorConfig.WithTraceURI(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (v *FloatRuleSet[T]) WithErrorCode(code errors.ErrorCode) *FloatRuleSet[T] {
	return v.clone(floatWithLabel[T](util.FormatErrorCodeLabel(code)), floatWithErrorConfig[T](v.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (v *FloatRuleSet[T]) WithErrorMeta(key string, value any) *FloatRuleSet[T] {
	return v.clone(floatWithLabel[T](util.FormatErrorMetaLabel(key, value)), floatWithErrorConfig[T](v.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (v *FloatRuleSet[T]) WithErrorCallback(fn errors.ErrorCallback) *FloatRuleSet[T] {
	return v.clone(floatWithLabel[T](util.FormatErrorCallbackLabel()), floatWithErrorConfig[T](v.errorConfig.WithCallback(fn)))
}
