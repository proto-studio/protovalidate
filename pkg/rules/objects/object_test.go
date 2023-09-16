package objects_test

import (
	"context"
	"testing"

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

func testStructMappedInit() *testStructMapped {
	return &testStructMapped{}
}

func TestObjectRuleSet(t *testing.T) {
	_, err := objects.New(testStructInit).
		Key("X", numbers.NewInt().Any()).
		Key("Y", numbers.NewInt().Any()).
		Validate(testMap())

	if err != nil {
		t.Error("Expected errors to be empty")
		return
	}

	ok := testhelpers.CheckRuleSetInterface[*testStruct](objects.New(testStructInit))
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}
}

func TestObjectFromMapToMap(t *testing.T) {
	in := testMap()

	out, err := objects.ObjectMap[any]().
		Key("X", numbers.NewInt().WithRuleFunc(mutateIntPlusOne).Any()).
		Key("Y", numbers.NewInt().Any()).
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

	out, err := objects.New(testStructInit).
		Key("X", numbers.NewInt().WithRuleFunc(mutateIntPlusOne).Any()).
		Key("Y", numbers.NewInt().Any()).
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

	out, err := objects.ObjectMap[any]().
		Key("X", numbers.NewInt().WithRuleFunc(mutateIntPlusOne).Any()).
		Key("Y", numbers.NewInt().Any()).
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

	out, err := objects.New(testStructInit).
		Key("X", numbers.NewInt().WithRuleFunc(mutateIntPlusOne).Any()).
		Key("Y", numbers.NewInt().Any()).
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

	objects.New(func() int { return 0 })
}

func TestPanicWhenAssigningRuleSetToMissingField(t *testing.T) {
	defer func() {
		err, ok := recover().(error)

		if err == nil || !ok {
			t.Error("Expected panic with error interface")
		} else if err.Error() != "missing mapping for key" {
			t.Errorf("Expected missing mapping error, got: %s", err)
		}
	}()

	objects.New(testStructInit).Key("a", strings.New().Any())
}

func TestObjectMapping(t *testing.T) {
	out, err := objects.New(testStructMappedInit).
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

func TestMissingField(t *testing.T) {
	out, err := objects.ObjectMap[int]().
		Key("A", numbers.NewInt().Any()).
		Key("B", numbers.NewInt().Any()).
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

func TestMissingRequiredField(t *testing.T) {
	_, err := objects.ObjectMap[int]().
		Key("A", numbers.NewInt().Any()).
		Key("B", numbers.NewInt().WithRequired().Any()).
		Validate(map[string]any{"A": 123})

	if err.Size() == 0 {
		t.Error("Expected errors to not be empty")
		return
	}
}

func TestWithRequired(t *testing.T) {
	ruleSet := objects.ObjectMap[int]()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}
}

func TestUnknownFields(t *testing.T) {
	ruleSet := objects.ObjectMap[int]().Key("A", numbers.NewInt().Any())
	value := map[string]any{"A": 123, "C": 456}

	testhelpers.MustBeInvalid(t, ruleSet.Any(), value, errors.CodeUnexpected)

	ruleSet = ruleSet.WithUknown()
	testhelpers.MustBeValidFunc(t, ruleSet.Any(), value, "", func(_, _ any) error { return nil })
}

func TestInputNotObjectLike(t *testing.T) {
	_, err := objects.New(testStructInit).
		Validate(123)

	if err == nil {
		t.Error("Expected errors to not be empty")
	}
}

func TestReturnsAllErrors(t *testing.T) {
	_, err := objects.ObjectMap[int]().
		Key("A", numbers.NewInt().WithMax(2).Any()).
		Key("B", numbers.NewInt().Any()).
		Key("C", strings.New().WithStrict().Any()).
		Validate(map[string]any{"A": 123, "B": 456, "C": 789})

	if err == nil {
		t.Errorf("Expected errors to not be nil")
	} else if err.Size() != 2 {
		t.Errorf("Expected 2 errors got %d: %s", err.Size(), err.Error())
	}
}

func TestReturnsCorrectPaths(t *testing.T) {
	ctx := rulecontext.WithPathString(context.Background(), "myobj")

	_, err := objects.ObjectMap[int]().
		Key("A", numbers.NewInt().WithMax(2).Any()).
		Key("B", numbers.NewInt().Any()).
		Key("C", strings.New().WithStrict().Any()).
		ValidateWithContext(map[string]any{"A": 123, "B": 456, "C": 789}, ctx)

	if err == nil {
		t.Errorf("Expected errors to not be nil")
	} else if err.Size() != 2 {
		t.Errorf("Expected 2 errors got %d: %s", err.Size(), err.Error())
		return
	}

	errA := err.For("myobj.A")
	if errA == nil {
		t.Errorf("Expected error for myobj.A to not be nil")
	} else if errA.Size() != 1 {
		t.Errorf("Expected exactly 1 error for myobj.A got %d", errA.Size())
	} else if errA.First().Path() != "myobj.A" {
		t.Errorf("Expected error path to be `%s` got `%s`", "myobj.A", errA.First().Path())
	}

	errC := err.For("myobj.C")
	if errC == nil {
		t.Errorf("Expected error for myobj.C to not be nil")
	} else if errC.Size() != 1 {
		t.Errorf("Expected exactly 1 error for myobj.C got %d", errC.Size())
	} else if errC.First().Path() != "myobj.C" {
		t.Errorf("Expected error path to be `%s` got `%s`", "myobj.C", errC.First().Path())
	}
}

func TesMixedMap(t *testing.T) {
	_, err := objects.ObjectMap[any]().
		Key("A", numbers.NewInt().Any()).
		Key("B", numbers.NewInt().Any()).
		Key("C", strings.New().Any()).
		Validate(map[string]any{"A": 123, "B": 456, "C": "789"})

	if err != nil {
		t.Errorf("Expected errors to be empty %s", err.Error())
		return
	}
}

func TestCustom(t *testing.T) {
	_, err := objects.New(testStructInit).
		WithRuleFunc(testhelpers.MockCustomRule(testStructInit(), 1)).
		WithRuleFunc(testhelpers.MockCustomRule(testStructInit(), 1)).
		Validate(map[string]any{"A": 123, "B": 456, "C": "789"})

	if err == nil {
		t.Error("Expected errors to not be nil")
	} else if err.Size() == 0 {
		t.Error("Expected errors to not be empty")
	}
}

func TestCustomMutation(t *testing.T) {

	result := testStructInit()
	result.z = 123

	obj, err := objects.New(testStructInit).
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
	ruleSet := objects.New(testStructInit).Key("W", numbers.NewInt().Any())

	obj, err := ruleSet.Validate(map[string]any{"W": 123})

	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	} else if obj.W == nil {
		t.Errorf("Expected obj.W to not be nil")
	} else if *obj.W != 123 {
		t.Errorf("Expected obj.W to be 123, got: %d", obj.W)
	}
}
