package time_test

import (
	"context"
	"fmt"
	"testing"
	internalTime "time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/time"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// TestTimeRuleSet_Apply tests:
// - Default configuration doesn't return errors on valid value.
// - Implements interface.
func TestTimeRuleSet_Apply(t *testing.T) {
	now := internalTime.Now()

	// Prepare an output variable for Apply
	var output internalTime.Time

	// Use Apply to validate the current time
	err := time.Time().Apply(context.TODO(), now, &output)

	if err != nil {
		t.Fatal("Expected errors to be empty")
	}

	if output != now {
		t.Fatal("Expected test time to be returned")
	}

	// Check if the rule set implements the expected interface
	ok := testhelpers.CheckRuleSetInterface[internalTime.Time](time.Time())
	if !ok {
		t.Fatal("Expected rule set to be implemented")
	}

	testhelpers.MustApplyTypes[internalTime.Time](t, time.Time(), now)
}

// TestTimeRuleSet_Apply_RFC3339 tests:
// - Will coerce time from RFC 3339
func TestTimeRuleSet_Apply_RFC3339(t *testing.T) {
	s := "2023-09-29T18:57:42.108Z"

	tm, err := internalTime.Parse(internalTime.RFC3339, s)
	if err != nil {
		t.Fatalf("Unable to parse test string: %s", err)
	}

	ruleSet := time.Time()
	testhelpers.MustNotApply(t, ruleSet.Any(), s, errors.CodeType)

	ruleSet = ruleSet.WithLayouts(internalTime.RFC3339)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), s, tm)
}

// TestTimeRuleSet_Apply_MultiLayout tests:
// - Will coerce from multiple layouts
func TestTimeRuleSet_Apply_MultiLayout(t *testing.T) {
	s := "2023-09-29"

	tm, err := internalTime.Parse(internalTime.DateOnly, s)
	if err != nil {
		t.Fatalf("Unable to parse test string: %s", err)
	}

	ruleSet := time.Time().WithLayouts(internalTime.RFC3339)
	testhelpers.MustNotApply(t, ruleSet.Any(), s, errors.CodeType)

	ruleSet = ruleSet.WithLayouts(internalTime.RFC3339, internalTime.DateOnly)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), s, tm)

	ruleSet = ruleSet.WithLayouts(internalTime.DateOnly, internalTime.RFC3339)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), s, tm)
}

// TestTimeRuleSet_WithRequired tests:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
func TestTimeRuleSet_WithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[internalTime.Time](t, time.Time())
}

// TestTimeCustom tests:
func TestTimeCustom(t *testing.T) {
	now := internalTime.Now()

	ruleSet := time.Time().WithRuleFunc(testhelpers.NewMockRuleWithErrors[internalTime.Time](1).Function()).Any()
	testhelpers.MustNotApply(t, ruleSet, now, errors.CodeUnknown)

	rule := testhelpers.NewMockRule[internalTime.Time]()
	ruleSet = time.Time().WithRuleFunc(rule.Function()).Any()
	testhelpers.MustApply(t, ruleSet, now)

	if c := rule.EvaluateCallCount(); c != 1 {
		t.Errorf("Expected rule to be called once, got %d", c)
		return
	}
}

// TestTimeRuleSet_Any tests:
func TestTimeRuleSet_Any(t *testing.T) {
	ruleSet := time.Time().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}

// TestTimeRuleSet_Apply_Pointer tests:
func TestTimeRuleSet_Apply_Pointer(t *testing.T) {
	now := internalTime.Now()

	ruleSet := time.Time()
	testhelpers.MustApplyMutation(t, ruleSet.Any(), &now, now)
}

// TestTimeRuleSet_Apply_BadType tests:
func TestTimeRuleSet_Apply_BadType(t *testing.T) {
	ruleSet := time.Time()
	type x struct{}

	testhelpers.MustNotApply(t, ruleSet.Any(), new(x), errors.CodeType)
}

// TestTimeRuleSet_String_WithLayouts tests:
// - WithLayouts will serialize up to 3 layouts.
// - Layouts are comma separated.
// - Layout values are quoted.
// - If there are more than 3, the test " ... and X more" is used.
func TestTimeRuleSet_String_WithLayouts(t *testing.T) {
	layouts := []string{
		internalTime.DateOnly,
		internalTime.TimeOnly,
		internalTime.Stamp,
		internalTime.RFC3339,
		internalTime.RFC1123,
	}

	ruleSet := time.Time().WithLayouts(layouts[0], layouts[1])
	expected := fmt.Sprintf("TimeRuleSet.WithLayouts(\"%s\", \"%s\")", layouts[0], layouts[1])
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = time.Time().WithLayouts(layouts[0], layouts[1:3]...)
	expected = fmt.Sprintf("TimeRuleSet.WithLayouts(\"%s\", \"%s\", \"%s\")", layouts[0], layouts[1], layouts[2])
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}

	ruleSet = time.Time().WithLayouts(layouts[0], layouts[1:]...)
	expected = fmt.Sprintf("TimeRuleSet.WithLayouts(\"%s\", \"%s\", \"%s\" ... and 2 more)", layouts[0], layouts[1], layouts[2])
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestTimeRuleSet_String_WithRequired tests:
// - Serializes to WithRequired()
func TestTimeRuleSet_String_WithRequired(t *testing.T) {
	ruleSet := time.Time().WithRequired()

	expected := "TimeRuleSet.WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestTimeRuleSet_Apply_String tests:
// - Apply must convert to time.RFC3339 if output is a string.
// - Apply must maintain input format if output and input are strings.
// - Apply must allow the user to override the string output format.
// - WithOutputLayout is idempotent.
func TestTimeRuleSet_Apply_String(t *testing.T) {
	now := internalTime.Now()
	ctx := context.TODO()

	rfcTime := now.Format(internalTime.RFC3339)
	dateOnly := now.Format(internalTime.DateOnly)

	ruleSet := time.Time()

	var output string
	errs := ruleSet.Apply(ctx, now, &output)
	if errs != nil {
		t.Errorf("Expected errors to be nil, got: %s", errs)
	} else if output != rfcTime {
		t.Errorf(`Expected output to be "%s", got "%s"`, rfcTime, output)
	}
	ruleSet = ruleSet.WithLayouts(internalTime.RFC3339, internalTime.DateOnly)

	errs = ruleSet.Apply(ctx, dateOnly, &output)
	if errs != nil {
		t.Errorf("Expected errors to be nil, got: %s", errs)
	} else if output != dateOnly {
		t.Errorf(`Expected output to be "%s", got "%s"`, dateOnly, output)
	}

	ruleSetWithOuputLayout := ruleSet.WithOutputLayout(internalTime.DateOnly)

	if ruleSet == ruleSetWithOuputLayout {
		t.Errorf("Expected ruleSetWithOuputLayout to not equal ruleSet")
	}

	errs = ruleSetWithOuputLayout.Apply(ctx, rfcTime, &output)
	if errs != nil {
		t.Errorf("Expected errors to be nil, got: %s", errs)
	} else if output != dateOnly {
		t.Errorf(`Expected output to be "%s", got "%s"`, dateOnly, output)
	}

	ruleSet = ruleSetWithOuputLayout.WithOutputLayout(internalTime.DateOnly)

	if ruleSet != ruleSetWithOuputLayout {
		t.Errorf("Expected ruleSetWithOuputLayout to equal ruleSet")
	}

}

// TestTimeRuleSet_WithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestTimeRuleSet_WithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[internalTime.Time](t, time.Time())
}
