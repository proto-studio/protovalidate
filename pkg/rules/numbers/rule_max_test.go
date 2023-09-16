package numbers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMaxInt(t *testing.T) {
	ruleSet := numbers.NewInt().WithMax(10).Any()

	testhelpers.MustBeValid(t, ruleSet, 9, 9)
	testhelpers.MustBeValid(t, ruleSet, 10, 10)
	testhelpers.MustBeInvalid(t, ruleSet, 11, errors.CodeMax)
}

func TestWithMaxFloat(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithMax(10.0).Any()

	testhelpers.MustBeValid(t, ruleSet, 9.9, 9.9)
	testhelpers.MustBeValid(t, ruleSet, 10.0, 10.0)
	testhelpers.MustBeInvalid(t, ruleSet, 10.1, errors.CodeMax)
}
