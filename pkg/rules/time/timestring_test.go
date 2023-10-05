package time_test

import (
	"fmt"
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
func TestTimeStringRuleSet(t *testing.T) {
	now := internalTime.Now()

	tm, err := time.NewTimeString(internalTime.RFC3339).Validate(now)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if tm != now.Format(internalTime.RFC3339) {
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
func TestTimeStringRFC3339(t *testing.T) {
	s := "2023-09-29T18:57:42Z"

	ruleSet := time.NewTimeString(internalTime.RFC3339)
	testhelpers.MustBeValid(t, ruleSet.Any(), s, s)
}

// Requirements:
// - WithLayouts overrides default layout.
func TestTimeLayoutChange(t *testing.T) {
	s := "2023-09-29T18:57:42Z"

	ruleSet := time.NewTimeString(internalTime.RFC3339)
	testhelpers.MustBeValid(t, ruleSet.Any(), s, s)

	ruleSet = ruleSet.WithLayouts(internalTime.TimeOnly)
	testhelpers.MustBeInvalid(t, ruleSet.Any(), s, errors.CodeType)
}

// Requirements:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
func TestTimeStringRequired(t *testing.T) {
	ruleSet := time.NewTimeString(internalTime.RFC3339)

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestTimeStringCustom(t *testing.T) {
	now := internalTime.Now()
	s := now.Format(internalTime.RFC3339)

	ruleSet := time.NewTimeString(internalTime.RFC3339).WithRuleFunc(testhelpers.MockCustomRule(now, 1)).Any()
	testhelpers.MustBeInvalid(t, ruleSet, now, errors.CodeUnknown)

	ruleSet = time.NewTimeString(internalTime.RFC3339).WithRuleFunc(testhelpers.MockCustomRule(now, 0)).Any()
	testhelpers.MustBeValid(t, ruleSet, now, s)
}

func TestTimeStringAny(t *testing.T) {
	ruleSet := time.NewTimeString(internalTime.RFC3339).Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}

// Requirements:
// - Serializes the same as Time but with TimeString instead
// - Serialized value contains WithLayouts()
func TestTimeStringSerialize(t *testing.T) {
	ruleSet := time.NewTimeString(internalTime.RFC3339)

	expected := fmt.Sprintf("TimeStringRuleSet.WithLayouts(\"%s\")", internalTime.RFC3339)
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
