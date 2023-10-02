package time_test

import (
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/time"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMaxTime(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.NewTime().WithMax(now).Any()

	testhelpers.MustBeValid(t, ruleSet, before, before)
	testhelpers.MustBeValid(t, ruleSet, now, now)
	testhelpers.MustBeInvalid(t, ruleSet, after, errors.CodeMax)
}

func TestWithMaxTimeString(t *testing.T) {
	now := internalTime.Now()
	before := now.Add(-1 * internalTime.Minute)
	after := now.Add(1 * internalTime.Minute)

	ruleSet := time.NewTimeString(internalTime.RFC3339).WithMax(now).Any()

	testhelpers.MustBeValid(t, ruleSet, before, before.Format(internalTime.RFC3339))
	testhelpers.MustBeValid(t, ruleSet, now, now.Format(internalTime.RFC3339))
	testhelpers.MustBeInvalid(t, ruleSet, after, errors.CodeMax)
}
