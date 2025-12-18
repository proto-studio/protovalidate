package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestSliceRuleSet(t *testing.T) {
	// Prepare an output variable for Apply
	var output []string

	// Apply with a valid array, expecting no error
	err := rules.Slice[string]().Apply(context.TODO(), []string{"a", "b", "c"}, &output)
	if err != nil {
		t.Fatalf("Expected errors to be empty. Got: %v", err)
	}

	if len(output) != 3 {
		t.Fatalf("Expected returned array to have length 3 but got %d", len(output))
	}

	// Check if the rule set implements the expected interface
	ok := testhelpers.CheckRuleSetInterface[[]string](rules.Slice[string]())
	if !ok {
		t.Fatalf("Expected rule set to be implemented")
	}

	testhelpers.MustApplyTypes[[]string](t, rules.Slice[string](), []string{"a", "b", "c"})
}

func TestSliceRuleSetTypeError(t *testing.T) {
	// Prepare an output variable for Apply
	var output []string

	// Apply with an invalid input type, expecting an error
	err := rules.Slice[string]().Apply(context.TODO(), 123, &output)
	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func TestSliceItemRuleSetSuccess(t *testing.T) {
	// Prepare an output variable for Apply
	var output []string

	// Apply with a valid array and item rule set, expecting no error
	err := rules.Slice[string]().WithItemRuleSet(rules.String()).Apply(context.TODO(), []string{"a", "b", "c"}, &output)
	if err != nil {
		t.Errorf("Expected errors to be empty. Got: %v", err)
		return
	}
}

func TestSliceItemCastError(t *testing.T) {
	// Prepare an output variable for Apply
	var output []string

	// Apply with an array of incorrect types, expecting an error
	err := rules.Slice[string]().Apply(context.TODO(), []int{1, 2, 3}, &output)
	if len(err) == 0 {
		t.Errorf("Expected errors to not be empty.")
		return
	}
}

func TestSliceItemRuleSetError(t *testing.T) {
	// Prepare an output variable for Apply
	var output []string

	// Apply with a valid array but with an item rule set that will fail, expecting 2 errors
	err := rules.Slice[string]().WithItemRuleSet(rules.String().WithMinLen(2)).Apply(context.TODO(), []string{"", "a", "ab", "abc"}, &output)
	if len(err) != 2 {
		t.Errorf("Expected 2 errors and got %d.", len(err))
		return
	}
}

func TestWithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[[]string](t, rules.Slice[string]())
}

func TestCustom(t *testing.T) {
	mock := testhelpers.NewMockRuleWithErrors[[]int](1)

	// Prepare an output variable for Apply
	var output []int

	// Apply with the mock rules, expecting errors
	err := rules.Slice[int]().
		WithRuleFunc(mock.Function()).
		WithRuleFunc(mock.Function()).
		Apply(context.TODO(), []int{1, 2, 3}, &output)

	if err == nil {
		t.Error("Expected errors to not be nil")
		return
	}

	if len(err) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(err))
		return
	}

	if mock.EvaluateCallCount() != 2 {
		t.Errorf("Expected rule to be called 2 times, got %d", mock.EvaluateCallCount())
		return
	}
}

func TestReturnsCorrectPaths(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "myarray")

	// Prepare an output variable for Apply
	var output []string

	// Apply with an array and a context, expecting errors
	err := rules.Slice[string]().
		WithItemRuleSet(rules.String().WithMinLen(2)).
		Apply(ctx, []string{"", "a", "ab", "abc"}, &output)

	if err == nil {
		t.Errorf("Expected errors to not be nil")
	} else if len(err) != 2 {
		t.Errorf("Expected 2 errors got %d: %s", len(err), err.Error())
		return
	}

	// Check for the first error path (/myarray/0)
	errA := err.For("/myarray/0")
	if errA == nil {
		t.Errorf("Expected error for /myarray/0 to not be nil")
	} else if len(errA) != 1 {
		t.Errorf("Expected exactly 1 error for /myarray/0 got %d", len(errA))
	} else if errA.First().Path() != "/myarray/0" {
		t.Errorf("Expected error path to be `%s` got `%s`", "/myarray/0", errA.First().Path())
	}

	// Check for the second error path (/myarray/1)
	errC := err.For("/myarray/1")
	if errC == nil {
		t.Errorf("Expected error for /myarray/1 to not be nil")
	} else if len(errC) != 1 {
		t.Errorf("Expected exactly 1 error for /myarray/1 got %d", len(errC))
	} else if errC.First().Path() != "/myarray/1" {
		t.Errorf("Expected error path to be `%s` got `%s`", "/myarray/1", errC.First().Path())
	}
}

func TestAny(t *testing.T) {
	ruleSet := rules.Slice[int]().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestRequiredSlice(t *testing.T) {
	ruleSet := rules.Slice[int]().WithRequired()

	expected := "SliceRuleSet[int].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithItemRuleSet()
func TestWithItemRuleSetString(t *testing.T) {
	ruleSet := rules.Slice[int]().WithItemRuleSet(rules.Int().WithMin(2))

	expected := "SliceRuleSet[int].WithItemRuleSet(IntRuleSet[int].WithMin(2))"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Evaluate behaves like ValidateWithContext
func TestEvaluate(t *testing.T) {
	v := []int{123, 456}
	ctx := context.Background()

	ruleSet := rules.Slice[int]().WithItemRuleSet(rules.Int().WithMin(2))

	// Evaluate the array directly using Evaluate
	err1 := ruleSet.Evaluate(ctx, v)

	// Prepare an output variable for Apply
	var output []int

	// Validate the array using Apply
	err2 := ruleSet.Apply(ctx, v, &output)

	// Check if both methods result in no errors
	if err1 != nil || err2 != nil {
		t.Errorf("Expected errors to both be nil, got %s and %s", err1, err2)
	}
}

// Requirements:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestSliceWithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[[]string](t, rules.Slice[string]())
}
