package rules

import "fmt"

// Rounding type is used to specify how a floating point number should be converted to
// a number with lower precision.
//
// These values are not guaranteed to be unchanged between versions. Don't use for serialization or cross-process communication.
type Rounding int

const (
	RoundingNone     Rounding = iota // Default. No rounding will be performed and return an error if the number is not already rounded.
	RoundingUp                       // Ceil. Always round up.
	RoundingDown                     // Floor. Always round down.
	RoundingHalfEven                 // "Bankers rounding." Round to the nearest even number.
	RoundingHalfUp                   // Always round half values up.
)

// String returns the string value for the rounding. Useful for debugging.
func (r Rounding) String() string {
	switch r {
	case RoundingNone:
		return "None"
	case RoundingUp:
		return "Up"
	case RoundingDown:
		return "Down"
	case RoundingHalfEven:
		return "HalfEven"
	case RoundingHalfUp:
		return "HalfUp"
	}
	return "Unknown"
}

// WithRounding returns a new child RuleSet that applies the specified rounding method when converting floating point numbers to integers.
//
// Notes on floating point numbers:
// The RuleSet will attempt to convert floating point numbers to integers even if rounding is not enabled.
// If the number is not within tolerance (1e-9) of a whole number, an error will be returned.
func (v *IntRuleSet[T]) WithRounding(rounding Rounding) *IntRuleSet[T] {
	return &IntRuleSet[T]{
		strict:   v.strict,
		parent:   v,
		base:     v.base,
		required: v.required,
		rounding: rounding,
		label:    fmt.Sprintf("WithRounding(%s)", rounding.String()),
	}
}

// WithRounding returns a new child RuleSet that applies the specified rounding method and precision to floating point numbers.
//
// Standard warnings for floating point numbers apply:
// - Some numbers cannot be represented precisely with floating points.
// - Sometimes the rounded result may have additional precision when the rounded number cannot be exactly represented.
// - For best results, consider using int for your math and data storage/transfer.
func (v *FloatRuleSet[T]) WithRounding(rounding Rounding, precision int) *FloatRuleSet[T] {
	return &FloatRuleSet[T]{
		strict:    v.strict,
		parent:    v,
		required:  v.required,
		rounding:  rounding,
		precision: precision,
		label:     fmt.Sprintf("WithRounding(%s, %d)", rounding.String(), precision),
	}
}
