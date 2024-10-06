package runfile

import (
	"fmt"
	"io"
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

func createCommand(ctx Context, args cmdArgs) *exec.Cmd {
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
	c.Env = append(os.Environ(), args.env...)
	c.Stdout = args.stdout
	c.Stderr = args.stderr

	return c
}

type runTaskArgs struct {
	taskName     string
	envOverrides map[string]string
}

func (rf *Runfile) runTask(ctx Context, args runTaskArgs) errors.Message {
	attr := []any{"task", args.taskName, "runfile", rf.attrs.RunfilePath}

	logger := ctx.With("runfile", rf.attrs.RunfilePath, "task", args.taskName, "env:overrides", args.envOverrides)
	logger.Debug("running task")
	task, ok := rf.Tasks[args.taskName]
	if !ok {
		return errors.TaskNotFound.WithMetadata(attr)
	}

	task.Name = args.taskName
	if task.Env == nil {
		task.Env = make(EnvVar)
	}
	for k, v := range args.envOverrides {
		task.Env[k] = v
	}
	pt, err := ParseTask(ctx, rf, task)
	if err != nil {
		return err
	}

	// envVars := append(pt.Environ, args.envOverrides...)
	ctx.Debug("debugging env", "pt.environ", pt.Env, "overrides", args.envOverrides, "task", args.taskName)
	for _, command := range pt.Commands {
		if command.Run != "" {
			if err := rf.runTask(ctx, runTaskArgs{
				taskName:     command.Run,
				envOverrides: pt.Env,
				// envOverrides: append(pt.Environ, args.envOverrides...),
			}); err != nil {
				return err
			}
			continue
		}

		cmd := createCommand(ctx, cmdArgs{
			shell:      pt.Shell,
			env:        ToEnviron(pt.Env),
			cmd:        command.Command,
			workingDir: pt.WorkingDir,
		})
		if err := cmd.Run(); err != nil {
			return errors.CommandFailed.WithErr(err).WithMetadata(attr)
		}
	}

	return nil
}

type RunArgs struct {
	Tasks             []string
	ExecuteInParallel bool
	Watch             bool
	Debug             bool
	KVs               map[string]string
}

func (rf *Runfile) Run(ctx Context, args RunArgs) errors.Message {
	includes, err := rf.ParseIncludes()
	if err != nil {
		return err
	}

	for _, taskName := range args.Tasks {
		for k, v := range includes {
			for tn := range v.Runfile.Tasks {
				if taskName == fmt.Sprintf("%s:%s", k, tn) {
					return v.Runfile.runTask(ctx, runTaskArgs{taskName: tn})
				}
			}
		}

		task, ok := rf.Tasks[taskName]
		if !ok {
			return errors.TaskNotFound.WithMetadata("task", taskName, "runfile", rf.attrs.RunfilePath)
		}

		// INFO: adding parsed KVs as environments to the specified tasks
		for k, v := range args.KVs {
			if task.Env == nil {
				task.Env = EnvVar{}
			}
			task.Env[k] = v
		}

		rf.Tasks[taskName] = task
	}

	if args.ExecuteInParallel {
		ctx.Debug("running in parallel mode", "tasks", args.Tasks)
		g := new(errgroup.Group)

		for _, _tn := range args.Tasks {
			tn := _tn
			g.Go(func() error {
				return rf.runTask(ctx, runTaskArgs{taskName: tn})
			})
		}

		// Wait for all tasks to finish
		if err := g.Wait(); err != nil {
			return errors.TaskFailed.WithErr(err).WithMetadata("task", args.Tasks, "runfile", rf.attrs.RunfilePath)
		}

		return nil
	}

	for _, tn := range args.Tasks {
		if err := rf.runTask(ctx, runTaskArgs{taskName: tn}); err != nil {
			return err
		}
	}

	return nil
}
