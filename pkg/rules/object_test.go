package rules_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	stringsHelper "strings"
	"sync/atomic"
	"testing"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func testMap() map[string]any {
	return map[string]any{"X": 10, "Y": 20}
}

func jsonTestValidator(x, y any) error {
	if m, ok := y.(map[string]any); !ok || m["X"] != 123 {
		return fmt.Errorf("Expected X to be 123")
	}
	return nil
}

type testStruct struct {
	W *int
	X int
	Y int
	z int //lint:ignore U1000 Used in reflection testing but not code
}

func testStructInit() *testStruct {
	return &testStruct{}
}

type testStructMapped struct {
	A int
	B int `validate:"C"`
	C int // Should never be written to since X is mapped to B and takes priority
	D int `validate:""` // Empty tag, ignore
}

// TestObjectRuleSet tests:
func TestObjectRuleSet(t *testing.T) {
	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of Validate
	err := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().Any()).
		WithKey("Y", rules.Int().Any()).
		Apply(context.TODO(), testMap(), &out)

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	// Verify that the rule set interface is implemented correctly
	ok := testhelpers.CheckRuleSetInterface[*testStruct](rules.Struct[*testStruct]())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	// Test both pointer and non-pointer.
	// These cases are tested in more detail in other tests.
	testhelpers.MustApplyTypes[testStruct](t, rules.Struct[testStruct](), testStruct{})
	testhelpers.MustApplyTypes[*testStruct](t, rules.Struct[*testStruct](), &testStruct{})
}

// TestObjectOutput_Apply tests:
// - Correctly applies validation to object output
// - Preserves existing field values
// - Works with pointer and non-pointer outputs
// - Works with any interface outputs
func TestObjectOutput_Apply(t *testing.T) {
	type outStruct struct {
		Name string
		// Age is not in the validator and should not be modified from its existing value
		// A modified Age means that Apply created a brand new outStruct instead of using
		// the existing one.
		Age int
	}

	ruleSet := rules.Struct[outStruct]().WithJson().WithKey("Name", rules.String().Any())
	ctx := context.Background()

	input := `{"Name": "Test"}`
	expected := "Test"
	// Correct type
	out1 := outStruct{Age: 1}
	err := ruleSet.Apply(ctx, input, &out1)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	} else if out1.Age != 1 {
		t.Errorf("Expected out1.Age to be 1, got: %d", out1.Age)
	} else if out1.Name != expected {
		t.Errorf(`Expected out1.Name to be "%s", got: "%s"`, expected, out1.Name)
	}

	// Non pointer
	err = ruleSet.Apply(ctx, input, out1)
	if err == nil || err.Code() != errors.CodeInternal {
		t.Errorf("Expected error to not be internal")
	}

	// Any
	var out3 any
	err = ruleSet.Apply(ctx, input, &out3)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	} else {
		out3struct, ok := out3.(outStruct)
		if !ok {
			t.Errorf(`Expected output to be outStruct, got %T`, out3)

		} else if out3struct.Name != expected {
			t.Errorf(`Expected out3struct.Name to be "%s", got: "%s"`, expected, out3struct.Name)
		}
	}

	// Pointer to incorrect type
	var out4 int
	err = ruleSet.Apply(ctx, input, out4)
	if err == nil || err.Code() != errors.CodeInternal {
		t.Errorf("Expected error to not be internal")
	}

	// Nil pointer to correct type
	var out5 *outStruct
	err = ruleSet.Apply(ctx, input, &out5)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	} else if out5 == nil {
		t.Error("Expected out5 to not be nil")
	} else if out5.Name != expected {
		t.Errorf(`Expected out5.Name to be "%s", got: "%s"`, expected, out5.Name)
	}

	// Non-nil pointer to correct type
	out5 = &outStruct{Age: 1}
	err = ruleSet.Apply(ctx, input, &out5)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	} else if out5 == nil {
		t.Error("Expected out5 to not be nil")
	} else if out5.Age != 1 {
		t.Errorf("Expected out5.Age to be 1, got: %d", out5.Age)
	} else if out5.Name != expected {
		t.Errorf(`Expected out5.Name to be "%s", got: "%s"`, expected, out5.Name)
	}

	// Non-empty interface with assignable type
	// Currently in this case Age will be lost because we cannot assign to
	var out6 any = outStruct{Age: 1}
	err = ruleSet.Apply(ctx, input, &out6)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	} else if out6 == nil {
		t.Error("Expected out6 to not be nil")
	} else {
		out6struct, ok := out6.(outStruct)

		if !ok {
			t.Errorf(`Expected output to be outStruct, got %T`, out3)
		} else if out6struct.Name != expected {
			t.Errorf(`Expected out6struct.Name to be "%s", got: "%s"`, expected, out6struct.Name)
		} else if out6struct.Age != 0 {
			t.Errorf("Expected out6struct.Age to be 0, got: %d", out6struct.Age)
		}
	}

	// Incompatible interface
	var out7 MyTestInterface
	err = ruleSet.Apply(ctx, input, &out7)
	if err == nil {
		t.Errorf("Expected error to not be nil")
	} else if out7 != nil {
		t.Error("Expected out7 to be nil")
	} else if c := err.Code(); c != errors.CodeInternal {
		t.Errorf("Expected error to be %s (errors.CodeInternal), got: %s", errors.CodeInternal, c)
	}
}

// TestObjectOutputPointer_Apply tests:
// - Correctly applies validation to pointer object output
// - Works with double pointer outputs
// - Works with any interface outputs
// - Returns error for incorrect output types
func TestObjectOutputPointer_Apply(t *testing.T) {
	type outStruct struct {
		Name string
		// Age is not in the validator and should not be modified from its existing value
		// A modified Age means that Apply created a brand new outStruct instead of using
		// the existing one.
		Age int
	}

	ruleSet := rules.Struct[*outStruct]().WithJson().WithKey("Name", rules.String().Any())
	ctx := context.Background()

	input := `{"Name": "Test"}`
	expected := "Test"

	// Correct type, interface to non-pointer
	out1 := outStruct{Age: 1}
	err := ruleSet.Apply(ctx, input, &out1)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	} else if out1.Age != 1 {
		t.Errorf("Expected out1.Age to be 1, got: %d", out1.Age)
	} else if out1.Name != expected {
		t.Errorf(`Expected out1.Name to be "%s", got: "%s"`, expected, out1.Name)
	}

	// Non pointer
	err = ruleSet.Apply(ctx, input, out1)
	if err == nil || err.Code() != errors.CodeInternal {
		t.Errorf("Expected error to not be internal")
	}

	// Double pointer to correct type, nil
	var out2 *outStruct
	err = ruleSet.Apply(ctx, input, &out2)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	} else if out2.Age != 0 {
		t.Errorf("Expected out2.Age to be 0, got: %d", out2.Age)
	} else if out2.Name != expected {
		t.Errorf(`Expected out2.Name to be "%s", got: "%s"`, expected, out2.Name)
	}

	// Pointer to correct type, non-nil
	out2 = &outStruct{Age: 1}
	err = ruleSet.Apply(ctx, input, &out2)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	} else if out2.Age != 1 {
		t.Errorf("Expected out2.Age to be 1, got: %d", out2.Age)
	} else if out2.Name != expected {
		t.Errorf(`Expected out2.Name to be "%s", got: "%s"`, expected, out2.Name)
	}

	// Any
	var out3 any
	err = ruleSet.Apply(ctx, input, &out3)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %v", err)
	} else {
		out3struct, ok := out3.(*outStruct)
		if !ok {
			t.Errorf(`Expected output to be *outStruct, got %T`, out3)
		} else if out3struct.Name != expected {
			t.Errorf(`Expected name to be "%s", got: "%s"`, expected, out3struct.Name)
		}
	}

	// Pointer to incorrect type
	var out4 int
	err = ruleSet.Apply(ctx, input, out4)
	if err == nil || err.Code() != errors.CodeInternal {
		t.Errorf("Expected error to not be internal")
	}
}

// TestObjectFromMapToMap tests:
// - Correctly validates and converts map to map
func TestObjectFromMapToMap(t *testing.T) {
	in := testMap()

	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of Validate
	err := rules.StringMap[any]().
		WithKey("X", rules.Int().Any()).
		WithKey("Y", rules.Int().Any()).
		Apply(context.TODO(), in, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out["X"] != 10 {
		t.Errorf("Expected output X to be 10 but got %v", out["X"])
		return
	}

	if out["Y"] != 20 {
		t.Errorf("Expected output Y to be 20 but got %v", out["Y"])
		return
	}
}

// TestObjectFromMapToStruct tests:
// - Correctly validates and converts map to struct
func TestObjectFromMapToStruct(t *testing.T) {
	in := testMap()

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of Validate
	err := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().Any()).
		WithKey("Y", rules.Int().Any()).
		Apply(context.TODO(), in, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out.X != 10 {
		t.Errorf("Expected output X to be 10 but got %v", out.X)
		return
	}

	if out.Y != 20 {
		t.Errorf("Expected output Y to be 20 but got %v", out.Y)
		return
	}
}

// TestObjectFromStructToMap tests:
// - Correctly validates and converts struct to map
func TestObjectFromStructToMap(t *testing.T) {
	in := testStructInit()
	in.X = 10
	in.Y = 20

	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of Validate
	err := rules.StringMap[any]().
		WithKey("X", rules.Int().Any()).
		WithKey("Y", rules.Int().Any()).
		Apply(context.TODO(), in, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out["X"] != 10 {
		t.Errorf("Expected output X to be 10 but got %v", out["X"])
		return
	}

	if out["Y"] != 20 {
		t.Errorf("Expected output Y to be 20 but got %v", out["Y"])
		return
	}
}

// TestObjectFromStructToStruct tests:
// - Correctly validates and converts struct to struct
func TestObjectFromStructToStruct(t *testing.T) {
	in := testStructInit()
	in.X = 10
	in.Y = 20

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of Validate
	err := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().Any()).
		WithKey("Y", rules.Int().Any()).
		Apply(context.TODO(), in, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out.X != 10 {
		t.Errorf("Expected output X to be 10 but got %v", out.X)
		return
	}

	if out.Y != 20 {
		t.Errorf("Expected output Y to be 20 but got %v", out.Y)
		return
	}
}

// TestPanicWhenOutputNotObjectLike tests:
// - Panics when output type is not object-like
func TestPanicWhenOutputNotObjectLike(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	rules.Struct[int]()
}

// TestPanicWhenAssigningRuleSetToMissingField tests:
// - Panics when trying to assign rule set to a missing field
func TestPanicWhenAssigningRuleSetToMissingField(t *testing.T) {
	defer func() {
		err, ok := recover().(error)

		if err == nil || !ok {
			t.Error("Expected panic with error interface")
		} else if err.Error() != `missing mapping for key: a` {
			t.Errorf("Expected missing mapping error, got: %s", err)
		}
	}()

	rules.Struct[*testStruct]().WithKey("a", rules.String().Any())
}

// TestKeyFunction tests:
// - WithKey function works correctly for key-specific rules
func TestKeyFunction(t *testing.T) {
	// Prepare the output variable for Apply
	var out *testStructMapped

	// Use Apply with WithKey for key-specific rules
	err := rules.Struct[*testStructMapped]().
		WithKey("A", rules.Int().Any()).
		WithKey("C", rules.Int().Any()).
		Apply(context.TODO(), map[string]any{"A": 123, "C": 456}, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out.A != 123 {
		t.Errorf("Expected output A to be 123 but got %v", out.A)
		return
	}

	if out.B != 456 {
		t.Errorf("Expected output B to be 456 but got %v", out.B)
		return
	}

	if out.C != 0 {
		t.Errorf("Expected output C to be 0 but got %v", out.C)
		return
	}
}

// TestObjectMapping tests:
// - Field mappings work correctly
func TestObjectMapping(t *testing.T) {
	// Prepare the output variable for Apply
	var out *testStructMapped

	// Use Apply instead of Validate
	err := rules.Struct[*testStructMapped]().
		WithKey("A", rules.Int().Any()).
		WithKey("C", rules.Int().Any()).
		Apply(context.TODO(), map[string]any{"A": 123, "C": 456}, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out.A != 123 {
		t.Errorf("Expected output A to be 123 but got %v", out.A)
		return
	}

	if out.B != 456 {
		t.Errorf("Expected output B to be 456 but got %v", out.B)
		return
	}

	if out.C != 0 {
		t.Errorf("Expected output C to be 0 but got %v", out.C)
		return
	}
}

// TestMissingField tests:
// - Missing optional fields do not cause errors
func TestMissingField(t *testing.T) {
	// Prepare the output variable for Apply
	var out map[string]int

	// Use Apply instead of Validate
	err := rules.StringMap[int]().
		WithKey("A", rules.Int()).
		WithKey("B", rules.Int()).
		Apply(context.TODO(), map[string]any{"A": 123}, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out["A"] != 123 {
		t.Errorf("Expected output A to be 123 but got %v", out["A"])
		return
	}

	b, ok := out["B"]
	if ok {
		t.Errorf("Expected output B to be missing but got %v", b)
		return
	}
}

// TestUnderlyingMapField tests:
// - Works when the input is a type whose underlying implementation is a map with string keys
func TestUnderlyingMapField(t *testing.T) {
	type underlyingMap map[string]string
	input := underlyingMap(map[string]string{"A": "123"})

	// Prepare the output variable for Apply
	var out map[string]int

	// Use Apply instead of Validate
	err := rules.StringMap[int]().
		WithKey("A", rules.Int()).
		WithKey("B", rules.Int()).
		Apply(context.TODO(), input, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out["A"] != 123 {
		t.Errorf("Expected output A to be 123 but got %v", out["A"])
		return
	}

	b, ok := out["B"]
	if ok {
		t.Errorf("Expected output B to be missing but got %v", b)
		return
	}
}

// TestMissingRequiredField tests:
// - Missing required fields cause errors
func TestMissingRequiredField(t *testing.T) {
	// Prepare the output variable for Apply
	var out map[string]int

	// Use Apply instead of Validate
	err := rules.StringMap[int]().
		WithKey("A", rules.Int()).
		WithKey("B", rules.Int().WithRequired()).
		Apply(context.TODO(), map[string]any{"A": 123}, &out)

	if len(errors.Unwrap(err)) == 0 {
		t.Errorf("Expected errors to not be empty")
	}
}

// TestObjectWithRequired tests:
// - WithRequired is correctly implemented for objects
func TestObjectWithRequired(t *testing.T) {
	testhelpers.MustImplementWithRequired[map[string]int](t, rules.StringMap[int]())
}

// TestUnknownFields tests:
// - Unknown fields cause errors by default
// - Unknown fields are allowed when WithUnknown is set
func TestUnknownFields(t *testing.T) {
	ruleSet := rules.StringMap[int]().WithKey("A", rules.Int())
	value := map[string]any{"A": 123, "C": 456}

	testhelpers.MustNotApply(t, ruleSet.Any(), value, errors.CodeUnexpected)

	ruleSet = ruleSet.WithUnknown()
	testhelpers.MustApplyFunc(t, ruleSet.Any(), value, "", func(_, _ any) error { return nil })
}

// TestInputNotObjectLike tests:
// - Returns error when input is not an object or map
func TestInputNotObjectLike(t *testing.T) {
	// Prepare the output variable for Apply
	var out *testStruct

	err := rules.Struct[*testStruct]().
		Apply(context.TODO(), 123, &out)

	if err == nil {
		t.Error("Expected errors to not be empty")
	}
}

// TestReturnsAllErrors tests:
// - Returns all validation errors, not just the first one
func TestReturnsAllErrors(t *testing.T) {
	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of Validate
	err := rules.StringMap[any]().
		WithKey("A", rules.Int().WithMax(2).Any()).
		WithKey("B", rules.Int().Any()).
		WithKey("C", rules.String().WithStrict().Any()).
		Apply(context.TODO(), map[string]any{"A": 123, "B": 456, "C": 789}, &out)

	if err == nil {
		t.Errorf("Expected errors to not be nil")
	} else if len(errors.Unwrap(err)) != 2 {
		t.Errorf("Expected 2 errors got %d: %s", len(errors.Unwrap(err)), err.Error())
	}
}

// TestObjectReturnsCorrectPaths tests:
// - Error paths correctly reflect nested object structure
func TestObjectReturnsCorrectPaths(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "myobj")

	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of ValidateWithContext
	err := rules.StringMap[any]().
		WithKey("A", rules.Int().WithMax(2).Any()).
		WithKey("B", rules.Int().Any()).
		WithKey("C", rules.String().WithStrict().Any()).
		Apply(ctx, map[string]any{"A": 123, "B": 456, "C": 789}, &out)

	if err == nil {
		t.Errorf("Expected errors to not be nil")
	} else if len(errors.Unwrap(err)) != 2 {
		t.Errorf("Expected 2 errors got %d: %s", len(errors.Unwrap(err)), err.Error())
		return
	}

	errA := errors.For(err, "/myobj/A")
	if errA == nil {
		t.Errorf("Expected error for /myobj/A to not be nil")
	} else if len(errors.Unwrap(errA)) != 1 {
		t.Errorf("Expected exactly 1 error for /myobj/A got %d: %s", len(errors.Unwrap(err)), err)
	} else if errA.Path() != "/myobj/A" {
		t.Errorf("Expected error path to be `%s` got `%s`", "/myobj/A", errA.Path())
	}

	errC := errors.For(err, "/myobj/C")
	if errC == nil {
		t.Errorf("Expected error for /myobj/C to not be nil")
	} else if len(errors.Unwrap(errC)) != 1 {
		t.Errorf("Expected exactly 1 error for /myobj/C got %d: %s", len(errors.Unwrap(err)), err)
	} else if errC.Path() != "/myobj/C" {
		t.Errorf("Expected error path to be `%s` got `%s`", "/myobj/C", errC.Path())
	}
}

// TestMixedMap tests:
// - Handles maps with mixed value types
func TestMixedMap(t *testing.T) {
	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of Validate
	err := rules.StringMap[any]().
		WithKey("A", rules.Int().Any()).
		WithKey("B", rules.Int().Any()).
		WithKey("C", rules.String().Any()).
		Apply(context.TODO(), map[string]any{"A": 123, "B": 456, "C": "789"}, &out)

	if err != nil {
		t.Errorf("Expected errors to be empty %s", err.Error())
		return
	}
}

// TestObjectCustom tests:
// - Custom rule functions are executed
// - Multiple custom rules are all executed
func TestObjectCustom(t *testing.T) {
	mock := testhelpers.NewMockRuleWithErrors[*testStruct](1)

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of Validate
	err := rules.Struct[*testStruct]().
		WithRuleFunc(mock.Function()).
		WithRuleFunc(mock.Function()).
		Apply(context.TODO(), map[string]any{"A": 123, "B": 456, "C": "789"}, &out)

	if err == nil {
		t.Error("Expected errors to not be nil")
	} else if len(errors.Unwrap(err)) != 5 {
		// The two custom errors + 3 unexpected keys
		t.Errorf("Expected 5 errors, got: %d", len(errors.Unwrap(err)))
	}

	if mock.EvaluateCallCount() != 2 {
		t.Errorf("Expected rule to be called 2 times, got %d", mock.EvaluateCallCount())
	}
}

// TestObjectAny tests:
// - Any returns a RuleSet[any] implementation
func TestObjectAny(t *testing.T) {
	ruleSet := rules.Float64().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	}
}

// TestPointer tests:
// - Handles pointer fields correctly
func TestPointer(t *testing.T) {
	// W is a pointer to an int
	ruleSet := rules.Struct[*testStruct]().WithKey("W", rules.Int().Any())

	// Prepare the output variable for Apply
	var obj *testStruct

	// Use Apply instead of Validate
	err := ruleSet.Apply(context.TODO(), map[string]any{"W": 123}, &obj)

	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	} else if obj.W == nil {
		t.Errorf("Expected obj.W to not be nil")
	} else if *obj.W != 123 {
		t.Errorf("Expected obj.W to be 123, got: %d", *obj.W)
	}
}

type testStructMappedBug struct {
	Email string `validate:"email"`
}

// This tests for an issue where the value could be set when the struct is not a pointer.
//
// See: https://github.com/proto-studio/protovalidate/issues/1
func TestBug001(t *testing.T) {
	ruleSet := rules.Struct[testStructMappedBug]().
		WithKey("email", rules.String().Any())

	expected := "hello@example.com"

	// Prepare the output variable for Apply
	var out testStructMappedBug

	// Use Apply instead of Validate
	err := ruleSet.Apply(context.TODO(), map[string]any{
		"email": expected,
	}, &out)

	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	} else if out.Email != expected {
		t.Errorf("Expected email to be '%s', got: '%s'", expected, out.Email)
	}
}

// TestObjectRequiredString tests:
// - Serializes to WithRequired()
func TestObjectRequiredString(t *testing.T) {
	ruleSet := rules.Struct[*testStruct]().WithRequired()

	expected := "ObjectRuleSet[*rules_test.testStruct].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestAllowUnknownString tests:
// - Serializes to WithUnknown()
func TestAllowUnknownString(t *testing.T) {
	ruleSet := rules.Struct[*testStruct]().WithUnknown()

	expected := "ObjectRuleSet[*rules_test.testStruct].WithUnknown()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestObjectWithItemRuleSetString tests:
// - Serializes to WithItemRuleSet()
func TestObjectWithItemRuleSetString(t *testing.T) {
	ruleSet := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().Any()).
		WithKey("Y", rules.Int().Any())

	expected := "ObjectRuleSet[*rules_test.testStruct].WithKey(\"X\", IntRuleSet[int].Any()).WithKey(\"Y\", IntRuleSet[int].Any())"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestWithRuleString tests:
// - Serializes to WithRule()
func TestWithRuleString(t *testing.T) {
	ruleSet := rules.Struct[*testStruct]().
		WithRuleFunc(testhelpers.NewMockRule[*testStruct]().Function())

	expected := "ObjectRuleSet[*rules_test.testStruct].WithRuleFunc(<function>)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// TestObjectEvaluate tests:
// - Evaluate behaves like ValidateWithContext
func TestObjectEvaluate(t *testing.T) {
	ctx := context.Background()

	ruleSet := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().Any()).
		WithKey("Y", rules.Int().Any())

	input := testStructInit()
	input.X = 12
	input.Y = 34

	// Evaluate directly using Evaluate method
	err1 := ruleSet.Evaluate(ctx, input)

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of ValidateWithContext
	err2 := ruleSet.Apply(ctx, input, &out)

	if err1 != nil || err2 != nil {
		t.Errorf("Expected errors to both be nil, got %s and %s", err1, err2)
	}
}

// TestMultipleRules tests:
// - Multiple rules on the same key all evaluate
func TestMultipleRules(t *testing.T) {
	ruleSet := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().WithMin(2).Any()).
		WithKey("X", rules.Int().WithMax(4).Any()).
		Any()

	testhelpers.MustApplyFunc(t, ruleSet, &testStruct{X: 3}, &testStruct{X: 3}, func(a, b any) error {
		if a.(*testStruct).X != b.(*testStruct).X {
			return fmt.Errorf("Expected X to be %d, got: %d", b.(*testStruct).X, a.(*testStruct).X)
		}
		return nil
	})
	testhelpers.MustNotApply(t, ruleSet, &testStruct{X: 1}, errors.CodeMin)
	testhelpers.MustNotApply(t, ruleSet, &testStruct{X: 5}, errors.CodeMax)
}

// TestTimeoutInObjectRule tests:
// - RuleSet times out if context does
// - Timeout error is returned
// - This test is specifically for a timeout while performing an object rule (as opposed to a key rule)
func TestTimeoutInObjectRule(t *testing.T) {
	// Set up a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	ruleSet := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().WithMin(2).Any()).
		WithRuleFunc(func(_ context.Context, x *testStruct) errors.ValidationError {
			// Simulate a delay that exceeds the timeout
			time.Sleep(1 * time.Second)
			return nil
		})

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of ValidateWithContext
	errs := ruleSet.Apply(ctx, &testStruct{}, &out)

	if errs == nil {
		t.Error("Expected errors to not be nil")
	} else if all := errors.Unwrap(errs); len(all) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(all))
	} else {
		codes := map[errors.ErrorCode]bool{}
		for _, e := range all {
			if ve, ok := e.(errors.ValidationError); ok {
				codes[ve.Code()] = true
			}
		}
		if !codes[errors.CodeTimeout] || !codes[errors.CodeMin] {
			t.Errorf("Expected one CodeTimeout and one CodeMin, got: %s", errs)
		}
	}
}

// TestTimeoutInKeyRule tests:
// - RuleSet times out if context does
// - Timeout error is returned
// - This test is specifically for a timeout while performing a key rule (as opposed to an object rule)
func TestTimeoutInKeyRule(t *testing.T) {
	// Set up a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ruleSet := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().
			WithRuleFunc(func(_ context.Context, x int) errors.ValidationError {
				// Simulate a delay that exceeds the timeout
				time.Sleep(1 * time.Second)
				return nil
			}).Any())

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of ValidateWithContext
	errs := ruleSet.Apply(ctx, &testStruct{}, &out)

	if errs == nil {
		t.Error("Expected errors to not be nil")
	} else if len(errors.Unwrap(errs)) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(errors.Unwrap(errs)), errs)
	} else if c := errs.Code(); c != errors.CodeTimeout {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeTimeout, c, errs)
	}
}

// TestCancelled tests:
// - Rules stop running after the context is cancelled
func TestCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var intCallCount int32 = 0
	var structCallCount int32 = 0

	intRule := func(_ context.Context, x int) errors.ValidationError {
		atomic.AddInt32(&intCallCount, 1)
		cancel()
		time.Sleep(1 * time.Second) // Simulate a delay that allows cancellation
		return nil
	}

	structRule := func(_ context.Context, x *testStruct) errors.ValidationError {
		atomic.AddInt32(&structCallCount, 1)
		time.Sleep(1 * time.Second) // Simulate a delay that allows cancellation
		return nil
	}

	ruleSet := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().WithRuleFunc(intRule).Any()).
		WithKey("X", rules.Int().WithRuleFunc(intRule).Any()).
		WithRuleFunc(structRule).
		WithRuleFunc(structRule)

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of ValidateWithContext
	errs := ruleSet.Apply(ctx, &testStruct{}, &out)

	if errs == nil {
		t.Error("Expected errors to not be nil")
	} else if len(errors.Unwrap(errs)) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(errors.Unwrap(errs)), errs)
	} else if c := errs.Code(); c != errors.CodeCancelled {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeCancelled, c, errs)
	}

	finalCallCount := atomic.LoadInt32(&intCallCount)
	if finalCallCount != 1 {
		t.Errorf("Expected intRule to be called 1 time, got %d", finalCallCount)
	}

	finalCallCount = atomic.LoadInt32(&structCallCount)
	if finalCallCount != 0 {
		t.Errorf("Expected structRule to not be called, got %d", finalCallCount)
	}
}

// TestCancelledObjectRules tests:
// - Object rules stop running after a cancel
func TestCancelledObjectRules(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var structCallCount int32 = 0

	structRule := func(_ context.Context, x *testStruct) errors.ValidationError {
		atomic.AddInt32(&structCallCount, 1)
		cancel()
		time.Sleep(1 * time.Second) // Simulate a delay that allows cancellation
		return nil
	}

	ruleSet := rules.Struct[*testStruct]().
		WithRuleFunc(structRule).
		WithRuleFunc(structRule)

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of ValidateWithContext
	errs := ruleSet.Apply(ctx, &testStruct{}, &out)

	if errs == nil {
		t.Error("Expected errors to not be nil")
	} else if len(errors.Unwrap(errs)) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(errors.Unwrap(errs)), errs)
	} else if c := errs.Code(); c != errors.CodeCancelled {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeCancelled, c, errs)
	}

	finalCallCount := atomic.LoadInt32(&structCallCount)
	if finalCallCount != 1 {
		t.Errorf("Expected structRule to be called 1 time, got %d", finalCallCount)
	}
}

// Requirement:
// - Conditional rules are called only when the condition returns no errors
// - Conditional rules are not called until dependent keys are evaluated
func TestConditionalKey(t *testing.T) {
	// Values to make sure the functions get called in order
	var intState int32 = 0
	var condValue int32 = 0

	// If the condition is evaluated before this rule finishes then the value will be incorrect
	intRule := func(_ context.Context, x int) errors.ValidationError {
		atomic.StoreInt32(&intState, 1)
		time.Sleep(100 * time.Millisecond)
		atomic.StoreInt32(&intState, 2)
		return nil
	}

	condValueRule := func(_ context.Context, y int) errors.ValidationError {
		condValue = atomic.LoadInt32(&intState)
		return nil
	}

	// Only run the conditional rule if X is greater than 4. Which it should only be if the intRule
	// function ran.
	condKeyRuleSet := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().WithMin(4).Any())

	ruleSet := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().WithRuleFunc(intRule).Any()).
		WithKey("Y", rules.Int().Any()).
		WithConditionalKey("Y", condKeyRuleSet, rules.Int().WithRuleFunc(condValueRule).Any())

	checkFn := func(a, b any) error {
		if a.(*testStruct).Y != b.(*testStruct).Y {
			return fmt.Errorf("Expected Y to be %d, got: %d", a.(*testStruct).Y, b.(*testStruct).Y)
		}
		if a.(*testStruct).X != b.(*testStruct).X {
			return fmt.Errorf("Expected X to be %d, got: %d", a.(*testStruct).X, b.(*testStruct).X)
		}
		return nil
	}

	// Mock rule should not have been called
	testhelpers.MustApplyFunc(t, ruleSet.Any(), &testStruct{X: 3, Y: 3}, &testStruct{X: 3, Y: 3}, checkFn)
	if intState != 2 {
		t.Fatalf("Expected the int validator to be finished")
	}
	if condValue != 0 {
		t.Errorf("Expected conditional rules to not be called")
	}

	intState = 0
	condValue = 0

	// Mock rule should have been called
	testhelpers.MustApplyFunc(t, ruleSet.Any(), &testStruct{X: 1, Y: 3}, &testStruct{X: 1, Y: 3}, checkFn)
	if intState != 2 {
		t.Fatalf("Expected the int validator to be finished")
	}
	if condValue != 0 {
		t.Errorf("Expected conditional rules to be called after the dependency finished")
	}
}

// Requirement:
// - Returns all keys with rules
// - Does not return keys with no rules
// - Returns conditional keys
// - Only returns each key once
func TestKeyRules(t *testing.T) {

	ruleSet := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().Any()).
		WithKey("X", rules.Int().Any()).
		WithConditionalKey("Y", rules.Struct[*testStruct](), rules.Int().Any())

	keys := ruleSet.KeyRules()

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d (%s)", len(keys), keys)
	} else {
		key0, ok := keys[0].(*rules.ConstantRuleSet[string])
		if !ok {
			t.Error("Expected key 0 to be a constant rule set with type string")
		}
		key1, ok := keys[1].(*rules.ConstantRuleSet[string])
		if !ok {
			t.Error("Expected key 1 to be a constant rule set with type string")
		}

		key0v := key0.Value()
		key1v := key1.Value()

		if !((key0v == "X" && key1v == "Y") || (key0v == "Y" && key1v == "X")) {
			t.Errorf("Expected [X Y], got %s", keys)
		}
	}

}

// TestConditionalKeyCycle tests:
// - The code panics is a cycle is made directly with conditional keys
func TestConditionalKeyCycle(t *testing.T) {
	condX := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().WithMin(4).Any())

	condY := rules.Struct[*testStruct]().
		WithKey("Y", rules.Int().WithMin(4).Any())

	ruleSet := rules.Struct[*testStruct]().
		WithConditionalKey("X", condY, rules.Int().Any())

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	ruleSet.WithConditionalKey("Y", condX, rules.Int().Any())
}

// TestConditionalKeyIndirectCycle tests:
// - The code panics is a cycle is made indirectly with conditional keys
func TestConditionalKeyIndirectCycle(t *testing.T) {
	condX := rules.Struct[*testStruct]().
		WithKey("X", rules.Int().WithMin(4).Any())

	condY := rules.Struct[*testStruct]().
		WithKey("Y", rules.Int().WithMin(4).Any())

	condW := rules.Struct[*testStruct]().
		WithKey("W", rules.Int().WithMin(4).Any())

	ruleSet := rules.Struct[*testStruct]().
		WithConditionalKey("X", condY, rules.Int().Any()).
		WithConditionalKey("Y", condW, rules.Int().Any())

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	ruleSet.WithConditionalKey("W", condX, rules.Int().Any())
}

// TestConditionalKeyVisited tests:
// - No need to visit the same nodes twice
func TestConditionalKeyVisited(t *testing.T) {

	/**
	 * A -> B -> D
	 * A -> C -> D
	 */

	condB := rules.StringMap[int]().
		WithKey("B", rules.Int().WithMin(4))

	condC := rules.StringMap[int]().
		WithKey("C", rules.Int().WithMin(4))

	condD := rules.StringMap[int]().
		WithKey("D", rules.Int().WithMin(4))

	rules.StringMap[int]().
		WithConditionalKey("B", condD, rules.Int()).
		WithConditionalKey("C", condD, rules.Int()).
		WithConditionalKey("A", condB, rules.Int()).
		WithConditionalKey("A", condC, rules.Int())
}

// TestStructRightType tests:
// - When an object that is already the right type is passed in, it is validated
// - 1:1 mapped keys work
// - Mapped keys still work even if the struct property is different
// - Works with the input being both the struct and a pointer to the struct
// - C is mapped to B on input so a rule on C should act on B
func TestStructRightType(t *testing.T) {
	ruleSet := rules.Struct[*testStructMapped]().
		WithKey("A", rules.Int().WithMin(4).Any()).
		WithKey("C", rules.Int().WithMin(100).Any())

	in := &testStructMapped{
		A: 10,
		B: 150,
	}

	check := func(a, b any) error {
		aa := a.(*testStructMapped)
		bb := b.(*testStructMapped)

		if aa.A != bb.A {
			return fmt.Errorf("Expected A to be %d, got %d", aa.A, bb.A)
		}
		if aa.B != bb.B {
			return fmt.Errorf("Expected B to be %d, got %d", aa.B, bb.B)
		}
		return nil
	}

	testhelpers.MustApplyFunc(t, ruleSet.Any(), in, in, check)

	in.A = 3
	testhelpers.MustNotApply(t, ruleSet.Any(), in, errors.CodeMin)

	in.A = 5

	in.B = 50
	testhelpers.MustNotApply(t, ruleSet.Any(), in, errors.CodeMin)

	in.B = 150
	testhelpers.MustApplyFunc(t, ruleSet.Any(), *in, in, check)
}

// TestNestedPointer tests:
// - Will assign nested pointer structs to pointers
// - Fixes issue: **rules_test.testStructMapped is not assignable to type *rules_test.testStruct
func TestNestedPointer(t *testing.T) {

	type target struct {
		Test *testStruct
	}

	ruleSet := rules.Struct[*target]().
		WithKey("Test", rules.Struct[*testStruct]().WithUnknown().Any())

	in := map[string]any{
		"Test": &testStruct{},
	}

	testhelpers.MustApplyFunc(t, ruleSet.Any(), in, in, func(a, b any) error { return nil })
}

// TestObjectFromMapToMapUnknown tests:
// - When WithUnknown is set, the resulting map should contain unknown values
func TestObjectFromMapToMapUnknown(t *testing.T) {
	in := testMap()

	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of Validate
	err := rules.StringMap[any]().
		WithUnknown().
		WithKey("X", rules.Int().Any()).
		Apply(context.TODO(), in, &out)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out["X"] != 10 {
		t.Errorf("Expected output X to be 10 but got %v", out["X"])
		return
	}

	if out["Y"] != 20 {
		t.Errorf("Expected output Y to be 20 but got %v", out["Y"])
		return
	}
}

// Test for a bug with conditional keys. Validate incorrect errors if ALL these conditions are met:
// - A conditional key has the "required" flag set.
// - The conditional key is missing from the input.
// - The condition is NOT true.
//
// The reason for this bug is because the condition was originally evaluates in evaluateKeyRule which is
// not called at all if the field is missing from the input. But the validator still check for Required().
func TestConditionalKeyRequiredBug(t *testing.T) {

	type conditionalBugTest struct {
		Type string `validate:"type"`
		X    string `validate:"x"`
		Y    string `validate:"y"`
	}

	ruleSet := rules.Struct[*conditionalBugTest]().
		WithKey("type", rules.String().WithRequired().WithAllowedValues("X", "Y", "Z").Any()).
		WithUnknown().
		WithConditionalKey(
			"y",
			rules.Struct[*conditionalBugTest]().WithKey("type", rules.String().WithRequired().WithAllowedValues("Y").Any()),
			rules.String().WithRequired().Any(),
		)

	checkFn := func(a, b any) error {
		aa := a.(*conditionalBugTest)
		bb := b.(*conditionalBugTest)

		if aa.Type != bb.Type {
			return fmt.Errorf("Expected Type to be %s, got: %s", aa.Type, bb.Type)
		}
		if aa.Y != bb.Y {
			return fmt.Errorf("Expected Y to be %s, got: %s", aa.Y, bb.Y)
		}
		return nil
	}

	testhelpers.MustApplyFunc(t, ruleSet.Any(), map[string]string{"type": "Y", "y": "!"}, &conditionalBugTest{Type: "Y", Y: "!"}, checkFn)
	testhelpers.MustApplyFunc(t, ruleSet.Any(), map[string]string{"type": "X", "X": "!"}, &conditionalBugTest{Type: "X"}, checkFn)
	testhelpers.MustNotApply(t, ruleSet.Any(), map[string]string{"type": "Y"}, errors.CodeRequired)
}

// TestWithKeyStringify tests:
// - Stringified rule sets using WithConditionalKey should have WithConditionalKey in the string
// - WithKey should be in sets using that
// - The conditional RuleSet should serialized for WithConditionalKey
// - The key RuleSet should serialized for both
// - Key should be quoted
func TestWithKeyStringify(t *testing.T) {
	strRule := rules.String().WithMinLen(4).Any()
	strRuleStr := strRule.String()

	ruleSet := rules.Struct[*testStruct]().WithKey("X", strRule)
	ruleSetStr := ruleSet.String()

	if stringsHelper.Contains(ruleSetStr, "WithConditionalKey") {
		t.Errorf("Expected string to not contain WithConditionalKey")
	}
	if !stringsHelper.Contains(ruleSetStr, `WithKey("X",`) {
		t.Errorf("Expected string to contain WithKey")
	}
	if !stringsHelper.Contains(ruleSetStr, strRuleStr) {
		t.Errorf("Expected string to contain the nested rule")
	}

	condRuleSet := rules.Struct[*testStruct]().WithUnknown()
	condRuleSetStr := condRuleSet.String()

	ruleSet = rules.Struct[*testStruct]().WithConditionalKey("Y", condRuleSet, strRule)
	ruleSetStr = ruleSet.String()

	if !stringsHelper.Contains(ruleSetStr, "WithConditionalKey") {
		t.Errorf("Expected string to contain WithConditionalKey")
	}
	if !stringsHelper.Contains(ruleSetStr, condRuleSetStr) {
		t.Errorf("Expected string to contain the conditional rule")
	}
	if !stringsHelper.Contains(ruleSetStr, strRuleStr) {
		t.Errorf("Expected string to contain the nested rule")
	}
}

// TestWithKeyStringifyInt tests:
// - Maps with non-string keys should not be quoted in String() output.
func TestWithKeyStringifyInt(t *testing.T) {
	strRule := rules.String().WithMinLen(4)
	strRuleStr := strRule.String()

	ruleSet := rules.Map[int, string]().WithKey(1, strRule)
	ruleSetStr := ruleSet.String()

	if !stringsHelper.Contains(ruleSetStr, `WithKey(1,`) {
		t.Errorf("Expected string to contain WithKey")
	}
	if !stringsHelper.Contains(ruleSetStr, strRuleStr) {
		t.Errorf("Expected string to contain the nested rule")
	}
}

// TestUnexpectedKeyPath tests:
// - Correct path is returns on unexpected key
func TestUnexpectedKeyPath(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "myobj")

	// Prepare the output variable for Apply
	var out map[string]int

	// Use Apply instead of ValidateWithContext
	err := rules.StringMap[int]().Apply(ctx, map[string]any{"x": 20}, &out)

	if err == nil {
		t.Errorf("Expected errors to not be nil")
		return
	} else if len(errors.Unwrap(err)) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(errors.Unwrap(err)), err.Error())
		return
	}

	if err.Path() != "/myobj/x" {
		t.Errorf("Expected error path to be `%s` got `%s` (%s)", "/myobj/x", err.Path(), err)
	}

	errA := errors.For(err, "/myobj/x")
	if errA == nil {
		t.Errorf("Expected error for /myobj/x to not be nil")
	}
}

// TestJsonString tests:
// - Does not parse Json string by default
// - Can validate Json string
// - Must also work for pointers to strings
// - Non Json strings cannot be coerced
func TestJsonString(t *testing.T) {
	ruleSet := rules.StringMap[any]().
		WithKey("X", rules.Int().Any())

	j := `{"X": 123}`
	invalid := "x"

	testhelpers.MustNotApply(t, ruleSet.Any(), j, errors.CodeType)
	testhelpers.MustNotApply(t, ruleSet.Any(), &j, errors.CodeType)
	testhelpers.MustNotApply(t, ruleSet.Any(), &invalid, errors.CodeType)

	ruleSet = ruleSet.WithJson()

	testhelpers.MustApplyFunc(t, ruleSet.Any(), j, "", jsonTestValidator)
	testhelpers.MustApplyFunc(t, ruleSet.Any(), &j, "", jsonTestValidator)
	testhelpers.MustNotApply(t, ruleSet.Any(), &invalid, errors.CodeType)
}

// TestJsonBytes tests:
// - Does not parse Json []byte by default
// - Can validate Json []byte
func TestJsonBytes(t *testing.T) {
	ruleSet := rules.StringMap[any]().
		WithKey("X", rules.Int().Any())

	j := []byte(`{"X": 123}`)

	testhelpers.MustNotApply(t, ruleSet.Any(), j, errors.CodeType)

	ruleSet = ruleSet.WithJson()

	testhelpers.MustApplyFunc(t, ruleSet.Any(), j, "", jsonTestValidator)
}

// TestJsonRawMessage tests:
// - Does not parse json.RawMessage by default
// - Can validate json.RawMessage
// - Must also work with pointers to json.RawMessage
func TestJsonRawMessage(t *testing.T) {
	ruleSet := rules.StringMap[any]().
		WithKey("X", rules.Int().Any())

	j := json.RawMessage([]byte(`{"X": 123}`))

	testhelpers.MustNotApply(t, ruleSet.Any(), j, errors.CodeType)
	testhelpers.MustNotApply(t, ruleSet.Any(), &j, errors.CodeType)

	ruleSet = ruleSet.WithJson()

	testhelpers.MustApplyFunc(t, ruleSet.Any(), j, "", jsonTestValidator)
	testhelpers.MustApplyFunc(t, ruleSet.Any(), &j, "", jsonTestValidator)
}

// TestWithRequiredIdempotent tests:
// - WithRequired is idempotent.
func TestWithRequiredIdempotent(t *testing.T) {
	a := rules.StringMap[any]()
	b := a.WithRequired()
	c := b.WithRequired()

	if a.Required() {
		t.Error("Expected `a` to not be required")
	}
	if !b.Required() {
		t.Error("Expected `b` to be required")
	}
	if !c.Required() {
		t.Error("Expected `c` to be required")
	}

	if a == b {
		t.Error("Expected `a` to not equal `b`")
	}
	if b != c {
		t.Error("Expected `b` to equal `c`")
	}
}

// TestWithJsonIdempotent tests:
// - WithJson is idempotent.
func TestWithJsonIdempotent(t *testing.T) {
	a := rules.StringMap[any]()
	b := a.WithJson()
	c := b.WithJson()

	if a == b {
		t.Error("Expected `a` to not equal `b`")
	}
	if b != c {
		t.Error("Expected `b` to equal `c`")
	}
}

// TestWithUnknownIdempotent tests:
// - WithUnknown is idempotent.
func TestWithUnknownIdempotent(t *testing.T) {
	a := rules.StringMap[any]()
	b := a.WithUnknown()
	c := b.WithUnknown()

	if a == b {
		t.Error("Expected `a` to not equal `b`")
	}
	if b != c {
		t.Error("Expected `b` to equal `c`")
	}
}

// TestWithDynamicKeyToMap tests:
// - Dynamic keys are not considered "unknown"
// - Rule is run for each matching key
// - Errors are passed through
func TestWithDynamicKeyToMap(t *testing.T) {
	ruleSet := rules.StringMap[float64]().WithJson()

	validJson := `{"__abc": 123, "__xyz": 789}`

	testhelpers.MustNotApply(t, ruleSet.Any(), validJson, errors.CodeUnexpected)

	rule := testhelpers.NewMockRuleSet[float64]()

	ruleSet = ruleSet.WithDynamicKey(rules.String().WithRegexp(regexp.MustCompile("^__"), ""), rule)

	testhelpers.MustNotApply(t, ruleSet.Any(), `{"abc": 123, "__xyz": 789}`, errors.CodeUnexpected)
	testhelpers.MustApplyAny(t, ruleSet.Any(), validJson)
}

// TestWithDynamicBucketToMap tests:
// - Keys in dynamic buckets are not considered "unknown"
// - Value is copied into all matching buckets
// - If no fields match, bucket is not present
func TestWithDynamicBucketToMap(t *testing.T) {
	ruleSet := rules.StringMap[any]().WithJson()

	validJson := `{"__abc": "abc", "__123": 123}`

	testhelpers.MustNotApply(t, ruleSet.Any(), validJson, errors.CodeUnexpected)

	ruleSet = ruleSet.WithDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^__"), ""), "all")
	ruleSet = ruleSet.WithDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^__[0-9]+"), ""), "numbers")
	ruleSet = ruleSet.WithDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^__[a-z]+"), ""), "letters")
	ruleSet = ruleSet.WithDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^nomatch"), ""), "nomatch")

	testhelpers.MustNotApply(t, ruleSet.Any(), `{"abc": 123, "__xyz": 789}`, errors.CodeUnexpected)

	o, err := testhelpers.MustApplyAny(t, ruleSet.Any(), validJson)
	if err == nil {
		output, ok := o.(map[string]any)
		if !ok {
			t.Errorf("expected output to be a map of any")
			return
		}

		if _, ok := output["nomatch"].(map[string]any); ok {
			t.Errorf(`expect "nomatch" bucket to not be present`)
		}

		if m, ok := output["all"].(map[string]any); ok {
			if len(m) != 2 {
				t.Errorf(`expected "all" to have 2 items, got %d`, len(m))
			}
		} else {
			t.Errorf(`expected "all" to be map`)
		}

		if m, ok := output["letters"].(map[string]any); ok {
			if len(m) != 1 {
				t.Errorf(`expected "letters" to have 1 item, got %d`, len(m))
			}
			if v, ok := m["__abc"]; !ok || v.(string) != "abc" {
				t.Errorf(`expected letters["__abc"] to be "abc", got %v`, v)
			}
		} else {
			t.Errorf(`expected "letters" to be map`)
		}

		if m, ok := output["numbers"].(map[string]any); ok {
			if len(m) != 1 {
				t.Errorf(`expected "numbers" to have 1 item, got %d`, len(m))
			}
			if v, ok := m["__123"]; !ok || v.(float64) != 123.0 {
				t.Errorf(`expected letters["__123"] to be "123", got %v`, v)
			}
		} else {
			t.Errorf(`expected "numbers" to be map`)
		}
	}
}

// TestWithDynamicBucketToStruct tests:
// - Keys in dynamic buckets are not considered "unknown"
// - Value is copied into all matching buckets
// - If no fields match, bucket is nil
func TestWithDynamicBucketToStruct(t *testing.T) {

	type outputType struct {
		All     map[string]any
		Letters map[string]string
		Numbers map[string]float64
		NoMatch map[string]any
	}

	ruleSet := rules.Struct[outputType]().WithJson()

	validJson := `{"__abc": "abc", "__123": 123}`

	testhelpers.MustNotApply(t, ruleSet.Any(), validJson, errors.CodeUnexpected)

	ruleSet = ruleSet.WithDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^__"), ""), "All")
	ruleSet = ruleSet.WithDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^__[0-9]+"), ""), "Numbers")
	ruleSet = ruleSet.WithDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^__[a-z]+"), ""), "Letters")
	ruleSet = ruleSet.WithDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^nomatch"), ""), "NoMatch")

	testhelpers.MustNotApply(t, ruleSet.Any(), `{"abc": "abc", "__xyz": "xyz"}`, errors.CodeUnexpected)

	o, err := testhelpers.MustApplyAny(t, ruleSet.Any(), validJson)
	if err == nil {
		output, ok := o.(outputType)
		if !ok {
			t.Errorf("expected output to be a map of any")
			return
		}

		if output.NoMatch != nil {
			t.Errorf(`expect "nomatch" bucket to not be present`)
		}

		if output.All != nil {
			if len(output.All) != 2 {
				t.Errorf(`expected "output.All" to have 2 items, got %d`, len(output.All))
			}
		} else {
			t.Errorf(`expected "output.All" to not be nil`)
		}

		if output.Letters != nil {
			if len(output.Letters) != 1 {
				t.Errorf(`expected "output.Letters" to have 1 item, got %d`, len(output.Letters))
			}
			if v, ok := output.Letters["__abc"]; !ok || v != "abc" {
				t.Errorf(`expected output.Letters["__abc"] to be "abc", got %v`, v)
			}
		} else {
			t.Errorf(`expected "output.Letters" to not be nil`)
		}

		if output.Numbers != nil {
			if len(output.Numbers) != 1 {
				t.Errorf(`expected "output.Numbers" to have 1 item, got %d`, len(output.Numbers))
			}
			if v, ok := output.Numbers["__123"]; !ok || v != 123.0 {
				t.Errorf(`expected output.Numbers["__123"] to be "123", got %v`, v)
			}
		} else {
			t.Errorf(`expected "output.Numbers" to not be nil`)
		}
	}
}

// valueListForBucketTest is an interface used to reproduce the bug where WithDynamicBucket +
// WithDynamicKey(Interface.WithCast) + WithUnknown + struct output caused SetBucket to receive
// raw input ([]string) instead of validated output, leading to a reflect SetMapIndex panic.
type valueListForBucketTest interface {
	Values() []string
	doNotExtend()
}

type fieldListMapForTest map[string]bool

func (fl fieldListMapForTest) Values() []string {
	keys := make([]string, 0, len(fl))
	for k := range fl {
		keys = append(keys, k)
	}
	return keys
}

func (fieldListMapForTest) doNotExtend() {}

func newFieldListForTest(fields ...string) valueListForBucketTest {
	out := make(fieldListMapForTest, len(fields))
	for _, f := range fields {
		out[f] = true
	}
	return out
}

// TestWithDynamicBucketAndDynamicKeyInterfaceToStruct reproduces a bug where using WithDynamicKey
// with rules.Interface[T].WithCast(...) and WithDynamicBucket to a struct field map, plus
// WithUnknown(), caused the "unknown keys" path to call SetBucket with raw input (e.g. []string)
// instead of the validated output (e.g. ValueList), triggering:
//   panic: reflect.Value.SetMapIndex: value of type []string is not assignable to type T
// The fix is to track known keys when dynamic buckets exist so keys already handled by
// evaluateKeyRule are not re-processed in the unknown path.
func TestWithDynamicBucketAndDynamicKeyInterfaceToStruct(t *testing.T) {
	type queryData struct {
		Fields  map[string]valueListForBucketTest
		Filters map[string][]string
	}

	stringQueryValueRuleSet := rules.Slice[string]().WithItemRuleSet(rules.String()).WithMaxLen(1)
	fieldKeyRule := rules.String().WithRegexp(regexp.MustCompile(`^fields\[[^\]]+\]$`), "")
	filterKeyRule := rules.String().WithRegexp(regexp.MustCompile(`^filter\[[^\]]+\]$`), "")

	fieldsRuleSet := rules.Interface[valueListForBucketTest]().WithCast(
		func(ctx context.Context, value any) (valueListForBucketTest, errors.ValidationError) {
			var strs []string
			if errs := stringQueryValueRuleSet.Apply(ctx, value, &strs); errs != nil {
				return nil, errs
			}
			if len(strs) == 0 {
				return newFieldListForTest(), nil
			}
			return newFieldListForTest(stringsHelper.Split(strs[0], ",")...), nil
		},
	)
	filterRuleSet := rules.Slice[string]().WithItemRuleSet(rules.String())

	ruleSet := rules.Struct[queryData]().
		WithDynamicKey(fieldKeyRule, fieldsRuleSet.Any()).
		WithDynamicKey(filterKeyRule, filterRuleSet.Any()).
		WithDynamicBucket(fieldKeyRule, "Fields").
		WithDynamicBucket(filterKeyRule, "Filters").
		WithUnknown()

	parsed, err := url.ParseQuery(`fields[articles]=abc,xyz`)
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}

	var output queryData
	errs := ruleSet.Apply(context.Background(), parsed, &output)
	if errs != nil {
		t.Fatalf("Apply: %v", errs)
	}
	if output.Fields == nil {
		t.Fatal("output.Fields is nil")
	}
	vl, ok := output.Fields["fields[articles]"]
	if !ok {
		t.Fatal("fields[articles] not in output.Fields")
	}
	// If the bug were present we would have panicked in SetBucket ([]string not assignable to ValueList).
	vals := vl.Values()
	if len(vals) != 2 || (vals[0] != "abc" && vals[1] != "abc") || (vals[0] != "xyz" && vals[1] != "xyz") {
		t.Errorf("expected Values() to contain abc and xyz, got %v", vals)
	}
}

// TestWithConditionalDynamicBucket tests:
// - If no dynamic bucket is matched then the key is considered unknown
// - Dynamic buckets are not created unless condition is met
// - Values are not put in the bucket unless condition is met
func TestWithConditionalDynamicBucket(t *testing.T) {
	ruleSet := rules.StringMap[any]().WithJson()

	rootCondition := rules.StringMap[any]().WithUnknown()

	trueRule := rules.Constant(true).WithRequired().Any()

	ruleSet = ruleSet.WithKey("allowLetters", rules.Any()).WithKey("allowNumbers", rules.Any())
	ruleSet = ruleSet.WithConditionalDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^__[0-9]+"), ""), rootCondition.WithKey("allowNumbers", trueRule), "numbers")
	ruleSet = ruleSet.WithConditionalDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^__[a-z]+"), ""), rootCondition.WithKey("allowLetters", trueRule), "letters")

	// Conditions not met so these properties should still be unknown
	testhelpers.MustNotApply(t, ruleSet.Any(), `{"__abc": "abc", "__123": 123}`, errors.CodeUnexpected)

	// This will make it so the rule set always passes
	ruleSet = ruleSet.WithDynamicBucket(rules.String().WithRegexp(regexp.MustCompile("^__"), ""), "all")

	o, err := testhelpers.MustApplyAny(t, ruleSet.Any(), `{"__abc": "abc", "__123": 123}`)
	if err == nil {
		output, ok := o.(map[string]any)
		if !ok {
			t.Errorf("expected output to be a map of any")
			return
		}

		if m, ok := output["all"].(map[string]any); ok {
			if len(m) != 2 {
				t.Errorf(`expected "all" to have 2 items, got %d`, len(m))
			}
		} else {
			t.Errorf(`expected "all" to be map`)
		}

		if _, ok := output["letters"].(map[string]any); ok {
			t.Errorf(`expect "letters" bucket to not be present`)
		}

		if _, ok := output["numbers"].(map[string]any); ok {
			t.Errorf(`expect "numbers" bucket to not be present`)
		}
	}

	o, err = testhelpers.MustApplyAny(t, ruleSet.Any(), `{"__abc": "abc", "__123": 123, "allowLetters":true}`)
	if err == nil {
		output, ok := o.(map[string]any)
		if !ok {
			t.Errorf("expected output to be a map of any")
			return
		}

		if m, ok := output["all"].(map[string]any); ok {
			if len(m) != 2 {
				t.Errorf(`expected "all" to have 2 items, got %d`, len(m))
			}
		} else {
			t.Errorf(`expected "all" to be map`)
		}

		if m, ok := output["letters"].(map[string]any); ok {
			if len(m) != 1 {
				t.Errorf(`expected "letters" to have 1 item, got %d`, len(m))
			}
			if v, ok := m["__abc"]; !ok || v.(string) != "abc" {
				t.Errorf(`expected letters["__abc"] to be "abc", got %v`, v)
			}
		} else {
			t.Errorf(`expected "letters" to be map`)
		}

		if _, ok := output["numbers"].(map[string]any); ok {
			t.Errorf(`expect "numbers" bucket to not be present`)
		}
	}
}

// TestDynamicKeyWithBucket tests:
// - Keys are still added to dynamic buckets when they match a dynamic key rule.
// - Keys are not added to output map.
// - Keys have the correct data type.
func TestDynamicKeyWithBucket(t *testing.T) {
	keyRule := rules.String().WithRegexp(regexp.MustCompile("^__"), "")

	ruleSet := rules.StringMap[any]().
		WithJson().
		WithDynamicKey(keyRule, rules.Int().Any()).
		WithDynamicBucket(keyRule, "numbers")

	o, err := testhelpers.MustApplyAny(t, ruleSet.Any(), `{"__123": "123"}`)
	if err == nil {
		output, ok := o.(map[string]any)
		if !ok {
			t.Errorf("expected output to be a map of any")
			return
		}

		if _, ok := output["__123"]; ok {
			t.Errorf("expected __123 to be absent from output")
		}

		if m, ok := output["numbers"].(map[string]any); ok {
			if len(m) != 1 {
				t.Errorf(`expected "numbers" to have 1 item, got %d`, len(m))
			}
			if _, ok := m["__123"]; !ok {
				t.Errorf(`expected numbers["__123"] to be in the bucket`)
			} else if v, ok := m["__123"].(int); !ok || v != 123 {
				if ok {
					t.Errorf(`expected numbers["__123"] to be 123, got %v`, v)
				} else {
					t.Errorf(`expected numbers["__123"] to be an int`)
				}
			}

		} else {
			t.Errorf(`expected "numbers" to be map`)
		}
	}
}

// TestStaticKeyWithBucket tests:
// - Static keys are not added to buckets
// - NOTE: This is UNDEFINED behavior. Rule writers should avoid having dynamic keys overlap with static keys
func TestStaticKeyWithBucket(t *testing.T) {
	keyRule := rules.String().WithRegexp(regexp.MustCompile("^__"), "")

	ruleSet := rules.StringMap[any]().
		WithJson().
		WithKey("__xyz", rules.Any()).
		WithDynamicBucket(keyRule, "letters")

	o, err := testhelpers.MustApplyAny(t, ruleSet.Any(), `{"__abc": "abc", "__xyz": "xyz"}`)
	if err == nil {
		output, ok := o.(map[string]any)
		if !ok {
			t.Errorf("expected output to be a map of any")
			return
		}

		if _, ok := output["__abc"]; ok {
			t.Errorf("expected __abc to be absent from output")
		}

		if _, ok := output["__xyz"]; !ok {
			t.Errorf("expected __xyz to be present in output")
		}

		if m, ok := output["letters"].(map[string]any); ok {
			if len(m) != 1 {
				t.Errorf(`expected "letters" to have 1 item, got %d`, len(m))
			}
			if v, ok := m["__abc"]; !ok || v.(string) != "abc" {
				t.Errorf(`expected letters["__abc"] to be "abc", got %v`, v)
			}
		} else {
			t.Errorf(`expected "letters" to be map`)
		}
	}
}

// Setup:
// Two unconditional rule sets act on key "__abc". One is dynamic and the other is static. There is an additional
// static conditional rule set that depends on "__abc". The conditional rule should not run until both the dynamic
// and static unconditional rules run.
// TestDynamicKeyAsConditionalDependency tests:
// - Conditional rules are not run until after any dynamic keys that affect the keys they are dependent on
// - This test triggered a bug with reference counting and the initial dynamic key code
func TestDynamicKeyAsConditionalDependency(t *testing.T) {
	var callCount int32 = 0

	valueRule := rules.Any().WithRuleFunc(func(ctx context.Context, _ any) errors.ValidationError {
		if rulecontext.Path(ctx).String() == "__abc" {
			time.Sleep(200 * time.Millisecond)
			atomic.AddInt32(&callCount, 1)
		}
		return nil
	})

	finalValueRule := rules.Any().WithRuleFunc(func(ctx context.Context, _ any) errors.ValidationError {
		if count := atomic.LoadInt32(&callCount); count != 2 {
			return errors.Errorf(errors.CodeCancelled, ctx, "cancelled", "Expected count of %d, got %d", 2, count)
		}
		return nil
	})

	keyRule := rules.String().WithRegexp(regexp.MustCompile("^__"), "")

	ruleSet := rules.StringMap[any]().
		WithJson().
		WithKey("__abc", valueRule).
		WithDynamicKey(keyRule, valueRule).
		WithConditionalKey("__xyz", rules.StringMap[any]().WithUnknown().WithKey("__abc", rules.Any()), finalValueRule)

	testhelpers.MustApplyAny(t, ruleSet.Any(), `{"__abc": "abc", "__xyz": "xyz"}`)
}

// TestDynamicKeyAsDependentConditional tests:
// - Ref tracker should panic if you try to use a dynamic key in a conditional
// - In the future we may change this behavior but for now it would complicate the code too much
func TestDynamicKeyAsDependentConditional(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	keyRule := rules.String().WithRegexp(regexp.MustCompile("^__"), "")

	rules.StringMap[any]().
		WithConditionalKey("__xyz", rules.StringMap[any]().WithUnknown().WithDynamicKey(keyRule, rules.Any()), rules.Any())
}

// TestJsonEmptyOutputBug tests:
// - Bug: Passing a non-string into a Rule Set that supports JSON deserialization results in empty output
func TestJsonEmptyOutputBug(t *testing.T) {
	jsonIn := `{"Name":"Abc"}`
	mapIn := map[string]any{"Name": "Abc"}

	type outStruct struct {
		Name string
	}

	ruleSet := rules.Struct[outStruct]().WithJson().WithKey("Name", rules.String().Any())
	ctx := context.Background()

	expected := "Abc"

	// Prepare output variables for Apply
	var jsonOut outStruct
	var mapOut outStruct

	// Apply with JSON input
	errs := ruleSet.Apply(ctx, jsonIn, &jsonOut)
	if errs != nil {
		t.Errorf("Expected nil errors on Json input, got: %s", errs)
	} else if jsonOut.Name != expected {
		t.Errorf(`Expected "%s", got: "%s"`, expected, jsonOut.Name)
	}

	// Apply with map input
	errs = ruleSet.Apply(ctx, mapIn, &mapOut)
	if errs != nil {
		t.Errorf("Expected nil errors on map input, got: %s", errs)
	} else if mapOut.Name != expected {
		t.Errorf(`Expected "%s", got: "%s"`, expected, mapOut.Name)
	}
}

// TestQueryStringInput tests:
// - url.Values can be passed to object validator as input
// - This should always work because url.Values is simply a map[string][]string but this test serves as an example and also to guard against regressions
func TestQueryStringInput(t *testing.T) {
	qs := "abc=123&xyz=789"
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	itemRuleSet := rules.Slice[int]().WithItemRuleSet(rules.Int()).WithMaxLen(1)

	ruleSet := rules.StringMap[[]int]().
		WithKey("abc", itemRuleSet).
		WithKey("xyz", itemRuleSet)

	// Prepare the output variable for Apply
	var out map[string][]int

	// Use Apply instead of Run
	errs := ruleSet.Apply(context.Background(), parsed, &out)
	if errs != nil {
		t.Errorf("Expected nil errors on input, got: %s", errs)
	} else if v, ok := out["abc"]; !ok || len(v) != 1 {
		t.Errorf(`Expected "abc" to exist in output and have length 1`)
	}
}

// TestObjectWithNil tests:
// - Returns error with CodeNull when nil is provided and WithNil is not used
// - Does not error when nil is provided and WithNil is used
func TestObjectWithNil(t *testing.T) {
	testhelpers.MustImplementWithNil[*testStruct](t, rules.Struct[*testStruct]())
}

// Requirements:
//   - When outputting to a map and WithNil is used on a key's rule set, nil values in the input map
//     should result in nil values in the output map
//   - When outputting to a map and WithNil is NOT used on a key's rule set, nil values in the input map
//     should result in a CodeNull error
func TestObjectMapWithNilKeyValue(t *testing.T) {
	ctx := context.TODO()

	// Test with WithNil - should succeed and have nil in output map
	ruleSetWithNil := rules.StringMap[any]().
		WithKey("key", rules.Any().WithNil())

	var outputWithNil map[string]any
	inputWithNil := map[string]any{"key": nil}

	err := ruleSetWithNil.Apply(ctx, inputWithNil, &outputWithNil)
	if err != nil {
		t.Errorf("Expected no error when WithNil is used, got: %s", err)
		return
	}

	if outputWithNil == nil {
		t.Error("Expected output map to not be nil")
		return
	}

	val, ok := outputWithNil["key"]
	if !ok {
		t.Error("Expected 'key' to be present in output map")
		return
	}

	if val != nil {
		t.Errorf("Expected 'key' value to be nil, got: %v", val)
	}

	// Test without WithNil - should error with CodeNull
	ruleSetWithoutNil := rules.StringMap[any]().
		WithKey("key", rules.Any())

	var outputWithoutNil map[string]any
	inputWithoutNil := map[string]any{"key": nil}

	err = ruleSetWithoutNil.Apply(ctx, inputWithoutNil, &outputWithoutNil)
	if err == nil {
		t.Error("Expected error when WithNil is not used")
		return
	}

	if err.Code() != errors.CodeNull {
		t.Errorf("Expected error code to be CodeNull, got: %s", err.Code())
	}
}

// TestObjectRuleSet_ErrorConfig tests:
// - ObjectRuleSet implements error configuration methods
func TestObjectRuleSet_ErrorConfig(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"name"`
	}
	testhelpers.MustImplementErrorConfig[TestStruct, *rules.ObjectRuleSet[TestStruct, string, any]](t, rules.Struct[TestStruct]())
}
