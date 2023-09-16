package numbers

// Rounding type is used to specify how a floating point number should be convered to
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

// WithRounding returns a new child RuleSet with the rounding rule set to the supplied value.
//
// Notes on floating point numbers:
// The RuleSet will attempt to convert floating point numbers to integers even if rounding is not enabled.
// If the number is not within tollerence (1e-9) of a whole number, an error will be returned.
func (v *IntRuleSet[T]) WithRounding(rounding Rounding) *IntRuleSet[T] {
	return &IntRuleSet[T]{
		strict:   v.strict,
		parent:   v,
		base:     v.base,
		required: v.required,
		rounding: rounding,
	}
}

// WithRounding returns a new child RuleSet with the rounding rule set to the supplied value.
//
// Standard warnings for floating point numbers apply:
// - Some numbers cannot be represented precicely with floating points.
// - Sometimes the rounded result may have additional precision when the rounded number cannot be exactly represented.
// - For best results, consider using int for your math and data storage/transfer.
func (v *FloatRuleSet[T]) WithRounding(rounding Rounding, precision int) *FloatRuleSet[T] {
	return &FloatRuleSet[T]{
		strict:    v.strict,
		parent:    v,
		base:      v.base,
		required:  v.required,
		rounding:  rounding,
		precision: precision,
	}
}
