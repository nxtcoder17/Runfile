package runfile

import (
	"github.com/nxtcoder17/runfile/pkg/errors"
	"github.com/nxtcoder17/runfile/pkg/task"
	"github.com/nxtcoder17/runfile/pkg/types"
	"golang.org/x/sync/errgroup"
)

type RunOption struct {
	ExecuteInParallel bool
	Watch             bool
	Debug             bool
	KVs               map[string]string
}

func (rf *ParsedRunfile) Run(ctx *types.Context, tasks []string, opt RunOption) error {
	for k, v := range opt.KVs {
		if rf.Env == nil {
			rf.Env = make(map[string]string)
		}
		rf.Env[k] = v
	}

	for _, taskName := range tasks {
		if _, ok := rf.Tasks[taskName]; !ok {
			return errors.ErrTaskNotFound(taskName)
		}
	}

	if opt.ExecuteInParallel {
		ctx.Debug("running in parallel mode", "tasks", tasks)
		errg := new(errgroup.Group)

		for _, _tn := range tasks {
			name := _tn
			errg.Go(func() error {
				t := rf.Tasks[name]
				if err := t.Run(task.NewContext(ctx)); err != nil {
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
		t := rf.Tasks[tn]
		if err := t.Run(task.NewContext(ctx)); err != nil {
			return err
		}
	}

	return nil
}
