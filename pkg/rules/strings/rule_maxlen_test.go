package strings_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestMaxLen(t *testing.T) {
	ruleSet := strings.New().WithMaxLen(2).Any()

	testhelpers.MustBeValid(t, ruleSet, "a", "a")
	testhelpers.MustBeValid(t, ruleSet, "ab", "ab")
	testhelpers.MustBeInvalid(t, ruleSet, "abc", errors.CodeMax)

}
