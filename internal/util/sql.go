package util

import (
	"context"

	"github.com/pkg/errors"
	"github.com/wroge/superbasic"
)

func DoExpr[R any](
	ctx context.Context,
	f func(ctx context.Context, query string, args ...any) (R, error),
	expr superbasic.Expression,
) (R, error) {
	var zero R
	query, args, err := expr.ToSQL()
	if err != nil {
		return zero, errors.WithStack(err)
	}

	result, err := f(ctx, query, args...)
	return result, errors.WithStack(err)
}
