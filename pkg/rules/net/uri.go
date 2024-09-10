package net

// See: https://datatracker.ietf.org/doc/html/rfc3986

import (
	"context"
	"reflect"
	"regexp"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/numbers"
)

// Base rule set for all normal string portions of the URI.
func isHex(c rune) bool {
	return (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || (c >= '0' && c <= '9')
}

func percentEncodingRule(ctx context.Context, value string) errors.ValidationErrorCollection {
	runes := []rune(value)

	l := len(runes)
	for i, _ := range runes {
		if runes[i] != '%' {
			continue
		}

		if i >= l-2 || !isHex(runes[i+1]) || !isHex(runes[i+2]) {
			return errors.Collection(
				errors.Errorf(errors.CodeEncoding, ctx, "field is not URI encoded %d >= %d - 2", i, l),
			)
		}
	}

	return nil
}

var baseUriPartRuleSet *rules.StringRuleSet = rules.NewString().WithRuleFunc(percentEncodingRule)

// Scheme has special rules.
var defaultSchemaRuleSet *rules.StringRuleSet = rules.NewString().WithRegexpString("^[A-Za-z][A-Za-z0-9+\\-.]*$", "Invalid schema.")

// Terminal parts.
var defaultPathRuleSet *rules.StringRuleSet = baseUriPartRuleSet
var defaultQueryRuleSet *rules.StringRuleSet = baseUriPartRuleSet
var defaultFragmentRuleSet *rules.StringRuleSet = baseUriPartRuleSet
var defaultHostRuleSet *rules.StringRuleSet = baseUriPartRuleSet
var defaultUserRuleSet *rules.StringRuleSet = baseUriPartRuleSet
var defaultPasswordRuleSet *rules.StringRuleSet = baseUriPartRuleSet
var defaultPortRuleSet *numbers.IntRuleSet[int] = numbers.NewInt().WithMin(0).WithMax(65535)

// backgroundDomainRuleSet is the base domain rule set. Since rule sets are immutable.
var backgroundURIRuleSet URIRuleSet = URIRuleSet{
	label:           "URIRuleSet",
	schemeRuleSet:   defaultSchemaRuleSet,
	pathRuleSet:     defaultPathRuleSet,
	queryRuleSet:    defaultQueryRuleSet,
	fragmentRuleSet: defaultFragmentRuleSet,
	hostRuleSet:     defaultHostRuleSet,
	userRuleSet:     defaultUserRuleSet,
	passwordRuleSet: defaultPasswordRuleSet,
	portRuleSet:     defaultPortRuleSet,
}

// URIRuleSet implements the RuleSet interface for URIs.
//
// It is slightly less efficient than other URI validators because it focuses on being able to evaluate
// each part of the URI independently and return very specific errors rather than simple regular expressions.
// This leads to the ability to have modular and testable rules for individual parts of the URL.
type URIRuleSet struct {
	rules.NoConflict[string]
	required         bool
	deepErrors       bool
	relative         bool
	parent           *URIRuleSet
	schemeRuleSet    *rules.StringRuleSet
	authorityRuleSet *rules.StringRuleSet
	pathRuleSet      *rules.StringRuleSet
	queryRuleSet     *rules.StringRuleSet
	fragmentRuleSet  *rules.StringRuleSet
	hostRuleSet      *rules.StringRuleSet
	userinfoRuleSet  *rules.StringRuleSet
	userRuleSet      *rules.StringRuleSet
	passwordRuleSet  *rules.StringRuleSet
	portRuleSet      *numbers.IntRuleSet[int]

	rule  rules.Rule[string]
	label string
}

// NewDomain creates a new domain RuleSet
func NewURI() *URIRuleSet {
	return &backgroundURIRuleSet
}

// Required returns a boolean indicating if the value is allowed to be omitted when included in a nested object.
func (ruleSet *URIRuleSet) Required() bool {
	return ruleSet.required
}

// WithRequired returns a new rule set with the required flag set.
// Use WithRequired when nesting a RuleSet and the a value is not allowed to be omitted.
func (ruleSet *URIRuleSet) WithRequired() *URIRuleSet {
	if ruleSet.required {
		return ruleSet
	}

	newRuleSet := ruleSet.copyWithParent(ruleSet)
	newRuleSet.required = true
	newRuleSet.label = "WithRequired()"
	return newRuleSet
}

// WithUserRequired returns a new rule set with the user set to required.
// The user must be in the URI, however, it may be empty.
func (ruleSet *URIRuleSet) WithUserRequired() *URIRuleSet {
	if ruleSet.userRuleSet.Required() {
		return ruleSet
	}

	newRuleSet := ruleSet.copyWithParent(ruleSet)
	newRuleSet.userRuleSet = newRuleSet.userRuleSet.WithRequired()
	newRuleSet.label = "WithUserRequired()"
	return newRuleSet
}

// WithPasswordRequired returns a new rule set with the password set to required.
// The password must be in the URI, however, it may be empty.
func (ruleSet *URIRuleSet) WithPasswordRequired() *URIRuleSet {
	if ruleSet.passwordRuleSet.Required() {
		return ruleSet
	}

	newRuleSet := ruleSet.copyWithParent(ruleSet)
	newRuleSet.passwordRuleSet = newRuleSet.passwordRuleSet.WithRequired()
	newRuleSet.label = "WithPasswordRequired()"
	return newRuleSet
}

// WithHostRequired returns a new rule set with the host set to required.
// The host must be in the URI, however, it may be empty.
func (ruleSet *URIRuleSet) WithHostRequired() *URIRuleSet {
	if ruleSet.hostRuleSet.Required() {
		return ruleSet
	}

	newRuleSet := ruleSet.copyWithParent(ruleSet)
	newRuleSet.hostRuleSet = newRuleSet.hostRuleSet.WithRequired()
	newRuleSet.label = "WithHostRequired()"
	return newRuleSet
}

// WithPortRequired returns a new rule set with the port set to required.
// The port must be in the URI, however, it may be empty.
func (ruleSet *URIRuleSet) WithPortRequired() *URIRuleSet {
	if ruleSet.portRuleSet.Required() {
		return ruleSet
	}

	newRuleSet := ruleSet.copyWithParent(ruleSet)
	newRuleSet.portRuleSet = newRuleSet.portRuleSet.WithRequired()
	newRuleSet.label = "WithPortRequired()"
	return newRuleSet
}

// WithQueryRequired returns a new rule set with the query set to required.
// The query must be in the URI, however, it may be empty.
func (ruleSet *URIRuleSet) WithQueryRequired() *URIRuleSet {
	if ruleSet.queryRuleSet.Required() {
		return ruleSet
	}

	newRuleSet := ruleSet.copyWithParent(ruleSet)
	newRuleSet.queryRuleSet = newRuleSet.queryRuleSet.WithRequired()
	newRuleSet.label = "WithQueryRequired()"
	return newRuleSet
}

// WithFragmentRequired returns a new rule set with the fragment set to required.
// The fragment must be in the URI, however, it may be empty.
func (ruleSet *URIRuleSet) WithFragmentRequired() *URIRuleSet {
	if ruleSet.fragmentRuleSet.Required() {
		return ruleSet
	}

	newRuleSet := ruleSet.copyWithParent(ruleSet)
	newRuleSet.fragmentRuleSet = newRuleSet.fragmentRuleSet.WithRequired()
	newRuleSet.label = "WithFragmentRequired()"
	return newRuleSet
}

// deeoErrorContext creates a new context if deepErrors are enabled, otherwise it uses the same one.
func (ruleSet *URIRuleSet) deepErrorContext(ctx context.Context, name string) context.Context {
	if ruleSet.deepErrors {
		return rulecontext.WithPathString(ctx, name)
	}
	return ctx
}

// DeepErrors returns a boolean indicating if the the rule set is set to return deep errors.
// If deep errors are not set then the paths returned in validation errors should point to the string itself
// and not the segment within the string.
//
// See WithDeepErrors for examples.
func (ruleSet *URIRuleSet) DeepErrors() bool {
	return ruleSet.deepErrors
}

// WithDeepErrors returns a new rule set with the deep errors flag set.
// By default URIRuleSet will return the path to the string itself when returning errors. Setting deep errors
// will tell the rules to return the rull path to the error nested inside the string.
//
// For example,the URI https://example.com:-1/ has an invalid port numbers (ports can not be negative).
// By default the path may look like this: `/myobj/some_uri`
// With deep errors the path may look like this: `/myobj/some_uri/port`
func (ruleSet *URIRuleSet) WithDeepErrors() *URIRuleSet {
	if ruleSet.deepErrors {
		return ruleSet
	}

	newRuleSet := ruleSet.copyWithParent(ruleSet)
	newRuleSet.deepErrors = true
	newRuleSet.label = "WithDeepErrors()"
	return newRuleSet
}

// Relative returns a boolean indicating if the the rule set is set to allow relative URIs.
func (ruleSet *URIRuleSet) Relative() bool {
	return ruleSet.relative
}

// WithRelative returns a new rule set with the relative flag set.
// By default URIRuleSet requires all parts of the URI to be specified. WithRelative will allow some
// parts of the URI to be omitted.
//
// Scheme is normally required for URIs but is optional if relative URIs are enabled.
func (ruleSet *URIRuleSet) WithRelative() *URIRuleSet {
	if ruleSet.relative {
		return ruleSet
	}

	newRuleSet := ruleSet.copyWithParent(ruleSet)
	newRuleSet.relative = true
	newRuleSet.label = "WithRelative()"
	return newRuleSet
}

// Apply performs a validation of a RuleSet against a value and assigns the result to the output parameter.
// It returns a ValidationErrorCollection if any validation errors occur.
func (ruleSet *URIRuleSet) Apply(ctx context.Context, input any, output any) errors.ValidationErrorCollection {
	// Attempt to cast the input to a string
	valueStr, ok := input.(string)
	if !ok {
		return errors.Collection(errors.NewCoercionError(ctx, "string", reflect.ValueOf(input).Kind().String()))
	}

	// Perform the validation
	if err := ruleSet.Evaluate(ctx, valueStr); err != nil {
		return err
	}

	outputVal := reflect.ValueOf(output)

	// Check if the output is a non-nil pointer
	if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Output must be a non-nil pointer",
		))
	}

	// Dereference the pointer to get the actual value that needs to be set
	outputElem := outputVal.Elem()

	switch outputElem.Kind() {
	case reflect.String:
		outputElem.SetString(valueStr)
	case reflect.Interface:
		outputElem.Set(reflect.ValueOf(valueStr))
	default:
		return errors.Collection(errors.Errorf(
			errors.CodeInternal, ctx, "Cannot assign string to %T", output,
		))
	}

	return nil
}

// evaluateScheme evaluates the scheme portion of the URI and also returns a context with the scheme set.
func (ruleSet *URIRuleSet) evaluateScheme(ctx context.Context, value string) (context.Context, errors.ValidationErrorCollection) {
	newCtx := context.WithValue(ctx, "scheme", value)
	subContext := ruleSet.deepErrorContext(newCtx, "scheme")

	if value == "" {
		if !ruleSet.relative {
			return newCtx, errors.Collection(errors.Errorf(errors.CodeRequired, subContext, "Scheme is required."))
		}
		return newCtx, nil
	}

	return newCtx, ruleSet.schemeRuleSet.Evaluate(subContext, value)
}

// evaluateUser evaluates the user portion of the userinfo in the URI and also returns a context with the user set.
func (ruleSet *URIRuleSet) evaluateUser(ctx context.Context, value string) (context.Context, errors.ValidationErrorCollection) {
	newCtx := context.WithValue(ctx, "user", value)
	subContext := ruleSet.deepErrorContext(newCtx, "user")

	return newCtx, ruleSet.userRuleSet.Evaluate(subContext, value)
}

// evaluatePassword evaluates the password portion of the userinfo in the URI and also returns a context with the password set.
func (ruleSet *URIRuleSet) evaluatePassword(ctx context.Context, value string) (context.Context, errors.ValidationErrorCollection) {
	newCtx := context.WithValue(ctx, "password", value)

	if value == "" && !ruleSet.passwordRuleSet.Required() {
		return newCtx, nil
	}

	subContext := ruleSet.deepErrorContext(newCtx, "password")

	return newCtx, ruleSet.passwordRuleSet.Evaluate(subContext, value)
}

// evaluateAuthorityPart takes a context, a authority part name, and its value and returns any validation errors and a modified context.
func (ruleSet *URIRuleSet) evaluateUserinfoPart(ctx context.Context, name, value string) (context.Context, errors.ValidationErrorCollection) {
	switch name {
	case "user":
		return ruleSet.evaluateUser(ctx, value)
	case "password":
		return ruleSet.evaluatePassword(ctx, value)
	}
	return ctx, nil
}

// evaluateUserinfo evaluates the userinfo portion of the URI and also returns a context with the userinfo set.
func (ruleSet *URIRuleSet) evaluateUserinfo(ctx context.Context, value string) (context.Context, errors.ValidationErrorCollection) {
	const userinfoRegex = `^` +
		`(?P<user>[^:]*)` + // User
		`([:]?)(?P<password>.*)` + // Password
		`$`

	newCtx := context.WithValue(ctx, "userinfo", value)

	if value == "" {
		var verr errors.ValidationErrorCollection

		if ruleSet.passwordRuleSet.Required() {
			subContext := ruleSet.deepErrorContext(newCtx, "password")
			verr = append(verr, errors.Errorf(errors.CodeRequired, subContext, "Password is required."))
		}
		if ruleSet.userRuleSet.Required() {
			subContext := ruleSet.deepErrorContext(newCtx, "user")
			verr = append(verr, errors.Errorf(errors.CodeRequired, subContext, "User is required."))
		}

		if len(verr) > 0 {
			return newCtx, verr
		}
		return newCtx, nil
	}

	allErrors := errors.Collection()
	r := regexp.MustCompile(userinfoRegex)
	match := r.FindStringSubmatch(value)

	var verr errors.ValidationErrorCollection

	// Regex always matches
	for i, name := range r.SubexpNames() {
		// User is implicit but if there is no ':' we treat password as missing.
		// The match right before password should be a colon or empty
		if name == "password" && match[i-1] == "" {
			if ruleSet.passwordRuleSet.Required() {
				subContext := ruleSet.deepErrorContext(newCtx, "password")
				return newCtx, errors.Collection(errors.Errorf(errors.CodeRequired, subContext, "Password is required."))
			}
		}

		newCtx, verr = ruleSet.evaluateUserinfoPart(newCtx, name, match[i])
		allErrors = append(allErrors, verr...)
	}

	if len(allErrors) > 0 {
		return newCtx, allErrors
	}

	return newCtx, nil
}

// evaluateHost evaluates the host portion of the URI and also returns a context with the host set.
func (ruleSet *URIRuleSet) evaluateHost(ctx context.Context, value string) (context.Context, errors.ValidationErrorCollection) {
	newCtx := context.WithValue(ctx, "host", value)
	subContext := ruleSet.deepErrorContext(newCtx, "host")

	return newCtx, ruleSet.hostRuleSet.Evaluate(subContext, value)
}

// evaluatePort evaluates the port portion of the URI and also returns a context with the port set.
func (ruleSet *URIRuleSet) evaluatePort(ctx context.Context, value string) (context.Context, errors.ValidationErrorCollection) {
	newCtx := context.WithValue(ctx, "port", value)

	if value == "" && !ruleSet.portRuleSet.Required() {
		return newCtx, nil
	}

	subContext := ruleSet.deepErrorContext(newCtx, "port")

	var output int
	err := ruleSet.portRuleSet.Apply(subContext, value, &output)
	return newCtx, err
}

// evaluateAuthorityPart takes a context, a authority part name, and its value and returns any validation errors and a modified context.
func (ruleSet *URIRuleSet) evaluateAuthorityPart(ctx context.Context, name, value string) (context.Context, errors.ValidationErrorCollection) {
	switch name {
	case "userinfo":
		return ruleSet.evaluateUserinfo(ctx, value)
	case "host":
		return ruleSet.evaluateHost(ctx, value)
	case "port":
		return ruleSet.evaluatePort(ctx, value)
	}
	return ctx, nil
}

// evaluateAuthority evaluates the authority portion of the URI and also returns a context with the authority, host, port, and userinfo set.
func (ruleSet *URIRuleSet) evaluateAuthority(ctx context.Context, value string, missing bool) (context.Context, errors.ValidationErrorCollection) {
	allErrors := errors.Collection()
	newCtx := context.WithValue(ctx, "authority", value)

	// Authority can be omitted from the URI.
	// If it is, that means that any required parts that are inside of the authority are missing.
	// That means that we should trigger validation errors for any missing but required parts.
	// Note: this is the ONLY way that host can be missing. All other parts are tested later as well.
	// Previous value should be "//" if the authority is present
	if missing {
		if ruleSet.userRuleSet.Required() {
			subContext := ruleSet.deepErrorContext(newCtx, "user")
			allErrors = append(allErrors, errors.Errorf(errors.CodeRequired, subContext, "User is required."))
		}
		if ruleSet.passwordRuleSet.Required() {
			subContext := ruleSet.deepErrorContext(newCtx, "password")
			allErrors = append(allErrors, errors.Errorf(errors.CodeRequired, subContext, "Password is required."))
		}
		if ruleSet.hostRuleSet.Required() {
			subContext := ruleSet.deepErrorContext(newCtx, "host")
			allErrors = append(allErrors, errors.Errorf(errors.CodeRequired, subContext, "Host is required."))
		}
		if ruleSet.portRuleSet.Required() {
			subContext := ruleSet.deepErrorContext(newCtx, "port")
			allErrors = append(allErrors, errors.Errorf(errors.CodeRequired, subContext, "Port is required."))
		}

		// These are usually set in evaluateURIPart but we are skipping that
		newCtx = context.WithValue(newCtx, "userinfo", "")
		newCtx = context.WithValue(newCtx, "user", "")
		newCtx = context.WithValue(newCtx, "password", "")
		newCtx = context.WithValue(newCtx, "host", "")
		newCtx = context.WithValue(newCtx, "port", "")
		return newCtx, allErrors
	}

	// Authority can be empty
	const authorityRegex = `^` +
		`(:?(?P<userinfo>[^@]*)@)?` + // Userinfo
		`(?P<host>[^:]*)` + // Host
		`([:]?)(?P<port>.*)` + // Port
		`$`

	r := regexp.MustCompile(authorityRegex)
	match := r.FindStringSubmatch(value)

	var verr errors.ValidationErrorCollection

	// Regex always matches since all parts are optional
	for i, name := range r.SubexpNames() {
		if name == "port" && match[i-1] == "" {
			if ruleSet.portRuleSet.Required() {
				subContext := ruleSet.deepErrorContext(newCtx, "port")
				allErrors = append(allErrors, errors.Errorf(errors.CodeRequired, subContext, "Port is required."))
				continue
			}
		}

		newCtx, verr = ruleSet.evaluateAuthorityPart(newCtx, name, match[i])
		allErrors = append(allErrors, verr...)
	}

	if len(allErrors) > 0 {
		return newCtx, allErrors
	}

	return newCtx, nil
}

// evaluatePath evaluates the path portion of the URI and also returns a context with the path set.
func (ruleSet *URIRuleSet) evaluatePath(ctx context.Context, value string) (context.Context, errors.ValidationErrorCollection) {
	newCtx := context.WithValue(ctx, "path", value)
	subContext := ruleSet.deepErrorContext(newCtx, "path")

	return newCtx, ruleSet.pathRuleSet.Evaluate(subContext, value)
}

// evaluateQuery evaluates the fragment portion of the URI and also returns a context with the fragment set.
func (ruleSet *URIRuleSet) evaluateQuery(ctx context.Context, value string, missing bool) (context.Context, errors.ValidationErrorCollection) {
	newCtx := context.WithValue(ctx, "query", value)
	subContext := ruleSet.deepErrorContext(newCtx, "query")

	if missing {
		if ruleSet.queryRuleSet.Required() {
			return newCtx, errors.Collection(
				errors.Errorf(errors.CodeRequired, subContext, "Query is required."),
			)
		}
		return newCtx, nil
	}

	return newCtx, ruleSet.queryRuleSet.Evaluate(subContext, value)
}

// evaluateFragment evaluates the fragment portion of the URI and also returns a context with the fragment set.
func (ruleSet *URIRuleSet) evaluateFragment(ctx context.Context, value string, missing bool) (context.Context, errors.ValidationErrorCollection) {
	newCtx := context.WithValue(ctx, "fragment", value)
	subContext := ruleSet.deepErrorContext(newCtx, "fragment")

	if missing {
		if ruleSet.fragmentRuleSet.Required() {
			return newCtx, errors.Collection(
				errors.Errorf(errors.CodeRequired, subContext, "Fragment is required."),
			)
		}
		return newCtx, nil
	}

	return newCtx, ruleSet.fragmentRuleSet.Evaluate(subContext, value)
}

// evaluateURIPart takes a context, a URI part name, and its value and returns any validation errors and a modified context.
func (ruleSet *URIRuleSet) evaluateURIPart(ctx context.Context, name, value, previousValue string) (context.Context, errors.ValidationErrorCollection) {
	switch name {
	case "scheme":
		return ruleSet.evaluateScheme(ctx, value)
	case "authority":
		return ruleSet.evaluateAuthority(ctx, value, previousValue == "")
	case "path":
		return ruleSet.evaluatePath(ctx, value)
	case "query":
		return ruleSet.evaluateQuery(ctx, value, previousValue == "")
	case "fragment":
		return ruleSet.evaluateFragment(ctx, value, previousValue == "")
	}
	return ctx, nil
}

// Evaluate performs a validation of a RuleSet against a string and returns an object value of the
// same type or a ValidationErrorCollection.
func (ruleSet *URIRuleSet) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	const URIRegex = `^` +
		`(?:(?P<scheme>[^:/?#]+):)?` + // Scheme
		`(?:(//)(?P<authority>[^/?#]*))?` + // Authority
		`(?P<path>[^?#]*)` + // Path
		`(?:(\?)(?P<query>[^#]*))?` + // Query
		`(?:(#)(?P<fragment>.*))?` + // Fragment
		`$`

	r := regexp.MustCompile(URIRegex)
	match := r.FindStringSubmatch(value)

	allErrors := errors.Collection()

	currentRuleSet := ruleSet
	ctx = rulecontext.WithRuleSet(ctx, ruleSet)

	var verr errors.ValidationErrorCollection

	// Regex always matches
	prevMatch := ""
	for i, name := range r.SubexpNames() {
		ctx, verr = ruleSet.evaluateURIPart(ctx, name, match[i], prevMatch)
		allErrors = append(allErrors, verr...)
		prevMatch = match[i]
	}

	for currentRuleSet != nil {
		if currentRuleSet.rule != nil {
			if errs := currentRuleSet.rule.Evaluate(ctx, value); errs != nil {
				allErrors = append(allErrors, errs...)
			}
		}

		currentRuleSet = currentRuleSet.parent
	}

	if len(allErrors) > 0 {
		return allErrors
	}

	return nil
}

// noConflict returns the new array rule set with all conflicting rules removed.
// Does not mutate the existing rule sets.
func (ruleSet *URIRuleSet) noConflict(rule rules.Rule[string]) *URIRuleSet {
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

	newRuleSet := ruleSet.copyWithParent(newParent)
	newRuleSet.rule = ruleSet.rule
	newRuleSet.label = ruleSet.label
	return newRuleSet
}

// WithRule returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRule takes an implementation of the Rule interface
// for the string type.
//
// Use this when implementing custom rules.
//
// In addition to the normal context values available to all rules, for URI rules
// the following values will always be set (but may be empty strings):
// - scheme
// - authority
// - path
// - query
// - fragment
// - port
// - userinfo
// - user
// - password
func (ruleSet *URIRuleSet) WithRule(rule rules.Rule[string]) *URIRuleSet {
	newRuleSet := ruleSet.copyWithParent(ruleSet.noConflict(rule))
	newRuleSet.rule = rule
	return newRuleSet
}

// WithRuleFunc returns a new child rule set with a rule added to the list of
// rules to evaluate. WithRuleFunc takes an implementation of the RuleFunc interface
// for the string type.
//
// Use this when implementing custom rules.
//
// In addition to the normal context values available to all rules, for URI rules
// the following values will always be set (but may be empty strings):
// - scheme
// - authority
// - path
// - query
// - fragment
// - port
// - userinfo
// - user
// - password
func (ruleSet *URIRuleSet) WithRuleFunc(rule rules.RuleFunc[string]) *URIRuleSet {
	return ruleSet.WithRule(rule)
}

// String returns a string representation of the rule set suitable for debugging.
func (ruleSet *URIRuleSet) String() string {
	label := ruleSet.label

	if ruleSet.parent != nil {
		return ruleSet.parent.String() + "." + label
	}
	return label
}

// Any returns a new RuleSet that wraps the URI RuleSet in any Any rule set
// which can then be used in nested validation.
func (ruleSet *URIRuleSet) Any() rules.RuleSet[any] {
	return rules.WrapAny[string](ruleSet)
}

// copy creates a rule set with all the appropriate fields copied and the parent set.
func (ruleSet *URIRuleSet) copyWithParent(newParent *URIRuleSet) *URIRuleSet {
	return &URIRuleSet{
		parent:           newParent,
		schemeRuleSet:    ruleSet.schemeRuleSet,
		authorityRuleSet: ruleSet.authorityRuleSet,
		pathRuleSet:      ruleSet.pathRuleSet,
		queryRuleSet:     ruleSet.queryRuleSet,
		fragmentRuleSet:  ruleSet.fragmentRuleSet,
		hostRuleSet:      ruleSet.hostRuleSet,
		portRuleSet:      ruleSet.portRuleSet,
		userinfoRuleSet:  ruleSet.userinfoRuleSet,
		userRuleSet:      ruleSet.userRuleSet,
		passwordRuleSet:  ruleSet.passwordRuleSet,
		required:         ruleSet.required,
		deepErrors:       ruleSet.deepErrors,
		relative:         ruleSet.relative,
	}
}
