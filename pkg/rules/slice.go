package rules

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// conflictType identifies the type of method that was called on a ruleset.
// Used for fast conflict checking instead of slow string prefix matching.
type conflictType int

const (
	conflictTypeNone conflictType = iota
	conflictTypeRequired
	conflictTypeNil
	conflictTypeMinLen
	conflictTypeMaxLen
)

// Conflict returns true if this conflict type conflicts with the other conflict type.
// Two conflict types conflict if they are the same (non-zero) value.
func (ct conflictType) Conflict(other conflictType) bool {
	return ct != conflictTypeNone && ct == other
}

// Replaces returns true if this conflict type replaces the given rule.
// It attempts to cast the rule to *SliceRuleSet and checks if the conflictType conflicts.
// Since SliceRuleSet is generic, we use an interface to access the conflictType field.
func (ct conflictType) Replaces(r any) bool {
	// Use an interface to access the conflictType field
	if rs, ok := r.(interface{ getConflictType() conflictType }); ok {
		return ct.Conflict(rs.getConflictType())
	}
	return false
}

// SliceStreamResult is a single result from ApplyStream or EvaluateStream.
// Index is the item index (0-based). Use Index == -1 for slice-level errors (e.g. min/max length, whole-slice rules).
// Value is the validated item (zero value when Err is set for that item, or for slice-level errors).
// Err is the validation error for this result, if any.
type SliceStreamResult[T any] struct {
	Index int
	Value T
	Err   errors.ValidationError
}

// Implementation of RuleSet for arrays of a given type.
type SliceRuleSet[T any] struct {
	NoConflict[[]T]
	itemRules    RuleSet[T]
	rule         Rule[[]T]
	maxLen       int // maxLen > 0 means max length is set, 0 means no limit
	minLen       int // minLen > 0 means min length is set, 0 means no limit
	required     bool
	withNil      bool
	parent       *SliceRuleSet[T]
	label        string
	conflictType conflictType
	errorConfig  *errors.ErrorConfig
}

// Slice creates a new slice RuleSet.
func Slice[T any]() *SliceRuleSet[T] {
	var empty [0]T

	return &SliceRuleSet[T]{
		label: fmt.Sprintf("SliceRuleSet[%s]", reflect.TypeOf(empty).Elem().Kind()),
	}
}

// sliceCloneOption is a functional option for cloning SliceRuleSet.
type sliceCloneOption[T any] func(*SliceRuleSet[T])

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (v *SliceRuleSet[T]) clone(options ...sliceCloneOption[T]) *SliceRuleSet[T] {
	newRuleSet := &SliceRuleSet[T]{
		itemRules:   v.itemRules,
		maxLen:      v.maxLen,
		minLen:      v.minLen,
		required:    v.required,
		withNil:     v.withNil,
		parent:      v,
		errorConfig: v.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func sliceWithLabel[T any](label string) sliceCloneOption[T] {
	return func(rs *SliceRuleSet[T]) { rs.label = label }
}

func sliceWithErrorConfig[T any](config *errors.ErrorConfig) sliceCloneOption[T] {
	return func(rs *SliceRuleSet[T]) { rs.errorConfig = config }
}

func sliceWithConflictType[T any](ct conflictType) sliceCloneOption[T] {
	return func(rs *SliceRuleSet[T]) {
		// Check for conflicts and update parent if needed
		if rs.parent != nil {
			rs.parent = rs.parent.noConflict(sliceConflictTypeReplacesWrapper[T]{ct: ct})
		}
		rs.conflictType = ct
	}
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *SliceRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
// WithRequired has no effect on slices if the RuleSet is strict since nil is not a valid slice.
func (v *SliceRuleSet[T]) WithRequired() *SliceRuleSet[T] {
	newRuleSet := v.clone(sliceWithLabel[T]("WithRequired()"), sliceWithConflictType[T](conflictTypeRequired))
	newRuleSet.required = true
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *SliceRuleSet[T]) WithNil() *SliceRuleSet[T] {
	newRuleSet := v.clone(sliceWithLabel[T]("WithNil()"), sliceWithConflictType[T](conflictTypeNil))
	newRuleSet.withNil = true
	return newRuleSet
}

// WithItemRuleSet takes a new rule set to use to validate array items and returns a new child rule set.
//
// If this function is called more than once, only the most recent one will be used to validate the items.
// If you don't set an item rule set then the validator will attempt to cast the items to the correct type
// and perform no additional validation.
func (v *SliceRuleSet[T]) WithItemRuleSet(itemRules RuleSet[T]) *SliceRuleSet[T] {
	newRuleSet := v.clone()
	newRuleSet.itemRules = itemRules
	return newRuleSet
}

// finishApply merges coercion errors, applies top-level rules, and returns the final error.
func (v *SliceRuleSet[T]) finishApply(ctx context.Context, outputItems []T, itemErrors errors.ValidationError, coercionErrors []errors.ValidationError) errors.ValidationError {
	var errs errors.ValidationError
	if itemErrors != nil {
		errs = errors.Join(errs, itemErrors)
	}
	for _, ce := range coercionErrors {
		if ce != nil {
			errs = errors.Join(errs, ce)
		}
	}
	if v.minLen > 0 {
		actualLen := len(outputItems)
		if actualLen < v.minLen {
			errs = errors.Join(errs, errors.Error(errors.CodeMinLen, ctx, v.minLen))
		}
	}
	if len(outputItems) > 0 {
		for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
			if currentRuleSet.rule != nil {
				if err := currentRuleSet.rule.Evaluate(ctx, outputItems); err != nil {
					errs = errors.Join(errs, err)
				}
			}
		}
	}
	return errs
}

// newInputChan converts a slice or array to a channel and returns the channel, original items, and coercion errors by index.
// originalItems is populated when itemRuleSet exists, allowing it to process items that couldn't be cast to T.
// coercionByIndex has length valueOf.Len(); coercionByIndex[i] is non-nil when coercion failed for index i.
func (v *SliceRuleSet[T]) newInputChan(ctx context.Context, valueOf reflect.Value) (<-chan T, []any, []errors.ValidationError) {
	n := valueOf.Len()
	ch := make(chan T)
	var itemRuleSet RuleSet[T]
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.itemRules != nil {
			itemRuleSet = currentRuleSet.itemRules
			break
		}
	}

	var originalItems []any
	coercionByIndex := make([]errors.ValidationError, n)

	if itemRuleSet != nil {
		originalItems = make([]any, n)
	}

	go func() {
		defer close(ch)
		for i := 0; i < n; i++ {
			item := valueOf.Index(i)
			itemInterface := item.Interface()
			var castItem T
			if c, ok := itemInterface.(T); ok {
				castItem = c
			} else {
				if originalItems != nil {
					originalItems[i] = itemInterface
				}
			}
			select {
			case <-ctx.Done():
				return
			case ch <- castItem:
			}
		}
	}()

	if itemRuleSet == nil {
		expectedType := reflect.TypeOf((*T)(nil)).Elem()
		for i := 0; i < n; i++ {
			item := valueOf.Index(i)
			itemInterface := item.Interface()
			if _, ok := itemInterface.(T); !ok {
				subContext := rulecontext.WithPathString(ctx, strconv.Itoa(i))
				actual := item.Kind().String()
				coercionByIndex[i] = errors.Error(errors.CodeType, subContext, expectedType.Name(), actual)
			}
		}
	}

	return ch, originalItems, coercionByIndex
}

// applyChan performs streaming validation from an input channel.
// All errors are collected and returned at once. Always returns the collected slice.
// originalItems is optional - if provided, it contains the original items before casting.
// applyChan does NOT close channels - they are managed by the caller.
func (v *SliceRuleSet[T]) applyChan(ctx context.Context, input <-chan T, originalItems []any) ([]T, errors.ValidationError) {
	var errs errors.ValidationError
	outputItems := make([]T, 0)
	var index int

	maxLen := v.maxLen

	// Check for an item RuleSet
	var itemRuleSet RuleSet[T]
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.itemRules != nil {
			itemRuleSet = currentRuleSet.itemRules
			break
		}
	}

	// Stream process items - stop processing after maxLen if set
	for {
		select {
		case <-ctx.Done():
			errs = errors.Join(errs, contextErrorToValidation(ctx))
			return outputItems, errs
		case item, ok := <-input:
			if !ok {
				if ctx.Err() != nil {
					errs = errors.Join(errs, contextErrorToValidation(ctx))
				}
				return outputItems, errs
			}

			// Check maxLen proactively - stop applying item rules after maxLen
			// Item rules are applied up to maxLen, after which we stop processing items
			if maxLen > 0 && index >= maxLen {
				errs = errors.Join(errs, errors.Error(errors.CodeMaxLen, ctx, maxLen))
				return outputItems, errs
			}

			var itemOutput T
			var itemErr errors.ValidationError
			if itemRuleSet != nil {
				subContext := rulecontext.WithPathIndex(ctx, index)
				var itemInput any = item
				if originalItems != nil && index < len(originalItems) && originalItems[index] != nil {
					itemInput = originalItems[index]
				}
				itemOutput, itemErr = itemRuleSet.Apply(subContext, itemInput)
				if itemErr != nil {
					itemOutput = item
					errs = errors.Join(errs, itemErr)
				}
			} else {
				// No item rules
				itemOutput = item
			}

			outputItems = append(outputItems, itemOutput)
			index++
		}
	}
}

// applyChanStream reads from input, validates each item, and sends SliceStreamResult to output.
// coercionByIndex[i] is the coercion error for index i (nil if no error); may be nil when no coercion errors.
// output is closed when done. Slice-level errors are sent with Index == -1.
func (v *SliceRuleSet[T]) applyChanStream(ctx context.Context, input <-chan T, originalItems []any, coercionByIndex []errors.ValidationError, output chan<- SliceStreamResult[T]) {
	var outputItems []T
	index := 0
	maxLen := v.maxLen

	var itemRuleSet RuleSet[T]
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.itemRules != nil {
			itemRuleSet = currentRuleSet.itemRules
			break
		}
	}

	defer close(output)

	for {
		select {
		case <-ctx.Done():
			output <- SliceStreamResult[T]{Index: -1, Err: contextErrorToValidation(ctx)}
			return
		case item, ok := <-input:
			if !ok {
				// Input exhausted; run slice-level validation and send any errors with Index -1
				var sliceLevelErr errors.ValidationError
				if v.minLen > 0 && len(outputItems) < v.minLen {
					sliceLevelErr = errors.Join(sliceLevelErr, errors.Error(errors.CodeMinLen, ctx, v.minLen))
				}
				if len(outputItems) > 0 {
					for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
						if currentRuleSet.rule != nil {
							if err := currentRuleSet.rule.Evaluate(ctx, outputItems); err != nil {
								sliceLevelErr = errors.Join(sliceLevelErr, err)
							}
						}
					}
				}
				if sliceLevelErr != nil {
					output <- SliceStreamResult[T]{Index: -1, Err: sliceLevelErr}
				}
				return
			}

			if maxLen > 0 && index >= maxLen {
				output <- SliceStreamResult[T]{Index: -1, Err: errors.Error(errors.CodeMaxLen, ctx, maxLen)}
				return
			}

			if coercionByIndex != nil && index < len(coercionByIndex) && coercionByIndex[index] != nil {
				var zero T
				output <- SliceStreamResult[T]{Index: index, Err: coercionByIndex[index]}
				outputItems = append(outputItems, zero)
			} else {
				var itemOutput T
				var itemErr errors.ValidationError
				if itemRuleSet != nil {
					subContext := rulecontext.WithPathIndex(ctx, index)
					itemInput := any(item)
					if originalItems != nil && index < len(originalItems) && originalItems[index] != nil {
						itemInput = originalItems[index]
					}
					itemOutput, itemErr = itemRuleSet.Apply(subContext, itemInput)
					if itemErr != nil {
						itemOutput = item
					}
				} else {
					itemOutput = item
				}
				output <- SliceStreamResult[T]{Index: index, Value: itemOutput, Err: itemErr}
				outputItems = append(outputItems, itemOutput)
			}
			index++
		}
	}
}

// Apply coerces input to []T, evaluates item and slice-level rules, and returns the result.
// Input may be a slice, array, or receive-only channel of T.
func (v *SliceRuleSet[T]) Apply(ctx context.Context, input any) ([]T, errors.ValidationError) {
	ctx = errors.WithErrorConfig(ctx, v.errorConfig)

	if handled, err := util.TryNilIfAllowed(ctx, v.withNil, input); handled {
		return nil, err
	}

	valueOf := reflect.ValueOf(input)
	typeOf := valueOf.Type()
	inputKind := typeOf.Kind()

	var inputChan <-chan T
	var coercionErrors []errors.ValidationError
	var originalItems []any

	switch inputKind {
	case reflect.Chan:
		inputVal := reflect.ValueOf(input)
		if inputVal.IsNil() {
			return nil, errors.Errorf(errors.CodeInternal, ctx, "internal error", "Input channel cannot be nil")
		}
		var recvChan <-chan T
		switch ch := input.(type) {
		case <-chan T:
			recvChan = ch
		case chan T:
			recvChan = ch
		default:
			expectedType := reflect.TypeOf((*T)(nil)).Elem()
			actualType := inputVal.Type().Elem()
			return nil, errors.Error(errors.CodeType, ctx, expectedType.String(), actualType.String())
		}
		inputChan = recvChan
	case reflect.Slice, reflect.Array:
		inputChan, originalItems, coercionErrors = v.newInputChan(ctx, valueOf)
	default:
		return nil, errors.Error(errors.CodeType, ctx, "array", inputKind.String())
	}

	outputItems, itemErrors := v.applyChan(ctx, inputChan, originalItems)
	errs := v.finishApply(ctx, outputItems, itemErrors, coercionErrors)
	return outputItems, errs
}

// Evaluate performs validation of a RuleSet against a slice type and returns a ValidationError.
func (ruleSet *SliceRuleSet[T]) Evaluate(ctx context.Context, value []T) errors.ValidationError {
	_, err := ruleSet.Apply(ctx, value)
	return err
}

// ApplyStream coerces input to a stream of T, validates each item, and sends results on the returned channel.
// Each result includes Index (0-based), Value, and Err. Slice-level errors use Index == -1.
// The channel is closed when processing is complete. Input may be a slice, array, or receive-only channel of T.
func (v *SliceRuleSet[T]) ApplyStream(ctx context.Context, input any) (<-chan SliceStreamResult[T], errors.ValidationError) {
	ctx = errors.WithErrorConfig(ctx, v.errorConfig)

	if handled, err := util.TryNilIfAllowed(ctx, v.withNil, input); handled {
		return nil, err
	}

	valueOf := reflect.ValueOf(input)
	typeOf := valueOf.Type()
	inputKind := typeOf.Kind()

	var inputChan <-chan T
	var coercionByIndex []errors.ValidationError
	var originalItems []any

	switch inputKind {
	case reflect.Chan:
		inputVal := reflect.ValueOf(input)
		if inputVal.IsNil() {
			return nil, errors.Errorf(errors.CodeInternal, ctx, "internal error", "Input channel cannot be nil")
		}
		switch ch := input.(type) {
		case chan T:
			inputChan = ch
		case <-chan T:
			inputChan = ch
		default:
			expectedType := reflect.TypeOf((*T)(nil)).Elem()
			actualType := inputVal.Type().Elem()
			return nil, errors.Error(errors.CodeType, ctx, expectedType.String(), actualType.String())
		}
	case reflect.Slice, reflect.Array:
		inputChan, originalItems, coercionByIndex = v.newInputChan(ctx, valueOf)
	default:
		return nil, errors.Error(errors.CodeType, ctx, "array", inputKind.String())
	}

	output := make(chan SliceStreamResult[T])
	go v.applyChanStream(ctx, inputChan, originalItems, coercionByIndex, output)
	return output, nil
}

// EvaluateStream validates a stream of T (from the given channel) and sends results on the returned channel.
// Each result includes Index (0-based), Value, and Err. Slice-level errors use Index == -1.
// The returned channel is closed when the input channel is closed and slice-level validation has run.
func (v *SliceRuleSet[T]) EvaluateStream(ctx context.Context, input <-chan T) (<-chan SliceStreamResult[T], errors.ValidationError) {
	ctx = errors.WithErrorConfig(ctx, v.errorConfig)
	if input == nil {
		return nil, errors.Errorf(errors.CodeInternal, ctx, "internal error", "Input channel cannot be nil")
	}
	output := make(chan SliceStreamResult[T])
	go v.applyChanStream(ctx, input, nil, nil, output)
	return output, nil
}

// sliceConflictTypeReplacesWrapper wraps a conflict type to implement Replaces[[]T]
type sliceConflictTypeReplacesWrapper[T any] struct {
	ct conflictType
}

func (w sliceConflictTypeReplacesWrapper[T]) Replaces(r Rule[[]T]) bool {
	// Try to cast to SliceRuleSet to access conflictType
	if rs, ok := r.(interface{ getConflictType() conflictType }); ok {
		return w.ct.Conflict(rs.getConflictType())
	}
	return false
}

// getConflictType returns the conflict type of the rule set.
// This is used by the conflict type wrapper to check for conflicts.
func (v *SliceRuleSet[T]) getConflictType() conflictType {
	return v.conflictType
}

// noConflict returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *SliceRuleSet[T]) noConflict(checker Replaces[[]T]) *SliceRuleSet[T] {
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
// Note: Adding a rule at this level will result in the whole output being buffered in memory,
// which could have performance implications on larger slices. Top-level rules are applied after
// all items are processed, requiring all validated items to be collected before rule evaluation.
func (v *SliceRuleSet[T]) WithRule(rule Rule[[]T]) *SliceRuleSet[T] {
	newRuleSet := v.clone()
	newRuleSet.rule = rule
	newRuleSet.parent = v.noConflict(rule)
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
//
// Note: Adding a rule at this level will result in the whole output being buffered in memory,
// which could have performance implications on larger slices. Top-level rules are applied after
// all items are processed, requiring all validated items to be collected before rule evaluation.
func (v *SliceRuleSet[T]) WithRuleFunc(rule RuleFunc[[]T]) *SliceRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the slice RuleSet in an Any rule set
// which can then be used in nested validation.
func (v *SliceRuleSet[T]) Any() RuleSet[any] {
	return WrapAny(v)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *SliceRuleSet[T]) String() string {
	label := ruleSet.label

	if label == "" {
		if ruleSet.rule != nil {
			label = ruleSet.rule.String()
		} else if ruleSet.itemRules != nil {
			label = fmt.Sprintf("WithItemRuleSet(%s)", ruleSet.itemRules)
		}
	}

	if ruleSet.parent != nil {
		return ruleSet.parent.String() + "." + label
	}
	return label
}

// WithErrorMessage returns a new RuleSet with custom short and long error messages.
func (v *SliceRuleSet[T]) WithErrorMessage(short, long string) *SliceRuleSet[T] {
	return v.clone(sliceWithLabel[T](util.FormatErrorMessageLabel(short, long)), sliceWithErrorConfig[T](v.errorConfig.WithErrorMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (v *SliceRuleSet[T]) WithDocsURI(uri string) *SliceRuleSet[T] {
	return v.clone(sliceWithLabel[T](util.FormatStringArgLabel("WithDocsURI", uri)), sliceWithErrorConfig[T](v.errorConfig.WithDocsURI(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (v *SliceRuleSet[T]) WithTraceURI(uri string) *SliceRuleSet[T] {
	return v.clone(sliceWithLabel[T](util.FormatStringArgLabel("WithTraceURI", uri)), sliceWithErrorConfig[T](v.errorConfig.WithTraceURI(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (v *SliceRuleSet[T]) WithErrorCode(code errors.ErrorCode) *SliceRuleSet[T] {
	return v.clone(sliceWithLabel[T](util.FormatErrorCodeLabel(code)), sliceWithErrorConfig[T](v.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (v *SliceRuleSet[T]) WithErrorMeta(key string, value any) *SliceRuleSet[T] {
	return v.clone(sliceWithLabel[T](util.FormatErrorMetaLabel(key, value)), sliceWithErrorConfig[T](v.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (v *SliceRuleSet[T]) WithErrorCallback(fn errors.ErrorCallback) *SliceRuleSet[T] {
	return v.clone(sliceWithLabel[T](util.FormatErrorCallbackLabel()), sliceWithErrorConfig[T](v.errorConfig.WithCallback(fn)))
}
