package runfile

import (
	"context"
	"log/slog"

	"github.com/nxtcoder17/runfile/pkg/runfile/resolver"
)

type Context struct {
	context.Context
	RunfilePath string
}

func RunTask(ctx context.Context, runfile string, task string) error {
	slog.Debug("[run-task] START", "task", task)
	defer slog.Debug("[run-task] FINISH", "task", task)
	r, err := resolver.Load(ctx, runfile)
	if err != nil {
		return err
	}

	return r.RunTask(ctx, task)
}
