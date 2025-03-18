package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
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

func padString(str string, padWith string) string {
	sp := strings.Split(str, "\n")
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

func hasANSISupport() bool {
	term := os.Getenv("TERM")
	return strings.Contains(term, "xterm") || strings.Contains(term, "screen") || strings.Contains(term, "vt100")
}

func printCommand(writer io.Writer, prefix, lang, cmd string) {
	if isTTY() {
		borderColor := "#4388cc"
		if !isDarkTheme() {
			borderColor = "#3d5485"
		}
		s := lipgloss.NewStyle().BorderForeground(lipgloss.Color(borderColor)).PaddingLeft(1).PaddingRight(1).Border(lipgloss.RoundedBorder(), true, true, true, true)

		hlCode := new(bytes.Buffer)
		// choose colorschemes from `https://swapoff.org/chroma/playground/`
		colorscheme := "catppuccin-macchiato"
		if !isDarkTheme() {
			colorscheme = "monokailight"
		}
		_ = colorscheme

		cmdStr := strings.TrimSpace(cmd)

		quick.Highlight(hlCode, cmdStr, lang, "terminal16m", colorscheme)

		fmt.Fprintf(writer, "\r%s%s\n", s.Render(padString(hlCode.String(), prefix)), s.UnsetBorderStyle())
	}
}

type CreateCommandGroupArgs struct {
	Runfile *types.ParsedRunfile
	Task    *types.ParsedTask

	Trail []string

	Stdout *LogWriter
	Stderr *LogWriter

	EnvOverrides map[string]string
}

func createCommandGroups(ctx types.Context, args CreateCommandGroupArgs) ([]executor.CommandGroup, error) {
	var groups []executor.CommandGroup

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
					Runfile:      args.Runfile,
					Task:         rtp,
					Trail:        append(append([]string{}, args.Trail...), rtp.Name),
					Stdout:       args.Stdout,
					Stderr:       args.Stderr,
					EnvOverrides: cmd.Env,
				})
				if err != nil {
					return nil, errors.WithErr(err).KV("env-vars", args.Runfile.Env)
				}

				cg := executor.CommandGroup{
					Groups:   rtCommands,
					Parallel: rtp.Parallel,
					PreExecCommand: func(c *exec.Cmd) {
						str := c.String()
						sp := strings.SplitN(str, " ", 3)
						args.Stderr.WithDimmedPrefix(*cmd.Run).Write([]byte(sp[2]))
					},
				}

				groups = append(groups, cg)
			}

		case cmd.Command != nil:
			{
				cg := executor.CommandGroup{Parallel: args.Task.Parallel}

				cg.PreExecCommand = func(cmd *exec.Cmd) {
					str := strings.TrimSpace(cmd.String())
					sp := strings.SplitN(str, " ", len(args.Task.Shell)+1)
					printCommand(args.Stderr, args.Task.Name, "bash", sp[2])
				}

				cg.Commands = append(
					cg.Commands,
					func(c context.Context) *exec.Cmd {
						return CreateCommand(ctx, CmdArgs{
							Shell:       args.Task.Shell,
							Env:         fn.ToEnviron(fn.MapMerge(args.Task.Env, args.EnvOverrides)),
							Cmd:         *cmd.Command,
							WorkingDir:  args.Task.WorkingDir,
							interactive: args.Task.Interactive,
							Stdout: func() io.Writer {
								if args.Task.Interactive {
									return os.Stdout
								}
								return args.Stdout.WithPrefix(args.Task.Name)
							}(),
							Stderr: func() io.Writer {
								if args.Task.Interactive {
									return os.Stderr
								}
								return args.Stderr.WithPrefix(args.Task.Name)
							}(),
						})
					})

				ctx.Debug("HERE", "cmd", *cmd.Command, "parallel", args.Task.Parallel)

				groups = append(groups, cg)
			}
		}
	}

	return groups, nil
}

func runTask(ctx types.Context, prf *types.ParsedRunfile, args runTaskArgs) error {
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
