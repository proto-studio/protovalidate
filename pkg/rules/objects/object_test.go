package objects_test

import (
	"context"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/rules/objects"
	"proto.zip/studio/validate/pkg/rules/strings"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func mutateIntPlusOne(_ context.Context, x int) (int, errors.ValidationErrorCollection) {
	return x + 1, nil
}

func testMap() map[string]any {
	return map[string]any{"X": 10, "Y": 20}
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
	_, err := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("Y", numbers.NewInt().Any()).
		Validate(testMap())

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	ok := testhelpers.CheckRuleSetInterface[*testStruct](objects.New[*testStruct]())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

func TestObjectFromMapToMap(t *testing.T) {
	in := testMap()

	out, err := objects.NewObjectMap[any]().
		WithKey("X", numbers.NewInt().WithRuleFunc(mutateIntPlusOne).Any()).
		WithKey("Y", numbers.NewInt().Any()).
		Validate(in)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out["X"] != 11 {
		t.Errorf("Expected output X to be 11 but got %v", out["X"])
		return
	}

	if out["Y"] != 20 {
		t.Errorf("Expected output Y to be 20 but got %v", out["Y"])
		return
	}

	if in["X"] != 10 {
		t.Errorf("Expected input X to still be 10 but got %v", in["X"])
		return
	}
}

func TestObjectFromMapToStruct(t *testing.T) {
	in := testMap()

	out, err := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().WithRuleFunc(mutateIntPlusOne).Any()).
		WithKey("Y", numbers.NewInt().Any()).
		Validate(in)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out.X != 11 {
		t.Errorf("Expected output X to be 11 but got %v", out.X)
		return
	}

	if out.Y != 20 {
		t.Errorf("Expected output Y to be 20 but got %v", out.Y)
		return
	}

	if in["X"] != 10 {
		t.Errorf("Expected input X to still be 10 but got %v", in["X"])
		return
	}
}

func TestObjectFromStructToMap(t *testing.T) {
	in := testStructInit()
	in.X = 10
	in.Y = 20

	out, err := objects.NewObjectMap[any]().
		WithKey("X", numbers.NewInt().WithRuleFunc(mutateIntPlusOne).Any()).
		WithKey("Y", numbers.NewInt().Any()).
		Validate(in)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out["X"] != 11 {
		t.Errorf("Expected output X to be 11 but got %v", out["X"])
		return
	}

	if out["Y"] != 20 {
		t.Errorf("Expected output Y to be 20 but got %v", out["Y"])
		return
	}

	if in.X != 10 {
		t.Errorf("Expected input X to still be 10 but got %v", in.X)
		return
	}
}

func TestObjectFromStructToStruct(t *testing.T) {
	in := testStructInit()
	in.X = 10
	in.Y = 20

	out, err := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().WithRuleFunc(mutateIntPlusOne).Any()).
		WithKey("Y", numbers.NewInt().Any()).
		Validate(in)

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	if out == nil {
		t.Error("Expected output to not be nil")
		return
	}

	if out.X != 11 {
		t.Errorf("Expected output X to be 11 but got %v", out.X)
		return
	}

	if out.Y != 20 {
		t.Errorf("Expected output Y to be 20 but got %v", out.Y)
		return
	}

	if in.X != 10 {
		t.Errorf("Expected input X to still be 10 but got %v", in.X)
		return
	}
}

func TestPanicWhenOutputNotObjectLike(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	objects.New[int]()
}

func TestPanicWhenAssigningRuleSetToMissingField(t *testing.T) {
	defer func() {
		err, ok := recover().(error)

		if err == nil || !ok {
			t.Error("Expected panic with error interface")
		} else if err.Error() != "missing mapping for key: a" {
			t.Errorf("Expected missing mapping error, got: %s", err)
		}
	}()

	objects.New[*testStruct]().WithKey("a", strings.New().Any())
}

// This function is deprecated and will be removed in v1.0.0.
// Until then, make sure it still works.
func TestKeyFunction(t *testing.T) {
	out, err := objects.New[*testStructMapped]().
		Key("A", numbers.NewInt().Any()).
		Key("C", numbers.NewInt().Any()).
		Validate(map[string]any{"A": 123, "C": 456})

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
	out, err := objects.New[*testStructMapped]().
		WithKey("A", numbers.NewInt().Any()).
		WithKey("C", numbers.NewInt().Any()).
		Validate(map[string]any{"A": 123, "C": 456})

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
	out, err := objects.NewObjectMap[int]().
		WithKey("A", numbers.NewInt().Any()).
		WithKey("B", numbers.NewInt().Any()).
		Validate(map[string]any{"A": 123})

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

	out, err := objects.NewObjectMap[int]().
		WithKey("A", numbers.NewInt().Any()).
		WithKey("B", numbers.NewInt().Any()).
		Validate(input)

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
	_, err := objects.NewObjectMap[int]().
		WithKey("A", numbers.NewInt().Any()).
		WithKey("B", numbers.NewInt().WithRequired().Any()).
		Validate(map[string]any{"A": 123})

	if len(err) == 0 {
		t.Errorf("Expected errors to not be empty")
	}
}

func TestWithRequired(t *testing.T) {
	ruleSet := objects.NewObjectMap[int]()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestUnknownFields(t *testing.T) {
	ruleSet := objects.NewObjectMap[int]().WithKey("A", numbers.NewInt().Any())
	value := map[string]any{"A": 123, "C": 456}

	testhelpers.MustBeInvalid(t, ruleSet.Any(), value, errors.CodeUnexpected)

	ruleSet = ruleSet.WithUnknown()
	testhelpers.MustBeValidFunc(t, ruleSet.Any(), value, "", func(_, _ any) error { return nil })
}

func TestInputNotObjectLike(t *testing.T) {
	_, err := objects.New[*testStruct]().
		Validate(123)

	if err == nil {
		t.Error("Expected errors to not be empty")
	}
}

func TestReturnsAllErrors(t *testing.T) {
	_, err := objects.NewObjectMap[int]().
		WithKey("A", numbers.NewInt().WithMax(2).Any()).
		WithKey("B", numbers.NewInt().Any()).
		WithKey("C", strings.New().WithStrict().Any()).
		Validate(map[string]any{"A": 123, "B": 456, "C": 789})

	if err == nil {
		t.Errorf("Expected errors to not be nil")
	} else if len(err) != 2 {
		t.Errorf("Expected 2 errors got %d: %s", len(err), err.Error())
	}
}

func TestReturnsCorrectPaths(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "myobj")

	_, err := objects.NewObjectMap[int]().
		WithKey("A", numbers.NewInt().WithMax(2).Any()).
		WithKey("B", numbers.NewInt().Any()).
		WithKey("C", strings.New().WithStrict().Any()).
		ValidateWithContext(map[string]any{"A": 123, "B": 456, "C": 789}, ctx)

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

func TesMixedMap(t *testing.T) {
	_, err := objects.NewObjectMap[any]().
		WithKey("A", numbers.NewInt().Any()).
		WithKey("B", numbers.NewInt().Any()).
		WithKey("C", strings.New().Any()).
		Validate(map[string]any{"A": 123, "B": 456, "C": "789"})

	if err != nil {
		t.Errorf("Expected errors to be empty %s", err.Error())
		return
	}
}

func TestCustom(t *testing.T) {
	_, err := objects.New[*testStruct]().
		WithRuleFunc(testhelpers.MockCustomRule(testStructInit(), 1)).
		WithRuleFunc(testhelpers.MockCustomRule(testStructInit(), 1)).
		Validate(map[string]any{"A": 123, "B": 456, "C": "789"})

	if err == nil {
		t.Error("Expected errors to not be nil")
	} else if len(err) == 0 {
		t.Error("Expected errors to not be empty")
	}
}

func TestCustomMutation(t *testing.T) {

	result := testStructInit()
	result.z = 123

	obj, err := objects.New[*testStruct]().
		WithRuleFunc(testhelpers.MockCustomRule(result, 0)).
		Validate(map[string]any{})

	if err != nil {
		t.Errorf("Expected errors to be nil, got: %s", err)
	} else if obj.z != 123 {
		t.Errorf("Expected obj.z to be 123, got: %d", obj.z)
	}
}

func TestAny(t *testing.T) {
	ruleSet := numbers.NewFloat64().Any()

	if ruleSet == nil {
		t.Error("Expected Any not be nil")
	} else if _, ok := ruleSet.(rules.RuleSet[any]); !ok {
		t.Error("Expected Any not implement RuleSet[any]")
	}
}

func TestPointer(t *testing.T) {
	// W is a pointer to an int
	ruleSet := objects.New[*testStruct]().WithKey("W", numbers.NewInt().Any())

	obj, err := ruleSet.Validate(map[string]any{"W": 123})

	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	} else if obj.W == nil {
		t.Errorf("Expected obj.W to not be nil")
	} else if *obj.W != 123 {
		t.Errorf("Expected obj.W to be 123, got: %d", obj.W)
	}
}

type testStructMappedBug struct {
	Email string `validate:"email"`
}

// This tests for an issue where the value could be set when the struct is not a pointer.
//
// See: https://github.com/proto-studio/protovalidate/issues/1
func TestBug001(t *testing.T) {
	ruleSet := objects.New[testStructMappedBug]().
		WithKey("email", strings.New().Any())

	expected := "hello@example.com"

	out, err := ruleSet.Validate(map[string]any{
		"email": expected,
	})

	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	} else if out.Email != expected {
		t.Errorf("Expected email to  be '%s', got: '%s'", expected, out.Email)
	}
}

// Requirements:
// - Serializes to WithRequired()
func TestRequiredString(t *testing.T) {
	ruleSet := objects.New[*testStruct]().WithRequired()

	expected := "ObjectRuleSet[*objects_test.testStruct].WithRequired()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithUnknown()
func TestAllowUnknownString(t *testing.T) {
	ruleSet := objects.New[*testStruct]().WithUnknown()

	expected := "ObjectRuleSet[*objects_test.testStruct].WithUnknown()"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithItemRuleSet()
func TestWithItemRuleSetString(t *testing.T) {
	ruleSet := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("Y", numbers.NewInt().Any())

	expected := "ObjectRuleSet[*objects_test.testStruct].WithKey(\"X\", IntRuleSet[int].Any()).WithKey(\"Y\", IntRuleSet[int].Any())"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Serializes to WithRule()
func TestWithRuleString(t *testing.T) {
	ruleSet := objects.New[*testStruct]().
		WithRuleFunc(testhelpers.MockCustomRule(testStructInit(), 0))

	expected := "ObjectRuleSet[*objects_test.testStruct].WithRuleFunc(...)"
	if s := ruleSet.String(); s != expected {
		t.Errorf("Expected rule set to be %s, got %s", expected, s)
	}
}

// Requirements:
// - Evaluate behaves like ValidateWithContext
func TestEvaluate(t *testing.T) {
	ctx := context.Background()

	ruleSet := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("Y", numbers.NewInt().Any())

	input := testStructInit()
	input.X = 12
	input.Y = 34

	v1, err1 := ruleSet.Evaluate(ctx, input)
	v2, err2 := ruleSet.ValidateWithContext(input, ctx)

	if !reflect.DeepEqual(v1, v2) {
		t.Errorf("Expected values to match, got %v and %v", v1, v2)
	}

	if err1 != nil || err2 != nil {
		t.Errorf("Expected errors to both be nil, got %s and %s", err1, err2)
	}
}

// Requirements:
// - Multiple rules on the same key all evaluate
func TestMultipleRules(t *testing.T) {
	ruleSet := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().WithMin(2).Any()).
		WithKey("X", numbers.NewInt().WithMax(4).Any()).
		Any()

	testhelpers.MustBeValidFunc(t, ruleSet, &testStruct{X: 3}, &testStruct{X: 3}, func(a, b any) error {
		if a.(*testStruct).X != b.(*testStruct).X {
			return fmt.Errorf("Expected X to be %d, got: %d", b.(*testStruct).X, a.(*testStruct).X)
		}
		return nil
	})
	testhelpers.MustBeInvalid(t, ruleSet, &testStruct{X: 1}, errors.CodeMin)
	testhelpers.MustBeInvalid(t, ruleSet, &testStruct{X: 5}, errors.CodeMax)
}

// Requirement:
// This test is specifically for a timeout while performing an object rule (as opposed to a key rule)
// - RuleSet times out if context does
// - Timeout error is returned
func TestTimeoutInObjectRule(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	ruleSet := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().WithMin(2).Any()).
		WithRuleFunc(func(_ context.Context, x *testStruct) (*testStruct, errors.ValidationErrorCollection) {
			time.Sleep(1 * time.Second)
			return x, nil
		})

	_, errs := ruleSet.ValidateWithContext(&testStruct{}, ctx)

	if errs == nil {
		t.Error("Expected errors to be nil")
	} else if len(errs) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errs))
	} else if c := errs.For("").First().Code(); c != errors.CodeTimeout {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeTimeout, errs, c)
	}
}

// Requirement:
// This test is specifically for a timeout while performing an key rule (as opposed to an object rule)
// - RuleSet times out if context does
// - Timeout error is returned
func TestTimeoutInKeyRule(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ruleSet := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().WithRuleFunc(func(_ context.Context, x int) (int, errors.ValidationErrorCollection) {
			time.Sleep(1 * time.Second)
			return x, nil
		}).Any())

	_, errs := ruleSet.ValidateWithContext(&testStruct{}, ctx)

	if errs == nil {
		t.Error("Expected errors to be nil")
	} else if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(errs), errs)
	} else if c := errs.For("").First().Code(); c != errors.CodeTimeout {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeTimeout, errs, c)
	}
}

// Requirement:
// - Rules stop running after the context is cancelled
func TestCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var intCallCount int32 = 0
	var structCallCount int32 = 0

	intRule := func(_ context.Context, x int) (int, errors.ValidationErrorCollection) {
		atomic.AddInt32(&intCallCount, 1)
		cancel()
		time.Sleep(1 * time.Second)
		return x, nil
	}

	structRule := func(_ context.Context, x *testStruct) (*testStruct, errors.ValidationErrorCollection) {
		atomic.AddInt32(&structCallCount, 1)
		time.Sleep(1 * time.Second)
		return x, nil
	}

	ruleSet := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().WithRuleFunc(intRule).Any()).
		WithKey("X", numbers.NewInt().WithRuleFunc(intRule).Any()).
		WithRuleFunc(structRule).
		WithRuleFunc(structRule)

	_, errs := ruleSet.ValidateWithContext(&testStruct{}, ctx)

	if errs == nil {
		t.Error("Expected errors to be nil")
	} else if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(errs), errs)
	} else if c := errs.First().Code(); c != errors.CodeCancelled {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeCancelled, errs, c)
	}

	// If these two rules succeed but the ones above fail, check to make sure "wait" is only called once

	finalCallCount := atomic.LoadInt32(&intCallCount)
	if finalCallCount != 1 {
		t.Errorf("Expected a call count of 1, got %d", finalCallCount)
	}

	finalCallCount = atomic.LoadInt32(&structCallCount)
	if finalCallCount != 0 {
		t.Errorf("Expected a call count of 0, got %d", finalCallCount)
	}
}

// Requirement:
// - Object rules stop running after a cancel
func TestCancelledObjectRules(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var structCallCount int32 = 0

	structRule := func(_ context.Context, x *testStruct) (*testStruct, errors.ValidationErrorCollection) {
		atomic.AddInt32(&structCallCount, 1)
		cancel()
		time.Sleep(1 * time.Second)
		return x, nil
	}

	ruleSet := objects.New[*testStruct]().
		WithRuleFunc(structRule).
		WithRuleFunc(structRule)

	_, errs := ruleSet.ValidateWithContext(&testStruct{}, ctx)

	if errs == nil {
		t.Error("Expected errors to be nil")
	} else if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d: %s", len(errs), errs)
	} else if c := errs.First().Code(); c != errors.CodeCancelled {
		t.Errorf("Expected error to be %s, got %s (%s)", errors.CodeCancelled, errs, c)
	}

	finalCallCount := atomic.LoadInt32(&structCallCount)
	if finalCallCount != 1 {
		t.Errorf("Expected a call count of 1, got %d", finalCallCount)
	}
}

// Requirement:
// - Conditional rules are called only when the condition returns no errors
// - Conditional rules are not called until dependent keys are evaluated
func TestConditionalKey(t *testing.T) {
	// This rule mutates the value of X.
	// If the condition is evaluated before this rule finishes then the value will be incorrect
	intRule := func(_ context.Context, x int) (int, errors.ValidationErrorCollection) {
		time.Sleep(100 * time.Millisecond)
		return x * 2, nil
	}

	condValueRule := func(_ context.Context, y int) (int, errors.ValidationErrorCollection) {
		return y * 3, nil
	}

	// Only run the conditional rule if X is greater than 4. Which it should only be if the intRule
	// function ran.
	condKeyRuleSet := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().WithMin(4).Any())

	ruleSet := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().WithRuleFunc(intRule).Any()).
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

	// Both X and Y should be mutated
	testhelpers.MustBeValidFunc(t, ruleSet.Any(), &testStruct{X: 3, Y: 3}, &testStruct{X: 6, Y: 9}, checkFn)

	// Only X should be mutated
	testhelpers.MustBeValidFunc(t, ruleSet.Any(), &testStruct{X: 1, Y: 3}, &testStruct{X: 2, Y: 0}, checkFn)
}

// Requirement:
// - Returns all keys with rules
// - Does not return keys with no rules
// - Returns conditional keys
// - Only returns each key once
func TestKeys(t *testing.T) {

	ruleSet := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().Any()).
		WithKey("X", numbers.NewInt().Any()).
		WithConditionalKey("Y", objects.New[*testStruct](), numbers.NewInt().Any())

	keys := ruleSet.Keys()

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d (%s)", len(keys), keys)
	} else if !((keys[0] == "X" && keys[1] == "Y") || (keys[0] == "Y" && keys[1] == "X")) {
		t.Errorf("Expected [X Y], got %s", keys)
	}

}

// Requirement:
// - The code panics is a cycle is made directly with conditional keys
func TestConditionalKeyCycle(t *testing.T) {
	condX := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().WithMin(4).Any())

	condY := objects.New[*testStruct]().
		WithKey("Y", numbers.NewInt().WithMin(4).Any())

	ruleSet := objects.New[*testStruct]().
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
	condX := objects.New[*testStruct]().
		WithKey("X", numbers.NewInt().WithMin(4).Any())

	condY := objects.New[*testStruct]().
		WithKey("Y", numbers.NewInt().WithMin(4).Any())

	condW := objects.New[*testStruct]().
		WithKey("W", numbers.NewInt().WithMin(4).Any())

	ruleSet := objects.New[*testStruct]().
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

	condB := objects.NewObjectMap[int]().
		WithKey("B", numbers.NewInt().WithMin(4).Any())

	condC := objects.NewObjectMap[int]().
		WithKey("C", numbers.NewInt().WithMin(4).Any())

	condD := objects.NewObjectMap[int]().
		WithKey("D", numbers.NewInt().WithMin(4).Any())

	objects.NewObjectMap[int]().
		WithConditionalKey("B", condD, numbers.NewInt().Any()).
		WithConditionalKey("C", condD, numbers.NewInt().Any()).
		WithConditionalKey("A", condB, numbers.NewInt().Any()).
		WithConditionalKey("A", condC, numbers.NewInt().Any())
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
	ruleSet := objects.New[*testStructMapped]().
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

	testhelpers.MustBeValidFunc(t, ruleSet.Any(), in, in, check)

	in.A = 3
	testhelpers.MustBeInvalid(t, ruleSet.Any(), in, errors.CodeMin)

	in.A = 5

	in.B = 50
	testhelpers.MustBeInvalid(t, ruleSet.Any(), in, errors.CodeMin)

	in.B = 150
	testhelpers.MustBeValidFunc(t, ruleSet.Any(), *in, in, check)
}

// Requirements:
// - Will assign nested pointer structs to pointers
//
// Fixes issue:
// **objects_test.testStructMapped is not assignable to type *objects_test.testStruct
func TestNestedPointer(t *testing.T) {

	type target struct {
		Test *testStruct
	}

	ruleSet := objects.New[*target]().
		WithKey("Test", objects.New[*testStruct]().WithUnknown().Any())

	in := map[string]any{
		"Test": &testStruct{},
	}

	testhelpers.MustBeValidFunc(t, ruleSet.Any(), in, in, func(a, b any) error { return nil })
}

// Requirement:
// - When WithUnkown is set, the resulting map should contain unknown values
func TestObjectFromMapToMapUknown(t *testing.T) {
	in := testMap()

	out, err := objects.NewObjectMap[any]().
		WithUnknown().
		WithKey("X", numbers.NewInt().Any()).
		Validate(in)

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
