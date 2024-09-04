package numbers_test

import (
	"context"
	"math"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/numbers"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestRoundingIntNone(t *testing.T) {
	expected := 123
	ruleSet := numbers.NewInt().Any()

	testhelpers.MustNotApply(t, ruleSet, float32(123.12), errors.CodeType)

	testhelpers.MustNotApply(t, ruleSet, float64(123.12), errors.CodeType)

	// Within tolerance

	testhelpers.MustApplyMutation(t, ruleSet, float32(123+1e-10), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(123-1e-10), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(123-(1e-10)), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(123+(1e-10)), expected)
}

func TestRoundingFloatNone(t *testing.T) {
	expected := float64(123.12)
	ruleSet := numbers.NewFloat64()

	var output float64

	// Apply the rule set with the input value and check for errors
	err := ruleSet.Any().Apply(context.TODO(), float32(expected), &output)
	if err != nil {
		t.Errorf("Expected err to be nil, got: %s", err)
	} else if delta := math.Abs(output - expected); delta > 10e-5 {
		t.Errorf("Expected result to be within tolerance, got: %f (%f - %f)", delta, expected, output)
	}

	// Use the MustRunMutation helper to validate the mutation
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(expected), expected)
}

func TestRoundingIntFloor(t *testing.T) {
	ruleSet := numbers.NewInt().WithRounding(numbers.RoundingDown).Any()

	// Positive numbers
	expected := 123

	testhelpers.MustApplyMutation(t, ruleSet, float32(123.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(123.6), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(123.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(123.6), expected)

	// Negative numbers
	expected = -123

	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.6), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.6), expected)

	// Out of range

	int8RuleSet := numbers.NewInt8().WithRounding(numbers.RoundingDown).Any()

	testhelpers.MustNotApply(t, int8RuleSet, float32(1024.6), errors.CodeRange)

	testhelpers.MustNotApply(t, int8RuleSet, float32(1064.6), errors.CodeRange)
}

func TestRoundingFloatFloor(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingDown, 0).Any()

	// Positive numbers
	expected := 123.0

	testhelpers.MustApplyMutation(t, ruleSet, float32(123.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(123.6), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(123.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(123.6), expected)

	// Negative numbers
	expected = -123.0

	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.6), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.6), expected)
}

func TestRoundingFloatFloorPrecision2(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingDown, 2).Any()

	// Positive numbers
	expected := 123.12

	testhelpers.MustApplyMutation(t, ruleSet, float32(123.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(123.126), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(123.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(123.126), expected)

	// Negative numbers
	expected = -122.13

	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.126), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.126), expected)
}

func TestRoundingIntCeil(t *testing.T) {
	ruleSet := numbers.NewInt().WithRounding(numbers.RoundingUp).Any()

	// Positive numbers
	expected := 124

	testhelpers.MustApplyMutation(t, ruleSet, float32(123.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(123.6), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(123.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(123.6), expected)

	// Negative numbers
	expected = -122

	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.6), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.6), expected)

	// Out of range

	int8RuleSet := numbers.NewInt8().WithRounding(numbers.RoundingUp).Any()

	testhelpers.MustNotApply(t, int8RuleSet, float32(1024.6), errors.CodeRange)

	testhelpers.MustNotApply(t, int8RuleSet, float32(1064.6), errors.CodeRange)
}

func TestRoundingFloatCeil(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingUp, 0).Any()

	// Positive numbers
	expected := 124.0

	testhelpers.MustApplyMutation(t, ruleSet, float32(123.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(123.6), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(123.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(123.6), expected)

	// Negative numbers
	expected = -122.0

	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.6), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.6), expected)
}

func TestRoundingFloatCeilPrecision2(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingUp, 2).Any()

	// Positive numbers
	expected := 123.13

	testhelpers.MustApplyMutation(t, ruleSet, float32(123.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(123.126), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(123.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(123.126), expected)

	// Negative numbers
	expected = -122.12

	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-122.126), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-122.126), expected)
}

func TestRoundingIntHalfUp(t *testing.T) {
	ruleSet := numbers.NewInt().WithRounding(numbers.RoundingHalfUp).Any()

	// Positive numbers
	expected := 124
	expectedUp := 125

	testhelpers.MustApplyMutation(t, ruleSet, float32(124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(124.5), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float32(124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float64(124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(124.5), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float64(124.6), expectedUp)

	// Negative numbers
	expected = -124
	expectedUp = -125

	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.5), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.5), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.6), expectedUp)

	// Out of range

	int8RuleSet := numbers.NewInt8().WithRounding(numbers.RoundingHalfUp).Any()

	testhelpers.MustNotApply(t, int8RuleSet, float32(1024.6), errors.CodeRange)

	testhelpers.MustNotApply(t, int8RuleSet, float32(1064.6), errors.CodeRange)
}

func TestRoundingFloatHalfUp(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingHalfUp, 0).Any()

	// Positive numbers
	expected := 124.0
	expectedUp := 125.0

	testhelpers.MustApplyMutation(t, ruleSet, float32(124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(124.5), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float32(124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float64(124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(124.5), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float64(124.6), expectedUp)

	// Negative numbers
	expected = -124
	expectedUp = -125

	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.5), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.5), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.6), expectedUp)
}

func TestRoundingFloatHalfUpPrecision2(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingHalfUp, 2).Any()

	// Positive numbers
	expected := 124.12
	expectedUp := 124.13

	testhelpers.MustApplyMutation(t, ruleSet, float32(124.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(124.125), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float32(124.126), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float64(124.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(124.125), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float64(124.126), expectedUp)

	// Negative numbers
	expected = -124.12
	expectedUp = -124.13

	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.125), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.126), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.125), expectedUp)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.126), expectedUp)
}

func TestRoundingIntHalfEven(t *testing.T) {
	ruleSet := numbers.NewInt().WithRounding(numbers.RoundingHalfEven).Any()

	// Positive numbers
	expected := 124
	expectedUp := 125

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(124.5), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(124.5), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(123.5), expected)

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(123.5), expected)

	// Negative numbers
	expected = -124
	expectedUp = -125

	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.5), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.5), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float32(-123.5), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-123.5), expected)

	// Out of range

	int8RuleSet := numbers.NewInt8().WithRounding(numbers.RoundingHalfEven).Any()

	testhelpers.MustNotApply(t, int8RuleSet, float32(1024.6), errors.CodeRange)

	testhelpers.MustNotApply(t, int8RuleSet, float32(1064.6), errors.CodeRange)
}

func TestRoundingFloatHalfEven(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingHalfEven, 0).Any()

	// Positive numbers
	expected := 124.0
	expectedUp := 125.0

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(124.5), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(124.5), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(123.5), expected)

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(123.5), expected)

	// Negative numbers

	expected = -124.0
	expectedUp = -125.0

	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.5), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.4), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.5), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.6), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float32(-123.5), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-123.5), expected)
}

func TestRoundingFloatHalfEvenPrecision2(t *testing.T) {
	ruleSet := numbers.NewFloat64().WithRounding(numbers.RoundingHalfEven, 2).Any()

	// Positive numbers
	expected := 124.12
	expectedUp := 124.13

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(124.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(124.125), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(124.126), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(124.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(124.125), expected)
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(124.126), expectedUp)

	// Note that "124.115" will fial here due to float32 precision. So we'll use "124.1155"
	testhelpers.MustApplyMutation(t, ruleSet.Any(), float32(124.1155), expected)

	testhelpers.MustApplyMutation(t, ruleSet.Any(), float64(124.115), expected)

	// Negative numbers

	expected = -124.12
	expectedUp = -124.13

	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.125), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.126), expectedUp)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.124), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.125), expected)
	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.126), expectedUp)

	// See previous comment
	testhelpers.MustApplyMutation(t, ruleSet, float32(-124.1155), expected)

	testhelpers.MustApplyMutation(t, ruleSet, float64(-124.115), expected)
}

// Requirements:
// - Serializes all the rounding levels
func TestRoundingSerialization(t *testing.T) {
	expected := "None"
	if s := numbers.RoundingNone.String(); s != expected {
		t.Errorf("Expected %s, got %s", expected, s)
	}

	expected = "Up"
	if s := numbers.RoundingUp.String(); s != expected {
		t.Errorf("Expected %s, got %s", expected, s)
	}

	expected = "Down"
	if s := numbers.RoundingDown.String(); s != expected {
		t.Errorf("Expected %s, got %s", expected, s)
	}

	expected = "HalfUp"
	if s := numbers.RoundingHalfUp.String(); s != expected {
		t.Errorf("Expected %s, got %s", expected, s)
	}

	expected = "HalfEven"
	if s := numbers.RoundingHalfEven.String(); s != expected {
		t.Errorf("Expected %s, got %s", expected, s)
	}

	expected = "Unknown"
	r := numbers.Rounding(-1)
	if s := r.String(); s != expected {
		t.Errorf("Expected %s, got %s", expected, s)
	}
}
