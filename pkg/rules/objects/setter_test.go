package objects

import (
	"reflect"
	"testing"
)

func TestStructSetter_Set(t *testing.T) {
	type outType struct {
		X *string
		Y string
	}

	out := &outType{}

	setter := structSetter[string]{
		out: reflect.Indirect(reflect.ValueOf(out)),
		mapping: map[string]string{
			"A": "X",
			"B": "Y",
		},
	}

	// Can set pointer key
	expected := "hello"
	setter.Set("A", expected)
	if out.X == nil {
		t.Errorf(`Expected out.X to not be nil`)
	} else if *out.X != expected {
		t.Errorf(`Expected out.X to be "%s", got: "%s"`, expected, *out.X)
	}

	// Can set static key
	setter.Set("B", expected)
	if out.Y != expected {
		t.Errorf(`Expected out.Y to be "%s", got: "%s"`, expected, out.Y)
	}

	// Can set nil
	setter.Set("A", nil)
	if out.X != nil {
		t.Errorf(`Expected out.X to be nil, got: "%s"`, *out.X)
	}
}

func TestStructSetter_SetBucket(t *testing.T) {
	type outType struct {
		B map[string]any
	}

	out := &outType{}

	setter := structSetter[string]{
		out:     reflect.Indirect(reflect.ValueOf(out)),
		mapping: map[string]string{},
	}

	// Creates bucket if it is not already created
	setter.SetBucket("B", "A", 123)
	if out.B == nil {
		t.Fatal("Expected bucket to not be nil")
	}
	setter.SetBucket("B", "B", nil)

	if out.B["A"] != 123 {
		t.Errorf(`Expected out.B["A"] to be 123, got %v`, out.B["A"])
	}
	if out.B["B"] != nil {
		t.Errorf(`Expected out.B["B"] to be nil, got %v`, out.B["B"])
	}
}

func TestStructSetter_SetBucket_IncorrectType(t *testing.T) {
	type outType struct {
		B string
	}

	out := &outType{}

	setter := structSetter[string]{
		out:     reflect.Indirect(reflect.ValueOf(out)),
		mapping: map[string]string{},
	}

	// Should not error
	setter.SetBucket("B", "A", 123)
}

func TestMapSetter_Set(t *testing.T) {
	out := make(map[string]*string)

	setter := mapSetter[string]{
		out: reflect.Indirect(reflect.ValueOf(out)),
	}

	// Can set key
	expected := "hello"
	setter.Set("X", &expected)
	if _, ok := out["X"]; !ok {
		t.Errorf(`Expected out["X"] to be set`)
	} else if out["X"] == nil {
		t.Errorf(`Expected out["X"] to not be nil`)
	} else if *out["X"] != expected {
		t.Errorf(`Expected out["X"] to be "%s", got: "%s"`, expected, *out["X"])
	}

	// Can set nil
	setter.Set("X", nil)
	if out["X"] != nil {
		t.Errorf(`Expected out.X to be nil, got: "%s"`, *out["X"])
	}
}

func TestMapSetter_SetBucket(t *testing.T) {
	out := make(map[string]any)

	setter := mapSetter[string]{
		out: reflect.Indirect(reflect.ValueOf(out)),
	}

	// Creates bucket if it is not already created
	setter.SetBucket("B", "A", 123)

	bucket, ok := out["B"].(map[string]any)

	if !ok {
		t.Fatalf(`Expected out["B"] to be a bucket of type map[string] any, got: %T`, out["B"])
	}
	setter.SetBucket("B", "B", nil)

	if bucket["A"] != 123 {
		t.Errorf(`Expected out.B["A"] to be 123, got %v`, bucket["A"])
	}
	if bucket["B"] != nil {
		t.Errorf(`Expected out.B["B"] to be nil, got %v`, bucket["B"])
	}
}
