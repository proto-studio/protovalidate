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

// Implementation of RuleSet for arrays of a given type.
type SliceRuleSet[T any] struct {
	NoConflict[[]T]
	itemRules RuleSet[T]
	rule      Rule[[]T]
	required  bool
	withNil   bool
	parent    *SliceRuleSet[T]
	label     string
}

// Slice creates a new slice RuleSet.
func Slice[T any]() *SliceRuleSet[T] {
	var empty [0]T

	return &SliceRuleSet[T]{
		label: fmt.Sprintf("SliceRuleSet[%s]", reflect.TypeOf(empty).Elem().Kind()),
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
	return &SliceRuleSet[T]{
		parent:   v,
		required: true,
		withNil:  v.withNil,
		label:    "WithRequired()",
	}
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *SliceRuleSet[T]) WithNil() *SliceRuleSet[T] {
	return &SliceRuleSet[T]{
		parent:   v,
		required: v.required,
		withNil:  true,
		label:    "WithNil()",
	}
}

// WithItemRuleSet takes a new rule set to use to validate array items and returns a new child rule set.
//
// If this function is called more than once, only the most recent one will be used to validate the items.
// If you don't set an item rule set then the validator will attempt to cast the items to the correct type
// and perform no additional validation.
func (v *SliceRuleSet[T]) WithItemRuleSet(itemRules RuleSet[T]) *SliceRuleSet[T] {
	return &SliceRuleSet[T]{
		itemRules: itemRules,
		parent:    v,
		required:  v.required,
		withNil:   v.withNil,
	}
}

// finishApply merges coercion errors, applies top-level rules, and returns the final error collection.
func (v *SliceRuleSet[T]) finishApply(ctx context.Context, outputItems []T, itemErrors errors.ValidationErrorCollection, coercionErrors []errors.ValidationErrorCollection) errors.ValidationErrorCollection {
	// Merge coercion errors if any
	allErrors := itemErrors
	if len(coercionErrors) > 0 {
		for _, ce := range coercionErrors {
			if ce != nil {
				allErrors = append(allErrors, ce...)
			}
		}
	}

	// Apply top-level rules (including maxLen) on collected output
	if len(outputItems) > 0 {
		for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
			if currentRuleSet.rule != nil {
				if err := currentRuleSet.rule.Evaluate(ctx, outputItems); err != nil {
					allErrors = append(allErrors, err...)
				}
			}
		}
	}

	if len(allErrors) != 0 {
		return allErrors
	}
	return nil
}

// newInputChan converts a slice or array to a channel and returns the channel, original items, and coercion errors.
// originalItems is populated when itemRuleSet exists, allowing it to process items that couldn't be cast to T.
// coercionErrors is populated when no itemRuleSet exists, tracking items that couldn't be cast.
func (v *SliceRuleSet[T]) newInputChan(ctx context.Context, valueOf reflect.Value) (<-chan T, []any, []errors.ValidationErrorCollection, errors.ValidationErrorCollection) {
	// Convert slice/array to channel
	// Note: maxLen is checked at the end as a top-level rule (after all items are processed)
	// Send all items - if they can't be cast to T, send zero value
	// Track original items for itemRuleSet processing
	// Use unbuffered channel (size 0) - no need to buffer all items upfront
	ch := make(chan T)
	var itemRuleSet RuleSet[T]
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.itemRules != nil {
			itemRuleSet = currentRuleSet.itemRules
			break
		}
	}

	var originalItems []any
	var coercionErrors []errors.ValidationErrorCollection

	// If we have itemRuleSet, track original items for items that can't be cast
	if itemRuleSet != nil {
		originalItems = make([]any, valueOf.Len())
	}

	go func() {
		defer close(ch)
		for i := 0; i < valueOf.Len(); i++ {
			item := valueOf.Index(i)
			itemInterface := item.Interface()
			var castItem T
			if c, ok := itemInterface.(T); ok {
				castItem = c
			} else {
				// Cast failed - send zero value
				// Store original item for itemRuleSet processing
				if originalItems != nil {
					originalItems[i] = itemInterface
				}
			}
			// Always send, even if cast failed (zero value)
			select {
			case <-ctx.Done():
				return
			case ch <- castItem:
			}
		}
	}()

	// If no itemRuleSet, track coercion errors during conversion
	if itemRuleSet == nil {
		expectedType := reflect.TypeOf((*T)(nil)).Elem()
		for i := 0; i < valueOf.Len(); i++ {
			item := valueOf.Index(i)
			itemInterface := item.Interface()
			if _, ok := itemInterface.(T); !ok {
				subContext := rulecontext.WithPathString(ctx, strconv.Itoa(i))
				actual := item.Kind().String()
				coercionErrors = append(coercionErrors, errors.Collection(errors.NewCoercionError(subContext, expectedType.Name(), actual)))
			}
		}
	}

	return ch, originalItems, coercionErrors, nil
}

// applyChan performs streaming validation from an input channel to an output channel.
// Items are validated and written to output as they are read from input.
// All errors are collected and returned at once.
// originalItems is optional - if provided, it contains the original items before casting
// (used when itemRuleSet needs to process original items that couldn't be cast to T)
// applyChan does NOT close channels - they are managed by the caller.
// applyChan returns the collected items and errors. Top-level rules are NOT applied here.
func (v *SliceRuleSet[T]) applyChan(ctx context.Context, input <-chan T, output chan<- T, originalItems []any) ([]T, errors.ValidationErrorCollection) {
	var allErrors = errors.Collection()
	var outputItems []T
	var index int

	// Check if we need to collect items for top-level rules
	var hasTopLevelRules bool
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule != nil {
			hasTopLevelRules = true
			break
		}
	}

	// Only allocate outputItems if we have top-level rules
	if hasTopLevelRules {
		outputItems = make([]T, 0)
	}

	// Check for an item RuleSet
	var itemRuleSet RuleSet[T]
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.itemRules != nil {
			itemRuleSet = currentRuleSet.itemRules
			break
		}
	}

	// Stream process items - process ALL items (no early maxLen check)
	for {
		select {
		case <-ctx.Done():
			allErrors = append(allErrors, contextErrorToValidation(ctx))
			return outputItems, allErrors
		case item, ok := <-input:
			if !ok {
				// Input channel closed - all items processed
				// Return items and errors (top-level rules will be applied in Apply)
				return outputItems, allErrors
			}

			// Validate item
			var itemOutput T
			var itemErr errors.ValidationErrorCollection

			if itemRuleSet != nil {
				subContext := rulecontext.WithPathIndex(ctx, index)
				// Use original item if available (for items that couldn't be cast to T)
				var itemInput any = item
				if originalItems != nil && index < len(originalItems) && originalItems[index] != nil {
					itemInput = originalItems[index]
				}
				itemErr = itemRuleSet.Apply(subContext, itemInput, &itemOutput)
				if itemErr != nil {
					// Try to use original item if validation fails
					itemOutput = item
					allErrors = append(allErrors, itemErr...)
				}
			} else {
				// No item rules
				itemOutput = item
			}

			// Write to output channel immediately
			select {
			case <-ctx.Done():
				allErrors = append(allErrors, contextErrorToValidation(ctx))
				return outputItems, allErrors
			case output <- itemOutput:
				if hasTopLevelRules {
					outputItems = append(outputItems, itemOutput)
				}
				index++
			}
		}
	}
}

// Apply performs validation of a RuleSet against a value and assigns the result to the output parameter.
// Apply returns a ValidationErrorCollection if any validation errors occur.
//
// Apply supports channels as both input and output. When using channels:
// - Input channel: reads values until closed, max length is hit, or context times out
// - Output channel: writes validated values in the same order as input
// - All errors are collected and returned at once
// - Items are streamed (validated and written immediately, not collected upfront)
func (v *SliceRuleSet[T]) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	// Check if withNil is enabled and input is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, v.withNil, input, output); handled {
		return err
	}

	// Ensure output is a non-nil pointer
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Output must be a non-nil pointer",
		))
	}

	outputElem := outputVal.Elem()
	outputElemKind := outputElem.Kind()

	// Validate output type early (before processing input)
	expectedType := reflect.TypeOf((*T)(nil)).Elem()
	expectedSliceType := reflect.TypeOf([]T(nil))

	if outputElemKind == reflect.Chan {
		// Validate channel element type
		if outputElem.IsNil() {
			return errors.Collection(errors.Errorf(
				errors.CodeInternal, ctx, "Output channel cannot be nil",
			))
		}
		actualType := outputElem.Type().Elem()
		if !actualType.AssignableTo(expectedType) {
			return errors.Collection(errors.Errorf(
				errors.CodeInternal, ctx, "Output channel element type %s is not compatible with %s",
				actualType.String(), expectedType.String(),
			))
		}
	} else if outputElemKind == reflect.Interface {
		// Interface output: check if []T is assignable to the interface type
		// If nil, it's valid (we'll set it). If not nil, check assignability.
		if !outputElem.IsNil() {
			if !expectedSliceType.AssignableTo(outputElem.Type()) {
				return errors.Collection(errors.Errorf(
					errors.CodeInternal, ctx, "Cannot assign %T to %T", []T(nil), outputElem.Interface(),
				))
			}
		}
	} else if outputElemKind == reflect.Slice {
		// Validate slice element type - check if []T is assignable to output slice type
		if !expectedSliceType.AssignableTo(outputElem.Type()) {
			return errors.Collection(errors.Errorf(
				errors.CodeInternal, ctx, "Cannot assign %T to %T", []T(nil), outputElem.Interface(),
			))
		}
	} else {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Output must be a slice or channel, got %s", outputElemKind,
		))
	}

	valueOf := reflect.ValueOf(input)
	typeOf := valueOf.Type()
	inputKind := typeOf.Kind()

	// Determine input channel
	var inputChan <-chan T
	var coercionErrors []errors.ValidationErrorCollection
	var originalItems []any

	if inputKind == reflect.Chan {
		// Input is already a channel
		inputVal := reflect.ValueOf(input)
		if inputVal.IsNil() {
			return errors.Collection(errors.Errorf(
				errors.CodeInternal, ctx, "Input channel cannot be nil",
			))
		}

		// Convert to receive-only channel
		var recvChan <-chan T
		switch ch := input.(type) {
		case <-chan T:
			recvChan = ch
		case chan T:
			recvChan = ch
		default:
			// Type assertion failed
			expectedType := reflect.TypeOf((*T)(nil)).Elem()
			actualType := inputVal.Type().Elem()
			return errors.Collection(errors.NewCoercionError(
				ctx, expectedType.String(), actualType.String(),
			))
		}
		inputChan = recvChan
	} else if inputKind == reflect.Slice || inputKind == reflect.Array {
		var err errors.ValidationErrorCollection
		inputChan, originalItems, coercionErrors, err = v.newInputChan(ctx, valueOf)
		if err != nil {
			return err
		}
	} else {
		return errors.Collection(errors.NewCoercionError(ctx, "array", inputKind.String()))
	}

	// Determine output channel and setup
	var outputChan chan<- T
	var outputSlice *[]T
	var outputSliceInterface []T
	var done chan struct{}
	var closeOutputChan bool

	if outputElemKind == reflect.Chan {
		// Output is already a channel - convert to send-only
		// We already validated the channel type earlier
		var sendChan chan<- T
		switch ch := outputElem.Interface().(type) {
		case chan<- T:
			sendChan = ch
		case chan T:
			sendChan = ch
		default:
			// Should not happen - we validated earlier, but handle gracefully
			return errors.Collection(errors.Errorf(
				errors.CodeInternal, ctx, "Output channel type assertion failed",
			))
		}
		outputChan = sendChan
		closeOutputChan = false // Caller manages the channel
	} else if outputElemKind == reflect.Interface {
		// For interface{} output, create a slice and assign it
		ch := make(chan T, 100) // Buffered to allow streaming
		outputChan = ch
		outputSliceInterface = make([]T, 0)
		closeOutputChan = true // We created it

		// Collect results synchronously in background
		done = make(chan struct{})
		go func() {
			defer close(done)
			for item := range ch {
				outputSliceInterface = append(outputSliceInterface, item)
			}
		}()
	} else {
		// Slice output
		ch := make(chan T, 100) // Buffered to allow streaming
		outputChan = ch
		outputSlice = outputElem.Addr().Interface().(*[]T)
		*outputSlice = make([]T, 0)
		closeOutputChan = true // We created it

		// Collect results synchronously in background
		done = make(chan struct{})
		go func() {
			defer close(done)
			for item := range ch {
				*outputSlice = append(*outputSlice, item)
			}
		}()
	}

	// Use applyChan for streaming validation
	outputItems, itemErrors := v.applyChan(ctx, inputChan, outputChan, originalItems)

	// Close output channel only if we created it
	// For caller-provided channels, we don't close - the caller manages it
	// Completion is signaled by returning from Apply, not by closing the channel
	if closeOutputChan {
		close(outputChan)
		// Wait for collection to complete
		<-done
	}

	// Assign the slice to interface{} if needed
	if outputElemKind == reflect.Interface {
		outputElem.Set(reflect.ValueOf(outputSliceInterface))
	}

	// Merge coercion errors and apply top-level rules (shared logic)
	return v.finishApply(ctx, outputItems, itemErrors, coercionErrors)
}

// Evaluate performs validation of a RuleSet against a slice type and returns a ValidationErrorCollection.
func (ruleSet *SliceRuleSet[T]) Evaluate(ctx context.Context, value []T) errors.ValidationErrorCollection {
	var out any
	return ruleSet.Apply(ctx, value, &out)
}

// noConflict returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *SliceRuleSet[T]) noConflict(rule Rule[[]T]) *SliceRuleSet[T] {

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

	return &SliceRuleSet[T]{
		rule:      ruleSet.rule,
		parent:    newParent,
		required:  ruleSet.required,
		withNil:   ruleSet.withNil,
		itemRules: ruleSet.itemRules,
		label:     ruleSet.label,
	}
}

// WithRule returns a new child rule set that applies a custom validation rule.
// The custom rule is evaluated during validation and any errors it returns are included in the validation result.
//
// Note: Adding a rule at this level will result in the whole output being buffered in memory,
// which could have performance implications on larger slices. Top-level rules are applied after
// all items are processed, requiring all validated items to be collected before rule evaluation.
func (v *SliceRuleSet[T]) WithRule(rule Rule[[]T]) *SliceRuleSet[T] {
	return &SliceRuleSet[T]{
		rule:     rule,
		parent:   v.noConflict(rule),
		required: v.required,
		withNil:  v.withNil,
	}
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
	return WrapAny[[]T](v)
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
