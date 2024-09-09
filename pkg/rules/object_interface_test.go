package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

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

	ruleSet := rules.NewStruct[InterfaceTest]().
		WithKey("IntTest", innerRuleSet.Any()).
		WithKey("StringTest", innerRuleSet.Any()).
		WithJson()

	// Prepare the output variable for Apply
	var out InterfaceTest

	// Use Apply instead of Run
	errs := ruleSet.Apply(context.TODO(), `{"IntTest":123, "StringTest":"abc"}`, &out)

	if errs != nil {
		t.Errorf("Expected errors to be empty %s", errs.Error())
		return
	}

	// Check the IntTest field
	if out.IntTest == nil {
		t.Errorf("Expected IntTest to not be nil")
	} else if _, ok := out.IntTest.(MyTestImplInt); !ok {
		t.Errorf("Expected IntTest to be a MyTestImplInt. Got: %v", out.IntTest)
	}

	// Check the StringTest field
	if out.StringTest == nil {
		t.Errorf("Expected StringTest to not be nil")
	} else if _, ok := out.StringTest.(MyTestImplStr); !ok {
		t.Errorf("Expected StringTest to be a MyTestImplStr. Got: %v", out.StringTest)
	}
}
