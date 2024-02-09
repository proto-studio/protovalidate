package testhelpers_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestMockRule(t *testing.T) {
	ctx := context.Background()

	rule1 := testhelpers.NewMockRuleWithValue(any(123)).Function()

	ret, err := rule1(ctx, 456)
	if err != nil {
		t.Errorf("Expected error to be nil, got: %s", err)
	} else if ret != 123 {
		t.Errorf("Expected return value to be %d, got: %d", 123, ret)
	}

	rule2 := testhelpers.NewMockRuleWithErrors[any](1).Function()

	_, err = rule2(ctx, 456)
	if err == nil {
		t.Error("Expected error to not be nil")
	} else if s := len(err); s != 1 {
		t.Errorf("Expected error collection size to be %d, got: %d", 1, s)
	}

	rule3 := testhelpers.NewMockRuleWithErrors[any](2).Function()

	_, err = rule3(ctx, 456)
	if err == nil {
		t.Error("Expected error to not be nil")
	} else if s := len(err); s != 2 {
		t.Errorf("Expected error collection size to be %d, got: %d", 2, s)
	}
}

func TestMockConflict(t *testing.T) {
	mockA := testhelpers.NewMockRuleWithValue[int](10)
	mockA.ConflictKey = "a"

	mockB := testhelpers.NewMockRuleWithValue[int](20)
	mockB.ConflictKey = "b"

	mockAA := testhelpers.NewMockRuleWithValue[int](20)
	mockAA.ConflictKey = "a"

	var mockD rules.RuleFunc[int] = testhelpers.NewMockRuleWithValue[int](30).Function()

	if mockA.Conflict(mockB) {
		t.Errorf("Expected mockA and mockB to not conflict")
	}

	if !mockA.Conflict(mockAA) {
		t.Errorf("Expected mockA and mockAA to conflict")
	}

	if mockA.Conflict(mockD) {
		t.Errorf("Expected mockA and mockD to not conflict")
	}
}

func TestMockString(t *testing.T) {
	str := testhelpers.NewMockRuleWithValue[int](10).String()

	if str != "WithMock()" {
		t.Errorf("Expected mock string to be `%s`, got: %s", "WithMock()", str)
	}
}

func TestMockCounter(t *testing.T) {
	mock := testhelpers.NewMockRuleWithValue[int](10)
	ctx := context.Background()

	if mock.CallCount() != 0 {
		t.Errorf("Expected call count to be %d, got %d", 0, mock.CallCount())
	}

	mock.Evaluate(ctx, 1)

	if mock.CallCount() != 1 {
		t.Errorf("Expected call count to be %d, got %d", 1, mock.CallCount())
	}

	mock.Evaluate(ctx, 1)
	mock.Evaluate(ctx, 1)

	if mock.CallCount() != 3 {
		t.Errorf("Expected call count to be %d, got %d", 3, mock.CallCount())
	}

	mock.Reset()

	if mock.CallCount() != 0 {
		t.Errorf("Expected call count to be %d, got %d", 0, mock.CallCount())
	}
}
