package rules

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"proto.zip/studio/validate/pkg/errors"
)

func (v *StringRuleSet) coerce(value any, ctx context.Context) (string, errors.ValidationError) {
	str, ok := value.(string)

	if ok {
		return str, nil
	}
	if v.strict {
		return "", errors.NewCoercionError(ctx, "string", reflect.TypeOf(value).String())
	}

	switch x := value.(type) {
	case int:
		return strconv.Itoa(x), nil
	case *int:
		return strconv.Itoa(*x), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case *int64:
		return strconv.FormatInt(*x, 10), nil
	case float64:
		return fmt.Sprintf("%v", x), nil
	case *float64:
		return fmt.Sprintf("%v", *x), nil
	case *string:
		return *x, nil
	}

	return "", errors.NewCoercionError(ctx, "string", reflect.TypeOf(value).String())
}
