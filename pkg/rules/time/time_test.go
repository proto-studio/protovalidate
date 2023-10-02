package time_test

import (
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/time"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Default configuration doesn't return errors on valid value.
// - Implements interface.
func TestTimeRuleSet(t *testing.T) {
	now := internalTime.Now()

	tm, err := time.NewTime().Validate(now)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if tm != now {
		t.Error("Expected test time to be returned")
		return
	}

	ok := testhelpers.CheckRuleSetInterface[internalTime.Time](time.NewTime())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

// Requirements:
// - Will coerce time from RFC 3339
func TestTimeRFC3339(t *testing.T) {
	s := "2023-09-29T18:57:42.108Z"

	tm, err := internalTime.Parse(internalTime.RFC3339, s)
	if err != nil {
		t.Fatalf("Unable to parse test string: %s", err)
	}

	ruleSet := time.NewTime()
	testhelpers.MustBeInvalid(t, ruleSet.Any(), s, errors.CodeType)

	ruleSet = ruleSet.WithLayouts(internalTime.RFC3339)
	testhelpers.MustBeValid(t, ruleSet.Any(), s, tm)
}

// Requirements:
// - Will coerce from multiple layouts
func TestTimeMultiLayout(t *testing.T) {
	s := "2023-09-29"

	tm, err := internalTime.Parse(internalTime.DateOnly, s)
	if err != nil {
		t.Fatalf("Unable to parse test string: %s", err)
	}

	ruleSet := time.NewTime().WithLayouts(internalTime.RFC3339)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), s, errors.CodeType)

	ruleSet = ruleSet.WithLayouts(internalTime.RFC3339, internalTime.DateOnly)
	testhelpers.MustBeValid(t, ruleSet.Any(), s, tm)

	ruleSet = ruleSet.WithLayouts(internalTime.DateOnly, internalTime.RFC3339)
	testhelpers.MustBeValid(t, ruleSet.Any(), s, tm)
}

// Requirements:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
func TestTimeRequired(t *testing.T) {
	ruleSet := time.NewTime()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestTimeCustom(t *testing.T) {
	now := internalTime.Now()

	ruleSet := time.NewTime().WithRuleFunc(testhelpers.MockCustomRule(now, 1)).Any()
	testhelpers.MustBeInvalid(t, ruleSet, now, errors.CodeUnknown)

	ruleSet = time.NewTime().WithRuleFunc(testhelpers.MockCustomRule(now, 0)).Any()
	testhelpers.MustBeValid(t, ruleSet, now, now)
}

func TestTimeAny(t *testing.T) {
	ruleSet := time.NewTime().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}

func TestTimePointer(t *testing.T) {
	now := internalTime.Now()

	ruleSet := time.NewTime()
	testhelpers.MustBeValid(t, ruleSet.Any(), &now, now)
}

func TestBadType(t *testing.T) {
	ruleSet := time.NewTime()
	type x struct{}

	testhelpers.MustBeInvalid(t, ruleSet.Any(), new(x), errors.CodeType)
}
