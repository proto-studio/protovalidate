package rules

import (
	"context"
	"fmt"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
)

var baseInt IntRuleSet[int] = IntRuleSet[int]{
	base:  10,
	label: "IntRuleSet[int]",
}

var baseUint IntRuleSet[uint] = IntRuleSet[uint]{
	base:  10,
	label: "IntRuleSet[uint]",
}

var baseInt8 IntRuleSet[int8] = IntRuleSet[int8]{
	base:  10,
	label: "IntRuleSet[int8]",
}

var baseUint8 IntRuleSet[uint8] = IntRuleSet[uint8]{
	base:  10,
	label: "IntRuleSet[uint8]",
}

var baseInt16 IntRuleSet[int16] = IntRuleSet[int16]{
	base:  10,
	label: "IntRuleSet[int16]",
}

var baseUint16 IntRuleSet[uint16] = IntRuleSet[uint16]{
	base:  10,
	label: "IntRuleSet[uint16]",
}

var baseInt32 IntRuleSet[int32] = IntRuleSet[int32]{
	base:  10,
	label: "IntRuleSet[int32]",
}

var baseUint32 IntRuleSet[uint32] = IntRuleSet[uint32]{
	base:  10,
	label: "IntRuleSet[uint32]",
}

var baseInt64 IntRuleSet[int64] = IntRuleSet[int64]{
	base:  10,
	label: "IntRuleSet[int64]",
}

var baseUint64 IntRuleSet[uint64] = IntRuleSet[uint64]{
	base:  10,
	label: "IntRuleSet[uint64]",
}

type integer interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~int | ~int8 | ~int16 | ~int32 | ~int64
}

// conflictType identifies the type of method that was called on a ruleset.
// Used for fast conflict checking instead of slow string prefix matching.
type intConflictType int

const (
	intConflictTypeNone intConflictType = iota
	intConflictTypeRequired
	intConflictTypeNil
	intConflictTypeStrict
	intConflictTypeBase
	intConflictTypeRounding
)

// Conflict returns true if this conflict type conflicts with the other conflict type.
// Two conflict types conflict if they are the same (non-zero) value.
func (ct intConflictType) Conflict(other intConflictType) bool {
	return ct != intConflictTypeNone && ct == other
}

// Implementation of RuleSet for integers.
type IntRuleSet[T integer] struct {
	NoConflict[T]
	strict       bool
	base         int
	rule         Rule[T]
	required     bool
	withNil      bool
	parent       *IntRuleSet[T]
	rounding     Rounding
	label        string
	conflictType intConflictType
	errorConfig  *errors.ErrorConfig
}

// Int creates a new integer RuleSet.
func Int() *IntRuleSet[int] {
	return &baseInt
}

// Uint creates a new unsigned integer RuleSet.
func Uint() *IntRuleSet[uint] {
	return &baseUint
}

// Int8 creates a new 8 bit integer RuleSet.
func Int8() *IntRuleSet[int8] {
	return &baseInt8
}

// Uint8 creates a new unsigned 8 bit integer RuleSet.
func Uint8() *IntRuleSet[uint8] {
	return &baseUint8
}

// Int16 creates a new 16 bit integer RuleSet.
func Int16() *IntRuleSet[int16] {
	return &baseInt16
}

// Uint16 creates a new unsigned 16 bit integer RuleSet.
func Uint16() *IntRuleSet[uint16] {
	return &baseUint16
}

// Int32 creates a new 32 bit integer RuleSet.
func Int32() *IntRuleSet[int32] {
	return &baseInt32
}

// Uint32 creates a new unsigned 32 bit integer RuleSet.
func Uint32() *IntRuleSet[uint32] {
	return &baseUint32
}

// Int64 creates a new int64 RuleSet.
func Int64() *IntRuleSet[int64] {
	return &baseInt64
}

// Uint64 creates a new unsigned 64 bit integer RuleSet.
func Uint64() *IntRuleSet[uint64] {
	return &baseUint64
}

// intCloneOption is a functional option for cloning IntRuleSet.
type intCloneOption[T integer] func(*IntRuleSet[T])

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (v *IntRuleSet[T]) clone(options ...intCloneOption[T]) *IntRuleSet[T] {
	newRuleSet := &IntRuleSet[T]{
		strict:       v.strict,
		base:         v.base,
		required:     v.required,
		withNil:      v.withNil,
		rounding:    v.rounding,
		parent:      v,
		errorConfig: v.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func intWithLabel[T integer](label string) intCloneOption[T] {
	return func(rs *IntRuleSet[T]) { rs.label = label }
}

func intWithErrorConfig[T integer](config *errors.ErrorConfig) intCloneOption[T] {
	return func(rs *IntRuleSet[T]) { rs.errorConfig = config }
}

func intWithConflictType[T integer](ct intConflictType) intCloneOption[T] {
	return func(rs *IntRuleSet[T]) {
		// Check for conflicts and update parent if needed
		if rs.parent != nil {
			rs.parent = rs.parent.noConflict(conflictTypeReplacesWrapper[T]{ct: ct})
		}
		rs.conflictType = ct
	}
}

// WithStrict returns a new child RuleSet that disables type coercion.
// When strict mode is enabled, validation only succeeds if the value is already the correct type.
//
// With number types, any type will work in strict mode as long as it can be converted
// deterministically and without loss.
func (v *IntRuleSet[T]) WithStrict() *IntRuleSet[T] {
	newRuleSet := v.clone(intWithLabel[T]("WithStrict()"), intWithConflictType[T](intConflictTypeStrict))
	newRuleSet.strict = true
	return newRuleSet
}

// WithBase returns a new child rule set that uses the specified base for string-to-number conversion and number-to-string conversion.
// The base determines how numeric strings are parsed from input (e.g., base 16 for hexadecimal) and how integers are formatted to strings in output.
// When outputting to a string type, the integer will be formatted using the specified base (e.g., base 16 will format as hexadecimal like "beef").
// The base has no effect if the RuleSet is strict since strict mode disables type conversion.
//
// The default is base 10.
func (v *IntRuleSet[T]) WithBase(base int) *IntRuleSet[T] {
	newRuleSet := v.clone(intWithLabel[T](fmt.Sprintf("WithBase(%d)", base)), intWithConflictType[T](intConflictTypeBase))
	newRuleSet.base = base
	return newRuleSet
}

// conflictTypeReplacesWrapper wraps a conflict type to implement Replaces[T]
type conflictTypeReplacesWrapper[T integer] struct {
	ct intConflictType
}

func (w conflictTypeReplacesWrapper[T]) Replaces(r Rule[T]) bool {
	// Try to cast to IntRuleSet to access conflictType
	if rs, ok := r.(interface{ getConflictType() intConflictType }); ok {
		return w.ct.Conflict(rs.getConflictType())
	}
	return false
}

// getConflictType returns the conflict type of the rule set.
// This is used by the conflict type wrapper to check for conflicts.
func (v *IntRuleSet[T]) getConflictType() intConflictType {
	return v.conflictType
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *IntRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (v *IntRuleSet[T]) WithRequired() *IntRuleSet[T] {
	newRuleSet := v.clone(intWithLabel[T]("WithRequired()"), intWithConflictType[T](intConflictTypeRequired))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *IntRuleSet[T]) WithNil() *IntRuleSet[T] {
	newRuleSet := v.clone(intWithLabel[T]("WithNil()"), intWithConflictType[T](intConflictTypeNil))
	newRuleSet.withNil = true
	return newRuleSet
}

// Apply performs validation of a RuleSet against a value and assigns the result to the output parameter.
// Apply returns a ValidationErrorCollection if any validation errors occur.
func (ruleSet *IntRuleSet[T]) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	// Add error config to context for error customization
	ctx = errors.WithErrorConfig(ctx, ruleSet.errorConfig)

	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, ruleSet.withNil, input, output); handled {
		return err
	}

	// Ensure output is a non-nil pointer
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "internal error", "Output must be a non-nil pointer",
		))
	}

	// Attempt to coerce the input value to an integer
	intval, validationErr := ruleSet.coerceInt(input, ctx)
	if validationErr != nil {
		return errors.Collection(validationErr)
	}

	// Handle setting the value in output
	outputElem := outputVal.Elem()

	var assignable bool

	// Format the integer as a string using the same base as input parsing
	strVal := formatInt(intval, ruleSet.base)

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
		outputElem.SetBool(intval != 0)
		assignable = true
	} else if (outputElem.Kind() == reflect.Interface && outputElem.IsNil()) ||
		(outputElem.Kind() == reflect.Int || outputElem.Kind() == reflect.Int8 ||
			outputElem.Kind() == reflect.Int16 || outputElem.Kind() == reflect.Int32 ||
			outputElem.Kind() == reflect.Int64 || outputElem.Type().AssignableTo(reflect.TypeOf(intval))) {

		// If output is a nil interface, or an assignable type, set it directly to the new integer value
		outputElem.Set(reflect.ValueOf(intval))
		assignable = true
	}

	// If the types are incompatible, return an error
	if !assignable {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "internal error", "Cannot assign %T to %T", intval, outputElem.Interface(),
		))
	}

	allErrors := errors.Collection()

	for currentRuleSet := ruleSet; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			if err := currentRuleSet.rule.Evaluate(ctx, intval); err != nil {
				allErrors = append(allErrors, err...)
			}
		}
	}

	if len(allErrors) != 0 {
		return allErrors
	}
	return nil
}

// Evaluate performs validation of a RuleSet against an integer value and returns a ValidationErrorCollection.
func (v *IntRuleSet[T]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	allErrors := errors.Collection()

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			if err := currentRuleSet.rule.Evaluate(ctx, value); err != nil {
				allErrors = append(allErrors, err...)
			}
		}
	}

	if len(allErrors) != 0 {
		return allErrors
	} else {
		return nil
	}
}

// noConflict returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *IntRuleSet[T]) noConflict(checker Replaces[T]) *IntRuleSet[T] {
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
func (ruleSet *IntRuleSet[T]) WithRule(rule Rule[T]) *IntRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.rule = rule
	newRuleSet.parent = ruleSet.noConflict(rule)
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
func (v *IntRuleSet[T]) WithRuleFunc(rule RuleFunc[T]) *IntRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the number RuleSet in an Any rule set
// which can then be used in nested validation.
func (v *IntRuleSet[T]) Any() RuleSet[any] {
	return WrapAny(v)
}

// typeName returns the name for the target integer type.
// Used for error message formatting.
func (v *IntRuleSet[T]) typeName() string {
	return reflect.ValueOf(*new(T)).Kind().String()
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *IntRuleSet[T]) String() string {
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
func (v *IntRuleSet[T]) WithErrorMessage(short, long string) *IntRuleSet[T] {
	return v.clone(intWithLabel[T](util.FormatErrorMessageLabel(short, long)), intWithErrorConfig[T](v.errorConfig.WithErrorMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (v *IntRuleSet[T]) WithDocsURI(uri string) *IntRuleSet[T] {
	return v.clone(intWithLabel[T](util.FormatStringArgLabel("WithDocsURI", uri)), intWithErrorConfig[T](v.errorConfig.WithDocsURI(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (v *IntRuleSet[T]) WithTraceURI(uri string) *IntRuleSet[T] {
	return v.clone(intWithLabel[T](util.FormatStringArgLabel("WithTraceURI", uri)), intWithErrorConfig[T](v.errorConfig.WithTraceURI(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (v *IntRuleSet[T]) WithErrorCode(code errors.ErrorCode) *IntRuleSet[T] {
	return v.clone(intWithLabel[T](util.FormatErrorCodeLabel(code)), intWithErrorConfig[T](v.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (v *IntRuleSet[T]) WithErrorMeta(key string, value any) *IntRuleSet[T] {
	return v.clone(intWithLabel[T](util.FormatErrorMetaLabel(key, value)), intWithErrorConfig[T](v.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (v *IntRuleSet[T]) WithErrorCallback(fn errors.ErrorCallback) *IntRuleSet[T] {
	return v.clone(intWithLabel[T](util.FormatErrorCallbackLabel()), intWithErrorConfig[T](v.errorConfig.WithCallback(fn)))
}
