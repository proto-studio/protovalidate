package rules_test

import (
	"fmt"
	"strings"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestCoerceToInt(t *testing.T) {
	expected := int(123)
	ruleSet := rules.Int().Any()

	testhelpers.MustApplyMutation(t, ruleSet, int(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int8(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int16(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int32(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int64(123), expected)

	testhelpers.MustApplyMutation(t, ruleSet, uint(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint8(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint16(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint32(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint64(123), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float32(123.0), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(123.0), expected)
}

func TestCoerceToInt8(t *testing.T) {
	expected := int8(12)
	ruleSet := rules.Int8().Any()

	testhelpers.MustApplyMutation(t, ruleSet, int(12), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int8(12), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int16(12), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int32(12), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int64(12), expected)

	testhelpers.MustApplyMutation(t, ruleSet, uint(12), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint8(12), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint16(12), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint32(12), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint64(12), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float32(12.0), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(12.0), expected)
}

func TestOutOfRangeInt8(t *testing.T) {
	ruleSet := rules.Int8().Any()

	testhelpers.MustNotApply(t, ruleSet, int16(1024), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, int32(1024), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, int64(1024), errors.CodeRange)

	testhelpers.MustNotApply(t, ruleSet, float32(1024), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, float64(1024), errors.CodeRange)

	testhelpers.MustNotApply(t, ruleSet, "1024", errors.CodeRange)
}

func TestOutOfRangeUInt8(t *testing.T) {
	ruleSet := rules.Uint8().Any()

	testhelpers.MustNotApply(t, ruleSet, int16(1024), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, int32(1024), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, int64(1024), errors.CodeRange)

	testhelpers.MustNotApply(t, ruleSet, int16(-1024), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, int32(-1024), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, int64(-1024), errors.CodeRange)

	testhelpers.MustNotApply(t, ruleSet, float32(1024), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, float64(1024), errors.CodeRange)

	testhelpers.MustNotApply(t, ruleSet, float32(-1024), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, float64(-1024), errors.CodeRange)
}

func TestStringToInt(t *testing.T) {
	ruleSetBase10 := rules.Int().Any()
	expected := int(123)

	testhelpers.MustApplyMutation(t, ruleSetBase10, "123", expected)

	err := testhelpers.MustNotApply(t, ruleSetBase10, "7B", errors.CodeType)

	if !strings.Contains(err.Error(), "string to int") {
		t.Errorf("Expected error to contain 'string to int', got: %s", err)
	}

	ruleSetBase16 := rules.Int().WithBase(16).Any()
	testhelpers.MustApplyMutation(t, ruleSetBase16, "7B", expected)

	err = testhelpers.MustNotApply(t, ruleSetBase10, "7x", errors.CodeType)

	if !strings.Contains(err.Error(), "string to int") {
		t.Errorf("Expected error to contain 'string to int', got: %s", err)
	}
}

func TestStringToIntOutOfRange(t *testing.T) {
	ruleSetSigned := rules.Int8().Any()
	testhelpers.MustNotApply(t, ruleSetSigned, "128", errors.CodeRange)

	ruleSetUnsigned := rules.Uint8().Any()
	testhelpers.MustApplyMutation(t, ruleSetUnsigned, "128", uint8(128))
	testhelpers.MustNotApply(t, ruleSetUnsigned, "256", errors.CodeRange)
}

func TestStringToIntInvalid(t *testing.T) {
	ruleSetUnsigned := rules.Int().Any()
	testhelpers.MustNotApply(t, ruleSetUnsigned, "hello", errors.CodeType)
}

func TestUnknownToInt(t *testing.T) {
	from := new(struct{})

	ruleSetSigned := rules.Int8().Any()
	testhelpers.MustNotApply(t, ruleSetSigned, &from, errors.CodeType)

	ruleSetUnsigned := rules.Uint8().Any()
	testhelpers.MustNotApply(t, ruleSetUnsigned, &from, errors.CodeType)
}

func TestCoerceToFloat64(t *testing.T) {
	expected := float64(123.0)
	ruleSet := rules.Float64().Any()

	testhelpers.MustApplyMutation(t, ruleSet, int(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int8(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int16(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int32(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int64(123), expected)

	testhelpers.MustApplyMutation(t, ruleSet, uint(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint8(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint16(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint32(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint64(123), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float32(123.0), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(123.0), expected)
}

func TestOutOfRangeFloat32(t *testing.T) {
	ruleSet := rules.Float32().Any()

	testhelpers.MustNotApply(t, ruleSet, int32(0x7FFFFFFF), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, int64(0x7FFFFFFFFFFFFFFF), errors.CodeRange)

	testhelpers.MustNotApply(t, ruleSet, int32(-0x7FFFFFFF), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, int64(-0x7FFFFFFFFFFFFFFF), errors.CodeRange)

	testhelpers.MustNotApply(t, ruleSet, uint32(0xFFFFFFFF), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, uint64(0xFFFFFFFFFFFFFFFF), errors.CodeRange)

	testhelpers.MustNotApply(t, ruleSet, float64(1.7e+308), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, float64(-1.7e+308), errors.CodeRange)

	testhelpers.MustNotApply(t, ruleSet, fmt.Sprintf("%f", 1.7e+308), errors.CodeRange)
}

func TestOutOfRangeFloat64(t *testing.T) {
	ruleSet := rules.Float64().Any()

	// float64 can represent integers exactly up to 2^53 = 9007199254740992
	// Test values just above this limit
	maxExactInt64 := int64(9007199254740992) // 2^53
	testhelpers.MustNotApply(t, ruleSet, maxExactInt64+1, errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, maxExactInt64+100, errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, -maxExactInt64-1, errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, -maxExactInt64-100, errors.CodeRange)

	// Test very large int64 values
	testhelpers.MustNotApply(t, ruleSet, int64(0x7FFFFFFFFFFFFFFF), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, int64(-0x7FFFFFFFFFFFFFFF), errors.CodeRange)

	// Test large uint64 values
	testhelpers.MustNotApply(t, ruleSet, uint64(0xFFFFFFFFFFFFFFFF), errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, uint64(maxExactInt64+1), errors.CodeRange)
}

func TestFloat32BoundaryValues(t *testing.T) {
	ruleSet := rules.Float32().Any()

	// float32 can represent integers exactly up to 2^24 = 16777216
	// Test values at and just above the boundary
	maxExactInt32 := int32(16777216) // 2^24
	maxExactInt64 := int64(16777216) // 2^24

	// Values just below the limit should work
	testhelpers.MustApplyMutation(t, ruleSet, maxExactInt32-1, float32(maxExactInt32-1))
	testhelpers.MustApplyMutation(t, ruleSet, maxExactInt64-1, float32(maxExactInt64-1))
	testhelpers.MustApplyMutation(t, ruleSet, -maxExactInt32+1, float32(-maxExactInt32+1))
	testhelpers.MustApplyMutation(t, ruleSet, -maxExactInt64+1, float32(-maxExactInt64+1))

	// Values at the limit should work
	testhelpers.MustApplyMutation(t, ruleSet, maxExactInt32, float32(maxExactInt32))
	testhelpers.MustApplyMutation(t, ruleSet, maxExactInt64, float32(maxExactInt64))
	testhelpers.MustApplyMutation(t, ruleSet, -maxExactInt32, float32(-maxExactInt32))
	testhelpers.MustApplyMutation(t, ruleSet, -maxExactInt64, float32(-maxExactInt64))

	// Values just above the limit should fail
	testhelpers.MustNotApply(t, ruleSet, maxExactInt32+1, errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, maxExactInt64+1, errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, -maxExactInt32-1, errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, -maxExactInt64-1, errors.CodeRange)
}

func TestFloat64BoundaryValues(t *testing.T) {
	ruleSet := rules.Float64().Any()

	// float64 can represent integers exactly up to 2^53 = 9007199254740992
	// Test values at and just above the boundary
	maxExactInt64 := int64(9007199254740992) // 2^53

	// Values just below the limit should work
	testhelpers.MustApplyMutation(t, ruleSet, maxExactInt64-1, float64(maxExactInt64-1))
	testhelpers.MustApplyMutation(t, ruleSet, -maxExactInt64+1, float64(-maxExactInt64+1))

	// Values at the limit should work
	testhelpers.MustApplyMutation(t, ruleSet, maxExactInt64, float64(maxExactInt64))
	testhelpers.MustApplyMutation(t, ruleSet, -maxExactInt64, float64(-maxExactInt64))

	// Values just above the limit should fail
	testhelpers.MustNotApply(t, ruleSet, maxExactInt64+1, errors.CodeRange)
	testhelpers.MustNotApply(t, ruleSet, -maxExactInt64-1, errors.CodeRange)
}

func TestCoerceToFloat32(t *testing.T) {
	expected := float32(123.0)
	ruleSet := rules.Float32().Any()

	testhelpers.MustApplyMutation(t, ruleSet, int(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int8(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int16(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int32(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, int64(123), expected)

	testhelpers.MustApplyMutation(t, ruleSet, uint(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint8(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint16(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint32(123), expected)
	testhelpers.MustApplyMutation(t, ruleSet, uint64(123), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float32(123.0), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(123.0), expected)
}

func TestFloat32EqualityCheckFailure(t *testing.T) {
	ruleSet := rules.Float32().Any()

	// These values pass the range check (abs(float32(value)) <= 2^24)
	// but fail the equality check (int64(float32(value)) != value)
	// This happens because float32 rounds values just above 2^24 down to 2^24
	testhelpers.MustNotApply(t, ruleSet, int64(16777217), errors.CodeRange)  // 2^24 + 1
	testhelpers.MustNotApply(t, ruleSet, int64(-16777217), errors.CodeRange) // -(2^24 + 1)
	testhelpers.MustNotApply(t, ruleSet, uint64(16777217), errors.CodeRange) // 2^24 + 1
}

func TestFloat64EqualityCheckFailure(t *testing.T) {
	ruleSet := rules.Float64().Any()

	// These values pass the range check (abs(float64(value)) <= 2^53)
	// but fail the equality check (int64(float64(value)) != value)
	// This happens because float64 rounds values just above 2^53 down to 2^53
	testhelpers.MustNotApply(t, ruleSet, int64(9007199254740993), errors.CodeRange)  // 2^53 + 1
	testhelpers.MustNotApply(t, ruleSet, int64(-9007199254740993), errors.CodeRange) // -(2^53 + 1)
	testhelpers.MustNotApply(t, ruleSet, uint64(9007199254740993), errors.CodeRange) // 2^53 + 1
}

func TestStringToFloat(t *testing.T) {
	ruleSet := rules.Float64().Any()
	expected := float64(123.456)

	testhelpers.MustApplyMutation(t, ruleSet, "123.456", expected)
}

func TestStringToFloatInvalid(t *testing.T) {
	ruleSetUnsigned := rules.Float64().Any()
	testhelpers.MustNotApply(t, ruleSetUnsigned, "hello", errors.CodeType)
}

func TestUnknownToFloat(t *testing.T) {
	from := new(struct{})

	ruleSetSigned := rules.Float64().Any()
	testhelpers.MustNotApply(t, ruleSetSigned, &from, errors.CodeType)

	ruleSetUnsigned := rules.Float64().Any()
	testhelpers.MustNotApply(t, ruleSetUnsigned, &from, errors.CodeType)
}
