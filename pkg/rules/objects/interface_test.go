package objects_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/objects"
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
}

func InitInterfaceRuleSet() rules.RuleSet[MyTestInterface] {
	return rules.Interface[MyTestInterface]().
		WithCast(func(ctx context.Context, v any) (MyTestInterface, errors.ValidationErrorCollection) {
			if v == nil {
				return nil, nil
			}

			switch vcast := v.(type) {
			case float64:
				return MyTestImplInt(int(vcast)), nil
			case string:
				return MyTestImplStr(vcast), nil
			}
			return nil, nil
		})
}

// Requirements:
// - Can cast to interface.
func TestInterfaceStruct(t *testing.T) {
	innerRuleSet := InitInterfaceRuleSet()

	ruleSet := objects.New[InterfaceTest]().
		WithKey("IntTest", innerRuleSet.Any()).
		WithKey("StringTest", innerRuleSet.Any()).
		WithJson()

	out, errs := ruleSet.Run(context.TODO(), `{"IntTest":123, "StringTest":"abc"}`)

	if errs != nil {
		t.Errorf("Expected errors to be empty %s", errs.Error())
		return
	}

	if out.IntTest == nil {
		t.Errorf("Expected IntTest to not be nil")
	} else if _, ok := out.IntTest.(MyTestImplInt); !ok {
		t.Errorf("Expected IntTest to be an MyTestImplInt. Got: %v", out.IntTest)
	}

	if out.StringTest == nil {
		t.Errorf("Expected StringTest to not be nil")
	} else if _, ok := out.StringTest.(MyTestImplStr); !ok {
		t.Errorf("Expected StringTest to be an MyTestImplStr. Got: %v", out.StringTest)
	}
}
