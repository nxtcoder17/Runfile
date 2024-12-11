package runner

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"github.com/nxtcoder17/runfile/errors"
	"github.com/nxtcoder17/runfile/types"
	"golang.org/x/sync/errgroup"
)

type CmdArgs struct {
	Shell      []string
	Env        []string // [key=value, key=value, ...]
	WorkingDir string

	Cmd string

	interactive bool
	Stdout      io.Writer
	Stderr      io.Writer
}

func CreateCommand(ctx context.Context, args CmdArgs) *exec.Cmd {
	if args.Shell == nil {
		args.Shell = []string{"sh", "-c"}
	}

	if args.Stdout == nil {
		args.Stdout = os.Stdout
	}

	if args.Stderr == nil {
		args.Stderr = os.Stderr
	}

	shell := args.Shell[0]

	cargs := append(args.Shell[1:], args.Cmd)
	c := exec.CommandContext(ctx, shell, cargs...)
	c.Dir = args.WorkingDir
	c.Env = args.Env
	c.Stdout = args.Stdout
	c.Stderr = args.Stderr

	if args.interactive {
		c.Stdin = os.Stdin
	}

	return c
}

type RunArgs struct {
	Tasks             []string
	ExecuteInParallel bool
	Watch             bool
	Debug             bool
	KVs               map[string]string
}

type Context struct {
	context.Context
	*slog.Logger
}

func NewContext(ctx context.Context, logger *slog.Logger) Context {
	return Context{Context: ctx, Logger: logger}
}

func Run(ctx Context, prf *types.ParsedRunfile, args RunArgs) error {
	// INFO: adding parsed KVs as environments to the specified tasks
	for k, v := range args.KVs {
		if prf.Env == nil {
			prf.Env = make(map[string]string)
		}
		prf.Env[k] = v
	}

	attr := func(taskName string) []any {
		return []any{
			"runfile", prf.Metadata.RunfilePath,
			"task", taskName,
		}
	}

	for _, taskName := range args.Tasks {
		if _, ok := prf.Tasks[taskName]; !ok {
			return errors.ErrTaskNotFound.KV(attr(taskName)...)
		}
	}

	if args.ExecuteInParallel {
		ctx.Debug("running in parallel mode", "tasks", args.Tasks)
		g := new(errgroup.Group)

		for _, _tn := range args.Tasks {
			tn := _tn
			g.Go(func() error {
				if err := runTask(ctx, prf, runTaskArgs{taskName: tn}); err != nil {
					return errors.ErrTaskFailed.Wrap(err).KV(attr(tn)...)
				}
				return nil
			})
		}

		// Wait for all tasks to finish
		if err := g.Wait(); err != nil {
			return err
		}

		return nil
	}

	for _, tn := range args.Tasks {
		if err := runTask(ctx, prf, runTaskArgs{taskName: tn}); err != nil {
			return errors.ErrTaskFailed.Wrap(err).KV(attr(tn)...)
		}
	}

	return nil
}
