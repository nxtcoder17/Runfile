package types

import (
	"context"

	"github.com/nxtcoder17/fastlog"
)

type Context struct {
	context.Context
	*fastlog.Logger
	TaskName      string
	TaskNamespace string
}

func NewContext(ctx context.Context, logger *fastlog.Logger) Context {
	return Context{Context: ctx, Logger: logger}
}
