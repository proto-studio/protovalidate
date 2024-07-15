package arrays_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/arrays"
)

type MyTestInterface interface {
	internal()
}

type MyTestImplInt int

func (x MyTestImplInt) internal() {}

type MyTestImplStr string

func (x MyTestImplStr) internal() {}

type InterfaceTest struct {
	IntTest    MyTestInterface
	StringTest MyTestInterface
	NilTest    MyTestInterface
}

// Requirements:
// - Can cast to interface.
func TestInterfaceStruct(t *testing.T) {
	innerRuleSet := rules.Interface[MyTestInterface]().
		WithCast(func(ctx context.Context, v any) (MyTestInterface, errors.ValidationErrorCollection) {
			if v == nil {
				return nil, nil
			}

			switch vcast := v.(type) {
			case int:
				return MyTestImplInt(int(vcast)), nil
			case string:
				return MyTestImplStr(vcast), nil
			}
			return nil, nil
		})

	ruleSet := arrays.New[MyTestInterface]().WithItemRuleSet(innerRuleSet)

	out, errs := ruleSet.Run(context.TODO(), []any{123, "abc"})

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
}
