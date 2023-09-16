package numbers_test

import (
	"fmt"
	"strings"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestCoerceToInt(t *testing.T) {
	expected := int(123)
	ruleSet := numbers.NewInt().Any()

	testhelpers.MustBeValid(t, ruleSet, int(123), expected)
	testhelpers.MustBeValid(t, ruleSet, int8(123), expected)
	testhelpers.MustBeValid(t, ruleSet, int16(123), expected)
	testhelpers.MustBeValid(t, ruleSet, int32(123), expected)
	testhelpers.MustBeValid(t, ruleSet, int64(123), expected)

	testhelpers.MustBeValid(t, ruleSet, uint(123), expected)
	testhelpers.MustBeValid(t, ruleSet, uint8(123), expected)
	testhelpers.MustBeValid(t, ruleSet, uint16(123), expected)
	testhelpers.MustBeValid(t, ruleSet, uint32(123), expected)
	testhelpers.MustBeValid(t, ruleSet, uint64(123), expected)

	testhelpers.MustBeValid(t, ruleSet, float32(123.0), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(123.0), expected)
}

func TestCoerceToInt8(t *testing.T) {
	expected := int8(12)
	ruleSet := numbers.NewInt8().Any()

	testhelpers.MustBeValid(t, ruleSet, int(12), expected)
	testhelpers.MustBeValid(t, ruleSet, int8(12), expected)
	testhelpers.MustBeValid(t, ruleSet, int16(12), expected)
	testhelpers.MustBeValid(t, ruleSet, int32(12), expected)
	testhelpers.MustBeValid(t, ruleSet, int64(12), expected)

	testhelpers.MustBeValid(t, ruleSet, uint(12), expected)
	testhelpers.MustBeValid(t, ruleSet, uint8(12), expected)
	testhelpers.MustBeValid(t, ruleSet, uint16(12), expected)
	testhelpers.MustBeValid(t, ruleSet, uint32(12), expected)
	testhelpers.MustBeValid(t, ruleSet, uint64(12), expected)

	testhelpers.MustBeValid(t, ruleSet, float32(12.0), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(12.0), expected)
}

func TestOutOfRangeInt8(t *testing.T) {
	ruleSet := numbers.NewInt8().Any()

	testhelpers.MustBeInvalid(t, ruleSet, int16(1024), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, int32(1024), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, int64(1024), errors.CodeRange)

	testhelpers.MustBeInvalid(t, ruleSet, float32(1024), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, float64(1024), errors.CodeRange)

	testhelpers.MustBeInvalid(t, ruleSet, "1024", errors.CodeRange)
}

func TestOutOfRangeUInt8(t *testing.T) {
	ruleSet := numbers.NewUint8().Any()

	testhelpers.MustBeInvalid(t, ruleSet, int16(1024), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, int32(1024), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, int64(1024), errors.CodeRange)

	testhelpers.MustBeInvalid(t, ruleSet, int16(-1024), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, int32(-1024), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, int64(-1024), errors.CodeRange)

	testhelpers.MustBeInvalid(t, ruleSet, float32(1024), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, float64(1024), errors.CodeRange)

	testhelpers.MustBeInvalid(t, ruleSet, float32(-1024), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, float64(-1024), errors.CodeRange)
}

func TestStringToInt(t *testing.T) {
	ruleSetBase10 := numbers.NewInt().Any()
	expected := int(123)

	testhelpers.MustBeValid(t, ruleSetBase10, "123", expected)

	err := testhelpers.MustBeInvalid(t, ruleSetBase10, "7B", errors.CodeType)

	if !strings.Contains(err.Error(), "string to int") {
		t.Errorf("Expected error to contain 'string to int', got: %s", err)
	}

	ruleSetBase16 := numbers.NewInt().WithBase(16).Any()
	testhelpers.MustBeValid(t, ruleSetBase16, "7B", expected)

	err = testhelpers.MustBeInvalid(t, ruleSetBase10, "7x", errors.CodeType)

	if !strings.Contains(err.Error(), "string to int") {
		t.Errorf("Expected error to contain 'string to int', got: %s", err)
	}
}

func TestStringToIntOutOfRange(t *testing.T) {
	ruleSetSigned := numbers.NewInt8().Any()
	testhelpers.MustBeInvalid(t, ruleSetSigned, "128", errors.CodeRange)

	ruleSetUnsigned := numbers.NewUint8().Any()
	testhelpers.MustBeValid(t, ruleSetUnsigned, "128", uint8(128))
	testhelpers.MustBeInvalid(t, ruleSetUnsigned, "256", errors.CodeRange)
}

func TestStringToIntInvalid(t *testing.T) {
	ruleSetUnsigned := numbers.NewInt().Any()
	testhelpers.MustBeInvalid(t, ruleSetUnsigned, "hello", errors.CodeType)
}

func TestUnknownToInt(t *testing.T) {
	from := new(struct{})

	ruleSetSigned := numbers.NewInt8().Any()
	testhelpers.MustBeInvalid(t, ruleSetSigned, &from, errors.CodeType)

	ruleSetUnsigned := numbers.NewUint8().Any()
	testhelpers.MustBeInvalid(t, ruleSetUnsigned, &from, errors.CodeType)
}

func TestCoerceToFloat64(t *testing.T) {
	expected := float64(123.0)
	ruleSet := numbers.NewFloat64().Any()

	testhelpers.MustBeValid(t, ruleSet, int(123), expected)
	testhelpers.MustBeValid(t, ruleSet, int8(123), expected)
	testhelpers.MustBeValid(t, ruleSet, int16(123), expected)
	testhelpers.MustBeValid(t, ruleSet, int32(123), expected)
	testhelpers.MustBeValid(t, ruleSet, int64(123), expected)

	testhelpers.MustBeValid(t, ruleSet, uint(123), expected)
	testhelpers.MustBeValid(t, ruleSet, uint8(123), expected)
	testhelpers.MustBeValid(t, ruleSet, uint16(123), expected)
	testhelpers.MustBeValid(t, ruleSet, uint32(123), expected)
	testhelpers.MustBeValid(t, ruleSet, uint64(123), expected)

	testhelpers.MustBeValid(t, ruleSet, float32(123.0), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(123.0), expected)
}

func TestOutOfRangeFloat32(t *testing.T) {
	ruleSet := numbers.NewFloat32().Any()

	testhelpers.MustBeInvalid(t, ruleSet, int32(0x7FFFFFFF), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, int64(0x7FFFFFFFFFFFFFFF), errors.CodeRange)

	testhelpers.MustBeInvalid(t, ruleSet, int32(-0x7FFFFFFF), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, int64(-0x7FFFFFFFFFFFFFFF), errors.CodeRange)

	testhelpers.MustBeInvalid(t, ruleSet, uint32(0xFFFFFFFF), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, uint64(0xFFFFFFFFFFFFFFFF), errors.CodeRange)

	testhelpers.MustBeInvalid(t, ruleSet, float64(1.7e+308), errors.CodeRange)
	testhelpers.MustBeInvalid(t, ruleSet, float64(-1.7e+308), errors.CodeRange)

	testhelpers.MustBeInvalid(t, ruleSet, fmt.Sprintf("%f", 1.7e+308), errors.CodeRange)
}

func TestStringToFloat(t *testing.T) {
	ruleSet := numbers.NewFloat64().Any()
	expected := float64(123.456)

	testhelpers.MustBeValid(t, ruleSet, "123.456", expected)
}

func TestStringToFloatInvalid(t *testing.T) {
	ruleSetUnsigned := numbers.NewFloat64().Any()
	testhelpers.MustBeInvalid(t, ruleSetUnsigned, "hello", errors.CodeType)
}

func TestUnknownToFloat(t *testing.T) {
	from := new(struct{})

	ruleSetSigned := numbers.NewFloat64().Any()
	testhelpers.MustBeInvalid(t, ruleSetSigned, &from, errors.CodeType)

	ruleSetUnsigned := numbers.NewFloat64().Any()
	testhelpers.MustBeInvalid(t, ruleSetUnsigned, &from, errors.CodeType)
}
