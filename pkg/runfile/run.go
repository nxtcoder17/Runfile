package runfile

import (
	"github.com/nxtcoder17/runfile/pkg/errors"
	"github.com/nxtcoder17/runfile/pkg/types"
	"golang.org/x/sync/errgroup"
)

type RunOption struct {
	ExecuteInParallel bool
	Watch             bool
	Debug             bool
	KVs               map[string]string
}

func (r *ParsedRunfile) Run(ctx *types.Context, tasks []string, opt RunOption) error {
	for k, v := range opt.KVs {
		if r.Env == nil {
			r.Env = make(map[string]string)
		}
		r.Env[k] = v
	}

	for _, taskName := range tasks {
		if _, ok := r.Tasks[taskName]; !ok {
			return errors.ErrTaskNotFound(taskName)
		}
	}

	if opt.ExecuteInParallel {
		ctx.Debug("running in parallel mode", "tasks", tasks)
		errg := new(errgroup.Group)

		for _, _tn := range tasks {
			name := _tn
			errg.Go(func() error {
				if err := r.RunTask(NewContext(ctx), name); err != nil {
					return err
				}
				return nil
			})
		}

		// Wait for all tasks to finish
		if err := errg.Wait(); err != nil {
			return err
		}

		return nil
	}

	for _, tn := range tasks {
		if err := r.RunTask(NewContext(ctx), tn); err != nil {
			return err
		}
	}

	return nil
}
