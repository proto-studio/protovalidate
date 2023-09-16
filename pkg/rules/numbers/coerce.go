package numbers

import (
	"context"
	"math"
	"reflect"
	"strconv"

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
		return tryCoerceIntToInt[int, T](ruleSet, x, ctx)
	case int8:
		return tryCoerceIntToInt[int8, T](ruleSet, x, ctx)
	case int16:
		return tryCoerceIntToInt[int16, T](ruleSet, x, ctx)
	case int32:
		return tryCoerceIntToInt[int32, T](ruleSet, x, ctx)
	case int64:
		return tryCoerceIntToInt[int64, T](ruleSet, x, ctx)
	case uint:
		return tryCoerceIntToInt[uint, T](ruleSet, x, ctx)
	case uint8:
		return tryCoerceIntToInt[uint8, T](ruleSet, x, ctx)
	case uint16:
		return tryCoerceIntToInt[uint16, T](ruleSet, x, ctx)
	case uint32:
		return tryCoerceIntToInt[uint32, T](ruleSet, x, ctx)
	case uint64:
		return tryCoerceIntToInt[uint64, T](ruleSet, x, ctx)
	case float32:
		return tryCoerceFloatToInt[float32, T](ruleSet, x, ctx)
	case float64:
		return tryCoerceFloatToInt[float64, T](ruleSet, x, ctx)
	default:
		return tryCoerceIntDefault[T](ruleSet, value, ctx)
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

// tryCoerceFloatToFloat attempts to coerce a float from one type to another and checks that no data was lost in the process.
func tryCoerceIntToFloat[From integer, To floating](value From, ctx context.Context) (To, errors.ValidationError) {
	floatval := To(value)
	if From(floatval) != value {
		target := reflect.ValueOf(*new(To)).Kind().String()
		return 0, errors.NewRangeError(ctx, target)
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
		return tryCoerceIntToFloat[int, T](x, ctx)
	case int8:
		return tryCoerceIntToFloat[int8, T](x, ctx)
	case int16:
		return tryCoerceIntToFloat[int16, T](x, ctx)
	case int32:
		return tryCoerceIntToFloat[int32, T](x, ctx)
	case int64:
		return tryCoerceIntToFloat[int64, T](x, ctx)
	case uint:
		return tryCoerceIntToFloat[uint, T](x, ctx)
	case uint8:
		return tryCoerceIntToFloat[uint8, T](x, ctx)
	case uint16:
		return tryCoerceIntToFloat[uint16, T](x, ctx)
	case uint32:
		return tryCoerceIntToFloat[uint32, T](x, ctx)
	case uint64:
		return tryCoerceIntToFloat[uint64, T](x, ctx)
	case float32:
		return tryCoerceFloatToFloat[float32, T](x, ctx)
	case float64:
		return tryCoerceFloatToFloat[float64, T](x, ctx)
	default:
		return tryCoerceFloatDefault[T](v, value, ctx)
	}
}
