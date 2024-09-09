package rules_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

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
				return MyTestImplInt(vcast), nil
			case string:
				return MyTestImplStr(vcast), nil
			}
			return nil, nil
		})

	ruleSet := rules.NewSlice[MyTestInterface]().WithItemRuleSet(innerRuleSet)

	// Prepare an output variable for Apply
	var output []MyTestInterface

	// Use Apply instead of Run
	errs := ruleSet.Apply(context.TODO(), []any{123, "abc"}, &output)

	if errs != nil {
		t.Errorf("Expected errors to be empty %s", errs.Error())
		return
	}

	// Check the output for the expected types
	if len(output) < 1 || output[0] == nil {
		t.Errorf("Expected IntTest to not be nil")
	} else if _, ok := output[0].(MyTestImplInt); !ok {
		t.Errorf("Expected IntTest to be a MyTestImplInt. Got: %v", output[0])
	}

	if len(output) < 2 || output[1] == nil {
		t.Errorf("Expected StringTest to not be nil")
	} else if _, ok := output[1].(MyTestImplStr); !ok {
		t.Errorf("Expected StringTest to be a MyTestImplStr. Got: %v", output[1])
	}
}
