package testhelpers

import (
	"context"
	"sync/atomic"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// MockRule is a mock implementation of the Rule interface that can be used for testing.
// They can be used to return errors, return mutated values, and simulate rule collisions.
//
// Call count tracks how many times the mock was evaluated.
//
// Because rule evaluation can happen in parallel (for example, with object keys or arrays)
// the call count is thread safe.
type MockRule[T any] struct {
	// Use int64 for atomic operations compatibility
	callCount int64

	// fn stores the function representation of the rule
	fn func(_ context.Context, _ T) (T, errors.ValidationErrorCollection)

	// Errors is used to return errors to the mock caller.
	Errors []errors.ValidationError

	// ConflictKey is used to determine if a MockCustomRule collides with another
	// If two rules have the same ConflictKey they will be treated as a collision.
	ConflictKey string

	// Value to return from Evaluate.
	ReturnValue *T
}

// NewMockRule creates a new MockRule with a specific return value.
func NewMockRuleWithValue[T any](value T) *MockRule[T] {
	return &MockRule[T]{
		ReturnValue: &value,
	}
}

// NewMockRule creates a new MockRule with errors set.
func NewMockRuleWithErrors[T any](count int) *MockRule[T] {
	return &MockRule[T]{
		Errors: NewMockErrors(count),
	}
}

// Evaluate takes a context and a value to evaluate.
// The return value will be different depending on the settings of the mock:
// - If errors are set then it will return all the errors.
// - If an override return value is set it will return that.
// - If neither, it will return the original value and no errors.
func (rule *MockRule[T]) Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection) {
	atomic.AddInt64(&rule.callCount, 1)

	retValue := value

	if rule.ReturnValue != nil {
		retValue = *rule.ReturnValue
	}

	if rule.Errors != nil && len(rule.Errors) > 0 {
		return retValue, errors.Collection(rule.Errors...)
	}

	return retValue, nil
}

// Conflict returns true for any MockCustomRule with the ConflictKey set to the same value.
func (rule *MockRule[T]) Conflict(x rules.Rule[T]) bool {
	y, ok := x.(*MockRule[T])
	if ok {
		return y.ConflictKey != "" && y.ConflictKey == rule.ConflictKey
	}
	return false
}

// String returns the string representation of the rule. Which is always WithMock() for mocks.
func (rule *MockRule[T]) String() string {
	return "WithMock()"
}

// CallCount returns the number of times the Evaluate function was called.
func (rule *MockRule[T]) CallCount() int64 {
	return atomic.LoadInt64(&rule.callCount)
}

// Reset resets the call count to 0.
func (rule *MockRule[T]) Reset() {
	atomic.StoreInt64(&rule.callCount, 0)
}

// Function returns a function rule implementation of the rule for testing WithCustomFunc implementations.
// Call count is shared so if you have a function and a struct representation of a mock rule, the counter
// will be synchronized. However, there is no way to get teh call count directly from the function so you should
// store a copy of the MockCustomRule if you wish to retrieve the count.
//
// Calling this function more than once will result in the same function being returned.
func (rule *MockRule[T]) Function() func(_ context.Context, _ T) (T, errors.ValidationErrorCollection) {
	if rule.fn == nil {
		rule.fn = func(ctx context.Context, value T) (T, errors.ValidationErrorCollection) {
			return rule.Evaluate(ctx, value)
		}
	}
	return rule.fn
}

// NewMockErrors creates a slice of 0 or more errors.
func NewMockErrors(count int) []errors.ValidationError {
	out := make([]errors.ValidationError, 0, count)

	for i := 0; i < count; i++ {
		out = append(out, errors.Errorf(errors.CodeUnknown, context.Background(), "test"))
	}

	return out
}
