package testhelpers_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestMockRule(t *testing.T) {
	ctx := context.Background()

	rule1 := testhelpers.NewMockRule[any]().Function()

	err := rule1(ctx, 456)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}

	rule2 := testhelpers.NewMockRuleWithErrors[any](1).Function()

	err = rule2(ctx, 456)
	if err == nil {
		t.Error("Expected error to not be nil")
	} else if s := len(err); s != 1 {
		t.Errorf("Expected error collection size to be %d, got: %d", 1, s)
	}

	rule3 := testhelpers.NewMockRuleWithErrors[any](2).Function()

	err = rule3(ctx, 456)
	if err == nil {
		t.Error("Expected error to not be nil")
	} else if s := len(err); s != 2 {
		t.Errorf("Expected error collection size to be %d, got: %d", 2, s)
	}
}

func TestMockConflict(t *testing.T) {
	mockA := testhelpers.NewMockRule[int]()
	mockA.ConflictKey = "a"

	mockB := testhelpers.NewMockRule[int]()
	mockB.ConflictKey = "b"

	mockAA := testhelpers.NewMockRule[int]()
	mockAA.ConflictKey = "a"

	var mockD rules.RuleFunc[int] = testhelpers.NewMockRule[int]().Function()

	if mockA.Conflict(mockB) {
		t.Errorf("Expected mockA and mockB to not conflict")
	}

	if !mockA.Conflict(mockAA) {
		t.Errorf("Expected mockA and mockAA to conflict")
	}

	if mockA.Conflict(mockD) {
		t.Errorf("Expected mockA and mockD to not conflict")
	}
}

func TestMockString(t *testing.T) {
	str := testhelpers.NewMockRule[int]().String()

	if str != "WithMock()" {
		t.Errorf("Expected mock string to be `%s`, got: %s", "WithMock()", str)
	}
}

func TestMockCounter(t *testing.T) {
	mock := testhelpers.NewMockRule[int]()
	ctx := context.Background()

	if mock.EvaluateCallCount() != 0 {
		t.Errorf("Expected call count to be %d, got %d", 0, mock.EvaluateCallCount())
	}

	mock.Evaluate(ctx, 1)

	if mock.EvaluateCallCount() != 1 {
		t.Errorf("Expected call count to be %d, got %d", 1, mock.EvaluateCallCount())
	}

	mock.Evaluate(ctx, 1)
	mock.Evaluate(ctx, 1)

	if mock.EvaluateCallCount() != 3 {
		t.Errorf("Expected call count to be %d, got %d", 3, mock.EvaluateCallCount())
	}

	mock.Reset()

	if mock.EvaluateCallCount() != 0 {
		t.Errorf("Expected call count to be %d, got %d", 0, mock.EvaluateCallCount())
	}
}

func TestMockRuleSet_Apply(t *testing.T) {
	ctx := context.Background()

	// Assigning mockRuleSet.OutputValue to output
	var outputInt int
	overrideValue := 456
	mockRuleSet := &testhelpers.MockRuleSet[int]{OutputValue: &overrideValue}
	err := mockRuleSet.Apply(ctx, 123, &outputInt)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if outputInt != overrideValue { // Should be 456 from mockRuleSet.OutputValue
		t.Errorf("expected outputInt to be 456, got %d", outputInt)
	}

	// mockRuleSet.OutputValue is nil, fallback to input assignment
	mockRuleSet = &testhelpers.MockRuleSet[int]{}
	err = mockRuleSet.Apply(ctx, 789, &outputInt)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if outputInt != 789 { // Should be 789 from input
		t.Errorf("expected outputInt to be 789, got %d", outputInt)
	}

	// Output is not a pointer
	outputNonPtr := 0
	err = mockRuleSet.Apply(ctx, 123, outputNonPtr)
	if err == nil {
		t.Errorf("expected an error when output is not a pointer, got nil")
	}

	// Error case when mockRuleSet.OutputValue is not assignable to output
	var outputMismatch string
	mockRuleSet = &testhelpers.MockRuleSet[int]{OutputValue: &overrideValue}
	err = mockRuleSet.Apply(ctx, 123, &outputMismatch)
	if err == nil {
		t.Errorf("expected an error when mockRuleSet.OutputValue is not assignable to output, got nil")
	}

	// Error case when input value is not assignable to output and mockRuleSet.OutputValue is nil
	mockRuleSet = &testhelpers.MockRuleSet[int]{}
	err = mockRuleSet.Apply(ctx, 123, &outputMismatch)
	if err == nil {
		t.Errorf("expected an error when mockRuleSet.OutputValue is not assignable to output, got nil")
	}

	// Output is nil
	mockRuleSet = &testhelpers.MockRuleSet[int]{}
	err = mockRuleSet.Apply(ctx, 123, nil)
	if err == nil {
		t.Errorf("expected non-nil error")
	}
}
