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

	"proto.zip/studio/validate"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/rules/strings"
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

func TestObjectRuleSet(t *testing.T) {
	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of Validate
	err := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("Y", numbers.NewInt().Any()).
		Apply(context.TODO(), testMap(), &out)

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	// Verify that the rule set interface is implemented correctly
	ok := testhelpers.CheckRuleSetInterface[*testStruct](rules.NewStruct[*testStruct]())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	// Test both pointer and non-pointer.
	// These cases are tested in more detail in other tests.
	testhelpers.MustApplyTypes[testStruct](t, rules.NewStruct[testStruct](), testStruct{})
	testhelpers.MustApplyTypes[*testStruct](t, rules.NewStruct[*testStruct](), &testStruct{})
}

func TestObjectOutput_Apply(t *testing.T) {
	type outStruct struct {
		Name string
		// Age is not in the validator and should not be modified from its existing value
		// A modified Age means that Apply created a brand new outStruct instead of using
		// the existing one.
		Age int
	}

	ruleSet := rules.NewStruct[outStruct]().WithJson().WithKey("Name", validate.String().Any())
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
	if err == nil || err.First().Code() != errors.CodeInternal {
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
	if err == nil || err.First().Code() != errors.CodeInternal {
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
	} else if c := err.First().Code(); c != errors.CodeInternal {
		t.Errorf("Expected error to be %s (errors.CodeInternal), got: %s", errors.CodeInternal, c)
	}
}

func TestObjectOutputPointer_Apply(t *testing.T) {
	type outStruct struct {
		Name string
		// Age is not in the validator and should not be modified from its existing value
		// A modified Age means that Apply created a brand new outStruct instead of using
		// the existing one.
		Age int
	}

	ruleSet := rules.NewStruct[*outStruct]().WithJson().WithKey("Name", validate.String().Any())
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
	if err == nil || err.First().Code() != errors.CodeInternal {
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
	if err == nil || err.First().Code() != errors.CodeInternal {
		t.Errorf("Expected error to not be internal")
	}
}

func TestObjectFromMapToMap(t *testing.T) {
	in := testMap()

	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of Validate
	err := rules.NewStringMap[any]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("Y", numbers.NewInt().Any()).
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

func TestObjectFromMapToStruct(t *testing.T) {
	in := testMap()

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of Validate
	err := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("Y", numbers.NewInt().Any()).
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

func TestObjectFromStructToMap(t *testing.T) {
	in := testStructInit()
	in.X = 10
	in.Y = 20

	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of Validate
	err := rules.NewStringMap[any]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("Y", numbers.NewInt().Any()).
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

func TestObjectFromStructToStruct(t *testing.T) {
	in := testStructInit()
	in.X = 10
	in.Y = 20

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of Validate
	err := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("Y", numbers.NewInt().Any()).
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

func TestPanicWhenOutputNotObjectLike(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	rules.NewStruct[int]()
}

func TestPanicWhenAssigningRuleSetToMissingField(t *testing.T) {
	defer func() {
		err, ok := recover().(error)

		if err == nil || !ok {
			t.Error("Expected panic with error interface")
		} else if err.Error() != `missing mapping for key: a` {
			t.Errorf("Expected missing mapping error, got: %s", err)
		}
	}()

	rules.NewStruct[*testStruct]().WithKey("a", strings.New().Any())
}

// This function is deprecated and will be removed in v1.0.0.
// Until then, make sure it still works.
func TestKeyFunction(t *testing.T) {
	// Prepare the output variable for Apply
	var out *testStructMapped

	// Use Apply instead of Validate
	err := rules.NewStruct[*testStructMapped]().
		Key("A", numbers.NewInt().Any()).
		Key("C", numbers.NewInt().Any()).
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

func TestObjectMapping(t *testing.T) {
	// Prepare the output variable for Apply
	var out *testStructMapped

	// Use Apply instead of Validate
	err := rules.NewStruct[*testStructMapped]().
		WithKey("A", numbers.NewInt().Any()).
		WithKey("C", numbers.NewInt().Any()).
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

func TestMissingField(t *testing.T) {
	// Prepare the output variable for Apply
	var out map[string]int

	// Use Apply instead of Validate
	err := rules.NewStringMap[int]().
		WithKey("A", numbers.NewInt()).
		WithKey("B", numbers.NewInt()).
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

// Requirements:
// - Works when the input is a type whose underlying implementation is a map with string keys
func TestUnderlyingMapField(t *testing.T) {
	type underlyingMap map[string]string
	input := underlyingMap(map[string]string{"A": "123"})

	// Prepare the output variable for Apply
	var out map[string]int

	// Use Apply instead of Validate
	err := rules.NewStringMap[int]().
		WithKey("A", numbers.NewInt()).
		WithKey("B", numbers.NewInt()).
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

func TestMissingRequiredField(t *testing.T) {
	// Prepare the output variable for Apply
	var out map[string]int

	// Use Apply instead of Validate
	err := rules.NewStringMap[int]().
		WithKey("A", numbers.NewInt()).
		WithKey("B", numbers.NewInt().WithRequired()).
		Apply(context.TODO(), map[string]any{"A": 123}, &out)

	if len(err) == 0 {
		t.Errorf("Expected errors to not be empty")
	}
}

func TestObjectWithRequired(t *testing.T) {
	ruleSet := rules.NewStringMap[int]()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestUnknownFields(t *testing.T) {
	ruleSet := rules.NewStringMap[int]().WithKey("A", numbers.NewInt())
	value := map[string]any{"A": 123, "C": 456}

	testhelpers.MustNotApply(t, ruleSet.Any(), value, errors.CodeUnexpected)

	ruleSet = ruleSet.WithUnknown()
	testhelpers.MustApplyFunc(t, ruleSet.Any(), value, "", func(_, _ any) error { return nil })
}

func TestInputNotObjectLike(t *testing.T) {
	// Prepare the output variable for Apply
	var out *testStruct

	err := rules.NewStruct[*testStruct]().
		Apply(context.TODO(), 123, &out)

	if err == nil {
		t.Error("Expected errors to not be empty")
	}
}

func TestReturnsAllErrors(t *testing.T) {
	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of Validate
	err := rules.NewStringMap[any]().
		WithKey("A", numbers.NewInt().WithMax(2).Any()).
		WithKey("B", numbers.NewInt().Any()).
		WithKey("C", strings.New().WithStrict().Any()).
		Apply(context.TODO(), map[string]any{"A": 123, "B": 456, "C": 789}, &out)

	if err == nil {
		t.Errorf("Expected errors to not be nil")
	} else if len(err) != 2 {
		t.Errorf("Expected 2 errors got %d: %s", len(err), err.Error())
	}
}

func TestObjectReturnsCorrectPaths(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "myobj")

	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of ValidateWithContext
	err := rules.NewStringMap[any]().
		WithKey("A", numbers.NewInt().WithMax(2).Any()).
		WithKey("B", numbers.NewInt().Any()).
		WithKey("C", strings.New().WithStrict().Any()).
		Apply(ctx, map[string]any{"A": 123, "B": 456, "C": 789}, &out)

	if err == nil {
		t.Errorf("Expected errors to not be nil")
	} else if len(err) != 2 {
		t.Errorf("Expected 2 errors got %d: %s", len(err), err.Error())
		return
	}

	errA := err.For("/myobj/A")
	if errA == nil {
		t.Errorf("Expected error for /myobj/A to not be nil")
	} else if len(errA) != 1 {
		t.Errorf("Expected exactly 1 error for /myobj/A got %d: %s", len(err), err)
	} else if errA.First().Path() != "/myobj/A" {
		t.Errorf("Expected error path to be `%s` got `%s`", "/myobj/A", errA.First().Path())
	}

	errC := err.For("/myobj/C")
	if errC == nil {
		t.Errorf("Expected error for /myobj/C to not be nil")
	} else if len(errC) != 1 {
		t.Errorf("Expected exactly 1 error for /myobj/C got %d: %s", len(err), err)
	} else if errC.First().Path() != "/myobj/C" {
		t.Errorf("Expected error path to be `%s` got `%s`", "/myobj/C", errC.First().Path())
	}
}

func TestMixedMap(t *testing.T) {
	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of Validate
	err := rules.NewStringMap[any]().
		WithKey("A", numbers.NewInt().Any()).
		WithKey("B", numbers.NewInt().Any()).
		WithKey("C", strings.New().Any()).
		Apply(context.TODO(), map[string]any{"A": 123, "B": 456, "C": "789"}, &out)

	if err != nil {
		t.Errorf("Expected errors to be empty %s", err.Error())
		return
	}
}

func TestObjectCustom(t *testing.T) {
	mock := testhelpers.NewMockRuleWithErrors[*testStruct](1)

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of Validate
	err := rules.NewStruct[*testStruct]().
		WithRuleFunc(mock.Function()).
		WithRuleFunc(mock.Function()).
		Apply(context.TODO(), map[string]any{"A": 123, "B": 456, "C": "789"}, &out)

	if err == nil {
		t.Error("Expected errors to not be nil")
	} else if len(err) != 5 {
		// The two custom errors + 3 unexpected keys
		t.Errorf("Expected 5 errors, got: %d", len(err))
	}

	if mock.EvaluateCallCount() != 2 {
		t.Errorf("Expected rule to be called 2 times, got %d", mock.EvaluateCallCount())
	}
}

func TestObjectAny(t *testing.T) {
	ruleSet := numbers.NewFloat64().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	}
}

func TestPointer(t *testing.T) {
	// W is a pointer to an int
	ruleSet := rules.NewStruct[*testStruct]().WithKey("W", numbers.NewInt().Any())

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
	ruleSet := rules.NewStruct[testStructMappedBug]().
		WithKey("email", strings.New().Any())

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

// Requirements:
// - Serializes to WithRequired()
func TestObjectRequiredString(t *testing.T) {
	ruleSet := rules.NewStruct[*testStruct]().WithRequired()

	expected := "ObjectRuleSet[*rules_test.testStruct].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithUnknown()
func TestAllowUnknownString(t *testing.T) {
	ruleSet := rules.NewStruct[*testStruct]().WithUnknown()

	expected := "ObjectRuleSet[*rules_test.testStruct].WithUnknown()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithItemRuleSet()
func TestObjectWithItemRuleSetString(t *testing.T) {
	ruleSet := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("Y", numbers.NewInt().Any())

	expected := "ObjectRuleSet[*rules_test.testStruct].WithKey(\"X\", IntRuleSet[int].Any()).WithKey(\"Y\", IntRuleSet[int].Any())"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithRule()
func TestWithRuleString(t *testing.T) {
	ruleSet := rules.NewStruct[*testStruct]().
		WithRuleFunc(testhelpers.NewMockRule[*testStruct]().Function())

	expected := "ObjectRuleSet[*rules_test.testStruct].WithRuleFunc(...)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Evaluate behaves like ValidateWithContext
func TestObjectEvaluate(t *testing.T) {
	ctx := context.Background()

	ruleSet := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("Y", numbers.NewInt().Any())

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

// Requirements:
// - Multiple rules on the same key all evaluate
func TestMultipleRules(t *testing.T) {
	ruleSet := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().WithMin(2).Any()).
		WithKey("X", numbers.NewInt().WithMax(4).Any()).
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

// Requirement:
// This test is specifically for a timeout while performing an object rule (as opposed to a key rule)
// - RuleSet times out if context does
// - Timeout error is returned
func TestTimeoutInObjectRule(t *testing.T) {
	// Set up a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	ruleSet := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().WithMin(2).Any()).
		WithRuleFunc(func(_ context.Context, x *testStruct) errors.ValidationErrorCollection {
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
	} else if len(errs) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errs))
	} else if c := errs.For("").First().Code(); c != errors.CodeTimeout {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeTimeout, c, errs.For("").First())
	}
}

// Requirement:
// This test is specifically for a timeout while performing an key rule (as opposed to an object rule)
// - RuleSet times out if context does
// - Timeout error is returned
func TestTimeoutInKeyRule(t *testing.T) {
	// Set up a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ruleSet := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().
			WithRuleFunc(func(_ context.Context, x int) errors.ValidationErrorCollection {
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
	} else if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(errs), errs)
	} else if c := errs.For("").First().Code(); c != errors.CodeTimeout {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeTimeout, c, errs.For("").First())
	}
}

// Requirement:
// - Rules stop running after the context is cancelled
func TestCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var intCallCount int32 = 0
	var structCallCount int32 = 0

	intRule := func(_ context.Context, x int) errors.ValidationErrorCollection {
		atomic.AddInt32(&intCallCount, 1)
		cancel()
		time.Sleep(1 * time.Second) // Simulate a delay that allows cancellation
		return nil
	}

	structRule := func(_ context.Context, x *testStruct) errors.ValidationErrorCollection {
		atomic.AddInt32(&structCallCount, 1)
		time.Sleep(1 * time.Second) // Simulate a delay that allows cancellation
		return nil
	}

	ruleSet := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().WithRuleFunc(intRule).Any()).
		WithKey("X", numbers.NewInt().WithRuleFunc(intRule).Any()).
		WithRuleFunc(structRule).
		WithRuleFunc(structRule)

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of ValidateWithContext
	errs := ruleSet.Apply(ctx, &testStruct{}, &out)

	if errs == nil {
		t.Error("Expected errors to not be nil")
	} else if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(errs), errs)
	} else if c := errs.First().Code(); c != errors.CodeCancelled {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeCancelled, c, errs.First())
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

// Requirement:
// - Object rules stop running after a cancel
func TestCancelledObjectRules(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var structCallCount int32 = 0

	structRule := func(_ context.Context, x *testStruct) errors.ValidationErrorCollection {
		atomic.AddInt32(&structCallCount, 1)
		cancel()
		time.Sleep(1 * time.Second) // Simulate a delay that allows cancellation
		return nil
	}

	ruleSet := rules.NewStruct[*testStruct]().
		WithRuleFunc(structRule).
		WithRuleFunc(structRule)

	// Prepare the output variable for Apply
	var out *testStruct

	// Use Apply instead of ValidateWithContext
	errs := ruleSet.Apply(ctx, &testStruct{}, &out)

	if errs == nil {
		t.Error("Expected errors to not be nil")
	} else if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(errs), errs)
	} else if c := errs.First().Code(); c != errors.CodeCancelled {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeCancelled, c, errs.First())
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
	intRule := func(_ context.Context, x int) errors.ValidationErrorCollection {
		atomic.StoreInt32(&intState, 1)
		time.Sleep(100 * time.Millisecond)
		atomic.StoreInt32(&intState, 2)
		return nil
	}

	condValueRule := func(_ context.Context, y int) errors.ValidationErrorCollection {
		condValue = atomic.LoadInt32(&intState)
		return nil
	}

	// Only run the conditional rule if X is greater than 4. Which it should only be if the intRule
	// function ran.
	condKeyRuleSet := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().WithMin(4).Any())

	ruleSet := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().WithRuleFunc(intRule).Any()).
		WithKey("Y", numbers.NewInt().Any()).
		WithConditionalKey("Y", condKeyRuleSet, numbers.NewInt().WithRuleFunc(condValueRule).Any())

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

	ruleSet := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("X", numbers.NewInt().Any()).
		WithConditionalKey("Y", rules.NewStruct[*testStruct](), numbers.NewInt().Any())

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

// Requirement:
// - The code panics is a cycle is made directly with conditional keys
func TestConditionalKeyCycle(t *testing.T) {
	condX := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().WithMin(4).Any())

	condY := rules.NewStruct[*testStruct]().
		WithKey("Y", numbers.NewInt().WithMin(4).Any())

	ruleSet := rules.NewStruct[*testStruct]().
		WithConditionalKey("X", condY, numbers.NewInt().Any())

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	ruleSet.WithConditionalKey("Y", condX, numbers.NewInt().Any())
}

// Requirement:
// - The code panics is a cycle is made indirectly with conditional keys
func TestConditionalKeyIndirectCycle(t *testing.T) {
	condX := rules.NewStruct[*testStruct]().
		WithKey("X", numbers.NewInt().WithMin(4).Any())

	condY := rules.NewStruct[*testStruct]().
		WithKey("Y", numbers.NewInt().WithMin(4).Any())

	condW := rules.NewStruct[*testStruct]().
		WithKey("W", numbers.NewInt().WithMin(4).Any())

	ruleSet := rules.NewStruct[*testStruct]().
		WithConditionalKey("X", condY, numbers.NewInt().Any()).
		WithConditionalKey("Y", condW, numbers.NewInt().Any())

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	ruleSet.WithConditionalKey("W", condX, numbers.NewInt().Any())
}

// Requirements:
// - No need to visit the same nodes twice
func TestConditionalKeyVisited(t *testing.T) {

	/**
	 * A -> B -> D
	 * A -> C -> D
	 */

	condB := rules.NewStringMap[int]().
		WithKey("B", numbers.NewInt().WithMin(4))

	condC := rules.NewStringMap[int]().
		WithKey("C", numbers.NewInt().WithMin(4))

	condD := rules.NewStringMap[int]().
		WithKey("D", numbers.NewInt().WithMin(4))

	rules.NewStringMap[int]().
		WithConditionalKey("B", condD, numbers.NewInt()).
		WithConditionalKey("C", condD, numbers.NewInt()).
		WithConditionalKey("A", condB, numbers.NewInt()).
		WithConditionalKey("A", condC, numbers.NewInt())
}

// Requirements:
// - When an object that is already the right type is passed in, it is validated.
// - 1:1 mapped keys works.
// - Mapped keys still work even if the struct property is different.
// - Works with the input being both the struct and a pointer to the struct
//
// C is mapped to B on input so a rule on C should act on B.
// The actual value of C should be ignored.
func TestStructRightType(t *testing.T) {
	ruleSet := rules.NewStruct[*testStructMapped]().
		WithKey("A", numbers.NewInt().WithMin(4).Any()).
		WithKey("C", numbers.NewInt().WithMin(100).Any())

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

// Requirements:
// - Will assign nested pointer structs to pointers
//
// Fixes issue:
// **rules_test.testStructMapped is not assignable to type *rules_test.testStruct
func TestNestedPointer(t *testing.T) {

	type target struct {
		Test *testStruct
	}

	ruleSet := rules.NewStruct[*target]().
		WithKey("Test", rules.NewStruct[*testStruct]().WithUnknown().Any())

	in := map[string]any{
		"Test": &testStruct{},
	}

	testhelpers.MustApplyFunc(t, ruleSet.Any(), in, in, func(a, b any) error { return nil })
}

// Requirement:
// - When WithUnknown is set, the resulting map should contain unknown values
func TestObjectFromMapToMapUnknown(t *testing.T) {
	in := testMap()

	// Prepare the output variable for Apply
	var out map[string]any

	// Use Apply instead of Validate
	err := rules.NewStringMap[any]().
		WithUnknown().
		WithKey("X", numbers.NewInt().Any()).
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

	ruleSet := rules.NewStruct[*conditionalBugTest]().
		WithKey("type", strings.New().WithRequired().WithAllowedValues("X", "Y", "Z").Any()).
		WithUnknown().
		WithConditionalKey(
			"y",
			rules.NewStruct[*conditionalBugTest]().WithKey("type", strings.New().WithRequired().WithAllowedValues("Y").Any()),
			strings.New().WithRequired().Any(),
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

// Requirements:
// - Stringified rule sets using WithConditionalKey should have WithConditionalKey in the string
// - WithKey should be in sets using that
// - The conditional RuleSet should serialized for WithConditionalKey
// - The key RuleSet should serialized for both
// - Key should be quoted
func TestWithKeyStringify(t *testing.T) {
	strRule := strings.New().WithMinLen(4).Any()
	strRuleStr := strRule.String()

	ruleSet := rules.NewStruct[*testStruct]().WithKey("X", strRule)
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

	condRuleSet := rules.NewStruct[*testStruct]().WithUnknown()
	condRuleSetStr := condRuleSet.String()

	ruleSet = rules.NewStruct[*testStruct]().WithConditionalKey("Y", condRuleSet, strRule)
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

// Requirements:
// - Maps with non-string keys should not be quoted in String() output.
func TestWithKeyStringifyInt(t *testing.T) {
	strRule := strings.New().WithMinLen(4)
	strRuleStr := strRule.String()

	ruleSet := rules.NewMap[int, string]().WithKey(1, strRule)
	ruleSetStr := ruleSet.String()

	if !stringsHelper.Contains(ruleSetStr, `WithKey(1,`) {
		t.Errorf("Expected string to contain WithKey")
	}
	if !stringsHelper.Contains(ruleSetStr, strRuleStr) {
		t.Errorf("Expected string to contain the nested rule")
	}
}

// Requirements:
// - Correct path is returns on unexpected key
func TestUnexpectedKeyPath(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "myobj")

	// Prepare the output variable for Apply
	var out map[string]int

	// Use Apply instead of ValidateWithContext
	err := rules.NewStringMap[int]().Apply(ctx, map[string]any{"x": 20}, &out)

	if err == nil {
		t.Errorf("Expected errors to not be nil")
		return
	} else if len(err) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(err), err.Error())
		return
	}

	if err.First().Path() != "/myobj/x" {
		t.Errorf("Expected error path to be `%s` got `%s` (%s)", "/myobj/x", err.First().Path(), err)
	}

	errA := err.For("/myobj/x")
	if errA == nil {
		t.Errorf("Expected error for /myobj/x to not be nil")
	}
}

// Requirements:
// - Does not parse Json string by default
// - Can validate Json string
// - Must also work for pointers to strings
// - Non Json strings cannot be coerced
func TestJsonString(t *testing.T) {
	ruleSet := rules.NewStringMap[any]().
		WithKey("X", numbers.NewInt().Any())

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

// Requirements:
// - Does not parse Json []byte by default
// - Can validate Json []byte
func TestJsonBytes(t *testing.T) {
	ruleSet := rules.NewStringMap[any]().
		WithKey("X", numbers.NewInt().Any())

	j := []byte(`{"X": 123}`)

	testhelpers.MustNotApply(t, ruleSet.Any(), j, errors.CodeType)

	ruleSet = ruleSet.WithJson()

	testhelpers.MustApplyFunc(t, ruleSet.Any(), j, "", jsonTestValidator)
}

// Requirements:
// - Does not parse json.RawMessage by default
// - Can validate json.RawMessage
// - Must also work with pointers to json.RawMessage
func TestJsonRawMessage(t *testing.T) {
	ruleSet := rules.NewStringMap[any]().
		WithKey("X", numbers.NewInt().Any())

	j := json.RawMessage([]byte(`{"X": 123}`))

	testhelpers.MustNotApply(t, ruleSet.Any(), j, errors.CodeType)
	testhelpers.MustNotApply(t, ruleSet.Any(), &j, errors.CodeType)

	ruleSet = ruleSet.WithJson()

	testhelpers.MustApplyFunc(t, ruleSet.Any(), j, "", jsonTestValidator)
	testhelpers.MustApplyFunc(t, ruleSet.Any(), &j, "", jsonTestValidator)
}

// Requirements:
// - WithRequired is idempotent.
func TestWithRequiredIdempotent(t *testing.T) {
	a := rules.NewStringMap[any]()
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

// Requirements:
// - WithJson is idempotent.
func TestWithJsonIdempotent(t *testing.T) {
	a := rules.NewStringMap[any]()
	b := a.WithJson()
	c := b.WithJson()

	if a == b {
		t.Error("Expected `a` to not equal `b`")
	}
	if b != c {
		t.Error("Expected `b` to equal `c`")
	}
}

// Requirements:
// - WithUnknown is idempotent.
func TestWithUnknownIdempotent(t *testing.T) {
	a := rules.NewStringMap[any]()
	b := a.WithUnknown()
	c := b.WithUnknown()

	if a == b {
		t.Error("Expected `a` to not equal `b`")
	}
	if b != c {
		t.Error("Expected `b` to equal `c`")
	}
}

// Requirements:
// - Dynamic keys are not considered "unknown"
// - Rule is run for each matching key
// - Errors are passed through
func TestWithDynamicKeyToMap(t *testing.T) {
	ruleSet := rules.NewStringMap[float64]().WithJson()

	validJson := `{"__abc": 123, "__xyz": 789}`

	testhelpers.MustNotApply(t, ruleSet.Any(), validJson, errors.CodeUnexpected)

	rule := testhelpers.NewMockRuleSet[float64]()

	ruleSet = ruleSet.WithDynamicKey(strings.New().WithRegexp(regexp.MustCompile("^__"), ""), rule)

	testhelpers.MustNotApply(t, ruleSet.Any(), `{"abc": 123, "__xyz": 789}`, errors.CodeUnexpected)
	testhelpers.MustApplyAny(t, ruleSet.Any(), validJson)
}

// Requirements:
// - Keys in dynamic buckets are not considered "unknown"
// - Value is copied into all matching buckets
// - If no fields match, bucket is not present
func TestWithDynamicBucketToMap(t *testing.T) {
	ruleSet := rules.NewStringMap[any]().WithJson()

	validJson := `{"__abc": "abc", "__123": 123}`

	testhelpers.MustNotApply(t, ruleSet.Any(), validJson, errors.CodeUnexpected)

	ruleSet = ruleSet.WithDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^__"), ""), "all")
	ruleSet = ruleSet.WithDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^__[0-9]+"), ""), "numbers")
	ruleSet = ruleSet.WithDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^__[a-z]+"), ""), "letters")
	ruleSet = ruleSet.WithDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^nomatch"), ""), "nomatch")

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

// Requirements:
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

	ruleSet := rules.NewStruct[outputType]().WithJson()

	validJson := `{"__abc": "abc", "__123": 123}`

	testhelpers.MustNotApply(t, ruleSet.Any(), validJson, errors.CodeUnexpected)

	ruleSet = ruleSet.WithDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^__"), ""), "All")
	ruleSet = ruleSet.WithDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^__[0-9]+"), ""), "Numbers")
	ruleSet = ruleSet.WithDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^__[a-z]+"), ""), "Letters")
	ruleSet = ruleSet.WithDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^nomatch"), ""), "NoMatch")

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

// Requirements:
// - If no dynamic bucket is matched then the key is considered unknown
// - Dynamic buckets are not created unless condition is met
// - Values are not put in the bucket unless condition is met
func TestWithConditionalDynamicBucket(t *testing.T) {
	ruleSet := rules.NewStringMap[any]().WithJson()

	rootCondition := rules.NewStringMap[any]().WithUnknown()

	trueRule := rules.Constant(true).WithRequired().Any()

	ruleSet = ruleSet.WithKey("allowLetters", rules.Any()).WithKey("allowNumbers", rules.Any())
	ruleSet = ruleSet.WithConditionalDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^__[0-9]+"), ""), rootCondition.WithKey("allowNumbers", trueRule), "numbers")
	ruleSet = ruleSet.WithConditionalDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^__[a-z]+"), ""), rootCondition.WithKey("allowLetters", trueRule), "letters")

	// Conditions not met so these properties should still be unknown
	testhelpers.MustNotApply(t, ruleSet.Any(), `{"__abc": "abc", "__123": 123}`, errors.CodeUnexpected)

	// This will make it so the rule set always passes
	ruleSet = ruleSet.WithDynamicBucket(strings.New().WithRegexp(regexp.MustCompile("^__"), ""), "all")

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

// Requirements:
// - Keys are still added to dynamic buckets when they match a dynamic key rule.
// - Keys are not added to output map.
// - Keys have the correct data type.
func TestDynamicKeyWithBucket(t *testing.T) {
	keyRule := strings.New().WithRegexp(regexp.MustCompile("^__"), "")

	ruleSet := rules.NewStringMap[any]().
		WithJson().
		WithDynamicKey(keyRule, validate.Int().Any()).
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

// Requirements:
// - Static keys are not added to buckets.
//
// NOTE: This is UNDEFINED behavior. Rule writers should avoid having dynamic keys overlap with static keys.
// The purpose of this test is just to let us know if this behavior unintentionally changes.
func TestStaticKeyWithBucket(t *testing.T) {
	keyRule := strings.New().WithRegexp(regexp.MustCompile("^__"), "")

	ruleSet := rules.NewStringMap[any]().
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
// Requirements:
// - Conditional rules are not run until after any dynamic keys that affect the keys they are dependent on.
//
// The order of this test depends heavily on weather WithDynamicKey or WithConditionalKey grab the lock first on the
// __xyz key (as of v0.3) of for now we are just checking to make sure the __abc rules run first and it's impossible to
// check that conditional rule does not block the __xyz rule without creating a race condition where the test sometimes
// fails.
//
// This test triggered a bug with reference counting and the initial dynamic key code. It is important that the dynamic
// key matches more than one input key to continue to test the reference bug.
func TestDynamicKeyAsConditionalDependency(t *testing.T) {
	var callCount int32 = 0

	valueRule := rules.Any().WithRuleFunc(func(ctx context.Context, _ any) errors.ValidationErrorCollection {
		if rulecontext.Path(ctx).String() == "__abc" {
			time.Sleep(200 * time.Millisecond)
			atomic.AddInt32(&callCount, 1)
		}
		return nil
	})

	finalValueRule := rules.Any().WithRuleFunc(func(ctx context.Context, _ any) errors.ValidationErrorCollection {
		if count := atomic.LoadInt32(&callCount); count != 2 {
			return errors.Collection(errors.Errorf(errors.CodeCancelled, ctx, "Expected count of %d, got %d", 2, count))
		}
		return nil
	})

	keyRule := strings.New().WithRegexp(regexp.MustCompile("^__"), "")

	ruleSet := rules.NewStringMap[any]().
		WithJson().
		WithKey("__abc", valueRule).
		WithDynamicKey(keyRule, valueRule).
		WithConditionalKey("__xyz", rules.NewStringMap[any]().WithUnknown().WithKey("__abc", rules.Any()), finalValueRule)

	testhelpers.MustApplyAny(t, ruleSet.Any(), `{"__abc": "abc", "__xyz": "xyz"}`)
}

// Requirements:
// - Ref tracker should panic if you try to use a dynamic key in a conditional.
// In the future we may change this behavior but for now it would complicate the code to much.
func TestDynamicKeyAsDependentConditional(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	keyRule := strings.New().WithRegexp(regexp.MustCompile("^__"), "")

	rules.NewStringMap[any]().
		WithConditionalKey("__xyz", rules.NewStringMap[any]().WithUnknown().WithDynamicKey(keyRule, rules.Any()), rules.Any())
}

// Bug: Passing a non-string into a Rule Set that supports Json deserialization results in empty output.
func TestJsonEmptyOutputBug(t *testing.T) {
	jsonIn := `{"Name":"Abc"}`
	mapIn := map[string]any{"Name": "Abc"}

	type outStruct struct {
		Name string
	}

	ruleSet := rules.NewStruct[outStruct]().WithJson().WithKey("Name", validate.String().Any())
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

// Requirement:
// - url.Values can be passed to object validator as input.
//
// This should always work because url.Values is simply a map[string] []string but this test
// serves as an example and also to guard against regressions.
func TestQueryStringInput(t *testing.T) {
	qs := "abc=123&xyz=789"
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	itemRuleSet := validate.Array[int]().WithItemRuleSet(validate.Int()).WithMaxLen(1)

	ruleSet := rules.NewStringMap[[]int]().
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
