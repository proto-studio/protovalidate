package rules

import (
	"context"

	"proto.zip/studio/validate/pkg/errors"
)

type Rule[T any] interface {
	Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection)
}

type RuleFunc[T any] func(tx context.Context, value T) (T, errors.ValidationErrorCollection)

func (rule RuleFunc[T]) Evaluate(ctx context.Context, value T) (T, errors.ValidationErrorCollection) {
	return rule(ctx, value)
}
