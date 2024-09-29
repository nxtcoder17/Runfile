package runfile

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"github.com/nxtcoder17/runfile/pkg/runfile/errors"
	"golang.org/x/sync/errgroup"
)

type cmdArgs struct {
	shell      []string
	env        []string // [key=value, key=value, ...]
	workingDir string

	cmd string

	stdout io.Writer
	stderr io.Writer
}

func createCommand(ctx context.Context, args cmdArgs) *exec.Cmd {
	if args.shell == nil {
		args.shell = []string{"sh", "-c"}
	}

	if args.stdout == nil {
		args.stdout = os.Stdout
	}

	if args.stderr == nil {
		args.stderr = os.Stderr
	}

	shell := args.shell[0]

	cargs := append(args.shell[1:], args.cmd)
	c := exec.CommandContext(ctx, shell, cargs...)
	c.Dir = args.workingDir
	c.Env = args.env
	c.Stdout = args.stdout
	c.Stderr = args.stderr
	return c
}

func (rf *Runfile) runTask(ctx context.Context, taskName string) error {
	pt, err := ParseTask(ctx, rf, taskName)
	if err != nil {
		return err
	}

	// slog.Default().Info("parsing", "task", pt)

	for _, command := range pt.Commands {
		if command.Run != "" {
			if err := rf.runTask(ctx, command.Run); err != nil {
				return err
			}
			continue
		}

		cmd := createCommand(ctx, cmdArgs{
			shell:      pt.Shell,
			env:        pt.Environ,
			cmd:        command.Command,
			workingDir: pt.WorkingDir,
		})
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

type RunArgs struct {
	Tasks             []string
	ExecuteInParallel bool
	Watch             bool
	Debug             bool
}

func (rf *Runfile) Run(ctx context.Context, args RunArgs) error {
	for _, v := range args.Tasks {
		if _, ok := rf.Tasks[v]; !ok {
			return errors.TaskNotFound{Context: errors.Context{
				Task:    v,
				Runfile: rf.attrs.RunfilePath,
			}}
		}
	}

	if args.ExecuteInParallel {
		slog.Default().Debug("running in parallel mode", "tasks", args.Tasks)
		g := new(errgroup.Group)

		for _, tn := range args.Tasks {
			g.Go(func() error {
				return rf.runTask(ctx, tn)
			})
		}

		// Wait for all tasks to finish
		if err := g.Wait(); err != nil {
			return err
		}

		return nil
	}

	for _, tn := range args.Tasks {
		if err := rf.runTask(ctx, tn); err != nil {
			return err
		}
	}

	return nil
}
