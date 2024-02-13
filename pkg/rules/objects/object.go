// Package objects provides a RuleSet implementation that can be used to validate object and map values.
//
// It implements standard rules and allows the developer to set a rule set to validate individual keys.
package objects

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

const annotation = "validate"

type objectRuleScope int

const (
	scopeKey          objectRuleScope = 0
	scopeUnknownKey   objectRuleScope = 1
	scopeUnknownValue objectRuleScope = 2
)

// Implementation of RuleSet for objects and maps.
type ObjectRuleSet[T any, TK comparable, TV any] struct {
	rules.NoConflict[T]
	allowUnknown bool
	key          TK
	rule         rules.RuleSet[TV]
	ruleScope    objectRuleScope
	objRule      rules.Rule[T]
	mapping      TK
	outputType   reflect.Type
	ptr          bool
	required     bool
	parent       *ObjectRuleSet[T, TK, TV]
	label        string
	condition    Conditional[T, TK]
	refs         *refTracker[TK]
	json         bool
}

// New returns a RuleSet that can be used to validate an object of an
// arbitrary data type.
//
// Using the "validate" annotation you can may input values to different
// properties of the object. This is useful for converting unstructured maps
// created from Json and converting to an object.
func New[T any]() *ObjectRuleSet[T, string, any] {
	var empty [0]T

	ruleSet := &ObjectRuleSet[T, string, any]{
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
		panic(fmt.Errorf("invalid output type for object rule set: %v", kind))
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

		ruleSet = &ObjectRuleSet[T, string, any]{
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
func NewObjectMap[T any]() *ObjectRuleSet[map[string]T, string, T] {
	var empty map[string]T

	return &ObjectRuleSet[map[string]T, string, T]{
		outputType: reflect.TypeOf(empty),
	}
}

// withParent is a helper function to assist in cloning object RuleSets.
func (v *ObjectRuleSet[T, TK, TV]) withParent() *ObjectRuleSet[T, TK, TV] {
	return &ObjectRuleSet[T, TK, TV]{
		allowUnknown: v.allowUnknown,
		required:     v.required,
		outputType:   v.outputType,
		ptr:          v.ptr,
		parent:       v,
		refs:         v.refs,
		json:         v.json,
	}
}

// WithUnknown returns a new RuleSet with the "unknown" flag set.
//
// By default if the validator fines an unknown key on a map it will return an error.
// Setting the unknown flag will allow keys that aren't defined to be present in the map.
// This is useful for parsing arbitrary Json where additional keys may be included.
func (v *ObjectRuleSet[T, TK, TV]) WithUnknown() *ObjectRuleSet[T, TK, TV] {
	if v.allowUnknown {
		return v
	}

	newRuleSet := v.withParent()
	newRuleSet.allowUnknown = true
	newRuleSet.label = "WithUnknown()"
	return newRuleSet
}

/*
// WithDynamicKey returns a new rule set that run evaluates the rule for any keys that
// match the key rule. If any rules match, the key is no longer considered "unknown" and
// will pass validation even if WithUnknown is not set.
//
// Dynamic key rules will be run for all keys that match the rule regardless of if they
// are already known keys.
//
// Note: Rules will be evaluates for both structs and maps. However, if you want to be
// able to access the dynamic values on structs you need to specific a receptor property
// for the values to be put into. Use WithDynamicKeyBucket instead.
func (v *ObjectRuleSet[T,TV]) WithUnknownKey(keyRule rules.RuleSet[string], valueRule rules.RuleSet[T]) *ObjectRuleSet[T,TV] {
{

}

*/

// WithUnknownKeyValue returns a new rule set with a validation function for unknown key values.
// Rules are cumulative. Unlike WithUnknownKey, all rules will always be run.
// If the unknown flag was not previously set, this will set it.
//
// The rules will not run if the key fails validation.
func (v *ObjectRuleSet[T, TK, TV]) WithUnknownKeyValue(rule rules.RuleSet[TV]) *ObjectRuleSet[T, TK, TV] {
	newRuleSet := v.withParent()
	newRuleSet.allowUnknown = true
	newRuleSet.rule = rule
	newRuleSet.ruleScope = scopeUnknownValue
	return newRuleSet
}

// fullMapping is a helper function that returns the full object field mappings as a map.
func (v *ObjectRuleSet[T, TK, TV]) fullMapping() map[TK]TK {
	mapping := make(map[TK]TK)
	empty := new(TK)

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key != *empty && currentRuleSet.mapping != *empty {
			mapping[currentRuleSet.key] = currentRuleSet.mapping
		}
	}
	return mapping
}

// mappingFor is a helper function that returns the key mapping given a specific key.
func (v *ObjectRuleSet[T, TK, TV]) mappingFor(key TK) (TK, bool) {
	empty := new(TK)

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key == key && currentRuleSet.mapping != *empty {
			return currentRuleSet.mapping, true
		}
	}
	return *empty, false
}

// WithKey returns a new RuleSet with a validation rule for the specified key.
//
// If more than one call is made with the same key than all will be evaluated. However, the order
// in which they are run is not guaranteed.
//
// Multiple rule sets may run in parallel but only one will run a time for each key.
func (v *ObjectRuleSet[T, TK, TV]) WithKey(key TK, ruleSet rules.RuleSet[TV]) *ObjectRuleSet[T, TK, TV] {
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
func (v *ObjectRuleSet[T, TK, TV]) Keys() []TK {
	mapping := make(map[TK]bool)
	keys := make([]TK, 0)
	empty := new(TK)

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key != *empty && currentRuleSet.rule != nil {
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
func (v *ObjectRuleSet[T, TK, TV]) WithConditionalKey(key TK, condition Conditional[T, TK], ruleSet rules.RuleSet[TV]) *ObjectRuleSet[T, TK, TV] {
	newRuleSet := v.withParent()

	// Only check mapping if output type is a struct (not a map)
	if v.outputType.Kind() != reflect.Map {
		destKey, ok := v.mappingFor(key)
		if !ok {
			panic(fmt.Errorf("missing mapping for key: %s", toPath(key)))
		}

		// Struct targets always have string as the key
		destKeyStr := any(destKey).(string)

		field, ok := v.outputType.FieldByName(destKeyStr)
		if !ok {
			// Should never get here since the only way to make mappings is in the New method.
			// But better to be defensive.
			panic(fmt.Errorf("missing destination mapping for field: %s", toPath(destKey)))
		}
		if !field.IsExported() {
			// Should also never get here since the only way to make mappings is in the New method
			// and New ignores unexported fields.
			panic(fmt.Errorf("field is not exported: %s", toPath(destKey)))
		}
		newRuleSet.mapping = destKey
	}

	newRuleSet.key = key
	newRuleSet.rule = ruleSet
	newRuleSet.ruleScope = scopeKey
	newRuleSet.condition = condition

	if condition != nil {
		if newRuleSet.refs == nil {
			newRuleSet.refs = newRefTracker[TK]()
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
func (v *ObjectRuleSet[T, TK, TV]) Key(key TK, ruleSet rules.RuleSet[TV]) *ObjectRuleSet[T, TK, TV] {
	return v.WithKey(key, ruleSet)
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *ObjectRuleSet[T, TK, TV]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (v *ObjectRuleSet[T, TK, TV]) WithRequired() *ObjectRuleSet[T, TK, TV] {
	if v.required {
		return v
	}

	newRuleSet := v.withParent()
	newRuleSet.required = true
	newRuleSet.label = "WithRequired()"
	return newRuleSet
}

// Validate performs a validation of a RuleSet against a value and returns a value of the correct type or
// a ValidationErrorCollection.
func (v *ObjectRuleSet[T, TK, TV]) Validate(value any) (T, errors.ValidationErrorCollection) {
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

// isDone checks if the context is done and returns a bool.
func done(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// evaluateKeyRule evaluates a single key rule.
// Note that this function is meant to be called on the rule set that contains the rule.
func (ruleSet *ObjectRuleSet[T, TK, TV]) evaluateKeyRule(ctx context.Context, out *T, wg *sync.WaitGroup, outValueMutex *sync.Mutex, errorsCh chan errors.ValidationErrorCollection, inFieldValue reflect.Value, s setter[TK], counters *counterSet[TK]) {
	defer wg.Done()
	counters.Lock(ruleSet.key)
	defer counters.Unlock(ruleSet.key)

	if done(ctx) {
		return
	}

	if ruleSet.condition != nil {
		keys := ruleSet.condition.Keys()
		counters.Wait(keys...)

		ok := func() bool {
			outValueMutex.Lock()
			defer outValueMutex.Unlock()
			_, err := ruleSet.condition.Evaluate(ctx, *out)
			return err == nil
		}()

		if !ok {
			return
		}
	}

	if inFieldValue.Kind() == reflect.Invalid {
		if ruleSet.rule.Required() {
			errorsCh <- errors.Collection(
				errors.Errorf(errors.CodeRequired, ctx, "field is required"),
			)
		}
		return
	}

	val, errs := ruleSet.rule.ValidateWithContext(inFieldValue.Interface(), ctx)
	if errs != nil {
		errorsCh <- errs
		return
	}

	outValueMutex.Lock()
	defer outValueMutex.Unlock()

	s.Set(ruleSet.key, val)
}

// evaluateKeyRules evaluates the rules for each key.
func (v *ObjectRuleSet[T, TK, TV]) evaluateKeyRules(ctx context.Context, out *T, inValue reflect.Value, s setter[TK], fromMap, fromSame bool) errors.ValidationErrorCollection {
	allErrors := errors.Collection()
	empty := new(TK)

	// Tracks which keys are known so we can create errors for unknown keys.
	knownKeys := newKnownKeys[TK]((!v.allowUnknown || s.Map()) && fromMap)

	// Create a table of how keys and a counter.
	// We need this because conditional keys cannot run.
	counters := newCounterSet[TK]()
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key != *empty && currentRuleSet.rule != nil {
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
	var unknownValueRuleSets []rules.RuleSet[TV]

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {

		if currentRuleSet.rule == nil {
			continue
		}

		switch currentRuleSet.ruleScope {
		case scopeUnknownValue:
			unknownValueRuleSets = append(unknownValueRuleSets, currentRuleSet.rule)
			continue
		}

		key := currentRuleSet.key

		var inFieldValue reflect.Value

		if fromMap {
			inFieldValue = inValue.MapIndex(reflect.ValueOf(key))
			knownKeys.Add(key)
		} else if fromSame {
			// From same always has string keys so we can cast it
			keyStr := any(currentRuleSet.mapping).(string)
			inFieldValue = inValue.FieldByName(keyStr)
		} else {
			// We know this isn't a map so the only option for a key is a string
			keyStr := any(key).(string)
			inFieldValue = inValue.FieldByName(keyStr)
		}

		subContext := rulecontext.WithPathString(ctx, toPath(key))

		wg.Add(1)
		go currentRuleSet.evaluateKeyRule(subContext, out, &wg, &outValueMutex, errorsCh, inFieldValue, s, counters)
	}

	// Unknown fields are not concurrent for now
	ruleErrors := wait(ctx, &wg, errorsCh, true)

	if !v.allowUnknown {
		knownKeyErrors := knownKeys.Check(ctx, inValue)
		allErrors = append(allErrors, knownKeyErrors...)
	} else if fromMap && s.Map() {
		for _, key := range knownKeys.Unknown(inValue) {
			validKey := true

			subContext := rulecontext.WithPathString(ctx, toPath(key))

			if validKey {
				val := inValue.MapIndex(reflect.ValueOf(key)).Interface()
				validValue := true
				var verr errors.ValidationErrorCollection

				for i := range unknownValueRuleSets {
					val, verr = unknownValueRuleSets[i].ValidateWithContext(val, subContext)
					if len(verr) > 0 {
						allErrors = append(allErrors, verr...)
						validValue = false
					}
				}

				if validValue {
					s.Set(key, val)
				}
			}
		}
	}

	return append(allErrors, ruleErrors...)
}

// evaluateObjectRules evaluates the object rules.
func (v *ObjectRuleSet[T, TK, TV]) evaluateObjectRules(ctx context.Context, out *T) (T, errors.ValidationErrorCollection) {
	final := *out

	var wg sync.WaitGroup
	var outValueMutex sync.Mutex
	errorsCh := make(chan errors.ValidationErrorCollection)
	defer close(errorsCh)

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.objRule != nil {
			if done(ctx) {
				break
			}

			wg.Add(1)
			go func(objRule rules.Rule[T]) {
				outValueMutex.Lock()
				defer outValueMutex.Unlock()
				defer wg.Done()

				if done(ctx) {
					return
				}

				newOutput, err := objRule.Evaluate(ctx, *out)
				if err != nil {
					errorsCh <- err
				} else {
					final = newOutput
				}

			}(currentRuleSet.objRule)
		}
	}

	return final, wait(ctx, &wg, errorsCh, !done(ctx))
}

// newSetter creates a new setter for the rule set
func (ruleSet *ObjectRuleSet[T, TK, TV]) newSetter(outValue reflect.Value) setter[TK] {
	if ruleSet.outputType.Kind() == reflect.Map {
		return &mapSetter[TK]{
			out: outValue,
		}
	}

	return &structSetter[TK]{
		out:     outValue,
		mapping: ruleSet.fullMapping(),
	}
}

// ValidateWithContext performs a validation of a RuleSet against a value and returns a value of the correct type or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by validation rules and error formatting.
func (v *ObjectRuleSet[T, TK, TV]) ValidateWithContext(in any, ctx context.Context) (T, errors.ValidationErrorCollection) {
	var out T

	if v.outputType.Kind() == reflect.Map {
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

	s := v.newSetter(outValue)

	inValue := reflect.Indirect(reflect.ValueOf(in))
	inKind := inValue.Kind()

	// Convert strings to Json
	if v.json {
		var result map[string]interface{}
		coerced := false
		attempted := false

		if inKind == reflect.String {
			attempted = true
			if err := json.Unmarshal([]byte(inValue.String()), &result); err == nil {
				coerced = true
			}

		} else if inKind == reflect.Slice && inValue.Type().Elem().Kind() == reflect.Uint8 {
			attempted = true

			// Note: this also matches types that have []byte as their underlying type, such as json.RawMessage
			if err := json.Unmarshal(inValue.Bytes(), &result); err == nil {
				coerced = true
			}
		}

		if !coerced && attempted {
			return out, errors.Collection(
				errors.NewCoercionError(ctx, "object, map, or Json string", inKind.String()),
			)
		}

		inValue = reflect.ValueOf(result)
		inKind = inValue.Kind()
	}

	fromMap := inKind == reflect.Map
	fromSame := !fromMap && inValue.Type() == v.outputType

	if !fromMap && inKind != reflect.Struct {
		return out, errors.Collection(
			errors.NewCoercionError(ctx, "object or map", inKind.String()),
		)
	}

	allErrors := errors.Collection()

	keyErrs := v.evaluateKeyRules(ctx, &out, inValue, s, fromMap, fromSame)
	allErrors = append(allErrors, keyErrs...)

	// This must be done after the key rules because we want to make sure all values are cast first.
	out, valErrs := v.evaluateObjectRules(ctx, &out)
	allErrors = append(allErrors, valErrs...)

	if len(allErrors) != 0 {
		return out, allErrors
	}
	return out, nil
}

// Evaluate performs a validation of a RuleSet against a value of the object type and returns an object value of the
// same type or a ValidationErrorCollection.
func (ruleSet *ObjectRuleSet[T, TK, TV]) Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection) {
	// We need to use reflection no matter what so the fact the input is already the right type doesn't help us
	return ruleSet.ValidateWithContext(value, ctx)
}

// WithJson allows the input to be a Json encoded string.
func (v *ObjectRuleSet[T, TK, TV]) WithJson() *ObjectRuleSet[T, TK, TV] {
	if v.json {
		return v
	}

	newRuleSet := v.withParent()
	newRuleSet.json = true
	return newRuleSet
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the given object type.
//
// Use this when implementing custom rules.
func (v *ObjectRuleSet[T, TK, TV]) WithRule(rule rules.Rule[T]) *ObjectRuleSet[T, TK, TV] {
	newRuleSet := v.withParent()
	newRuleSet.objRule = rule
	return newRuleSet
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the Rule function
// for the given object type.
//
// Use this when implementing custom rules.
func (v *ObjectRuleSet[T, TK, TV]) WithRuleFunc(rule rules.RuleFunc[T]) *ObjectRuleSet[T, TK, TV] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the object RuleSet in any Any rule set
// which can then be used in nested validation.
func (v *ObjectRuleSet[T, TK, TV]) Any() rules.RuleSet[any] {
	return rules.WrapAny[T](v)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *ObjectRuleSet[T, TK, TV]) String() string {
	// Pass through mappings with no rules
	empty := new(TK)

	if ruleSet.mapping != *empty && ruleSet.rule == nil {
		return ruleSet.parent.String()
	}

	label := ruleSet.label

	if label == "" {
		if ruleSet.rule != nil {
			if ruleSet.condition != nil {
				label = fmt.Sprintf("WithConditionalKey(\"%s\", %s, %s)", toPath(ruleSet.key), ruleSet.condition, ruleSet.rule)
			} else {
				label = fmt.Sprintf("WithKey(\"%s\", %s)", toPath(ruleSet.key), ruleSet.rule)
			}
		} else if ruleSet.objRule != nil {
			label = ruleSet.objRule.String()
		}
	}

	if ruleSet.parent != nil {
		return ruleSet.parent.String() + "." + label
	}
	return label
}
