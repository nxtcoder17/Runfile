package runfile

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/nxtcoder17/fwatcher/pkg/executor"
	"github.com/nxtcoder17/fwatcher/pkg/watcher"
	"github.com/nxtcoder17/runfile/pkg/errors"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"github.com/nxtcoder17/runfile/pkg/writer"
	"golang.org/x/term"
)

type CreateCommandGroupArgs struct {
	Trail []string

	Stdout *writer.LogWriter
	Stderr *writer.LogWriter

	Env map[string]string
}

func isDarkTheme() bool {
	return termenv.NewOutput(os.Stdout).HasDarkBackground()
}

func longestLineLen(str string) int {
	sp := strings.Split(str, "\n")
	l := len(sp[0])
	for i := 1; i < len(sp); i++ {
		if len(sp[i]) > l {
			l = len(sp[i])
		}
	}

	return l
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

type CmdArgs struct {
	Shell      []string
	Env        []string // [key=value, key=value, ...]
	WorkingDir *string

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
	c.Dir = func() string {
		if args.WorkingDir != nil {
			return *args.WorkingDir
		}
		return ""
	}()
	c.Env = args.Env
	c.Stdout = args.Stdout
	c.Stderr = args.Stderr

	if args.interactive {
		c.Stdin = os.Stdin
	}

	return c
}

func printCommand(writer io.Writer, prefix, lang, cmd string) {
	if isTTY() {
		borderColor := "#4388cc"
		if !isDarkTheme() {
			borderColor = "#3d5485"
		}

		s := lipgloss.NewStyle().BorderForeground(lipgloss.Color(borderColor)).PaddingLeft(1).PaddingRight(1).Border(lipgloss.RoundedBorder(), true, true, true, true)

		width := 0

		if term.IsTerminal(0) {
			width, _, _ = term.GetSize(0)
		}

		hlCode := new(bytes.Buffer)
		// choose colorschemes from `https://swapoff.org/chroma/playground/`
		colorscheme := "catppuccin-macchiato"
		if !isDarkTheme() {
			colorscheme = "xcode"
		}
		_ = colorscheme

		longestLen := longestLineLen(cmd) + len(prefix) + 2 // 2 for spaces around prefix

		cmdStr := strings.TrimSpace(cmd)

		quick.Highlight(hlCode, cmdStr, lang, "terminal16m", colorscheme)

		if width > 0 && longestLen >= width-2 {
			s = s.Width(width - 2)
		}
		fmt.Fprintf(writer, "\r%s%s\n", s.Render(padString(hlCode.String(), prefix)), s.UnsetBorderStyle())
	}
}

func (r *ParsedRunfile) createCommandGroups(ctx *Context, taskName string, args CreateCommandGroupArgs) ([]executor.CommandGroup, error) {
	task, ok := r.Tasks[taskName]
	if !ok {
		return nil, errors.ErrTaskNotFound(taskName)
	}

	env, err := r.ParseTaskEnv(ctx, taskName, args.Env)
	if err != nil {
		return nil, err
	}

	var groups []executor.CommandGroup

	for i := range task.Commands {
		cmd, err := parseCommand(ctx, task.Commands[i], env)
		if err != nil {
			return nil, err
		}

		ctx.Debug("debugging", "env", env, "task", taskName, "trail", args.Trail)

		switch {
		case cmd.Run != nil:
			{

				if task.Metadata.Namespace != "" {
					*cmd.Run = task.Metadata.Namespace + ":" + *cmd.Run
				}

				rt, ok := r.Tasks[*cmd.Run]
				if !ok {
					return nil, errors.ErrTaskNotFound(*cmd.Run).KV("all-tasks", fn.MapKeys(r.Tasks))
				}

				rtCommands, err := r.createCommandGroups(ctx, rt.Name, CreateCommandGroupArgs{
					Trail:  append(args.Trail, rt.Name),
					Stdout: args.Stdout,
					Stderr: args.Stderr,
					Env:    cmd.Env,
				})
				if err != nil {
					return nil, err
				}

				cg := executor.CommandGroup{
					Groups:   rtCommands,
					Parallel: rt.Parallel,
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
				shell, err := r.ParseTaskShell(taskName)
				if err != nil {
					return nil, err
				}

				task, ok := r.Tasks[taskName]
				if !ok {
					return nil, errors.ErrTaskNotFound(*cmd.Run).KV("all-tasks", fn.MapKeys(r.Tasks))
				}

				cg := executor.CommandGroup{
					Parallel: task.Parallel,
					PreExecCommand: func(cmd *exec.Cmd) {
						str := strings.TrimSpace(cmd.String())
						sp := strings.SplitN(str, " ", len(shell)+1)

						lang := "bash"
						if len(shell) > 0 {
							lang = shell[0]
						}
						printCommand(args.Stderr, task.Name, lang, sp[2])
					},

					Commands: []func(c context.Context) *exec.Cmd{
						func(c context.Context) *exec.Cmd {
							return CreateCommand(ctx, CmdArgs{
								Shell:       shell,
								Env:         fn.ToEnviron(cmd.Env),
								Cmd:         *cmd.Command,
								WorkingDir:  task.Dir,
								interactive: task.Interactive,
								Stdout: func() io.Writer {
									if task.Interactive {
										return os.Stdout
									}
									return args.Stdout.WithPrefix(task.Name)
								}(),
								Stderr: func() io.Writer {
									if task.Interactive {
										return os.Stderr
									}
									return args.Stderr.WithPrefix(task.Name)
								}(),
							})
						},
					},
				}

				groups = append(groups, cg)
			}
		}
	}

	return groups, nil
}

func (r *ParsedRunfile) RunTask(ctx *Context, name string) error {
	ctx.taskTrail = append(ctx.taskTrail, name)
	ctx.Debug("running", "task", name)

	logStdout := &writer.LogWriter{Writer: os.Stdout}

	commandGroups, err := r.createCommandGroups(ctx, name, CreateCommandGroupArgs{
		Trail:  []string{},
		Stdout: logStdout,
		Stderr: logStdout,
	})
	if err != nil {
		return err
	}

	task, ok := r.Tasks[name]
	if !ok {
		return errors.ErrTaskNotFound(name)
	}

	ex := executor.NewCmdExecutor(ctx, executor.CmdExecutorArgs{
		Logger:      ctx.Logger.Slog(),
		Interactive: task.Interactive,
		Commands:    commandGroups,
		Parallel:    task.Parallel,
	})

	switch task.Watch == nil {
	case true:
		{
			if err := ex.Start(); err != nil {
				ctx.Logger.Error("while running command, got", "err", err)
				return err
			}
			ctx.Debug("completed")
		}
	case false:
		{
			var wg sync.WaitGroup
			if task.Watch != nil && (task.Watch.Enable == nil || *task.Watch.Enable) {
				watch, err := watcher.NewWatcher(ctx, watcher.WatcherArgs{
					Logger: ctx.Logger.Slog(),
					// WatchDirs:            append(t.Watch.Dirs, t.Dir),
					WatchDirs:            task.Watch.Dirs,
					IgnoreDirs:           task.Watch.IgnoreDirs,
					WatchExtensions:      task.Watch.Extensions,
					IgnoreExtensions:     task.Watch.IgnoreExtensions,
					IgnoreList:           watcher.DefaultIgnoreList,
					Interactive:          task.Interactive,
					ShouldLogWatchEvents: false,
				})
				if err != nil {
					return err
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					<-ctx.Done()
					ctx.Logger.Info("fwatcher is closing ...")
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
