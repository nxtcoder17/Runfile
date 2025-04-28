package types

import (
	"context"

	"github.com/nxtcoder17/go.pkgs/log"
)

type Context struct {
	context.Context
	log.Logger
	TaskName      string
	TaskNamespace string
}

func NewContext(ctx context.Context, logger log.Logger) Context {
	return Context{Context: ctx, Logger: logger}
}
