package resolver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
	"github.com/nxtcoder17/fwatcher/pkg/watcher"
	"github.com/nxtcoder17/go.errors"
	"github.com/nxtcoder17/runfile/pkg/executor"
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
	Env   map[string]any

	Parallel    bool
	Interactive bool
	Silent      bool

	Watch *spec.TaskWatchSpec

	Commands []*Command
}

func (r *Resolver) GetTask(name string) (*ResolvedTask, error) {
	task, ok := r.Tasks[name]
	if !ok {
		return nil, errors.New("Task Not Found").KV("task", name, "all-tasks", fn.MapKeys(r.Tasks))
	}

	shell, err := ParseShell(task.Shell)
	if err != nil {
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
		Env:         task.Env,
		Commands:    commands,
		Parallel:    task.Parallel,
		Interactive: task.Interactive,
		Silent:      task.Silent,
		Watch:       task.Watch,
	}, nil
}

func (r *Resolver) RunTask(ctx context.Context, name string) error {
	rt, err := r.GetTask(name)
	if err != nil {
		return err
	}

	lw := &writer.LogWriter{Writer: os.Stderr}

	steps, err := r.createSteps(rt, createCommandGroupArgs{
		Stdout: lw,
		Stderr: lw,
	})
	if err != nil {
		return err
	}

	pipeline := executor.NewPipeline(slog.Default(), steps)

	notWatching := rt.Watch == nil || rt.Watch.Enabled == false

	if notWatching {
		return pipeline.Start(ctx)
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

	// executors := []executor.Executor{pipeline}
	//
	// if rt.Watch.SSE != nil && rt.Watch.SSE.Addr != "" {
	// 	executors = append(executors, executor.NewSSEExecutor(executor.SSEExecutorArgs{Addr: rt.Watch.SSE.Addr}))
	// }

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := pipeline.Start(ctx); err != nil {
			slog.Error("starting command", "err", err)
		}
		slog.Debug("final executor start finished")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		pipeline.Stop()
		slog.Debug("2. context cancelled")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		watch.Watch(ctx)
		slog.Debug("3. watcher closed")
	}()

	counter := 0
	for ev := range watch.GetEvents() {
		slog.Debug("received", "event", ev)
		counter += 1
		slog.Info(fmt.Sprintf("[RELOADING (%d)] due changes in %s", counter, ev.Name))
	}

	// if err := watch.WatchAndExecute(ctx, executors); err != nil {
	// 	return err
	// }
	//
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
				var jsonCmd spec.CommandObject

				b, err := json.Marshal(c)
				if err != nil {
					return nil, errors.New("failed to marshal command").Wrap(err).KV("cmd", c)
				}

				if err := json.Unmarshal(b, &jsonCmd); err != nil {
					return nil, errors.New("failed to unmarshal command into json command").Wrap(err).KV("b", b)
				}

				cmd := &Command{Env: jsonCmd.Env}

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
						return nil, errors.New("either 'run' or 'cmd' key, must be specified when setting command in json format")
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

func printCommand(w *writer.LogWriter, prefix, lang, cmd string) {
	borderColor := "#4388cc"
	switch os.Getenv("RUNFILE_THEME") {
	case "light":
		borderColor = "#3d5485"
	}

	myBorder := lipgloss.Border{
		Top:         "-+",
		Bottom:      "-+",
		Left:        "|",
		Right:       "|",
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
	}

	s := lipgloss.NewStyle().Border(myBorder).BorderForeground(lipgloss.Color(borderColor)).PaddingLeft(1).PaddingRight(1)
	defer s.UnsetBorderStyle()
	defer s.UnsetPadding()

	width := 0

	if term.IsTerminal(0) {
		width, _, _ = term.GetSize(0)
	}

	hlCode := new(bytes.Buffer)
	cmdStr := strings.TrimSpace(cmd)

	switch os.Getenv("RUNFILE_THEME") {
	case "dark":
		quick.Highlight(hlCode, cmdStr, lang, "terminal16m", "catppuccin-macchiato")
	case "light":
		quick.Highlight(hlCode, cmdStr, lang, "terminal16m", "xcode")
	default:
		hlCode.WriteString(cmdStr)
	}

	// INFO: 2 for spaces around prefix
	longestLen := longestLineLen(cmd) + len(prefix) + 2

	if width > 0 && longestLen >= width-2 {
		s = s.Width(width - 2)
	}

	// w.Mu.Lock()
	// defer w.Mu.Unlock()
	fmt.Fprintf(w, "\r\033[K%s%s\n", padString(s.Render(hlCode.String()), prefix), s.UnsetBorderStyle())
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

func padString(str string, withPrefix string) string {
	sp := strings.Split(str, "\n")
	for i := range sp {
		if i == 0 {
			sp[i] = fmt.Sprintf("%s %s", writer.GetStyledPrefix(withPrefix), sp[i])
			continue
		}
		sp[i] = fmt.Sprintf("%s %s", strings.Repeat(" ", len(withPrefix)+2), sp[i])
	}

	return strings.Join(sp, "\n")
}

type createCommandGroupArgs struct {
	Stdout    *writer.LogWriter
	Stderr    *writer.LogWriter
	Env       map[string]string
	Silent    bool
	TaskTrail []string
}

func (r *Resolver) createSteps(task *ResolvedTask, args createCommandGroupArgs) ([]executor.Step, error) {
	var steps []executor.Step

	taskTrail := make([]string, 0, len(args.TaskTrail)+1)
	for i := range args.TaskTrail {
		taskTrail = append(taskTrail, args.TaskTrail[i])
	}
	taskTrail = append(taskTrail, task.Name)

	slog.Debug("creating command groups", "task.name", task.Name, "task.trail", taskTrail)

	for _, cmd := range task.Commands {
		// INFO: env var overrides with args.Env takes priority
		envStore := fn.MapMerge(r.Env, args.Env)

		if task.Env != nil {
			if err := parseEnvInto(context.TODO(), envStore, task.Env); err != nil {
				return nil, err
			}
		}

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

			substeps, err := r.createSteps(rt, createCommandGroupArgs{
				Stdout:    args.Stdout,
				Stderr:    args.Stderr,
				Env:       envStore,
				Silent:    task.Silent || rt.Silent,
				TaskTrail: taskTrail,
			})
			if err != nil {
				return nil, err
			}

			steps = append(steps, executor.Step{
				SubSteps: substeps,
				Parallel: rt.Parallel,
			})
			continue
		}

		logPrefix := strings.Join(taskTrail, " ≫ ")

		step := executor.Step{Parallel: task.Parallel}

		cmdHandler := func(c context.Context) *exec.Cmd {
			return CreateCommand(c, CmdArgs{
				Shell:      task.Shell,
				Env:        fn.ToEnviron(envStore),
				Cmd:        cmd.Text,
				WorkingDir: task.Dir,
				Stdout:     args.Stdout.WithPrefix(logPrefix),
				Stderr:     args.Stderr.WithPrefix(logPrefix),
			})
		}

		preHook := func(c context.Context) error {
			if task.Silent || task.Interactive {
				return nil
			}
			str := strings.TrimSpace(cmd.Text)

			lang := "bash"
			if len(task.Shell) > 0 {
				lang = task.Shell[0]
			}
			printCommand(args.Stderr, logPrefix, lang, str)
			return nil
		}

		if task.Interactive {
			step.Commands = append(step.Commands, executor.NewInteractiveShellCommand(cmdHandler).AddPreHook(preHook))
		} else {
			step.Commands = append(step.Commands, executor.NewShellCommand(cmdHandler).AddPreHook(preHook))
		}

		steps = append(steps, step)
	}

	slog.Debug("created command groups", "len", len(steps))
	return steps, nil
}
