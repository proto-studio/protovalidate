package rules

import (
	"context"
	"math"
	"reflect"
	"strconv"
	"strings"

	"proto.zip/studio/validate/pkg/errors"
)

// Tolerance for floating point to int conversions
const tolerance = 1e-9

// tryCoerceIntToInt attempts to coerce an int from one type to another and checks that no data was lost in the process.
func tryCoerceIntToInt[From, To integer](ruleSet *IntRuleSet[To], value From, ctx context.Context) (To, errors.ValidationError) {
	intval := To(value)
	if From(intval) != value {
		return 0, errors.NewRangeError(ctx, ruleSet.typeName())
	}
	return intval, nil
}

// tryCoerceFloatToInt attempts to coerce a float into an int and checks that no data was lost in the process.
// Rounding rules are applied when appropriate.
func tryCoerceFloatToInt[From floating, To integer](ruleSet *IntRuleSet[To], value From, ctx context.Context) (To, errors.ValidationError) {
	var int64val int64
	float64val := float64(value)

	rounding := ruleSet.rounding

	switch rounding {
	case RoundingDown:
		int64val = int64(math.Floor(float64val))
	case RoundingUp:
		int64val = int64(math.Ceil(float64val))
	case RoundingHalfUp:
		int64val = int64(math.Round(float64val))
	case RoundingHalfEven:
		int64val = int64(math.RoundToEven(float64val))
	default:
		int64val = int64(math.Round(float64val))

		if math.Abs(float64(int64val)-float64val) > tolerance {
			return 0, errors.NewCoercionError(ctx, ruleSet.typeName(), reflect.ValueOf(value).Kind().String())
		}
	}

	intval := To(int64val)

	if int64(intval) != int64val {
		return 0, errors.NewRangeError(ctx, ruleSet.typeName())
	}

	return intval, nil
}

// parseInt attempts to parse an int from a string while using reflection to get the right type.
func parseInt[To integer](value string, base int) (To, error) {
	t := reflect.TypeOf(*new(To))

	switch t.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		intval, err := strconv.ParseUint(value, base, t.Bits())
		return To(intval), err
	default:
		intval, err := strconv.ParseInt(value, base, t.Bits())
		return To(intval), err
	}
}

// formatInt formats an integer to a string using the specified base.
func formatInt[T integer](value T, base int) string {
	t := reflect.TypeOf(value)

	switch t.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(uint64(value), base)
	default:
		return strconv.FormatInt(int64(value), base)
	}
}

// formatFloat formats a float to a string.
// If outputPrecision is set (>= 0) via WithFixedOutput, uses fixed-point format with zero-padding.
// If rounding was applied via WithRounding, caps output precision at the rounding precision (no zero-padding).
// Otherwise, uses a smart format that avoids floating point artifacts.
func formatFloat[T floating](ruleSet *FloatRuleSet[T], value T) string {
	float64val := float64(value)

	// Get the bit size for the float type
	bits := reflect.TypeOf(*new(T)).Bits()

	// Find output precision and rounding precision by traversing the ruleSet chain
	outputPrecision := -1   // -1 means not set
	roundingPrecision := -1 // -1 means no rounding applied

	for rs := ruleSet; rs != nil; rs = rs.parent {
		if outputPrecision < 0 && rs.outputPrecision >= 0 {
			outputPrecision = rs.outputPrecision
		}
		if roundingPrecision < 0 && rs.rounding != RoundingNone {
			roundingPrecision = rs.precision
		}
	}

	// If output precision is explicitly set via WithFixedOutput, use fixed-point format
	if outputPrecision >= 0 {
		return strconv.FormatFloat(float64val, 'f', outputPrecision, bits)
	}

	// If rounding was applied, cap output at the rounding precision (no zero-padding)
	if roundingPrecision >= 0 {
		formatted := strconv.FormatFloat(float64val, 'f', roundingPrecision, bits)
		return trimTrailingZeros(formatted)
	}

	// Default case: use 'g' format which automatically chooses the best representation
	// and avoids floating point artifacts
	sigDigits := 15 // float64 has ~15-17 decimal digits of precision, use 15 to be safe
	if bits == 32 {
		sigDigits = 7 // float32 has ~7 decimal digits of precision
	}

	return strconv.FormatFloat(float64val, 'g', sigDigits, bits)
}

// trimTrailingZeros removes trailing zeros after the decimal point.
func trimTrailingZeros(s string) string {
	dotIdx := strings.Index(s, ".")
	if dotIdx == -1 {
		return s
	}

	// Find last non-zero character
	lastNonZero := len(s) - 1
	for lastNonZero > dotIdx && s[lastNonZero] == '0' {
		lastNonZero--
	}

	// If we're at the decimal point, remove it too
	if lastNonZero == dotIdx {
		return s[:dotIdx]
	}

	return s[:lastNonZero+1]
}

// tryCoerceIntDefault attempts to convert to an int from a non-float and non-int type
func tryCoerceIntDefault[To integer](ruleSet *IntRuleSet[To], value any, ctx context.Context) (To, errors.ValidationError) {
	if ruleSet.strict {
		return 0, errors.NewCoercionError(ctx, ruleSet.typeName(), reflect.ValueOf(value).Kind().String())
	}

	if str, ok := value.(string); ok {
		var err error

		intval, err := parseInt[To](str, ruleSet.base)
		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrRange {
				return 0, errors.NewRangeError(ctx, ruleSet.typeName())
			}

			return 0, errors.NewCoercionError(ctx, ruleSet.typeName(), "string")
		}
		return To(intval), nil
	}

	return 0, errors.NewCoercionError(ctx, ruleSet.typeName(), reflect.ValueOf(value).Kind().String())
}

// coerceInt arrempts to convert the value to the appropriate number type and returns a validation error collection if it can't.
func (ruleSet *IntRuleSet[T]) coerceInt(value any, ctx context.Context) (T, errors.ValidationError) {
	switch x := value.(type) {
	case T:
		return x, nil
	case int:
		return tryCoerceIntToInt(ruleSet, x, ctx)
	case int8:
		return tryCoerceIntToInt(ruleSet, x, ctx)
	case int16:
		return tryCoerceIntToInt(ruleSet, x, ctx)
	case int32:
		return tryCoerceIntToInt(ruleSet, x, ctx)
	case int64:
		return tryCoerceIntToInt(ruleSet, x, ctx)
	case uint:
		return tryCoerceIntToInt(ruleSet, x, ctx)
	case uint8:
		return tryCoerceIntToInt(ruleSet, x, ctx)
	case uint16:
		return tryCoerceIntToInt(ruleSet, x, ctx)
	case uint32:
		return tryCoerceIntToInt(ruleSet, x, ctx)
	case uint64:
		return tryCoerceIntToInt(ruleSet, x, ctx)
	case float32:
		return tryCoerceFloatToInt(ruleSet, x, ctx)
	case float64:
		return tryCoerceFloatToInt(ruleSet, x, ctx)
	default:
		return tryCoerceIntDefault(ruleSet, value, ctx)
	}
}

// tryCoerceFloatToFloat attempts to coerce a float from one type to another and checks that no data was lost in the process.
func tryCoerceFloatToFloat[From, To floating](value From, ctx context.Context) (To, errors.ValidationError) {
	floatval := To(value)
	if From(floatval) != value {
		target := reflect.ValueOf(*new(To)).Kind().String()
		return 0, errors.NewRangeError(ctx, target)
	}
	return floatval, nil
}

// tryCoerceIntToFloat attempts to coerce an integer to a float type and checks that no data was lost.
// float32 can represent integers exactly up to 2^24 = 16777216
// float64 can represent integers exactly up to 2^53 = 9007199254740992
func tryCoerceIntToFloat[From integer, To floating](ruleSet *FloatRuleSet[To], value From, ctx context.Context) (To, errors.ValidationError) {
	floatval := To(value)

	// Determine the maximum exact integer value based on the target float type
	var maxExactFloat float64
	var zero To
	switch any(zero).(type) {
	case float32:
		maxExactFloat = 1 << 24 // 2^24
	case float64:
		maxExactFloat = 1 << 53 // 2^53
	}

	// Check if the absolute value of the float exceeds the exact representation limit
	absFloatVal := math.Abs(float64(floatval))
	if absFloatVal > maxExactFloat {
		return 0, errors.NewRangeError(ctx, ruleSet.typeName())
	}

	// Also verify the round-trip conversion works correctly
	// This catches edge cases where the float is within range but still can't represent the exact integer
	if From(floatval) != value {
		return 0, errors.NewRangeError(ctx, ruleSet.typeName())
	}

	return floatval, nil
}

// tryCoerceFloatDefault attempts to convert to a floar from a non-float and non-int type
func tryCoerceFloatDefault[To floating](ruleSet *FloatRuleSet[To], value any, ctx context.Context) (To, errors.ValidationError) {
	if ruleSet.strict {
		return 0, errors.NewCoercionError(ctx, ruleSet.typeName(), reflect.ValueOf(value).Kind().String())
	}

	if str, ok := value.(string); ok {
		var err error

		bits := reflect.TypeOf(*new(To)).Bits()
		floatval, err := strconv.ParseFloat(str, bits)

		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrRange {
				return 0, errors.NewRangeError(ctx, ruleSet.typeName())
			}

			return 0, errors.NewCoercionError(ctx, ruleSet.typeName(), "string")
		}

		return To(floatval), nil
	}

	return 0, errors.NewCoercionError(ctx, ruleSet.typeName(), reflect.ValueOf(value).Kind().String())
}

// coerceInt arrempts to convert the value to the appropriate number type and returns a validation error collection if it can't.
func (v *FloatRuleSet[T]) coerceFloat(value any, ctx context.Context) (T, errors.ValidationError) {
	switch x := value.(type) {
	case T:
		return x, nil
	case int:
		return tryCoerceIntToFloat(v, x, ctx)
	case int8:
		return tryCoerceIntToFloat(v, x, ctx)
	case int16:
		return tryCoerceIntToFloat(v, x, ctx)
	case int32:
		return tryCoerceIntToFloat(v, x, ctx)
	case int64:
		return tryCoerceIntToFloat(v, x, ctx)
	case uint:
		return tryCoerceIntToFloat(v, x, ctx)
	case uint8:
		return tryCoerceIntToFloat(v, x, ctx)
	case uint16:
		return tryCoerceIntToFloat(v, x, ctx)
	case uint32:
		return tryCoerceIntToFloat(v, x, ctx)
	case uint64:
		return tryCoerceIntToFloat(v, x, ctx)
	case float32:
		return tryCoerceFloatToFloat[float32, T](x, ctx)
	case float64:
		return tryCoerceFloatToFloat[float64, T](x, ctx)
	default:
		return tryCoerceFloatDefault(v, value, ctx)
	}
}
