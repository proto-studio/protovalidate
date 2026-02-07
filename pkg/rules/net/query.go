package net

import (
	"context"
	"net/url"
	"reflect"

	"proto.zip/studio/validate/internal/util"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// queryPercentEncodingRule validates that the query string is properly percent-encoded.
func queryPercentEncodingRule(ctx context.Context, value string) errors.ValidationErrorCollection {
	runes := []rune(value)
	l := len(runes)
	for i := range runes {
		if runes[i] != '%' {
			continue
		}
		if i >= l-2 || !isHex(runes[i+1]) || !isHex(runes[i+2]) {
			return errors.Collection(
				errors.Errorf(errors.CodeEncoding, ctx, "invalid encoding", "value is not properly URI encoded"),
			)
		}
	}
	return nil
}

// paramSpec holds the rule set for a single query parameter.
type paramSpec struct {
	ruleSet rules.RuleSet[any]
}

// QueryRuleSet implements RuleSet[url.Values] for validating the entire URI query string.
// Apply accepts string (parsed with url.ParseQuery; parse error becomes a validation error) or url.Values.
// Output may be *string, *url.Values, or *any.
var baseQueryRuleSet = QueryRuleSet{
	label: "QueryRuleSet",
}

// Query returns the base QueryRuleSet.
func Query() *QueryRuleSet {
	return &baseQueryRuleSet
}

// QueryRuleSet is a rule set for the entire query string (e.g. percent encoding and optional param rules).
// Native type is url.Values. Apply accepts string or url.Values; output may be *string, *url.Values, or *any.
type QueryRuleSet struct {
	rules.NoConflict[url.Values]
	paramRules  map[string]*paramSpec
	rule        rules.Rule[url.Values]
	parent      *QueryRuleSet
	label       string
	errorConfig *errors.ErrorConfig
	required    bool // true if any param rule set passed to WithParam was required
}

// Required returns true if the query string is required (i.e. WithRequired was called or any rule set passed to WithParam had Required() true).
func (q *QueryRuleSet) Required() bool {
	return q.required
}

// WithRequired returns a new rule set that requires the query string to be present.
func (q *QueryRuleSet) WithRequired() *QueryRuleSet {
	newRuleSet := q.clone()
	newRuleSet.required = true
	newRuleSet.label = "WithRequired()"
	return newRuleSet
}

// WithParam returns a new rule set that validates the named query parameter with the given rule set.
// If the rule set is required, the whole query rule set is marked required (query string must be present).
func (q *QueryRuleSet) WithParam(name string, ruleSet rules.RuleSet[any]) *QueryRuleSet {
	newRuleSet := q.clone()
	newRuleSet.paramRules = q.copyParamRules()
	if newRuleSet.paramRules[name] == nil {
		newRuleSet.paramRules[name] = &paramSpec{}
	}
	newRuleSet.paramRules[name].ruleSet = ruleSet
	if ruleSet != nil && ruleSet.Required() {
		newRuleSet.required = true
	}
	newRuleSet.label = util.FormatStringArgLabel("WithParam", name)
	return newRuleSet
}

// WithRule returns a new rule set that applies a custom validation rule to the entire query (url.Values).
func (q *QueryRuleSet) WithRule(rule rules.Rule[url.Values]) *QueryRuleSet {
	newRuleSet := q.clone()
	newRuleSet.rule = rule
	newRuleSet.label = "WithRule()"
	return newRuleSet
}

// WithRuleFunc returns a new rule set that applies a custom validation function to the entire query (url.Values).
func (q *QueryRuleSet) WithRuleFunc(rule rules.RuleFunc[url.Values]) *QueryRuleSet {
	return q.WithRule(rule)
}

func (q *QueryRuleSet) copyParamRules() map[string]*paramSpec {
	out := make(map[string]*paramSpec)
	if q.paramRules != nil {
		for k, v := range q.paramRules {
			spec := &paramSpec{}
			if v != nil {
				spec.ruleSet = v.ruleSet
			}
			out[k] = spec
		}
	}
	return out
}

type queryCloneOption func(*QueryRuleSet)

func (q *QueryRuleSet) clone(options ...queryCloneOption) *QueryRuleSet {
	newRuleSet := &QueryRuleSet{
		parent:      q,
		paramRules:  q.paramRules,
		rule:        q.rule,
		label:       q.label,
		errorConfig: q.errorConfig,
		required:    q.required,
	}
	for _, opt := range options {
		opt(newRuleSet)
	}
	return newRuleSet
}

// defaultQueryStringRuleSet validates the raw query string (e.g. percent encoding); run when we have a raw string.
// Tests may override to cover the error-return branch in Evaluate.
var defaultQueryStringRuleSet rules.RuleSet[string] = rules.String().WithRuleFunc(queryPercentEncodingRule)

// Evaluate validates the query (percent encoding on the encoded form), registered parameters, and top-level rules.
func (q *QueryRuleSet) Evaluate(ctx context.Context, values url.Values) errors.ValidationErrorCollection {
	queryStringForEncoding := values.Encode()
	if queryStringForEncoding != "" {
		if err := defaultQueryStringRuleSet.Evaluate(ctx, queryStringForEncoding); err != nil {
			return err
		}
	}
	allErrors := errors.Collection()
	if len(q.paramRules) > 0 {
		for name, spec := range q.paramRules {
			if spec == nil {
				continue
			}
			paramValues := values[name]
			paramPresent := len(paramValues) > 0
			paramVal := ""
			if paramPresent {
				paramVal = paramValues[0]
			}
			paramContext := rulecontext.WithPathString(ctx, "query["+name+"]")

			if spec.ruleSet != nil {
				if !paramPresent && spec.ruleSet.Required() {
					allErrors = append(allErrors, errors.Errorf(errors.CodeRequired, paramContext, "required", "query parameter %q is required", name))
					continue
				}
				if !paramPresent && !spec.ruleSet.Required() {
					continue
				}
				if err := spec.ruleSet.Evaluate(paramContext, any(paramVal)); err != nil {
					allErrors = append(allErrors, err...)
				}
			}
		}
		if len(allErrors) > 0 {
			return allErrors
		}
	}
	current := q
	ctx = rulecontext.WithRuleSet(ctx, q)
	for current != nil {
		if current.rule != nil {
			if errs := current.rule.Evaluate(ctx, values); errs != nil {
				allErrors = append(allErrors, errs...)
			}
		}
		current = current.parent
	}
	if len(allErrors) > 0 {
		return allErrors
	}
	return nil
}

// queryParser is used by Apply to parse a query string; tests may override to trigger the parse-error branch.
var queryParser = url.ParseQuery

// Apply coerces input to url.Values (string is parsed; parse error becomes a validation error), validates, and writes to output.
// Output may be *string, *url.Values, or *any.
func (q *QueryRuleSet) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	ctx = errors.WithErrorConfig(ctx, q.errorConfig)

	var values url.Values
	switch v := input.(type) {
	case string:
		var parseErr error
		values, parseErr = queryParser(v)
		if parseErr != nil {
			return errors.Collection(
				errors.Errorf(errors.CodeEncoding, ctx, "invalid query", "query string could not be parsed: %v", parseErr),
			)
		}
	case url.Values:
		values = v
	default:
		return errors.Collection(errors.Errorf(
			errors.CodeType, ctx, "string or url.Values", reflect.ValueOf(input).Kind().String(),
		))
	}

	if err := q.Evaluate(ctx, values); err != nil {
		return err
	}

	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "internal error", "output must be a non-nil pointer",
		))
	}
	elem := outputVal.Elem()

	switch elem.Kind() {
	case reflect.String:
		elem.SetString(values.Encode())
		return nil
	case reflect.Interface:
		elem.Set(reflect.ValueOf(values))
		return nil
	case reflect.Map:
		if elem.Type() == reflect.TypeOf(url.Values(nil)) {
			if elem.IsNil() {
				elem.Set(reflect.MakeMap(elem.Type()))
			}
			for k, v := range values {
				elem.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
			}
			return nil
		}
	}
	return errors.Collection(errors.Errorf(
		errors.CodeInternal, ctx, "internal error", "query output must be *string, *url.Values, or *any, got %T", output,
	))
}

// String returns a string representation of the rule set for debugging.
func (q *QueryRuleSet) String() string {
	if q.parent != nil {
		return q.parent.String() + "." + q.label
	}
	return q.label
}

// Any returns a RuleSet[any] wrapping this rule set.
func (q *QueryRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[url.Values](q)
}

// WithErrorMessage returns a new RuleSet with custom error messages.
func (q *QueryRuleSet) WithErrorMessage(short, long string) *QueryRuleSet {
	return q.clone(func(rs *QueryRuleSet) {
		rs.label = util.FormatErrorMessageLabel(short, long)
		rs.errorConfig = q.errorConfig.WithErrorMessage(short, long)
	})
}

// WithDocsURI returns a new RuleSet with a custom documentation URI.
func (q *QueryRuleSet) WithDocsURI(uri string) *QueryRuleSet {
	return q.clone(func(rs *QueryRuleSet) {
		rs.label = util.FormatStringArgLabel("WithDocsURI", uri)
		rs.errorConfig = q.errorConfig.WithDocsURI(uri)
	})
}

// WithTraceURI returns a new RuleSet with a custom trace URI.
func (q *QueryRuleSet) WithTraceURI(uri string) *QueryRuleSet {
	return q.clone(func(rs *QueryRuleSet) {
		rs.label = util.FormatStringArgLabel("WithTraceURI", uri)
		rs.errorConfig = q.errorConfig.WithTraceURI(uri)
	})
}

// WithErrorCode returns a new RuleSet with a custom error code.
func (q *QueryRuleSet) WithErrorCode(code errors.ErrorCode) *QueryRuleSet {
	return q.clone(func(rs *QueryRuleSet) {
		rs.label = util.FormatErrorCodeLabel(code)
		rs.errorConfig = q.errorConfig.WithCode(code)
	})
}

// WithErrorMeta returns a new RuleSet with additional error metadata.
func (q *QueryRuleSet) WithErrorMeta(key string, value any) *QueryRuleSet {
	return q.clone(func(rs *QueryRuleSet) {
		rs.label = util.FormatErrorMetaLabel(key, value)
		rs.errorConfig = q.errorConfig.WithMeta(key, value)
	})
}

// WithErrorCallback returns a new RuleSet with an error callback.
func (q *QueryRuleSet) WithErrorCallback(fn errors.ErrorCallback) *QueryRuleSet {
	return q.clone(func(rs *QueryRuleSet) {
		rs.label = util.FormatErrorCallbackLabel()
		rs.errorConfig = q.errorConfig.WithCallback(fn)
	})
}
