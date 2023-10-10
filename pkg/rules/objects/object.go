// Package objects provides a RuleSet implementation that can be used to validate object and map values.
//
// It implements standard rules and allows the developer to set a rule set to validate individual keys.
package objects

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

const annotation = "validate"

// Implementation of RuleSet for objects and maps.
type ObjectRuleSet[T any] struct {
	rules.NoConflict[T]
	allowUnknown bool
	key          string
	rule         rules.RuleSet[any]
	objRule      rules.Rule[T]
	mapping      string
	outputType   reflect.Type
	ptr          bool
	required     bool
	parent       *ObjectRuleSet[T]
	label        string
	condition    Conditional[T]
	refs         *refTracker
}

// New returns a RuleSet that can be used to validate an object of an
// arbitrary data type.
//
// Using the "validate" annotation you can may input values to different
// properties of the object. This is useful for converting unstructured maps
// created from Json and converting to an object.
func New[T any]() *ObjectRuleSet[T] {
	var empty [0]T

	ruleSet := &ObjectRuleSet[T]{
		outputType: reflect.TypeOf(empty).Elem(),
	}

	kind := ruleSet.outputType.Kind()

	ruleSet.ptr = kind == reflect.Pointer
	if ruleSet.ptr {
		ruleSet.outputType = ruleSet.outputType.Elem()
		kind = ruleSet.outputType.Kind()
		ruleSet.label = fmt.Sprintf("ObjectRuleSet[*%v]", ruleSet.outputType)
	} else {
		ruleSet.label = fmt.Sprintf("ObjectRuleSet[%v]", ruleSet.outputType)
	}

	if kind != reflect.Struct && kind != reflect.Map {
		panic(fmt.Errorf("invalid output type for object rule se: %v", kind))
	}

	mapped := make(map[string]bool)

	for i := 0; i < ruleSet.outputType.NumField(); i++ {
		field := ruleSet.outputType.Field(i)

		if !field.IsExported() {
			continue
		}

		tagValue, ok := field.Tag.Lookup(annotation)
		emptyTag := tagValue == ""

		// Ignore empty tags if they exist
		if ok && emptyTag {
			continue
		}

		var key string
		if emptyTag {
			key = field.Name

			// Don't allow the property names name to override the tagged mapping
			_, ok := mapped[key]
			if ok {
				continue
			}
		} else {
			key = tagValue
		}

		ruleSet = &ObjectRuleSet[T]{
			parent:     ruleSet,
			key:        key,
			mapping:    field.Name,
			outputType: ruleSet.outputType,
			ptr:        ruleSet.ptr,
		}

		mapped[key] = true
	}

	return ruleSet
}

// NewObjectMap returns a new RuleSet that can be used to validate maps with strings as the
// keys and the specified data type (which can be "any") as the values.
func NewObjectMap[T any]() *ObjectRuleSet[map[string]T] {
	var empty map[string]T

	return &ObjectRuleSet[map[string]T]{
		outputType: reflect.TypeOf(empty),
	}
}

// withParent is a helper function to assist in cloning object RuleSets.
func (v *ObjectRuleSet[T]) withParent() *ObjectRuleSet[T] {
	return &ObjectRuleSet[T]{
		allowUnknown: v.allowUnknown,
		required:     v.required,
		outputType:   v.outputType,
		ptr:          v.ptr,
		parent:       v,
		refs:         v.refs,
	}
}

// WithUnknown returns a new RuleSet with the "unknown" flag set.
//
// By default if the validator fines an unknown key on a map it will return an error.
// Setting the unknown flag will allow keys that aren't defined to be present in the map.
// This is useful for parsing arbitrary Json where additional keys may be included.
func (v *ObjectRuleSet[T]) WithUnknown() *ObjectRuleSet[T] {
	newRuleSet := v.withParent()
	newRuleSet.allowUnknown = true
	newRuleSet.label = "WithUnknown()"
	return newRuleSet
}

// fullMapping is a helper function that returns the full object field mappings as a map.
func (v *ObjectRuleSet[T]) fullMapping() map[string]string {
	mapping := make(map[string]string)

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key != "" && currentRuleSet.mapping != "" {
			mapping[currentRuleSet.key] = currentRuleSet.mapping
		}
	}
	return mapping
}

// mappingFor is a helper function that returns the key mapping given a specific key.
func (v *ObjectRuleSet[T]) mappingFor(key string) (string, bool) {
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key == key && currentRuleSet.mapping != "" {
			return currentRuleSet.mapping, true
		}
	}
	return "", false
}

// WithKey returns a new RuleSet with a validation rule for the specified key.
//
// If more than one call is made with the same key than all will be evaluated. However, the order
// in which they are run is not guaranteed.
//
// Multiple rule sets may run in parallel but only one will run a time for each key.
func (v *ObjectRuleSet[T]) WithKey(key string, ruleSet rules.RuleSet[any]) *ObjectRuleSet[T] {
	return v.WithConditionalKey(key, nil, ruleSet)
}

// Keys returns the keys names that have rule sets associated with them.
// This will not return keys that don't have rule sets (even if they do have a mapping).
//
// It also will not return keys that are referenced WithRule or WithRuleFund. To get around this
// you may want to consider moving your rule set to WithKey or putting a simple permissive validator
// inside WithKey.
//
// The results are not sorted. You should not depend on the order of the results.
func (v *ObjectRuleSet[T]) Keys() []string {
	mapping := make(map[string]bool)
	keys := make([]string, 0)

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key != "" && currentRuleSet.rule != nil {
			if !mapping[currentRuleSet.key] {
				mapping[currentRuleSet.key] = true
				keys = append(keys, currentRuleSet.key)
			}
		}
	}

	return keys
}

// WithConditionalKey returns a new Rule with a validation rule for the specified key.
//
// It takes as an argument a Rule that is used to evaluate the entire object or map. If it returns a nil error then
// the conditional key Rule will be evaluated.
//
// Errors returned from the conditional Rule are not considered validation failures and will not be returned from
// the Validate / Evaluate functions. Errors in the conditional are only used to determine if the Rule should be evaluated.
//
// Conditional rules will be run any time after all fields they depend on are evaluated. For example if the conditional
// rule set looks for keys X and Y then the conditional will not be evaluated until all the rules for both X and Y have
// also been evaluated. This includes conditional rules. So if X is also dependent on Z then Z will also need to be complete.
//
// If one or more of the fields has an error then the conditional rule will not be run.
//
// WithRule and WithRuleFunc are both evaluated after any keys or conditional keys. Because of this, it is not possible to
// have a conditional key that is dependent on data that is modified in those rules.
//
// If nil is passed in as the conditional then this method behaves identical to WithKey.
//
// This method will panic immediately if a circular dependency is detected.
func (v *ObjectRuleSet[T]) WithConditionalKey(key string, condition Conditional[T], ruleSet rules.RuleSet[any]) *ObjectRuleSet[T] {
	// Only check mapping if output type is a struct (not a map)
	if v.outputType.Kind() != reflect.Map {
		destKey, ok := v.mappingFor(key)
		if !ok {
			panic(fmt.Errorf("missing mapping for key: %s", key))
		}
		field, ok := v.outputType.FieldByName(destKey)
		if !ok {
			// Should never get here since the only way to make mappings is in the New method.
			// But better to be defensive.
			panic(fmt.Errorf("missing destination mapping for field: %s", destKey))
		}
		if !field.IsExported() {
			// Should also never get here since the only way to make mappings is in the New method
			// and New ignores unexported fields.
			panic(fmt.Errorf("field is not exported: %s", destKey))
		}
	}

	newRuleSet := v.withParent()
	newRuleSet.key = key
	newRuleSet.rule = ruleSet
	newRuleSet.condition = condition

	if condition != nil {
		if newRuleSet.refs == nil {
			newRuleSet.refs = newRefTracker()
		} else {
			newRuleSet.refs = newRuleSet.refs.Clone()
		}

		for _, dependsOn := range condition.Keys() {
			if err := newRuleSet.refs.Add(key, dependsOn); err != nil {
				panic(err)
			}
		}
	}

	return newRuleSet
}

// Deprecated: Key is deprecated and will be removed in v1.0.0. Use WithKey instead.
func (v *ObjectRuleSet[T]) Key(key string, ruleSet rules.RuleSet[any]) *ObjectRuleSet[T] {
	return v.WithKey(key, ruleSet)
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *ObjectRuleSet[T]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (v *ObjectRuleSet[T]) WithRequired() *ObjectRuleSet[T] {
	newRuleSet := v.withParent()
	newRuleSet.required = true
	newRuleSet.label = "WithRequired()"
	return newRuleSet
}

// Validate performs a validation of a RuleSet against a value and returns a value of the correct type or
// a ValidationErrorCollection.
func (v *ObjectRuleSet[T]) Validate(value any) (T, errors.ValidationErrorCollection) {
	return v.ValidateWithContext(value, context.Background())
}

// contextErrorToValidation takes a context error and returns a validation error.
func contextErrorToValidation(ctx context.Context) errors.ValidationError {
	switch ctx.Err() {
	case nil:
		return nil
	case context.DeadlineExceeded:
		return errors.Errorf(errors.CodeTimeout, ctx, "validation timed out before completing")
	case context.Canceled:
		return errors.Errorf(errors.CodeCancelled, ctx, "validation was cancelled")
	default:
		return errors.Errorf(errors.CodeInternal, ctx, "unknown context error: %v", ctx.Err())
	}
}

// wait blocks until either the context is cancelled or the wait group is done (all keys have been validated).
func wait(ctx context.Context, wg *sync.WaitGroup, errorsCh chan errors.ValidationErrorCollection, listenForCancelled bool) errors.ValidationErrorCollection {
	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	allErrors := errors.Collection()

	for {
		select {
		case err := <-errorsCh:
			allErrors = append(allErrors, err...)
		case <-ctx.Done():
			if listenForCancelled {
				wg.Wait()
				return append(allErrors, contextErrorToValidation(ctx))
			}
		case <-done:
			return allErrors
		}
	}
}

// ValidateWithContext performs a validation of a RuleSet against a value and returns a value of the correct type or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by validation rules and error formatting.
func (v *ObjectRuleSet[T]) ValidateWithContext(in any, ctx context.Context) (T, errors.ValidationErrorCollection) {
	done := func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}

	var out T

	var toMap bool

	if v.outputType.Kind() == reflect.Map {
		toMap = true
		out = reflect.MakeMap(v.outputType).Interface().(T)
	} else if v.ptr {
		out = reflect.New(v.outputType).Interface().(T)
	} else {
		out = reflect.New(v.outputType).Elem().Interface().(T)
	}

	// We can't use reflect.Set on a non-pointer struct so if the output is not a pointer
	// we want to make a pointer to work with.
	var outValue reflect.Value
	if v.ptr {
		outValue = reflect.Indirect(reflect.ValueOf(out))
	} else {
		outValue = reflect.Indirect(reflect.ValueOf(&out))
	}

	inValue := reflect.Indirect(reflect.ValueOf(in))
	inKind := inValue.Kind()

	fromMap := inKind == reflect.Map

	if !fromMap && inKind != reflect.Struct {
		return out, errors.Collection(
			errors.NewCoercionError(ctx, "object or map", inKind.String()),
		)
	}

	allErrors := errors.Collection()

	fieldMapping := v.fullMapping()

	// Only set if the input is a map and we don't allow unknown values
	var knownKeys map[string]bool
	if !v.allowUnknown && fromMap {
		knownKeys = make(map[string]bool)
	}

	// Create a table of how keys and a counter.
	// We need this because conditional keys cannot run.
	counters := newCounterSet()
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key != "" && currentRuleSet.rule != nil {
			counters.Increment(currentRuleSet.key)
		}
	}

	// Handle concurrency for the rule evaluation
	errorsCh := make(chan errors.ValidationErrorCollection)
	defer close(errorsCh)
	var outValueMutex sync.Mutex

	// Wait for all the rules to finish
	var wg sync.WaitGroup

	// Loop through the rule set and evaluate each one.
	// Run each rule set in a goroutine.
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key == "" || currentRuleSet.rule == nil {
			continue
		}

		key := currentRuleSet.key
		rule := currentRuleSet.rule

		var inFieldValue reflect.Value

		if fromMap {
			inFieldValue = inValue.MapIndex(reflect.ValueOf(key))

			if knownKeys != nil {
				knownKeys[key] = true
			}
		} else {
			inFieldValue = inValue.FieldByName(key)
		}

		subContext := rulecontext.WithPathString(ctx, key)

		if inFieldValue.Kind() == reflect.Invalid {
			if rule.Required() {
				allErrors = append(allErrors, errors.Errorf(errors.CodeRequired, subContext, "field is required"))
			}
			counters.Clear(key)

		} else {
			wg.Add(1)

			go func(key string, subContext context.Context, condition Conditional[T]) {
				defer wg.Done()
				counters.Lock(key)
				defer counters.Unlock(key)

				if done() {
					return
				}

				if condition != nil {
					keys := condition.Keys()
					counters.Wait(keys...)

					ok := func() bool {
						outValueMutex.Lock()
						defer outValueMutex.Unlock()
						_, err := condition.Evaluate(ctx, out)
						return err == nil
					}()

					if !ok {
						return
					}
				}

				val, errs := rule.ValidateWithContext(inFieldValue.Interface(), subContext)
				if errs != nil {
					errorsCh <- errs
					return
				}

				outValueMutex.Lock()
				defer outValueMutex.Unlock()

				if toMap {
					outValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
				} else {
					field := outValue.FieldByName(fieldMapping[key])
					if field.Kind() == reflect.Ptr {
						valPtr := reflect.New(reflect.TypeOf(val))
						valPtr.Elem().Set(reflect.ValueOf(val))
						field.Set(valPtr)
					} else {
						field.Set(reflect.ValueOf(val))
					}
				}
			}(key, subContext, currentRuleSet.condition)
		}
	}

	valErrs := wait(ctx, &wg, errorsCh, true)
	allErrors = append(allErrors, valErrs...)

	// Next apply object rules.
	// This must be done after the key rules because we want to make sure all values are cast first.
	var wg2 sync.WaitGroup

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.objRule != nil {
			if done() {
				break
			}

			wg2.Add(1)
			go func(objRule rules.Rule[T]) {
				outValueMutex.Lock()
				defer outValueMutex.Unlock()
				defer wg2.Done()

				if done() {
					return
				}

				newOutput, err := objRule.Evaluate(ctx, out)
				if err != nil {
					errorsCh <- err
				} else {
					out = newOutput
				}

			}(currentRuleSet.objRule)
		}
	}

	valErrs = wait(ctx, &wg2, errorsCh, !done())
	wg.Wait()
	allErrors = append(allErrors, valErrs...)

	if knownKeys != nil {
		for _, key := range inValue.MapKeys() {
			keyStr := key.String()
			_, ok := knownKeys[keyStr]
			if !ok {
				subContext := rulecontext.WithPathString(ctx, keyStr)
				allErrors = append(allErrors, errors.Errorf(errors.CodeUnexpected, subContext, "unexpected field"))
			}
		}
	}

	//	outValueMutex.Lock()
	//	defer outValueMutex.Unlock()

	if len(allErrors) != 0 {
		return out, allErrors
	} else {
		return out, nil
	}
}

// Evaluate performs a validation of a RuleSet against a value of the object type and returns an object value of the
// same type or a ValidationErrorCollection.
func (ruleSet *ObjectRuleSet[T]) Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection) {
	// We need to use reflection no matter what so the fact the input is already the right type doesn't help us
	return ruleSet.ValidateWithContext(value, ctx)
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the given object type.
//
// Use this when implementing custom rules.
func (v *ObjectRuleSet[T]) WithRule(rule rules.Rule[T]) *ObjectRuleSet[T] {
	newRuleSet := v.withParent()
	newRuleSet.objRule = rule
	return newRuleSet
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// for the given object type.
//
// Use this when implementing custom rules.
func (v *ObjectRuleSet[T]) WithRuleFunc(rule rules.RuleFunc[T]) *ObjectRuleSet[T] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the object RuleSet in any Any rule set
// which can then be used in nested validation.
func (v *ObjectRuleSet[T]) Any() rules.RuleSet[any] {
	return rules.WrapAny[T](v)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *ObjectRuleSet[T]) String() string {
	// Pass through mappings
	if ruleSet.mapping != "" {
		return ruleSet.parent.String()
	}

	label := ruleSet.label

	if label == "" {
		if ruleSet.rule != nil {
			label = fmt.Sprintf("WithKey(\"%s\", %s)", ruleSet.key, ruleSet.rule)
		} else if ruleSet.objRule != nil {
			label = ruleSet.objRule.String()
		}
	}

	if ruleSet.parent != nil {
		return ruleSet.parent.String() + "." + label
	}
	return label
}
