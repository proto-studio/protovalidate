package testhelpers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

type MockT struct {
	testing.T

	errorCount int
}

func (t *MockT) Error(...any) {
	t.errorCount++
}

func (t *MockT) Errorf(string, ...any) {
	t.errorCount++
}

func TestMustBeValid(t *testing.T) {
	ruleSet := rules.Any()

	mockT := &MockT{}
	if _, err := testhelpers.MustRun(mockT, ruleSet, 10); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0, got: %d", mockT.errorCount)
	}

	ruleSet = ruleSet.WithRule(testhelpers.NewMockRuleWithErrors[any](1))

	mockT = &MockT{}
	if _, err := testhelpers.MustRun(mockT, ruleSet, 10); err == nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}
}

func TestMustBeValidFunc(t *testing.T) {
	ruleSet := rules.Any()
	callCount := 0

	checkValid := func(a, b any) error {
		callCount++
		return nil
	}

	checkInvalid := func(a, b any) error {
		callCount++
		return errors.New(errors.CodeUnknown, "", "")
	}

	mockT := &MockT{}
	if _, err := testhelpers.MustRunFunc(mockT, ruleSet, 10, 10, checkValid); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0, got: %d", mockT.errorCount)
	}
	if callCount != 1 {
		t.Errorf("Expected check function call count to be 1, got: %d", callCount)
	}

	callCount = 0
	mockT = &MockT{}

	if _, err := testhelpers.MustRunFunc(mockT, ruleSet, 10, 10, checkInvalid); err == nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}
	if callCount != 1 {
		t.Errorf("Expected check function call count to be 1, got: %d", callCount)
	}
}

func TestMustBeInvalid(t *testing.T) {
	ruleSet := rules.Any().WithRule(testhelpers.NewMockRuleWithErrors[any](1))

	mockT := &MockT{}
	if err := testhelpers.MustNotRun(mockT, ruleSet, 10, errors.CodeUnknown); err == nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0, got: %d", mockT.errorCount)
	}

	mockT = &MockT{}
	// Wrong code
	if err := testhelpers.MustNotRun(mockT, ruleSet, 10, errors.CodeMin); err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}

	ruleSet = rules.Any()

	mockT = &MockT{}
	// Is actually valid
	if err := testhelpers.MustNotRun(mockT, ruleSet, 10, errors.CodeUnknown); err != nil {
		t.Error("Expected error to not be nil")
	}
	if mockT.errorCount != 1 {
		t.Errorf("Expected error count to be 1, got: %d", mockT.errorCount)
	}
}
