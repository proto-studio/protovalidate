package numbers_test

import (
	"math"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestRoundingIntNone(t *testing.T) {
	expected := 123
	ruleSet := numbers.NewInt().Any()

	testhelpers.MustBeInvalid(t, ruleSet, float32(123.12), errors.CodeType)

	testhelpers.MustBeInvalid(t, ruleSet, float64(123.12), errors.CodeType)

	// Within tolerance

	testhelpers.MustBeValid(t, ruleSet, float32(123+1e-10), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(123-1e-10), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(123-(1e-10)), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(123+(1e-10)), expected)
}

func TestRoundingFloatNone(t *testing.T) {
	expected := float64(123.12)
	ruleSet := numbers.NewFloat64()

	out, err := ruleSet.Any().Validate(float32(expected))
	if err != nil {
		t.Errorf("Expected err to be nil, got: %s", err)
	} else if delta := math.Abs(out.(float64) - expected); delta > 10e-5 {
		t.Errorf("Expected result to be within tolerance got: %f (%f - %f)", delta, expected, out)
	}

	testhelpers.MustBeValid(t, ruleSet.Any(), float64(expected), expected)
}

func TestRoundingIntFloor(t *testing.T) {
	ruleSet := numbers.NewInt().WithRounding(numbers.RoundingDown).Any()

	// Positive numbers
	expected := 123

	testhelpers.MustBeValid(t, ruleSet, float32(123.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(123.6), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(123.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(123.6), expected)

	// Negative numbers
	expected = -123

	testhelpers.MustBeValid(t, ruleSet, float32(-122.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-122.6), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(-122.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-122.6), expected)

	// Out of range

	int8RuleSet := numbers.NewInt8().WithRounding(numbers.RoundingDown).Any()

	testhelpers.MustBeInvalid(t, int8RuleSet, float32(1024.6), errors.CodeRange)

	testhelpers.MustBeInvalid(t, int8RuleSet, float32(1064.6), errors.CodeRange)
}

func TestRoundingFloatFloor(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingDown, 0).Any()

	// Positive numbers
	expected := 123.0

	testhelpers.MustBeValid(t, ruleSet, float32(123.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(123.6), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(123.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(123.6), expected)

	// Negative numbers
	expected = -123.0

	testhelpers.MustBeValid(t, ruleSet, float32(-122.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-122.6), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(-122.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-122.6), expected)
}

func TestRoundingFloatFloorPrecision2(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingDown, 2).Any()

	// Positive numbers
	expected := 123.12

	testhelpers.MustBeValid(t, ruleSet, float32(123.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(123.126), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(123.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(123.126), expected)

	// Negative numbers
	expected = -122.13

	testhelpers.MustBeValid(t, ruleSet, float32(-122.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-122.126), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(-122.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-122.126), expected)
}

func TestRoundingIntCeil(t *testing.T) {
	ruleSet := numbers.NewInt().WithRounding(numbers.RoundingUp).Any()

	// Positive numbers
	expected := 124

	testhelpers.MustBeValid(t, ruleSet, float32(123.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(123.6), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(123.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(123.6), expected)

	// Negative numbers
	expected = -122

	testhelpers.MustBeValid(t, ruleSet, float32(-122.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-122.6), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(-122.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-122.6), expected)

	// Out of range

	int8RuleSet := numbers.NewInt8().WithRounding(numbers.RoundingUp).Any()

	testhelpers.MustBeInvalid(t, int8RuleSet, float32(1024.6), errors.CodeRange)

	testhelpers.MustBeInvalid(t, int8RuleSet, float32(1064.6), errors.CodeRange)
}

func TestRoundingFloatCeil(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingUp, 0).Any()

	// Positive numbers
	expected := 124.0

	testhelpers.MustBeValid(t, ruleSet, float32(123.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(123.6), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(123.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(123.6), expected)

	// Negative numbers
	expected = -122.0

	testhelpers.MustBeValid(t, ruleSet, float32(-122.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-122.6), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(-122.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-122.6), expected)
}

func TestRoundingFloatCeilPrecision2(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingUp, 2).Any()

	// Positive numbers
	expected := 123.13

	testhelpers.MustBeValid(t, ruleSet, float32(123.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(123.126), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(123.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(123.126), expected)

	// Negative numbers
	expected = -122.12

	testhelpers.MustBeValid(t, ruleSet, float32(-122.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-122.126), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(-122.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-122.126), expected)
}

func TestRoundingIntHalfUp(t *testing.T) {
	ruleSet := numbers.NewInt().WithRounding(numbers.RoundingHalfUp).Any()

	// Positive numbers
	expected := 124
	expectedUp := 125

	testhelpers.MustBeValid(t, ruleSet, float32(124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(124.5), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float32(124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float64(124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(124.5), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float64(124.6), expectedUp)

	// Negative numbers
	expected = -124
	expectedUp = -125

	testhelpers.MustBeValid(t, ruleSet, float32(-124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.5), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float64(-124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.5), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.6), expectedUp)

	// Out of range

	int8RuleSet := numbers.NewInt8().WithRounding(numbers.RoundingHalfUp).Any()

	testhelpers.MustBeInvalid(t, int8RuleSet, float32(1024.6), errors.CodeRange)

	testhelpers.MustBeInvalid(t, int8RuleSet, float32(1064.6), errors.CodeRange)
}

func TestRoundingFloatHalfUp(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingHalfUp, 0).Any()

	// Positive numbers
	expected := 124.0
	expectedUp := 125.0

	testhelpers.MustBeValid(t, ruleSet, float32(124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(124.5), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float32(124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float64(124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(124.5), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float64(124.6), expectedUp)

	// Negative numbers
	expected = -124
	expectedUp = -125

	testhelpers.MustBeValid(t, ruleSet, float32(-124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.5), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float64(-124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.5), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.6), expectedUp)
}

func TestRoundingFloatHalfUpPrecision2(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingHalfUp, 2).Any()

	// Positive numbers
	expected := 124.12
	expectedUp := 124.13

	testhelpers.MustBeValid(t, ruleSet, float32(124.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(124.125), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float32(124.126), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float64(124.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(124.125), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float64(124.126), expectedUp)

	// Negative numbers
	expected = -124.12
	expectedUp = -124.13

	testhelpers.MustBeValid(t, ruleSet, float32(-124.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.125), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.126), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float64(-124.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.125), expectedUp)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.126), expectedUp)
}

func TestRoundingIntHalfEven(t *testing.T) {
	ruleSet := numbers.NewInt().WithRounding(numbers.RoundingHalfEven).Any()

	// Positive numbers
	expected := 124
	expectedUp := 125

	testhelpers.MustBeValid(t, ruleSet.Any(), float32(124.4), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float32(124.5), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float32(124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet.Any(), float64(124.4), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float64(124.5), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float64(124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet.Any(), float32(123.5), expected)

	testhelpers.MustBeValid(t, ruleSet.Any(), float64(123.5), expected)

	// Negative numbers
	expected = -124
	expectedUp = -125

	testhelpers.MustBeValid(t, ruleSet, float32(-124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.5), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float64(-124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.5), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float32(-123.5), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(-123.5), expected)

	// Out of range

	int8RuleSet := numbers.NewInt8().WithRounding(numbers.RoundingHalfEven).Any()

	testhelpers.MustBeInvalid(t, int8RuleSet, float32(1024.6), errors.CodeRange)

	testhelpers.MustBeInvalid(t, int8RuleSet, float32(1064.6), errors.CodeRange)
}

func TestRoundingFloatHalfEven(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingHalfEven, 0).Any()

	// Positive numbers
	expected := 124.0
	expectedUp := 125.0

	testhelpers.MustBeValid(t, ruleSet.Any(), float32(124.4), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float32(124.5), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float32(124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet.Any(), float64(124.4), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float64(124.5), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float64(124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet.Any(), float32(123.5), expected)

	testhelpers.MustBeValid(t, ruleSet.Any(), float64(123.5), expected)

	// Negative numbers

	expected = -124.0
	expectedUp = -125.0

	testhelpers.MustBeValid(t, ruleSet, float32(-124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.5), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float64(-124.4), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.5), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.6), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float32(-123.5), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(-123.5), expected)
}

func TestRoundingFloatHalfEvenPrecision2(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingHalfEven, 2).Any()

	// Positive numbers
	expected := 124.12
	expectedUp := 124.13

	testhelpers.MustBeValid(t, ruleSet.Any(), float32(124.124), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float32(124.125), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float32(124.126), expectedUp)

	testhelpers.MustBeValid(t, ruleSet.Any(), float64(124.124), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float64(124.125), expected)
	testhelpers.MustBeValid(t, ruleSet.Any(), float64(124.126), expectedUp)

	// Note that "124.115" will fial here due to float32 precision. So we'll use "124.1155"
	testhelpers.MustBeValid(t, ruleSet.Any(), float32(124.1155), expected)

	testhelpers.MustBeValid(t, ruleSet.Any(), float64(124.115), expected)

	// Negative numbers

	expected = -124.12
	expectedUp = -124.13

	testhelpers.MustBeValid(t, ruleSet, float32(-124.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.125), expected)
	testhelpers.MustBeValid(t, ruleSet, float32(-124.126), expectedUp)

	testhelpers.MustBeValid(t, ruleSet, float64(-124.124), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.125), expected)
	testhelpers.MustBeValid(t, ruleSet, float64(-124.126), expectedUp)

	// See previous comment
	testhelpers.MustBeValid(t, ruleSet, float32(-124.1155), expected)

	testhelpers.MustBeValid(t, ruleSet, float64(-124.115), expected)
}
