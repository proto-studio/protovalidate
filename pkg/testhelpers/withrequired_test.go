package testhelpers_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// MockNoRequired is a mock rule set that does not have a Required method
type MockNoRequired struct{ testhelpers.MockRuleSet[int] }

// MockNoWithRequired is a mock rule set that does not have a WithRequired method
type MockNoWithRequired struct{ testhelpers.MockRuleSet[int] }

func (m *MockNoWithRequired) Required() bool {
	return false
}

// MockRequiredAlwaysTrue is a mock rule set where Required always returns true
type MockRequiredAlwaysTrue struct{ testhelpers.MockRuleSet[int] }

func (m *MockRequiredAlwaysTrue) Required() bool {
	return true
}

func (m *MockRequiredAlwaysTrue) WithRequired() rules.RuleSet[int] {
	return &MockRequiredAlwaysTrueWithRequired{MockRuleSet: m.MockRuleSet}
}

// MockRequiredAlwaysTrueWithRequired is a mock rule set where Required always returns true and has WithRequired.
type MockRequiredAlwaysTrueWithRequired struct{ testhelpers.MockRuleSet[int] }

func (m *MockRequiredAlwaysTrueWithRequired) Required() bool {
	return true
}

// MockWithRequiredWrongReturnType is a mock rule set where WithRequired returns wrong type
type MockWithRequiredWrongReturnType struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredWrongReturnType) Required() bool {
	return false
}

func (m *MockWithRequiredWrongReturnType) WithRequired() string {
	return "wrong type"
}

// MockWithRequiredWrongReturnCount is a mock rule set where WithRequired returns wrong number of values
type MockWithRequiredWrongReturnCount struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredWrongReturnCount) Required() bool {
	return false
}

func (m *MockWithRequiredWrongReturnCount) WithRequired() (rules.RuleSet[int], string) {
	// Return 2 values instead of 1 - this is the bug we're testing
	return &MockWithRequiredWrongReturnCountWithRequired{MockRuleSet: m.MockRuleSet}, "extra value"
}

// MockWithRequiredWrongReturnCountWithRequired is a mock rule set where WithRequired returns wrong number of values.
type MockWithRequiredWrongReturnCountWithRequired struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredWrongReturnCountWithRequired) Required() bool {
	return true
}

// MockWithRequiredNotRequired is a mock rule set where WithRequired returns a rule set that is not required
type MockWithRequiredNotRequired struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredNotRequired) Required() bool {
	return false
}

func (m *MockWithRequiredNotRequired) WithRequired() rules.RuleSet[int] {
	return &MockWithRequiredNotRequiredWithRequired{MockRuleSet: m.MockRuleSet}
}

// MockWithRequiredNotRequiredWithRequired is a mock rule set where WithRequired returns a rule set that is not required.
type MockWithRequiredNotRequiredWithRequired struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredNotRequiredWithRequired) Required() bool {
	// Return false even though WithRequired was called - this is the bug we're testing
	return false
}

// TestMustImplementWithRequired tests:
// - MustImplementWithRequired correctly validates rule sets implement WithRequired
func TestMustImplementWithRequired(t *testing.T) {
	// Test with a real rule set that has WithRequired and Required - should pass
	mockT := &MockT{}
	testhelpers.MustImplementWithRequired[int](mockT, rules.Int())
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0 for valid rule set, got: %d", mockT.errorCount)
	}

	// Test with a rule set that doesn't have a Required method - should fail
	mockT = &MockT{}
	mockRuleSet := &MockNoRequired{}
	testhelpers.MustImplementWithRequired[int](mockT, mockRuleSet)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set without Required method, got: %d", mockT.errorCount)
	}

	// Test with a rule set that doesn't have a WithRequired method - should fail
	mockT = &MockT{}
	mockRuleSetNoWithRequired := &MockNoWithRequired{}
	testhelpers.MustImplementWithRequired[int](mockT, mockRuleSetNoWithRequired)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set without WithRequired method, got: %d", mockT.errorCount)
	}

	// Test with a rule set where Required always returns true - should fail
	mockT = &MockT{}
	mockRuleSetAlwaysTrue := &MockRequiredAlwaysTrue{}
	testhelpers.MustImplementWithRequired[int](mockT, mockRuleSetAlwaysTrue)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set where Required always returns true, got: %d", mockT.errorCount)
	}

	// Test with a rule set where WithRequired returns wrong type - should fail
	mockT = &MockT{}
	mockRuleSetWrongType := &MockWithRequiredWrongReturnType{}
	testhelpers.MustImplementWithRequired[int](mockT, mockRuleSetWrongType)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set with WithRequired returning wrong type, got: %d", mockT.errorCount)
	}

	// Test with a rule set where WithRequired returns wrong number of values - should fail
	mockT = &MockT{}
	mockRuleSetWrongCount := &MockWithRequiredWrongReturnCount{}
	testhelpers.MustImplementWithRequired[int](mockT, mockRuleSetWrongCount)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set with WithRequired returning wrong number of values, got: %d", mockT.errorCount)
	}

	// Test with a rule set where WithRequired returns a rule set that is not required - should fail
	mockT = &MockT{}
	mockRuleSetNotRequired := &MockWithRequiredNotRequired{}
	testhelpers.MustImplementWithRequired[int](mockT, mockRuleSetNotRequired)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set where WithRequired returns non-required rule set, got: %d", mockT.errorCount)
	}
}
