// Package objects provides a RuleSet implementation that can be used to validate object and map values.
//
// It implements standard rules and allows the developer to set a rule set to validate individual keys.
package objects

import (
	"context"
	standardErrors "errors"
	"fmt"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

const annotation = "validate"

// Implementation of RuleSet for objects and maps.
type ObjectRuleSet[T any] struct {
	init         func() T
	allowUnknown bool
	key          string
	rule         rules.RuleSet[any]
	objRule      rules.Rule[T]
	mapping      string
	outputType   reflect.Type
	required     bool
	parent       *ObjectRuleSet[T]
}

// New returns a validator that can be used to validate an object of an
// arbitrary data type.
//
// It takes a function as an argument that must return a new (zero) value
// for the struct.
//
// Using the "validate" annotation you can may input values to different
// properties of the object. This is useful for converting unstructured maps
// created from Json and converting to an object.
func New[T any](initFn func() T) *ObjectRuleSet[T] {
	templateValue := reflect.Indirect(reflect.ValueOf(initFn()))
	kind := templateValue.Kind()

	if kind != reflect.Struct && kind != reflect.Map {
		panic(standardErrors.New("invalid output type for object rule set"))
	}

	templateType := templateValue.Type()

	ruleSet := &ObjectRuleSet[T]{}

	mapped := make(map[string]bool)

	for i := 0; i < templateValue.NumField(); i++ {
		field := templateType.Field(i)

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
			parent:  ruleSet,
			key:     key,
			mapping: field.Name,
		}

		mapped[key] = true
	}

	ruleSet.init = initFn
	ruleSet.outputType = templateType

	return ruleSet
}

// NewObjectMap returns a new RuleSet that can be used to validate maps with strings as the
// keys and the specified data type (which can be "any") as the values.
func NewObjectMap[T any]() *ObjectRuleSet[map[string]T] {
	return &ObjectRuleSet[map[string]T]{
		init: func() map[string]T {
			return make(map[string]T)
		},
	}
}

// withParent is a helper function to assist in cloning object RuleSets.
func (v *ObjectRuleSet[T]) withParent() *ObjectRuleSet[T] {
	return &ObjectRuleSet[T]{
		init:         v.init,
		allowUnknown: v.allowUnknown,
		required:     v.required,
		outputType:   v.outputType,
		parent:       v,
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
// If more than one call is made to WithKey with the same key then only the final one will be used.
func (v *ObjectRuleSet[T]) WithKey(key string, ruleSet rules.RuleSet[any]) *ObjectRuleSet[T] {
	// Only check mapping if output type is a struct (not a map)
	if v.outputType != nil {
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
	return newRuleSet
}

// Validate performs a validation of a RuleSet against a value and returns a value of the correct type or
// a ValidationErrorCollection.
func (v *ObjectRuleSet[T]) Validate(value any) (T, errors.ValidationErrorCollection) {
	return v.ValidateWithContext(value, context.Background())
}

// ValidateWithContext performs a validation of a RuleSet against a value and returns a value of the correct type or
// a ValidationErrorCollection.
//
// Also, takes a Context which can be used by validaton rules and error formatting.
func (v *ObjectRuleSet[T]) ValidateWithContext(in any, ctx context.Context) (T, errors.ValidationErrorCollection) {
	out := v.init()

	var outValue reflect.Value

	// We can't use reflect.Set on a non-pointer struct so if the output is not a pointer
	// we want to make a pointer to work with.
	isPointer := reflect.ValueOf(out).Kind() == reflect.Ptr
	if isPointer {
		outValue = reflect.Indirect(reflect.ValueOf(out))
	} else {
		outValue = reflect.Indirect(reflect.ValueOf(&out))
	}

	outKind := outValue.Kind()

	inValue := reflect.Indirect(reflect.ValueOf(in))
	inKind := inValue.Kind()

	fromMap := inKind == reflect.Map

	if !fromMap && inKind != reflect.Struct {
		return out, errors.Collection(
			errors.NewCoercionError(ctx, "object or map", inKind.String()),
		)
	}

	toMap := outKind == reflect.Map

	allErrors := errors.Collection()

	fieldMapping := v.fullMapping()

	// Only set if the input is a map and we don't allow unknown values
	var knownKeys map[string]bool
	if !v.allowUnknown && fromMap {
		knownKeys = make(map[string]bool)
	}

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
				allErrors.Add(errors.Errorf(errors.CodeRequired, subContext, "field is required"))
			}

		} else {
			val, errs := rule.ValidateWithContext(inFieldValue.Interface(), subContext)

			if errs != nil {
				allErrors.Add(errs.All()...)
				continue
			}

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
		}
	}

	// Next apply object rules.
	// This must be done after the key rules because we want to make sure all values are cast first.
	for currentRuleSet := v; currentRuleSet != nil; currentRuleSet = currentRuleSet.parent {
		if currentRuleSet.objRule != nil {
			newOutput, err := currentRuleSet.objRule.Evaluate(ctx, out)
			if err != nil {
				allErrors.Add(err.All()...)
			} else {
				out = newOutput
			}
		}
	}

	if knownKeys != nil {
		for _, key := range inValue.MapKeys() {
			keyStr := key.String()
			_, ok := knownKeys[keyStr]
			if !ok {
				subContext := rulecontext.WithPathString(ctx, keyStr)
				allErrors.Add(errors.Errorf(errors.CodeUnexpected, subContext, "unexpected field"))
			}
		}
	}

	if allErrors.Size() != 0 {
		return out, allErrors
	} else {
		return out, nil
	}
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
