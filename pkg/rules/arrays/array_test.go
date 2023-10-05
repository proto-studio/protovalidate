package arrays_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/arrays"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestArrayRuleSet(t *testing.T) {
	arr, err := arrays.New[string]().Validate([]string{"a", "b", "c"})

	if err != nil {
		t.Errorf("Expected errors to be empty. Got: %v", err)
		return
	}

	if len(arr) != 3 {
		t.Errorf("Expected returned array to have length 3 but got %d", len(arr))
		return
	}

	ok := testhelpers.CheckRuleSetInterface[[]string](arrays.New[string]())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

func TestArrayRuleSetTypeError(t *testing.T) {
	_, err := arrays.New[string]().Validate(123)

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func TestArrayItemRuleSetSuccess(t *testing.T) {
	_, err := arrays.New[string]().WithItemRuleSet(strings.New()).Validate([]string{"a", "b", "c"})

	if err != nil {
		t.Errorf("Expected errors to be empty. Got: %v", err)
		return
	}
}

func TestArrayItemCastError(t *testing.T) {
	_, err := arrays.New[string]().Validate([]int{1, 2, 3})

	if len(err) == 0 {
		t.Errorf("Expected errors to not be empty.")
		return
	}
}

func TestArrayItemRuleSetError(t *testing.T) {
	_, err := arrays.New[string]().WithItemRuleSet(strings.New().WithMinLen(2)).Validate([]string{"", "a", "ab", "abc"})

	if len(err) != 2 {
		t.Errorf("Expected 2 errors and got %d.", len(err))
		return
	}
}

func TestWithRequired(t *testing.T) {
	ruleSet := arrays.New[string]()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestCustom(t *testing.T) {
	_, err := arrays.New[int]().
		WithRuleFunc(testhelpers.MockCustomRule([]int{1, 2, 3}, 1)).
		WithRuleFunc(testhelpers.MockCustomRule([]int{1, 2, 3}, 1)).
		Validate([]int{1, 2, 3})

	if err == nil {
		t.Error("Expected errors to not be nil")
		return
	}

	if len(err) == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func TestReturnsCorrectPaths(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "myarray")
	_, err := arrays.New[string]().WithItemRuleSet(strings.New().WithMinLen(2)).ValidateWithContext([]string{"", "a", "ab", "abc"}, ctx)

	if err == nil {
		t.Errorf("Expected errors to not be nil")
	} else if len(err) != 2 {
		t.Errorf("Expected 2 errors got %d: %s", len(err), err.Error())
		return
	}

	errA := err.For("/myarray/0")
	if errA == nil {
		t.Errorf("Expected error for /myarray/0 to not be nil")
	} else if len(errA) != 1 {
		t.Errorf("Expected exactly 1 error for /myarray/0 got %d", len(err))
	} else if errA.First().Path() != "/myarray/0" {
		t.Errorf("Expected error path to be `%s` got `%s`", "/myarray/0", errA.First().Path())
	}

	errC := err.For("/myarray/1")
	if errC == nil {
		t.Errorf("Expected error for /myarray/1 to not be nil")
	} else if len(errC) != 1 {
		t.Errorf("Expected exactly 1 error for /myarray/1 got %d", len(err))
	} else if errC.First().Path() != "/myarray/1" {
		t.Errorf("Expected error path to be `%s` got `%s`", "/myarray/1", errC.First().Path())
	}
}

func TestAny(t *testing.T) {
	ruleSet := arrays.New[int]().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestRequiredString(t *testing.T) {
	ruleSet := arrays.New[int]().WithRequired()

	expected := "ArrayRuleSet[int].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithItemRuleSet()
func TestWithItemRuleSetString(t *testing.T) {
	ruleSet := arrays.New[int]().WithItemRuleSet(numbers.NewInt().WithMin(2))

	expected := "ArrayRuleSet[int].WithItemRuleSet(IntRuleSet[int].WithMin(2))"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}
