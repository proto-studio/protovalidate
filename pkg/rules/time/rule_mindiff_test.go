package time_test

import (
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/time"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMinDiff(t *testing.T) {
	now := internalTime.Now()
	before14 := now.Add(-14 * internalTime.Minute)
	before16 := now.Add(-16 * internalTime.Minute)

	ruleSet := time.NewTime().WithMinDiff(-15 * internalTime.Minute).Any()

	testhelpers.MustBeInvalid(t, ruleSet, before16, errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet, before14, before14)
}

func TestStringWithMinDiff(t *testing.T) {
	now := internalTime.Now()
	before14 := now.Add(-14 * internalTime.Minute)
	before16 := now.Add(-16 * internalTime.Minute)

	ruleSet := time.NewTimeString(internalTime.RFC3339).WithMinDiff(-15 * internalTime.Minute).Any()

	testhelpers.MustBeInvalid(t, ruleSet, before16, errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet, before14, before14.Format(internalTime.RFC3339))
}
