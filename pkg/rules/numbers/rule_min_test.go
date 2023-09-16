package numbers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMinInt(t *testing.T) {
	ruleSet := numbers.NewInt().WithMin(10).Any()

	testhelpers.MustBeInvalid(t, ruleSet, 9, errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet, 10, 10)
	testhelpers.MustBeValid(t, ruleSet, 11, 11)
}

func TestWithMinFloat(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithMin(10.0).Any()

	testhelpers.MustBeInvalid(t, ruleSet, 9.9, errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet, 10.0, 10.0)
	testhelpers.MustBeValid(t, ruleSet, 10.1, 10.1)
}
