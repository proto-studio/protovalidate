package objects

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules/numbers"
)

type testStruct struct {
	X int
	Y int
	z int //lint:ignore U1000 Used in reflection testing but not code
}

func testStructInit() *testStruct {
	return &testStruct{}
}

func TestMissingMapping(t *testing.T) {
	ruleSet := New(testStructInit).withParent()

	// Manually create a mapping that is not on the struct
	ruleSet.key = "A"
	ruleSet.mapping = "A"

	// This should work
	ruleSet = ruleSet.Key("X", numbers.NewInt().Any())

	// This should panic

	defer func() {
		err, ok := recover().(error)

		if err == nil || !ok {
			t.Error("Expected panic with error interface")
		} else if err.Error() != "missing destination mapping for field" {
			t.Errorf("Expected missing mapping error, got: %s", err)
		}
	}()

	ruleSet = ruleSet.Key("A", numbers.NewInt().Any())
}

func TestUnexportedField(t *testing.T) {
	defer func() {
		err, ok := recover().(error)

		if err == nil || !ok {
			t.Error("Expected panic with error interface")
		} else if err.Error() != "field is not exported" {
			t.Errorf("Expected field is not exported error, got: %s", err)
		}
	}()

	ruleSet := New(testStructInit).withParent()

	// Manually create a mapping for the unexported field
	ruleSet.key = "z"
	ruleSet.mapping = "z"

	ruleSet = ruleSet.Key("z", numbers.NewInt().Any())
}
