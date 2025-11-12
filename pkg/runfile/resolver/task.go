package resolver

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/nxtcoder17/runfile/pkg/runfile/spec"
	"github.com/nxtcoder17/runfile/pkg/writer"
	"golang.org/x/term"
)

type TaskContext struct {
	context.Context
	PWD string
	Env map[string]string
}

type Command struct {
	Text        string
	IsRunTarget bool
	Env         map[string]any
}

type ResolvedTask struct {
	Name string

	Dir   string
	Shell []string
	Env   map[string]string

	Parallel    bool
	Interactive bool
	Watch       *spec.TaskWatchSpec

	Commands []*Command
}

func (r *Resolver) GetTask(name string) (*ResolvedTask, error) {
	task, ok := r.Tasks[name]
	if !ok {
		return nil, errors.New("Task Not Found").KV("task", name, "tasks", r.Tasks)
	}

	shell, err := ParseShell(task.Shell)
	if err != nil {
		return nil, err
	}

	envStore := fn.MapMerge(r.Env)
	if err := parseDotEnvFilesInto(envStore, task.DotEnv); err != nil {
		return nil, err
	}
	if err := parseEnvInto(context.TODO(), envStore, task.Env); err != nil {
		return nil, err
	}

	commands, err := parseCommands(task.Commands)
	if err != nil {
		return nil, err
	}

	if task.IsImported {
		for i := range commands {
			if commands[i].IsRunTarget {
				commands[i].Text = task.ImportPrefix + ":" + commands[i].Text
			}
		}
	}

	return &ResolvedTask{
		Name:        name,
		Dir:         task.Dir,
		Shell:       shell,
		Env:         envStore,
		Commands:    commands,
		Parallel:    task.Parallel,
		Interactive: task.Interactive,
		Watch:       task.Watch,
	}, nil
}

func (r *Resolver) RunTask(ctx context.Context, name string) error {
	rt, err := r.GetTask(name)
	if err != nil {
		return err
	}

	lw := &writer.LogWriter{Writer: os.Stderr}

	commandGroups, err := r.createCommandGroups(rt, createCommandGroupArgs{
		Stdout: lw,
		Stderr: lw,
	})
	if err != nil {
		return err
	}

	ex := executor.NewCmdExecutor(ctx, executor.CmdExecutorArgs{
		Interactive: rt.Interactive,
		Commands:    commandGroups,
		Parallel:    rt.Parallel,
	})

	notWatching := rt.Watch == nil || rt.Watch.Enabled == false

	if notWatching {
		return ex.Start()
	}

	var wg sync.WaitGroup
	watch, err := watcher.NewWatcher(ctx, watcher.WatcherArgs{
		WatchDirs:            rt.Watch.Dirs,
		IgnoreDirs:           rt.Watch.IgnoreDirs,
		WatchExtensions:      rt.Watch.Extensions,
		IgnoreExtensions:     rt.Watch.IgnoreExtensions,
		IgnoreList:           watcher.DefaultIgnoreList,
		Interactive:          rt.Interactive,
		ShouldLogWatchEvents: false,
	})
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		// ctx.Logger().Info("fwatcher is closing ...")
		watch.Close()
	}()

	executors := []executor.Executor{ex}

	if rt.Watch.SSE != nil && rt.Watch.SSE.Addr != "" {
		executors = append(executors, executor.NewSSEExecutor(executor.SSEExecutorArgs{Addr: rt.Watch.SSE.Addr}))
	}

	if err := watch.WatchAndExecute(ctx, executors); err != nil {
		return err
	}

	wg.Wait()
	return nil
}

func parseCommands(commands []any) ([]*Command, error) {
	result := make([]*Command, 0, len(commands))

	for _, command := range commands {
		switch c := command.(type) {
		case string:
			{
				if c == "" {
					continue
				}

				result = append(result, &Command{Text: c, IsRunTarget: false, Env: nil})
			}
		case map[string]any:
			{
				var jsonCmd struct {
					Run     *string        `json:"run"`
					Command *string        `json:"cmd"`
					Env     map[string]any `json:"env"`
				}

				b, err := json.Marshal(c)
				if err != nil {
					return nil, errors.New("failed to marshal command").Wrap(err).KV("cmd", c)
				}

				if err := json.Unmarshal(b, &jsonCmd); err != nil {
					return nil, errors.New("failed to unmarshal command into json command").Wrap(err).KV("b", b)
				}

				cmd := &Command{
					Env: jsonCmd.Env,
				}

				switch {
				case jsonCmd.Run != nil:
					{
						if *jsonCmd.Run == "" {
							return nil, errors.New("empty run target")
						}

						cmd.Text = *jsonCmd.Run
						cmd.IsRunTarget = true
					}
				case jsonCmd.Command != nil:
					{
						if *jsonCmd.Command == "" {
							return nil, errors.New("empty command")
						}
						cmd.Text = *jsonCmd.Command
					}
				default:
					{
						return nil, errors.WrapStr("either 'run' or 'cmd' key, must be specified when setting command in json format")
					}
				}
				result = append(result, cmd)
			}
		default:
			{
				return nil, errors.New("invalid command type, must be either a string or an object")
			}
		}
	}

	return result, nil
}

type CmdArgs struct {
	Shell      []string
	Env        []string // [key=value, key=value, ...]
	WorkingDir string

	Cmd string

	isInteractive bool
	Stdout        io.Writer
	Stderr        io.Writer
}

func CreateCommand(ctx context.Context, args CmdArgs) *exec.Cmd {
	shell := args.Shell[0]

	cargs := append(args.Shell[1:], args.Cmd)
	// #nosec G204 - shell is from a predefined map, command is passed via shell's -c flag
	c := exec.CommandContext(ctx, shell, cargs...)
	c.Dir = args.WorkingDir
	c.Env = args.Env
	c.Stdout = args.Stdout
	c.Stderr = args.Stderr

	if args.isInteractive {
		c.Stdin = os.Stdin
	}

	return c
}

func isDarkTheme() bool {
	return termenv.NewOutput(os.Stdout).HasDarkBackground()
}

func printCommand(w io.Writer, prefix, lang, cmd string) {
	if writer.IsANSITerminal() {
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
		fmt.Fprintf(w, "\r%s%s\n", s.Render(padString(hlCode.String(), prefix)), s.UnsetBorderStyle())
	}
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

type createCommandGroupArgs struct {
	Stdout *writer.LogWriter
	Stderr *writer.LogWriter
	Env    map[string]string
}

func (r *Resolver) createCommandGroups(task *ResolvedTask, args createCommandGroupArgs) ([]executor.CommandGroup, error) {
	var groups []executor.CommandGroup

	for _, cmd := range task.Commands {
		envStore := fn.MapMerge(r.Env, task.Env)

		if cmd.Env != nil {
			if err := parseEnvInto(context.TODO(), envStore, cmd.Env); err != nil {
				return nil, err
			}
		}

		if cmd.IsRunTarget {
			rt, err := r.GetTask(cmd.Text)
			if err != nil {
				return nil, err
			}

			rtCommands, err := r.createCommandGroups(rt, createCommandGroupArgs{
				Stdout: args.Stdout,
				Stderr: args.Stderr,
				Env:    envStore,
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
					args.Stderr.WithDimmedPrefix(cmd.Text).Write([]byte(sp[2]))
				},
			}

			groups = append(groups, cg)
			return groups, nil
		}

		cg := executor.CommandGroup{
			Parallel: task.Parallel,
			PreExecCommand: func(cmd *exec.Cmd) {
				str := strings.TrimSpace(cmd.String())
				sp := strings.SplitN(str, " ", len(task.Shell)+1)

				lang := "bash"
				if len(task.Shell) > 0 {
					lang = task.Shell[0]
				}
				printCommand(args.Stderr, task.Name, lang, sp[2])
			},

			Commands: []func(c context.Context) *exec.Cmd{
				func(c context.Context) *exec.Cmd {
					return CreateCommand(c, CmdArgs{
						Shell:         task.Shell,
						Env:           fn.ToEnviron(envStore),
						Cmd:           cmd.Text,
						WorkingDir:    task.Dir,
						isInteractive: task.Interactive,
						Stdout:        args.Stdout.WithPrefix(task.Name),
						Stderr:        args.Stderr.WithPrefix(task.Name),
					})
				},
			},
		}

		groups = append(groups, cg)
	}

	return groups, nil
}
