package rules

import (
	"context"
	"fmt"
	"regexp"

	"proto.zip/studio/validate/pkg/errors"
)

// Implements the Rule interface for regular expressions.
type regexpRule struct {
	NoConflict[string]
	exp *regexp.Regexp
	msg string
}

// Evaluate takes a context and string value and returns an error if it does not match the expected pattern.
func (rule *regexpRule) Evaluate(ctx context.Context, value string) errors.ValidationErrorCollection {
	if !rule.exp.MatchString(value) {
		return errors.Collection(
			errors.Errorf(errors.CodePattern, ctx, rule.msg),
		)
	}

	return nil
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
	return v.WithRegexp(regexp.MustCompile(exp), errorMsg)
}

// WithRegexp returns a new child RuleSet that is constrained to the provided regular expression.
// The second parameter is the error text, which will be localized if a translation is available.
func (v *StringRuleSet) WithRegexp(exp *regexp.Regexp, errorMsg string) *StringRuleSet {
	return v.WithRule(&regexpRule{
		exp: exp,
		msg: errorMsg,
	})
}
