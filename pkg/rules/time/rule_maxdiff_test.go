package time_test

import (
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/time"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMaxDiff(t *testing.T) {
	now := internalTime.Now()
	before14 := now.Add(-14 * internalTime.Minute)
	before16 := now.Add(-16 * internalTime.Minute)

	ruleSet := time.NewTime().WithMaxDiff(-15 * internalTime.Minute).Any()

	testhelpers.MustBeInvalid(t, ruleSet, before14, errors.CodeMax)
	testhelpers.MustBeValid(t, ruleSet, before16, before16)
}

func TestStringWithMaxDiff(t *testing.T) {
	now := internalTime.Now()
	before14 := now.Add(-14 * internalTime.Minute)
	before16 := now.Add(-16 * internalTime.Minute)

	ruleSet := time.NewTimeString(internalTime.RFC3339).WithMaxDiff(-15 * internalTime.Minute).Any()

	testhelpers.MustBeInvalid(t, ruleSet, before14, errors.CodeMax)
	testhelpers.MustBeValid(t, ruleSet, before16, before16.Format(internalTime.RFC3339))
}
