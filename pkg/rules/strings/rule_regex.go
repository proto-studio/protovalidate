package strings

import (
	"context"
	"fmt"
	"regexp"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// Implements the Rule interface for regular expressions.
type regexpRule struct {
	exp *regexp.Regexp
	msg string
}

// Evaluate takes a context and string value and returns an error if it does not match the expected pattern.
func (rule *regexpRule) Evaluate(ctx context.Context, value string) (string, errors.ValidationErrorCollection) {
	if !rule.exp.MatchString(value) {
		return value, errors.Collection(
			errors.Errorf(errors.CodePattern, ctx, rule.msg),
		)
	}

	return value, nil
}

// Conflict always returns false. Regex can be stacked.
func (rule *regexpRule) Conflict(x rules.Rule[string]) bool {
	return false
}

// String returns the string representation of the regex rule.
// Example: WithRegexp(2)
func (rule *regexpRule) String() string {
	return fmt.Sprintf("WithRegexp(%s)", rule.exp)
}

// WithRegexpString returns a new child RuleSet that is constrained to the provided regular expression.
// The second parameter is the error text, which will be localized if a translation is available.
//
// This method panics if the expression cannot be compiled.
func (v *StringRuleSet) WithRegexpString(exp, errorMsg string) *StringRuleSet {
	compiledExp := regexp.MustCompile(exp)

	return v.WithRule(&regexpRule{
		compiledExp,
		errorMsg,
	})
}

// WithRegexp returns a new child RuleSet that is constrained to the provided regular expression.
// The second parameter is the error text, which will be localized if a translation is available.
func (v *StringRuleSet) WithRegexp(exp *regexp.Regexp, errorMsg string) *StringRuleSet {
	return v.WithRule(&regexpRule{
		exp,
		errorMsg,
	})
}
