package time_test

import (
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/time"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMinTime(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.NewTime().WithMin(now).Any()

	testhelpers.MustBeInvalid(t, ruleSet, before, errors.CodeMin)

	testhelpers.MustBeValid(t, ruleSet, now, now)
	testhelpers.MustBeValid(t, ruleSet, after, after)
}

func TestWithMinTimeString(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.NewTimeString(internalTime.RFC3339).WithMin(now).Any()

	testhelpers.MustBeInvalid(t, ruleSet, before, errors.CodeMin)

	testhelpers.MustBeValid(t, ruleSet, now, now.Format(internalTime.RFC3339))
	testhelpers.MustBeValid(t, ruleSet, after, after.Format(internalTime.RFC3339))
}
