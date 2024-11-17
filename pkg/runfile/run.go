package runfile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"golang.org/x/sync/errgroup"
)

type cmdArgs struct {
	shell      []string
	env        []string // [key=value, key=value, ...]
	workingDir string

	cmd string

	interactive bool
	stdout      io.Writer
	stderr      io.Writer
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

	if args.interactive {
		c.Stdin = os.Stdin
	}

	return c
}

type runTaskArgs struct {
	taskTrail    []string
	taskName     string
	envOverrides map[string]string
}

func processOutput(writer io.Writer, reader io.Reader, prefix *string) {
	prevByte := byte('\n')
	msg := make([]byte, 1)
	for {
		n, err := reader.Read(msg)
		if err != nil {
			// logger.Info("stdout", "msg", string(msg[:n]), "err", err)
			if errors.Is(err, io.EOF) {
				writer.Write(msg[:n])
				return
			}
		}

		if n != 1 {
			continue
		}

		if prevByte == '\n' && prefix != nil {
			// os.Stdout.WriteString(fmt.Sprintf("HERE... msg: '%s'", msg[:n]))
			writer.Write([]byte(*prefix))
		}

		writer.Write(msg[:n])
		prevByte = msg[0]
	}
}

func runTask(ctx Context, rf *Runfile, args runTaskArgs) *Error {
	runfilePath := fn.Must(filepath.Rel(rf.attrs.RootRunfilePath, rf.attrs.RunfilePath))

	trail := append(args.taskTrail, args.taskName)

	formatErr := func(err *Error) *Error {
		if runfilePath != "." {
			return err.WithTask(strings.Join(trail, "/")).WithRunfile(runfilePath)
		}
		return err.WithTask(strings.Join(trail, "/"))
	}

	logger := ctx.With("task", args.taskName, "runfile", runfilePath)
	logger.Debug("running task")
	task, ok := rf.Tasks[args.taskName]
	if !ok {
		return formatErr(TaskNotFound)
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
		return formatErr(err)
	}

	logger.Debug("debugging env", "pt.environ", pt.Env, "overrides", args.envOverrides)
	for _, command := range pt.Commands {
		logger.Debug("running command task", "command.run", command.Run, "parent.task", args.taskName)

		if command.If != nil && !*command.If {
			logger.Debug("skipping execution for failed `if`", "command", command.Run)
			continue
		}

		if command.Run != "" {
			if err := runTask(ctx, rf, runTaskArgs{
				taskTrail:    trail,
				taskName:     command.Run,
				envOverrides: pt.Env,
			}); err != nil {
				return err
			}
			continue
		}

		stdoutR, stdoutW := io.Pipe()
		stderrR, stderrW := io.Pipe()

		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			logPrefix := fmt.Sprintf("%s ", ctx.theme.TaskPrefixStyle.Render(fmt.Sprintf("[%s]", strings.Join(trail, "/"))))
			processOutput(os.Stdout, stdoutR, &logPrefix)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			logPrefix := fmt.Sprintf("%s ", ctx.theme.TaskPrefixStyle.Render(fmt.Sprintf("[%s]", strings.Join(trail, "/"))))
			processOutput(os.Stderr, stderrR, &logPrefix)
		}()

		cmd := createCommand(ctx, cmdArgs{
			shell:       pt.Shell,
			env:         ToEnviron(pt.Env),
			cmd:         command.Command,
			workingDir:  pt.WorkingDir,
			interactive: pt.Interactive,
			stdout:      stdoutW,
			stderr:      stderrW,
		})
		if err := cmd.Run(); err != nil {
			return formatErr(CommandFailed).WithErr(err)
		}

		stdoutW.Close()
		stderrW.Close()

		wg.Wait()
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

func (rf *Runfile) Run(ctx Context, args RunArgs) *Error {
	includes, err := rf.ParseIncludes()
	if err != nil {
		return err
	}

	for _, taskName := range args.Tasks {
		for k, v := range includes {
			for tn := range v.Runfile.Tasks {
				if taskName == fmt.Sprintf("%s:%s", k, tn) {
					return runTask(ctx, v.Runfile, runTaskArgs{taskName: tn})
				}
			}
		}

		task, ok := rf.Tasks[taskName]
		if !ok {
			errAttr := []any{"task", taskName}
			if rf.attrs.RunfilePath != rf.attrs.RootRunfilePath {
				errAttr = append(errAttr, "runfile", fn.Must(filepath.Rel(rf.attrs.RootRunfilePath, rf.attrs.RunfilePath)))
			}
			return TaskNotFound.WithMetadata(errAttr...)
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
				return runTask(ctx, rf, runTaskArgs{taskName: tn})
			})
		}

		// Wait for all tasks to finish
		if err := g.Wait(); err != nil {
			err2, ok := err.(*Error)
			if ok {
				return err2
			}
			return TaskFailed.WithErr(err)
		}

		return nil
	}

	for _, tn := range args.Tasks {
		if err := runTask(ctx, rf, runTaskArgs{taskName: tn}); err != nil {
			return err
		}
	}

	return nil
}
