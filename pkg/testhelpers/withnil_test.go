package testhelpers_test

import (
	"context"
	"reflect"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// MockNoWithNilMethod is a mock rule set that has NO WithNil method at all
// It implements RuleSet[int] directly to avoid inheriting WithNil from MockRuleSet
type MockNoWithNilMethod struct{}

func (m *MockNoWithNilMethod) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	// Set output for valid input
	if outputPtr, ok := output.(*int); ok && outputPtr != nil {
		if val, ok := input.(int); ok {
			*outputPtr = val
		}
	}
	return nil
}

func (m *MockNoWithNilMethod) Evaluate(ctx context.Context, value int) errors.ValidationErrorCollection {
	return nil
}

func (m *MockNoWithNilMethod) Any() rules.RuleSet[any] {
	return &mockNoWithNilMethodAny{m}
}

func (m *MockNoWithNilMethod) Replaces(other rules.Rule[int]) bool {
	_, ok := other.(*MockNoWithNilMethod)
	return ok
}

func (m *MockNoWithNilMethod) Required() bool {
	return false
}

func (m *MockNoWithNilMethod) String() string {
	return "MockNoWithNilMethod"
}

// mockNoWithNilMethodAny wraps MockNoWithNilMethod to implement RuleSet[any]
type mockNoWithNilMethodAny struct{ inner *MockNoWithNilMethod }

func (m *mockNoWithNilMethodAny) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	return m.inner.Apply(ctx, input, output)
}

func (m *mockNoWithNilMethodAny) Evaluate(ctx context.Context, value any) errors.ValidationErrorCollection {
	return nil
}

func (m *mockNoWithNilMethodAny) Any() rules.RuleSet[any] {
	return m
}

func (m *mockNoWithNilMethodAny) Replaces(other rules.Rule[any]) bool {
	return false
}

func (m *mockNoWithNilMethodAny) Required() bool {
	return false
}

func (m *mockNoWithNilMethodAny) String() string {
	return m.inner.String()
}

// Note: MockNoWithNilMethod intentionally does NOT have a WithNil method

// MockNoWithNil is a mock rule set that has a broken WithNil method (returns itself unchanged)
type MockNoWithNil struct{ testhelpers.MockRuleSet[int] }

// WithNil returns itself without actually enabling nil handling, simulating a broken implementation
func (m *MockNoWithNil) WithNil() rules.RuleSet[int] {
	return m // Returns itself, not a version with nil handling enabled
}

// MockWrongNilErrorCode is a mock rule set that returns wrong error code for nil
type MockWrongNilErrorCode struct{ testhelpers.MockRuleSet[int] }

func (m *MockWrongNilErrorCode) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Return correct CodeNull for the first test (without WithNil)
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWrongNilErrorCode) WithNil() rules.RuleSet[int] {
	return &MockWrongNilErrorCodeWithNil{MockRuleSet: m.MockRuleSet}
}

// MockWrongNilErrorCodeWithNil is a mock rule set that returns wrong error code for nil even with WithNil.
type MockWrongNilErrorCodeWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockWrongNilErrorCodeWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Return wrong error code instead of CodeNull, and don't set output to nil
		return errors.Collection(errors.Errorf(errors.CodeUnknown, ctx, "unknown error", "value cannot be null"))
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

// MockNilNotSet is a mock rule set that doesn't set output to nil when WithNil is used
type MockNilNotSet struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilNotSet) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Return correct CodeNull for the first test (without WithNil)
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockNilNotSet) WithNil() rules.RuleSet[int] {
	return &MockNilNotSetWithNil{MockRuleSet: m.MockRuleSet}
}

// MockNilNotSetWithNil is a mock rule set that doesn't set output to nil when WithNil is used.
type MockNilNotSetWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilNotSetWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Don't set output to nil, just return success without setting output
		// This simulates a bug where WithNil is used but output isn't actually set to nil
		// We validate output is a pointer but don't set it to nil
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() != reflect.Ptr || outputVal.IsNil() {
			return errors.Collection(errors.Errorf(errors.CodeInternal, ctx, "internal error", "Output must be a non-nil pointer"))
		}
		// Intentionally don't set output to nil - this is the bug we're testing
		return nil
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

// MockNilWrongReturnType is a mock rule set where WithNil returns wrong type
type MockNilWrongReturnType struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilWrongReturnType) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Return correct CodeNull for the first test (without WithNil)
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockNilWrongReturnType) WithNil() string {
	return "wrong type"
}

// MockNilNoError is a mock rule set that doesn't return an error when nil is provided without WithNil
type MockNilNoError struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilNoError) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Don't return an error - this is the bug we're testing
		return nil
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockNilNoError) WithNil() rules.RuleSet[int] {
	return &MockNilNoErrorWithNil{MockRuleSet: m.MockRuleSet}
}

// MockNilNoErrorWithNil is a mock rule set that doesn't return an error when nil is provided even with WithNil.
type MockNilNoErrorWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilNoErrorWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Set output to nil correctly
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
			elem := outputVal.Elem()
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Map || elem.Kind() == reflect.Chan || elem.Kind() == reflect.Func {
				elem.Set(reflect.Zero(elem.Type()))
			}
		}
		return nil
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

// MockNilWrongCodeWithoutWithNil is a mock rule set that returns wrong error code when nil is provided without WithNil
type MockNilWrongCodeWithoutWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilWrongCodeWithoutWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Return wrong error code instead of CodeNull
		return errors.Collection(errors.Errorf(errors.CodeUnknown, ctx, "unknown error", "value cannot be null"))
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockNilWrongCodeWithoutWithNil) WithNil() rules.RuleSet[int] {
	return &MockNilWrongCodeWithoutWithNilWithNil{MockRuleSet: m.MockRuleSet}
}

// MockNilWrongCodeWithoutWithNilWithNil is a mock rule set that returns wrong error code when nil is provided even with WithNil.
type MockNilWrongCodeWithoutWithNilWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilWrongCodeWithoutWithNilWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Set output to nil correctly
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
			elem := outputVal.Elem()
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Map || elem.Kind() == reflect.Chan || elem.Kind() == reflect.Func {
				elem.Set(reflect.Zero(elem.Type()))
			}
		}
		return nil
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

// MockNilWrongReturnCount is a mock rule set where WithNil returns wrong number of values
type MockNilWrongReturnCount struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilWrongReturnCount) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Return correct CodeNull for the first test (without WithNil)
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockNilWrongReturnCount) WithNil() (rules.RuleSet[int], string) {
	// Return 2 values instead of 1 - this is the bug we're testing
	return &MockNilWrongReturnCountWithNil{MockRuleSet: m.MockRuleSet}, "extra value"
}

// MockNilWrongReturnCountWithNil is a mock rule set where WithNil returns wrong number of values.
type MockNilWrongReturnCountWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockNilWrongReturnCountWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Set output to nil correctly
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
			elem := outputVal.Elem()
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Map || elem.Kind() == reflect.Chan || elem.Kind() == reflect.Func {
				elem.Set(reflect.Zero(elem.Type()))
			}
		}
		return nil
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

// TestMustImplementWithNil tests:
// - MustImplementWithNil correctly validates rule sets implement WithNil
func TestMustImplementWithNil(t *testing.T) {
	// Test with a real rule set that has WithNil - should pass
	mockT := &MockT{}
	testhelpers.MustImplementWithNil[int](mockT, rules.Int())
	if mockT.errorCount != 0 {
		t.Errorf("Expected error count to be 0 for valid rule set, got: %d", mockT.errorCount)
	}

	// Test with a rule set that has NO WithNil method at all - should fail
	mockT = &MockT{}
	mockRuleSetNoMethod := &MockNoWithNilMethod{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetNoMethod)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set without WithNil method, got: %d", mockT.errorCount)
	}

	// Test with a rule set that has broken WithNil method - should fail
	// It returns 2 errors: one for returning an error when it should succeed, and one for not setting output to nil
	mockT = &MockT{}
	mockRuleSet := &MockNoWithNil{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSet)
	if mockT.errorCount != 2 {
		t.Errorf("Expected 2 errors on rule set with broken WithNil method, got: %d", mockT.errorCount)
	}

	// Test with a rule set that returns wrong error code for nil when WithNil is used - should fail
	// This will produce 2 errors: one for returning an error when it should succeed, and one for not setting output to nil
	mockT = &MockT{}
	mockRuleSetWrongCode := &MockWrongNilErrorCode{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWrongCode)
	if mockT.errorCount != 2 {
		t.Errorf("Expected 2 errors on rule set with wrong error code when WithNil is used, got: %d", mockT.errorCount)
	}

	// Test with a rule set that doesn't set output to nil - should fail
	mockT = &MockT{}
	mockRuleSetNotSet := &MockNilNotSet{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetNotSet)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set that doesn't set output to nil, got: %d", mockT.errorCount)
	}

	// Test with a rule set where WithNil returns wrong type - should fail
	mockT = &MockT{}
	mockRuleSetWrongType := &MockNilWrongReturnType{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWrongType)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set with WithNil returning wrong type, got: %d", mockT.errorCount)
	}

	// Test with a rule set that doesn't return an error when nil is provided without WithNil - should fail
	mockT = &MockT{}
	mockRuleSetNoError := &MockNilNoError{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetNoError)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set that doesn't return error for nil without WithNil, got: %d", mockT.errorCount)
	}

	// Test with a rule set that returns wrong error code when nil is provided without WithNil - should fail
	mockT = &MockT{}
	mockRuleSetWrongCodeWithoutWithNil := &MockNilWrongCodeWithoutWithNil{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWrongCodeWithoutWithNil)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set with wrong error code without WithNil, got: %d", mockT.errorCount)
	}

	// Test with a rule set where WithNil returns wrong number of values - should fail
	mockT = &MockT{}
	mockRuleSetWrongCount := &MockNilWrongReturnCount{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWrongCount)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set with WithNil returning wrong number of values, got: %d", mockT.errorCount)
	}

	// Test with a rule set that has WithNil but not WithRequired - should pass (WithRequired check is skipped)
	mockT = &MockT{}
	mockRuleSetWithNilOnly := &MockWithNilOnly{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWithNilOnly)
	if mockT.errorCount != 0 {
		t.Errorf("Expected 0 errors on rule set with WithNil but no WithRequired, got: %d", mockT.errorCount)
	}

	// Test with a rule set where WithRequired returns a rule set without WithNil - should pass (test is skipped)
	mockT = &MockT{}
	mockRuleSetWithRequiredNoWithNil := &MockWithRequiredNoWithNil{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWithRequiredNoWithNil)
	if mockT.errorCount != 0 {
		t.Errorf("Expected 0 errors on rule set where WithRequired returns rule set without WithNil, got: %d", mockT.errorCount)
	}

	// Test with a rule set where WithRequired returns wrong type - should fail
	mockT = &MockT{}
	mockRuleSetWithRequiredWrongType := &MockWithRequiredWrongType{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWithRequiredWrongType)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set where WithRequired returns wrong type, got: %d", mockT.errorCount)
	}

	// Test with a rule set where WithRequired returns wrong count - should fail
	mockT = &MockT{}
	mockRuleSetWithRequiredWrongCount := &MockWithRequiredWrongCount{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWithRequiredWrongCount)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set where WithRequired returns wrong count, got: %d", mockT.errorCount)
	}

	// Test with a rule set where WithNil on WithRequired rule set returns wrong type - should fail
	mockT = &MockT{}
	mockRuleSetWithNilWrongTypeOnRequired := &MockWithNilWrongTypeOnRequired{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWithNilWrongTypeOnRequired)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set where WithNil on WithRequired returns wrong type, got: %d", mockT.errorCount)
	}

	// Test with a rule set where WithNil on WithRequired rule set returns wrong count - should fail
	mockT = &MockT{}
	mockRuleSetWithNilWrongCountOnRequired := &MockWithNilWrongCountOnRequired{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWithNilWrongCountOnRequired)
	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error on rule set where WithNil on WithRequired returns wrong count, got: %d", mockT.errorCount)
	}

	// Test with a rule set where Apply returns error even with both WithNil and WithRequired - should fail
	// This will produce 2 errors: one for the Apply call returning an error, and one for output not being nil
	mockT = &MockT{}
	mockRuleSetWithBothButError := &MockWithBothButError{}
	testhelpers.MustImplementWithNil[int](mockT, mockRuleSetWithBothButError)
	if mockT.errorCount != 2 {
		t.Errorf("Expected 2 errors on rule set where Apply returns error with both WithNil and WithRequired, got: %d", mockT.errorCount)
	}
}

// MockWithNilOnly is a mock rule set that has WithNil but not WithRequired
type MockWithNilOnly struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithNilOnly) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Return correct CodeNull for the first test (without WithNil)
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithNilOnly) WithNil() rules.RuleSet[int] {
	return &MockWithNilOnlyWithNil{MockRuleSet: m.MockRuleSet}
}

// MockWithNilOnlyWithNil is a mock rule set that has WithNil but not WithRequired.
type MockWithNilOnlyWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithNilOnlyWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Set output to nil correctly
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
			elem := outputVal.Elem()
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Map || elem.Kind() == reflect.Chan || elem.Kind() == reflect.Func {
				elem.Set(reflect.Zero(elem.Type()))
			}
		}
		return nil
	}
	// For non-nil input, create a fresh mock and use its Apply
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

// MockWithRequiredNoWithNil is a mock rule set where WithRequired returns a rule set without WithNil
type MockWithRequiredNoWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredNoWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithRequiredNoWithNil) WithNil() rules.RuleSet[int] {
	return &MockWithRequiredNoWithNilWithNil{MockRuleSet: m.MockRuleSet}
}

// MockWithRequiredNoWithNilWithNil is a mock rule set where WithRequired returns a rule set without WithNil.
type MockWithRequiredNoWithNilWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredNoWithNilWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
			elem := outputVal.Elem()
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Map || elem.Kind() == reflect.Chan || elem.Kind() == reflect.Func {
				elem.Set(reflect.Zero(elem.Type()))
			}
		}
		return nil
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithRequiredNoWithNil) WithRequired() rules.RuleSet[int] {
	// Return a rule set that doesn't have WithNil
	return &MockWithRequiredNoWithNilRequired{MockRuleSet: m.MockRuleSet}
}

// MockWithRequiredNoWithNilRequired is a mock rule set where WithRequired returns a rule set without WithNil.
type MockWithRequiredNoWithNilRequired struct{ testhelpers.MockRuleSet[int] }

// Note: This intentionally doesn't implement WithNil() method

// MockWithRequiredWrongType is a mock rule set where WithRequired returns wrong type
type MockWithRequiredWrongType struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredWrongType) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithRequiredWrongType) WithNil() rules.RuleSet[int] {
	return &MockWithRequiredWrongTypeWithNil{MockRuleSet: m.MockRuleSet}
}

// MockWithRequiredWrongTypeWithNil is a mock rule set where WithRequired returns wrong type.
type MockWithRequiredWrongTypeWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredWrongTypeWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
			elem := outputVal.Elem()
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Map || elem.Kind() == reflect.Chan || elem.Kind() == reflect.Func {
				elem.Set(reflect.Zero(elem.Type()))
			}
		}
		return nil
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithRequiredWrongType) WithRequired() string {
	return "wrong type"
}

// MockWithRequiredWrongCount is a mock rule set where WithRequired returns wrong number of values
type MockWithRequiredWrongCount struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredWrongCount) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithRequiredWrongCount) WithNil() rules.RuleSet[int] {
	return &MockWithRequiredWrongCountWithNil{MockRuleSet: m.MockRuleSet}
}

// MockWithRequiredWrongCountWithNil is a mock rule set where WithRequired returns wrong number of values.
type MockWithRequiredWrongCountWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithRequiredWrongCountWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
			elem := outputVal.Elem()
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Map || elem.Kind() == reflect.Chan || elem.Kind() == reflect.Func {
				elem.Set(reflect.Zero(elem.Type()))
			}
		}
		return nil
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithRequiredWrongCount) WithRequired() (rules.RuleSet[int], string) {
	return &MockWithRequiredWrongCountWithNil{MockRuleSet: m.MockRuleSet}, "extra value"
}

// MockWithNilWrongTypeOnRequired is a mock where WithRequired returns a rule set whose WithNil returns wrong type
type MockWithNilWrongTypeOnRequired struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithNilWrongTypeOnRequired) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithNilWrongTypeOnRequired) WithNil() rules.RuleSet[int] {
	return &MockWithNilWrongTypeOnRequiredWithNil{MockRuleSet: m.MockRuleSet}
}

// MockWithNilWrongTypeOnRequiredWithNil is a mock where WithRequired returns a rule set whose WithNil returns wrong type.
type MockWithNilWrongTypeOnRequiredWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithNilWrongTypeOnRequiredWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
			elem := outputVal.Elem()
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Map || elem.Kind() == reflect.Chan || elem.Kind() == reflect.Func {
				elem.Set(reflect.Zero(elem.Type()))
			}
		}
		return nil
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithNilWrongTypeOnRequired) WithRequired() rules.RuleSet[int] {
	return &MockWithNilWrongTypeOnRequiredRequired{MockRuleSet: m.MockRuleSet}
}

// MockWithNilWrongTypeOnRequiredRequired is a mock where WithRequired returns a rule set whose WithNil returns wrong type.
type MockWithNilWrongTypeOnRequiredRequired struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithNilWrongTypeOnRequiredRequired) WithNil() string {
	return "wrong type"
}

// MockWithNilWrongCountOnRequired is a mock where WithRequired returns a rule set whose WithNil returns wrong count
type MockWithNilWrongCountOnRequired struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithNilWrongCountOnRequired) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithNilWrongCountOnRequired) WithNil() rules.RuleSet[int] {
	return &MockWithNilWrongCountOnRequiredWithNil{MockRuleSet: m.MockRuleSet}
}

// MockWithNilWrongCountOnRequiredWithNil is a mock where WithRequired returns a rule set whose WithNil returns wrong count.
type MockWithNilWrongCountOnRequiredWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithNilWrongCountOnRequiredWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
			elem := outputVal.Elem()
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Map || elem.Kind() == reflect.Chan || elem.Kind() == reflect.Func {
				elem.Set(reflect.Zero(elem.Type()))
			}
		}
		return nil
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithNilWrongCountOnRequired) WithRequired() rules.RuleSet[int] {
	return &MockWithNilWrongCountOnRequiredRequired{MockRuleSet: m.MockRuleSet}
}

// MockWithNilWrongCountOnRequiredRequired is a mock where WithRequired returns a rule set whose WithNil returns wrong count.
type MockWithNilWrongCountOnRequiredRequired struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithNilWrongCountOnRequiredRequired) WithNil() (rules.RuleSet[int], string) {
	return &MockWithNilWrongCountOnRequiredWithNil{MockRuleSet: m.MockRuleSet}, "extra value"
}

// MockWithBothButError is a mock where both WithNil and WithRequired are set but Apply returns error
type MockWithBothButError struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithBothButError) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		return errors.Collection(errors.Errorf(errors.CodeNull, ctx, "null not allowed", "value cannot be null"))
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithBothButError) WithNil() rules.RuleSet[int] {
	return &MockWithBothButErrorWithNil{MockRuleSet: m.MockRuleSet}
}

// MockWithBothButErrorWithNil is a mock where both WithNil and WithRequired are set but Apply returns error.
type MockWithBothButErrorWithNil struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithBothButErrorWithNil) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		outputVal := reflect.ValueOf(output)
		if outputVal.Kind() == reflect.Ptr && !outputVal.IsNil() {
			elem := outputVal.Elem()
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface || elem.Kind() == reflect.Slice ||
				elem.Kind() == reflect.Map || elem.Kind() == reflect.Chan || elem.Kind() == reflect.Func {
				elem.Set(reflect.Zero(elem.Type()))
			}
		}
		return nil
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}

func (m *MockWithBothButError) WithRequired() rules.RuleSet[int] {
	return &MockWithBothButErrorRequired{MockRuleSet: m.MockRuleSet}
}

// MockWithBothButErrorRequired is a mock where both WithNil and WithRequired are set but Apply returns error.
type MockWithBothButErrorRequired struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithBothButErrorRequired) WithNil() rules.RuleSet[int] {
	return &MockWithBothButErrorBoth{MockRuleSet: m.MockRuleSet}
}

// MockWithBothButErrorBoth is a mock where both WithNil and WithRequired are set but Apply returns error.
type MockWithBothButErrorBoth struct{ testhelpers.MockRuleSet[int] }

func (m *MockWithBothButErrorBoth) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	if input == nil {
		// Return an error even though both WithNil and WithRequired are set - this is the bug we're testing
		return errors.Collection(errors.Errorf(errors.CodeUnknown, ctx, "unknown error", "unexpected error"))
	}
	mockRuleSet := testhelpers.NewMockRuleSet[int]()
	return mockRuleSet.Apply(ctx, input, output)
}
