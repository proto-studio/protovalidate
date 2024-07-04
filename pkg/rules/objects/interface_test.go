package objects_test

import (
	"context"
	"fmt"
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/rules/objects"
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

func InitInterfaceRuleSet() rules.RuleSet[MyTestInterface] {
	return rules.Interface[MyTestInterface]().
		WithCast(func(v any) (MyTestInterface, bool) {
			if v == nil {
				return nil, true
			}

			switch vcast := v.(type) {
			case float64:
				return MyTestImplInt(int(vcast)), true
			case string:
				return MyTestImplStr(vcast), true
			}
			return nil, false
		})
}

// Requirements:
// - Can cast to interface.
// - Interface correctly assigns nil values.
func TestInterfaceStruct(t *testing.T) {
	innerRuleSet := InitInterfaceRuleSet()

	ruleSet := objects.New[InterfaceTest]().
		WithKey("IntTest", innerRuleSet.Any()).
		WithKey("StringTest", innerRuleSet.Any()).
		WithKey("NilTest", innerRuleSet.Any()).
		WithJson()

	out, errs := ruleSet.Run(context.TODO(), `{"IntTest":123, "StringTest":"abc", "NilTest": null}`)

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

	if out.NilTest != nil {
		t.Errorf("Expected NilTest to be an nil. Got: %v", out.NilTest)
	}
}

// Requirements:
// - Can cast to interface.
// - Interface correctly assigns nil values.
func TestInterfaceMap(t *testing.T) {
	innerRuleSet := InitInterfaceRuleSet()

	ruleSet := objects.NewObjectMap[MyTestInterface]().
		WithKey("IntTest", innerRuleSet).
		WithKey("StringTest", innerRuleSet).
		WithKey("NilTest", innerRuleSet).
		WithJson()

	out, errs := ruleSet.Run(context.TODO(), `{"IntTest":123, "StringTest":"abc", "NilTest": null}`)

	if errs != nil {
		t.Errorf("Expected errors to be empty %s", errs.Error())
		return
	}

	fmt.Printf("%v\n", out)

	if v, ok := out["IntTest"]; v == nil || !ok {
		t.Errorf("Expected IntTest to not be nil")
	} else if v, ok := out["IntTest"].(MyTestImplInt); !ok {
		t.Errorf("Expected IntTest to be an MyTestImplInt. Got: %v", v)
	}

	if v, ok := out["StringTest"]; v == nil || !ok {
		t.Errorf("Expected StringTest to not be nil")
	} else if v, ok := out["StringTest"].(MyTestImplStr); !ok {
		t.Errorf("Expected StringTest to be an MyTestImplStr. Got: %v", v)
	}

	v, ok := out["NilTest"]

	if v != nil {
		t.Errorf("Expected NilTest to be an nil. Got: %v", v)
	} else if !ok {
		t.Errorf("Expected NilTest to be in map.")
	}
}
