package error_helpers

import (
	"context"
	"errors"
	"strings"
)

func IsContextCanceled(ctx context.Context) bool {
	return IsContextCancelledError(ctx.Err())
}

func IsContextCancelledError(err error) bool {
	return err != nil && (errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "context canceled"))
}
