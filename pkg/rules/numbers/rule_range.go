package numbers

import "errors"

// WithRange returns a new child RuleSet that is constrained to the provided range of values.
// It is identical in every way to calling WithMin and WithMax separately.
func (v *IntRuleSet[T]) WithRange(min, max T) *IntRuleSet[T] {
	if min >= max {
		panic(errors.New("Minimum must be less than or equal to maximum."))
	}

	return v.WithMin(min).WithMax(max)
}

// WithRange returns a new child RuleSet that is constrained to the provided range of values.
// It is identical in every way to calling WithMin and WithMax separately.
func (v *FloatRuleSet[T]) WithRange(min, max T) *FloatRuleSet[T] {
	if min >= max {
		panic(errors.New("Minimum must be less than or equal to maximum."))
	}

	return v.WithMin(min).WithMax(max)
}
