package types

import (
	"context"

	"github.com/nxtcoder17/fastlog"
)

type (
	EnvExpr map[string]any
	Env     map[string]string
)

type Context struct {
	context.Context
	*fastlog.Logger
}
