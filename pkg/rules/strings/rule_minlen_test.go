package strings_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestMinLen(t *testing.T) {
	ruleSet := strings.New().WithMinLen(2).Any()

	testhelpers.MustBeValid(t, ruleSet, "abc", "abc")
	testhelpers.MustBeValid(t, ruleSet, "ab", "ab")
	testhelpers.MustBeInvalid(t, ruleSet, "a", errors.CodeMin)
}
