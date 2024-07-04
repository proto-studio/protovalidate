package arrays_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/arrays"
)

type MyTestInterface interface {
	Test()
}

type MyTestImplInt int

func (x MyTestImplInt) Test() {}

type MyTestImplStr string

func (x MyTestImplStr) Test() {}

type InterfaceTest struct {
	IntTest    MyTestInterface
	StringTest MyTestInterface
	NilTest    MyTestInterface
}

// Requirements:
// - Can cast to interface.
// - Interface correctly assigns nil values.
func TestInterfaceStruct(t *testing.T) {
	innerRuleSet := rules.Interface[MyTestInterface]().
		WithCast(func(v any) (MyTestInterface, bool) {
			if v == nil {
				return nil, true
			}

			switch vcast := v.(type) {
			case int:
				return MyTestImplInt(int(vcast)), true
			case string:
				return MyTestImplStr(vcast), true
			}
			return nil, false
		})

	ruleSet := arrays.New[MyTestInterface]().WithItemRuleSet(innerRuleSet)

	out, errs := ruleSet.Run(context.TODO(), []any{123, "abc", nil})

	if errs != nil {
		t.Errorf("Expected errors to be empty %s", errs.Error())
		return
	}

	if len(out) < 1 || out[0] == nil {
		t.Errorf("Expected IntTest to not be nil")
	} else if _, ok := out[0].(MyTestImplInt); !ok {
		t.Errorf("Expected IntTest to be an MyTestImplInt. Got: %v", out[0])
	}

	if len(out) < 2 || out[1] == nil {
		t.Errorf("Expected StringTest to not be nil")
	} else if _, ok := out[1].(MyTestImplStr); !ok {
		t.Errorf("Expected StringTest to be an MyTestImplStr. Got: %v", out[1])
	}

	if len(out) < 3 || out[2] != nil {
		t.Errorf("Expected NilTest to be an nil. Got: %v", out[2])
	}
}
