package runner

import (
	// "bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	// "github.com/alecthomas/chroma/v2/quick"
	// "github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/nxtcoder17/fwatcher/pkg/executor"
	"github.com/nxtcoder17/fwatcher/pkg/watcher"
	"github.com/nxtcoder17/runfile/errors"
	fn "github.com/nxtcoder17/runfile/functions"
	"github.com/nxtcoder17/runfile/parser"
	"github.com/nxtcoder17/runfile/types"
)

type runTaskArgs struct {
	taskTrail    []string
	taskName     string
	envOverrides map[string]string

	DebugEnv bool
}

func isDarkTheme() bool {
	return termenv.NewOutput(os.Stdout).HasDarkBackground()
}

func padString(v string, padWith string) string {
	sp := strings.Split(v, "\n")
	for i := range sp {
		if i == 0 {
			sp[i] = fmt.Sprintf("%s | %s", padWith, sp[i])
			continue
		}
		sp[i] = fmt.Sprintf("%s | %s", strings.Repeat(" ", len(padWith)), sp[i])
	}

	return strings.Join(sp, "\n")
}

// [snippet source](https://rderik.com/blog/identify-if-output-goes-to-the-terminal-or-is-being-redirected-in-golang/)
func isTTY() bool {
	stdout, _ := os.Stdout.Stat()
	stderr, _ := os.Stderr.Stat()
	return ((stdout.Mode() & os.ModeCharDevice) == os.ModeCharDevice) && ((stderr.Mode() & os.ModeCharDevice) == os.ModeCharDevice)
}

type CreateCommandGroupArgs struct {
	Runfile *types.ParsedRunfile
	Task    *types.ParsedTask

	Trail []string

	Stdout *LogWriter
	Stderr *LogWriter
}

func createCommandGroups(ctx Context, args CreateCommandGroupArgs) ([]executor.CommandGroup, error) {
	var cmds []executor.CommandGroup

	for _, cmd := range args.Task.Commands {
		switch {
		case cmd.Run != nil:
			{
				rt, ok := args.Runfile.Tasks[*cmd.Run]
				if !ok {
					return nil, fmt.Errorf("invalid run target")
				}

				rtp, err := parser.ParseTask(ctx, args.Runfile, rt)
				if err != nil {
					return nil, errors.WithErr(err).KV("env-vars", args.Runfile.Env)
				}

				rtCommands, err := createCommandGroups(ctx, CreateCommandGroupArgs{
					Runfile: args.Runfile,
					Task:    rtp,
					Trail:   append(append([]string{}, args.Trail...), rtp.Name),
					Stdout:  args.Stdout,
					Stderr:  args.Stderr,
				})
				if err != nil {
					return nil, errors.WithErr(err).KV("env-vars", args.Runfile.Env)
				}

				cg := executor.CommandGroup{
					Groups:   rtCommands,
					Parallel: rtp.Parallel,
				}

				cmds = append(cmds, cg)
			}

		case cmd.Command != nil:
			{
				cg := executor.CommandGroup{Parallel: args.Task.Parallel}

				cg.Commands = append(cg.Commands, func(c context.Context) *exec.Cmd {
					return CreateCommand(ctx, CmdArgs{
						Shell:       args.Task.Shell,
						Env:         fn.ToEnviron(args.Task.Env),
						Cmd:         *cmd.Command,
						WorkingDir:  args.Task.WorkingDir,
						interactive: args.Task.Interactive,
						Stdout:      args.Stdout.WithPrefix(strings.Join(args.Trail, "/")),
						Stderr:      args.Stderr.WithPrefix(strings.Join(args.Trail, "/")),
					})
				})

				ctx.Debug("HERE", "cmd", *cmd.Command, "parallel", args.Task.Parallel)

				cmds = append(cmds, cg)
			}
		}
	}

	return cmds, nil
}

func runTask(ctx Context, prf *types.ParsedRunfile, args runTaskArgs) error {
	runfilePath := prf.Metadata.RunfilePath
	task := prf.Tasks[args.taskName]

	if task.Metadata.RunfilePath != nil {
		runfilePath = *task.Metadata.RunfilePath
	}

	args.taskTrail = append(args.taskTrail, args.taskName)

	logger := ctx.With("task", args.taskName, "runfile", fn.Must(filepath.Rel(fn.Must(os.Getwd()), runfilePath)))
	logger.Debug("running task")

	task, ok := prf.Tasks[args.taskName]
	if !ok {
		return errors.ErrTaskNotFound
	}

	task.Name = args.taskName

	pt, err := parser.ParseTask(ctx, prf, task)
	if err != nil {
		return errors.WithErr(err)
	}

	logStdout := &LogWriter{w: os.Stdout}

	execCommands, err := createCommandGroups(ctx, CreateCommandGroupArgs{
		Runfile: prf,
		Task:    pt,
		Trail:   []string{pt.Name},
		Stdout:  logStdout,
		Stderr:  logStdout,
	})
	if err != nil {
		return err
	}

	ctx.Debug("top level command groups", "len", len(execCommands))

	for i := range execCommands {
		ctx.Debug("debugging execCommands", "i", execCommands[i].Parallel)
	}

	ex := executor.NewCmdExecutor(ctx, executor.CmdExecutorArgs{
		Logger:      logger,
		Interactive: pt.Interactive,
		Commands:    execCommands,
		Parallel:    pt.Parallel,
	})

	switch pt.Watch == nil {
	case true:
		{
			if err := ex.Start(); err != nil {
				logger.Error(err, "while running command")
				return err
			}
			logger.Debug("completed")
		}
	case false:
		{
			var wg sync.WaitGroup
			if pt.Watch != nil && (pt.Watch.Enable == nil || *pt.Watch.Enable) {
				watch, err := watcher.NewWatcher(ctx, watcher.WatcherArgs{
					Logger:               logger,
					WatchDirs:            append(task.Watch.Dirs, pt.WorkingDir),
					IgnoreDirs:           task.Watch.IgnoreDirs,
					WatchExtensions:      pt.Watch.Extensions,
					IgnoreExtensions:     pt.Watch.IgnoreExtensions,
					IgnoreList:           watcher.DefaultIgnoreList,
					Interactive:          pt.Interactive,
					ShouldLogWatchEvents: false,
				})
				if err != nil {
					return errors.WithErr(err)
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					<-ctx.Done()
					logger.Info("fwatcher is closing ...")
					watch.Close()
				}()

				executors := []executor.Executor{ex}

				if task.Watch.SSE != nil && task.Watch.SSE.Addr != "" {
					executors = append(executors, executor.NewSSEExecutor(executor.SSEExecutorArgs{Addr: task.Watch.SSE.Addr}))
				}

				if err := watch.WatchAndExecute(ctx, executors); err != nil {
					return err
				}
			}

			wg.Wait()
		}
	}

	return nil
}
