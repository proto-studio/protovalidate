package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

const annotation = "validate"

// Implementation of RuleSet for objects and maps.
type ObjectRuleSet[T any, TK comparable, TV any] struct {
	NoConflict[T]
	allowUnknown bool
	key          Rule[TK]
	rule         RuleSet[TV]
	objRule      Rule[T]
	mapping      TK
	outputType   reflect.Type
	ptr          bool
	required     bool
	withNil      bool
	parent       *ObjectRuleSet[T, TK, TV]
	label        string
	condition    Conditional[T, TK]
	refs         *refTracker[TK]
	bucket       TK
	json         bool
	errorConfig  *errors.ErrorConfig
}

// Struct returns a RuleSet that can be used to validate a struct of an
// arbitrary type.
//
// Struct uses the "validate" annotation to map input values to different
// properties of the object. This is useful for converting unstructured maps
// created from JSON and converting to an object.
func Struct[T any]() *ObjectRuleSet[T, string, any] {
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
			key:        Constant(key),
			mapping:    field.Name,
			outputType: ruleSet.outputType,
			ptr:        ruleSet.ptr,
		}

		mapped[key] = true
	}

	return ruleSet
}

// StringMap returns a new RuleSet that can be used to validate maps with strings as the
// keys and the specified data type (which can be "any") as the values.
func StringMap[T any]() *ObjectRuleSet[map[string]T, string, T] {
	var empty map[string]T

	return &ObjectRuleSet[map[string]T, string, T]{
		outputType: reflect.TypeOf(empty),
	}
}

// NewObjectMap returns a new RuleSet that can be used to validate maps with strings as the
// keys and the specified data type (which can be "any") as the values.
func Map[TK comparable, TV any]() *ObjectRuleSet[map[TK]TV, TK, TV] {
	var empty map[TK]TV

	return &ObjectRuleSet[map[TK]TV, TK, TV]{
		outputType: reflect.TypeOf(empty),
	}
}

// objectCloneOption is a functional option for cloning ObjectRuleSet.
type objectCloneOption[T any, TK comparable, TV any] func(*ObjectRuleSet[T, TK, TV])

// clone returns a shallow copy of the rule set with parent set to the current instance.
func (v *ObjectRuleSet[T, TK, TV]) clone(options ...objectCloneOption[T, TK, TV]) *ObjectRuleSet[T, TK, TV] {
	newRuleSet := &ObjectRuleSet[T, TK, TV]{
		allowUnknown: v.allowUnknown,
		required:     v.required,
		withNil:      v.withNil,
		outputType:   v.outputType,
		ptr:          v.ptr,
		parent:       v,
		refs:         v.refs,
		json:         v.json,
		errorConfig:  v.errorConfig,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

func objectWithLabel[T any, TK comparable, TV any](label string) objectCloneOption[T, TK, TV] {
	return func(rs *ObjectRuleSet[T, TK, TV]) { rs.label = label }
}

func objectWithErrorConfig[T any, TK comparable, TV any](config *errors.ErrorConfig) objectCloneOption[T, TK, TV] {
	return func(rs *ObjectRuleSet[T, TK, TV]) { rs.errorConfig = config }
}

// WithUnknown returns a new RuleSet that allows unknown keys in maps and objects.
//
// By default, if the validator finds an unknown key on a map it will return an error.
// WithUnknown allows keys that aren't defined to be present in the map.
// This is useful for parsing arbitrary JSON where additional keys may be included.
func (v *ObjectRuleSet[T, TK, TV]) WithUnknown() *ObjectRuleSet[T, TK, TV] {
	if v.allowUnknown {
		return v
	}

	newRuleSet := v.clone(objectWithLabel[T, TK, TV]("WithUnknown()"))
	newRuleSet.allowUnknown = true
	return newRuleSet
}

// fullMapping is a helper function that returns the full object field mappings as a map.
func (v *ObjectRuleSet[T, TK, TV]) fullMapping() map[TK]TK {
	mapping := make(map[TK]TK)
	empty := new(TK)

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key != nil && currentRuleSet.mapping != *empty {
			mapping[currentRuleSet.key.(*ConstantRuleSet[TK]).Value()] = currentRuleSet.mapping
		}
	}
	return mapping
}

// mappingFor is a helper function that returns the key mapping given a specific key.
func (v *ObjectRuleSet[T, TK, TV]) mappingFor(ctx context.Context, key TK) (TK, bool) {
	var empty TK

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key != nil && currentRuleSet.key.Evaluate(ctx, key) == nil && currentRuleSet.mapping != empty {
			return currentRuleSet.mapping, true
		}
	}
	return empty, false
}

// WithKey returns a new RuleSet with a validation rule for the specified key.
//
// If more than one call is made with the same key then all will be evaluated. However, the order
// in which they are run is not guaranteed.
//
// Multiple rule sets may run in parallel but only one will run at a time for each key since rule sets
// can return a mutated value.
func (v *ObjectRuleSet[T, TK, TV]) WithKey(key TK, ruleSet RuleSet[TV]) *ObjectRuleSet[T, TK, TV] {
	return v.WithConditionalKey(key, nil, ruleSet)
}

// WithDynamicKey returns a new RuleSet with a validation rule for any key that matches the key rule.
// WithDynamicKey runs dynamic rules even if they match a key that has an already defined rule. Mappings are not applied
// to dynamic keys.
//
// If more than one call is made with the same key or overlapping dynamic rules, then all will be evaluated.
// However, the order in which they are run is not guaranteed.
//
// Multiple rule sets may run in parallel but only one will run at a time for each key since rule sets
// can return a mutated value. This is true even for constant value keys and other dynamic rules if the
// patterns overlap.
//
// If a key matches the key rules of any unconditional dynamic rule it will no longer be considered an "unknown" key.
//
// With maps, the dynamic keys are directly set on the output map. For structs you must set a dynamic key
// bucket using WithDynamicBucket.
func (v *ObjectRuleSet[T, TK, TV]) WithDynamicKey(keyRule Rule[TK], ruleSet RuleSet[TV]) *ObjectRuleSet[T, TK, TV] {
	var empty TK

	return v.withKeyHelper(
		keyRule,
		empty,
		nil,
		ruleSet,
	)
}

// WithDynamicBucket tells the Rule Set to put matching keys into specific buckets. A bucket is expected to be a
// map with the key type (string for structs targets or variable for map) and a value type that matches the expected
// value.
//
// To avoid runtime errors it is usually best to also add a validation rule for the key using WithDynamic key to
// ensure the value is the correct type.
//
// This method is designed for unknown and dynamic keys only. If you have any explicit rules for your key, it will not
// be put into the dynamic bucket.
//
// If a key matches the dynamic bucket key rules then it will no longer be considered "unknown" and will not trigger an
// unknown key error. You are encouraged to add additional validation rules for the values.
//
// If a key belongs to more than one bucket it will be included in all of them.
//
// For structs:
//
//	When WithDynamicBucket is called this function will panic if the bucket property does not exist on the struct or
//	bucket property is not a map.
//	The value of the property will be nil until at least one key matches.
//
// For maps:
//
//	Running the rule set will panic if the value type is not "any" since any other type of value will not allow the bucket
//	map to be created.
//	The value of the bucket key in the map will not exist unless at least one key matches.
func (v *ObjectRuleSet[T, TK, TV]) WithDynamicBucket(keyRule Rule[TK], bucket TK) *ObjectRuleSet[T, TK, TV] {
	return v.WithConditionalDynamicBucket(keyRule, nil, bucket)
}

// WithConditionalDynamicBucket behaves like WithDynamicBucket except the value is not sorted into the bucket unless the
// condition is met.
//
// If the only dynamic rules are conditional, the key will be considered unknown if no conditions match.
func (v *ObjectRuleSet[T, TK, TV]) WithConditionalDynamicBucket(keyRule Rule[TK], condition Conditional[T, TK], bucket TK) *ObjectRuleSet[T, TK, TV] {
	newRuleSet := v.clone()

	newRuleSet.key = keyRule
	newRuleSet.condition = condition
	newRuleSet.bucket = bucket

	return newRuleSet
}

// Keys returns the keys names that have rule sets associated with them.
// This will not return keys that don't have rule sets (even if they do have a mapping).
//
// It also will not return keys that are referenced WithRule or WithRuleFund. To get around this
// you may want to consider moving your rule set to WithKey or putting a simple permissive validator
// inside WithKey.
//
// The results are not sorted. You should not depend on the order of the results.
func (v *ObjectRuleSet[T, TK, TV]) KeyRules() []Rule[TK] {
	// Don't return identical keys more than once
	mapping := make(map[Rule[TK]]bool)
	keys := make([]Rule[TK], 0)

	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key != nil && currentRuleSet.rule != nil {
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
// also been evaluated. This includes conditional  So if X is also dependent on Z then Z will also need to be complete.
//
// If one or more of the fields has an error then the conditional rule will not be run.
//
// WithRule and WithRuleFunc are both evaluated after any keys or conditional keys. Because of this, it is not possible to
// have a conditional key that is dependent on data that is modified in those
//
// If nil is passed in as the conditional then this method behaves identical to WithKey.
//
// This method will panic immediately if a circular dependency is detected.
func (v *ObjectRuleSet[T, TK, TV]) WithConditionalKey(key TK, condition Conditional[T, TK], ruleSet RuleSet[TV]) *ObjectRuleSet[T, TK, TV] {
	var destKey TK

	// Only check mapping if output type is a struct (not a map)
	if v.outputType.Kind() != reflect.Map {
		var ok bool
		destKey, ok = v.mappingFor(context.Background(), key)
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
	}

	return v.withKeyHelper(
		Constant(key),
		destKey,
		condition,
		ruleSet,
	)
}

// withKeyHelper returns a new rule set with the appropriate keys, conditions, and mappings set.
func (v *ObjectRuleSet[T, TK, TV]) withKeyHelper(key Rule[TK], destKey TK, condition Conditional[T, TK], ruleSet RuleSet[TV]) *ObjectRuleSet[T, TK, TV] {
	newRuleSet := v.clone()

	newRuleSet.mapping = destKey
	newRuleSet.key = key
	newRuleSet.rule = ruleSet
	newRuleSet.condition = condition

	if condition != nil {
		if newRuleSet.refs == nil {
			newRuleSet.refs = newRefTracker[TK]()
		} else {
			newRuleSet.refs = newRuleSet.refs.Clone()
		}

		for _, dependsOn := range condition.KeyRules() {
			if err := newRuleSet.refs.Add(newRuleSet.key, dependsOn); err != nil {
				panic(err)
			}
		}
	}

	return newRuleSet
}

// Deprecated: Key is deprecated and will be removed in v1.0.0. Use WithKey instead.
func (v *ObjectRuleSet[T, TK, TV]) Key(key TK, ruleSet RuleSet[TV]) *ObjectRuleSet[T, TK, TV] {
	return v.WithKey(key, ruleSet)
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (v *ObjectRuleSet[T, TK, TV]) Required() bool {
	return v.required
}

// WithRequired returns a new child rule set that requires the value to be present when nested in an object.
// When a required field is missing from the input, validation fails with an error.
func (v *ObjectRuleSet[T, TK, TV]) WithRequired() *ObjectRuleSet[T, TK, TV] {
	if v.required {
		return v
	}

	newRuleSet := v.clone()
	newRuleSet.required = true
	newRuleSet.label = "WithRequired()"
	return newRuleSet
}

// WithNil returns a new child rule set that allows nil input values.
// When nil input is provided, validation passes and the output is set to nil (if the output type supports nil values).
// By default, nil input values return a CodeNull error.
func (v *ObjectRuleSet[T, TK, TV]) WithNil() *ObjectRuleSet[T, TK, TV] {
	newRuleSet := v.clone()
	newRuleSet.withNil = true
	newRuleSet.label = "WithNil()"
	return newRuleSet
}

// contextErrorToValidation takes a context error and returns a validation error.
func contextErrorToValidation(ctx context.Context) errors.ValidationError {
	switch ctx.Err() {
	case nil:
		return nil
	case context.DeadlineExceeded:
		return errors.Error(errors.CodeTimeout, ctx)
	case context.Canceled:
		return errors.Error(errors.CodeCancelled, ctx)
	default:
		return errors.Error(errors.CodeInternal, ctx)
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

// done checks if the context is done and returns a bool.
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
func (ruleSet *ObjectRuleSet[T, TK, TV]) evaluateKeyRule(ctx context.Context, out *T, wg *sync.WaitGroup, outValueMutex *sync.Mutex, errorsCh chan errors.ValidationErrorCollection, key TK, inFieldValue reflect.Value, s setter[TK], counters *counterSet[TK], dynamicBuckets []*ObjectRuleSet[T, TK, TV]) {
	defer wg.Done()
	counters.Lock(key)
	defer counters.Unlock(key)

	// Don't keep evaluating if the context has been canceled.
	if done(ctx) {
		return
	}

	// Exit early if the condition is not met.
	if ruleSet.condition != nil {
		keys := ruleSet.condition.KeyRules()
		counters.Wait(keys...)

		ok := func() bool {
			outValueMutex.Lock()
			defer outValueMutex.Unlock()
			return ruleSet.condition.Evaluate(ctx, *out) == nil
		}()

		if !ok {
			return
		}
	}

	if inFieldValue.Kind() == reflect.Invalid {
		if ruleSet.rule.Required() {
			errorsCh <- errors.Collection(
				errors.Error(errors.CodeRequired, ctx),
			)
		}
		return
	}

	var val TV
	errs := ruleSet.rule.Apply(ctx, inFieldValue.Interface(), &val)
	if errs != nil {
		errorsCh <- errs
		return
	}

	outValueMutex.Lock()
	defer outValueMutex.Unlock()

	bucketMatched := false
	for _, bucketRuleSet := range dynamicBuckets {
		if bucketRuleSet.key.Evaluate(ctx, key) == nil && (bucketRuleSet.condition == nil || bucketRuleSet.condition.Evaluate(ctx, *out) == nil) {
			s.SetBucket(bucketRuleSet.bucket, key, val)
			bucketMatched = true
		}
	}

	if !bucketMatched {
		s.Set(key, val)
	}
}

// keyValue is a helper function that returns the name of a key for use in mapping and conditions
func (v *ObjectRuleSet[T, TK, TV]) keyValue(key TK, currentRuleSet *ObjectRuleSet[T, TK, TV], inValue reflect.Value, fromMap, fromSame bool) reflect.Value {
	var inFieldValue reflect.Value

	if fromMap {
		inFieldValue = inValue.MapIndex(reflect.ValueOf(key))
	} else if fromSame {
		// From same always has string keys since only structs would get this far so we can cast it.
		keyStr := any(currentRuleSet.mapping).(string)
		inFieldValue = inValue.FieldByName(keyStr)
	} else {
		// We know this isn't a map so the only option for a key is a string
		keyStr := any(key).(string)
		inFieldValue = inValue.FieldByName(keyStr)
	}

	return inFieldValue
}

// evaluateKeyRules evaluates the rules for each key and called evaluateKeyRule.
func (v *ObjectRuleSet[T, TK, TV]) evaluateKeyRules(ctx context.Context, out *T, inValue reflect.Value, s setter[TK], fromMap, fromSame bool) errors.ValidationErrorCollection {
	allErrors := errors.Collection()
	var emptyKey TK

	// Tracks which keys are known so we can create errors for unknown keys.
	knownKeys := newKnownKeys[TK]((!v.allowUnknown || s.Map()) && fromMap)

	// Add each key to the counter.
	// We need this because conditional keys cannot run until all rule sets are run since rule sets are able
	// to mutate values.
	// For dynamic keys we must increment for all matching keys.
	counters := newCounterSet[TK]()
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.key != nil && currentRuleSet.rule != nil {
			if c, ok := currentRuleSet.key.(*ConstantRuleSet[TK]); ok {
				counters.Increment(c.Value())
			} else if fromMap {
				// Dynamic keys only make sense if the source is a map.
				for _, mapKeyValue := range inValue.MapKeys() {
					key, ok := mapKeyValue.Interface().(TK)

					if ok && currentRuleSet.key.Evaluate(ctx, key) == nil {
						counters.Increment(key)
					}
				}
			}
		}
	}

	// Handle concurrency for the rule evaluation
	errorsCh := make(chan errors.ValidationErrorCollection)
	defer close(errorsCh)
	var outValueMutex sync.Mutex

	// Pre caching a list of dynamic buckets lets us avoid extra loops.
	// This method is faster in all cases where there is at least one bucket and the input has dynamic values
	dynamicBuckets := make([]*ObjectRuleSet[T, TK, TV], 0)
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.bucket != emptyKey {
			dynamicBuckets = append(dynamicBuckets, currentRuleSet)
		}
	}

	// Wait for all the rules to finish
	var wg sync.WaitGroup

	// Loop through all the rule sets and evaluate the rules
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.rule == nil {
			continue
		}

		if c, ok := currentRuleSet.key.(*ConstantRuleSet[TK]); ok {
			key := c.Value()
			inFieldValue := v.keyValue(key, currentRuleSet, inValue, fromMap, fromSame)
			knownKeys.Add(key)
			subContext := rulecontext.WithPathString(ctx, toPath(key))
			wg.Add(1)
			go currentRuleSet.evaluateKeyRule(subContext, out, &wg, &outValueMutex, errorsCh, key, inFieldValue, s, counters, nil)

		} else if fromMap {
			// Dynamic keys only make sense if the source is a map.
			for _, mapKeyValue := range inValue.MapKeys() {
				key, ok := mapKeyValue.Interface().(TK)

				if ok && currentRuleSet.key.Evaluate(ctx, key) == nil {
					inFieldValue := v.keyValue(key, currentRuleSet, inValue, fromMap, fromSame)
					subContext := rulecontext.WithPathString(ctx, toPath(key))
					knownKeys.Add(key)
					wg.Add(1)
					go currentRuleSet.evaluateKeyRule(subContext, out, &wg, &outValueMutex, errorsCh, key, inFieldValue, s, counters, dynamicBuckets)
				}
			}
		}
	}

	// Unknown fields are not concurrent for now so we need to wait for all rule evaluations to finish
	ruleErrors := wait(ctx, &wg, errorsCh, true)

	// Throw all applicable unknown keys into dynamic buckets.
	// Keys in dynamic buckets should not trigger an unknown key error.
	if len(dynamicBuckets) > 0 {
		unk := knownKeys.Unknown(inValue)
		for _, key := range unk {
			for _, bucketRuleSet := range dynamicBuckets {
				inFieldValue := v.keyValue(key, bucketRuleSet, inValue, fromMap, fromSame)

				if bucketRuleSet.key.Evaluate(ctx, key) == nil && (bucketRuleSet.condition == nil || bucketRuleSet.condition.Evaluate(ctx, *out) == nil) {
					knownKeys.Add(key)
					s.SetBucket(bucketRuleSet.bucket, key, inFieldValue.Interface())
				}
			}
		}
	}

	// Check for unknown values
	if !v.allowUnknown {
		// If allowUnknown is not set we want to error for each unknown value
		knownKeyErrors := knownKeys.Check(ctx, inValue)
		allErrors = append(allErrors, knownKeyErrors...)
	} else if fromMap && s.Map() {
		// If allowUnknown is set and the output is a map we want to assign each key to the map output.
		for _, key := range knownKeys.Unknown(inValue) {
			s.Set(key, inValue.MapIndex(reflect.ValueOf(key)).Interface())
		}
	}

	return append(allErrors, ruleErrors...)
}

// evaluateObjectRules evaluates the object
func (v *ObjectRuleSet[T, TK, TV]) evaluateObjectRules(ctx context.Context, out *T) errors.ValidationErrorCollection {
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
			go func(objRule Rule[T]) {
				outValueMutex.Lock()
				defer outValueMutex.Unlock()
				defer wg.Done()

				if done(ctx) {
					return
				}

				if err := objRule.Evaluate(ctx, *out); err != nil {
					errorsCh <- err
				}

			}(currentRuleSet.objRule)
		}
	}

	return wait(ctx, &wg, errorsCh, !done(ctx))
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

// Apply performs validation of a RuleSet against a value and assigns the result to the output parameter.
// Apply returns a ValidationErrorCollection if any validation errors occur.
func (v *ObjectRuleSet[T, TK, TV]) Apply(ctx context.Context, value any, output any) errors.ValidationErrorCollection {
	// Add error config to context for error customization
	ctx = errors.WithErrorConfig(ctx, v.errorConfig)

	// Check if withNil is enabled and value is nil
	if handled, err := util.TrySetNilIfAllowed(ctx, v.withNil, value, output); handled {
		return err
	}

	// Ensure output is a non-nil pointer
	rv := reflect.ValueOf(output)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "internal error", "output must be a non-nil pointer",
		))
	}

	// If this is true we need to assign the output at the end of the Apply since we can't assign it directly initially.
	assignLater := false

	var out *T

	// If output is the correct type, we use the pointer, otherwise we check if it can be assigned
	// so we can assign it later. We need an pointer to the correct output type regardless of the actual type of "output"
	// since the rules are strongly typed.

	elem := rv.Elem()

	if elem.Type() == v.outputType {
		// The output directly points to the type.
		// At this point we already know output is non-nil since we check earlier.
		if v.outputType.Kind() == reflect.Map && elem.IsNil() {
			elem.Set(reflect.MakeMap(v.outputType))
			out = output.(*T)
		} else if v.ptr {
			x := output.(T)
			out = &x
		} else {
			out = output.(*T)
		}

	} else if elem.Type() == reflect.PointerTo(v.outputType) {
		// Output is a pointer to the correct type (more specifically a double pointer).
		// This can happen a lot with generics because you are often taking a reference to &T and
		// T is already a pointer.
		// However, this can happen when T is not already a pointer as well by doing a double reference
		// on output so we need to handle both.

		if elem.IsNil() {
			out = new(T)
		} else if v.ptr {
			out = output.(*T)
		} else {
			tmp := *output.(**T)
			out = tmp
		}

		if v.ptr {
			indirectOutValue := reflect.Indirect(reflect.ValueOf(out))
			if indirectOutValue.IsNil() {
				// The pointer points to a pointer with a nil value so we need to initialize that too.
				indirectOutValue.Set(reflect.New(v.outputType))
				elem.Set(reflect.ValueOf(*out))
			}
		} else {
			elem.Set(reflect.ValueOf(out))
		}

	} else if elem.Kind() == reflect.Interface {
		// We're pointing to a nil interface{}
		// We can't set up the pointer now so we'll need to deal with it later
		if !reflect.ValueOf(out).Type().AssignableTo(elem.Type()) {
			return errors.Collection(errors.Errorf(errors.CodeInternal, ctx, "internal error", "cannot assign %T to %T", out, output))
		}

		assignLater = true
		out = new(T)

		outElem := reflect.ValueOf(out).Elem()
		if (outElem.Kind() == reflect.Pointer || outElem.Kind() == reflect.Map) && outElem.IsNil() {
			if v.outputType.Kind() == reflect.Map {
				newMap := reflect.MakeMap(v.outputType)
				elem.Set(newMap)
				reflect.ValueOf(out).Elem().Set(newMap)
			} else {
				newElem := reflect.New(v.outputType)
				elem.Set(newElem)
				reflect.ValueOf(out).Elem().Set(newElem)
			}
		}

	} else {
		return errors.Collection(errors.Errorf(errors.CodeInternal, ctx, "internal error", "cannot assign %T to %T", out, output))
	}

	var outValue reflect.Value
	if v.ptr {
		outValue = reflect.Indirect(reflect.ValueOf(*out))
	} else {
		outValue = reflect.Indirect(reflect.ValueOf(out))
	}

	s := v.newSetter(outValue)

	inValue := reflect.Indirect(reflect.ValueOf(value))
	inKind := inValue.Kind()

	// Convert strings to JSON if necessary
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
			if err := json.Unmarshal(inValue.Bytes(), &result); err == nil {
				coerced = true
			}
		}

		if !coerced && attempted {
			return errors.Collection(
				errors.Error(errors.CodeType, ctx, "object, map, or JSON string", inKind.String()),
			)
		}

		if attempted {
			inValue = reflect.ValueOf(result)
			inKind = inValue.Kind()
		}
	}

	fromMap := inKind == reflect.Map
	fromSame := !fromMap && inValue.Type() == v.outputType

	if !fromMap && inKind != reflect.Struct {
		return errors.Collection(
			errors.Error(errors.CodeType, ctx, "object or map", inKind.String()),
		)
	}

	allErrors := errors.Collection()

	// Evaluate key rules
	keyErrs := v.evaluateKeyRules(ctx, out, inValue, s, fromMap, fromSame)
	allErrors = append(allErrors, keyErrs...)

	// Evaluate object rules
	valErrs := v.evaluateObjectRules(ctx, out)
	allErrors = append(allErrors, valErrs...)

	if len(allErrors) > 0 {
		return allErrors
	}

	if assignLater {
		elem.Set(reflect.ValueOf(out).Elem())
	}

	return nil
}

// Evaluate performs validation of a RuleSet against a value of the object type and returns a ValidationErrorCollection.
func (ruleSet *ObjectRuleSet[T, TK, TV]) Evaluate(ctx context.Context, value T) errors.ValidationErrorCollection {
	// Prepare a variable to hold the output after applying the rule set
	var output T

	// Apply the rule set to the value within the provided context
	errs := ruleSet.Apply(ctx, value, &output)
	return errs
}

// WithJson allows the input to be a JSON encoded string.
func (v *ObjectRuleSet[T, TK, TV]) WithJson() *ObjectRuleSet[T, TK, TV] {
	if v.json {
		return v
	}

	newRuleSet := v.clone()
	newRuleSet.json = true
	return newRuleSet
}

// WithRule returns a new child rule set that applies a custom validation rule.
// The custom rule is evaluated during validation and any errors it returns are included in the validation result.
func (v *ObjectRuleSet[T, TK, TV]) WithRule(rule Rule[T]) *ObjectRuleSet[T, TK, TV] {
	newRuleSet := v.clone()
	newRuleSet.objRule = rule
	return newRuleSet
}

// WithRuleFunc returns a new child rule set that applies a custom validation function.
// The custom function is evaluated during validation and any errors it returns are included in the validation result.
func (v *ObjectRuleSet[T, TK, TV]) WithRuleFunc(rule RuleFunc[T]) *ObjectRuleSet[T, TK, TV] {
	return v.WithRule(rule)
}

// Any returns a new RuleSet that wraps the object RuleSet in an Any rule set
// which can then be used in nested validation.
func (v *ObjectRuleSet[T, TK, TV]) Any() RuleSet[any] {
	return WrapAny(v)
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
				path := "<dynamic>"
				if c, ok := ruleSet.key.(*ConstantRuleSet[TK]); ok {
					path = toQuotedPath(c.Value())
				}

				label = fmt.Sprintf("WithKey(%s, %s)", path, ruleSet.rule)
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

// WithErrorMessage returns a new RuleSet with custom short and long error messages.
func (v *ObjectRuleSet[T, TK, TV]) WithErrorMessage(short, long string) *ObjectRuleSet[T, TK, TV] {
	return v.clone(objectWithLabel[T, TK, TV](util.FormatErrorMessageLabel(short, long)), objectWithErrorConfig[T, TK, TV](v.errorConfig.WithMessage(short, long)))
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (v *ObjectRuleSet[T, TK, TV]) WithDocsURI(uri string) *ObjectRuleSet[T, TK, TV] {
	return v.clone(objectWithLabel[T, TK, TV](util.FormatStringArgLabel("WithDocsURI", uri)), objectWithErrorConfig[T, TK, TV](v.errorConfig.WithDocs(uri)))
}

// WithTraceURI returns a new RuleSet with a custom trace/debug URI.
func (v *ObjectRuleSet[T, TK, TV]) WithTraceURI(uri string) *ObjectRuleSet[T, TK, TV] {
	return v.clone(objectWithLabel[T, TK, TV](util.FormatStringArgLabel("WithTraceURI", uri)), objectWithErrorConfig[T, TK, TV](v.errorConfig.WithTrace(uri)))
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (v *ObjectRuleSet[T, TK, TV]) WithErrorCode(code errors.ErrorCode) *ObjectRuleSet[T, TK, TV] {
	return v.clone(objectWithLabel[T, TK, TV](util.FormatErrorCodeLabel(code)), objectWithErrorConfig[T, TK, TV](v.errorConfig.WithCode(code)))
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (v *ObjectRuleSet[T, TK, TV]) WithErrorMeta(key string, value any) *ObjectRuleSet[T, TK, TV] {
	return v.clone(objectWithLabel[T, TK, TV](util.FormatErrorMetaLabel(key, value)), objectWithErrorConfig[T, TK, TV](v.errorConfig.WithMeta(key, value)))
}

// WithErrorCallback returns a new RuleSet with an error callback for customization.
func (v *ObjectRuleSet[T, TK, TV]) WithErrorCallback(fn errors.ErrorCallback) *ObjectRuleSet[T, TK, TV] {
	return v.clone(objectWithLabel[T, TK, TV](util.FormatErrorCallbackLabel()), objectWithErrorConfig[T, TK, TV](v.errorConfig.WithCallback(fn)))
}
