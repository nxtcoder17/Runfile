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
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
	"github.com/nxtcoder17/fwatcher/pkg/watcher"
	"github.com/nxtcoder17/go.errors"
	"github.com/nxtcoder17/runfile/pkg/executor"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"github.com/nxtcoder17/runfile/pkg/runfile/spec"
	"github.com/nxtcoder17/runfile/pkg/writer"
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

	currentPipeline := executor.NewPipeline(slog.Default(), steps)

	notWatching := rt.Watch == nil || rt.Watch.Enabled == false

	if notWatching {
		return currentPipeline.Start(ctx)
	}

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

	go watch.Watch(ctx)
	defer watch.Close()

	var pMu sync.Mutex

	run := func() {
		pMu.Lock()
		defer pMu.Unlock()

		if currentPipeline != nil {
			currentPipeline.Stop()
		}

		currentPipeline = executor.NewPipeline(slog.Default(), steps)
		go func() {
			if err := currentPipeline.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
				// slog.Error("pipeline finished with error", "err", err)
			}
		}()
	}

	run()
	// // initial run
	// go func() {
	// 	if err := currentPipeline.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
	// 		// slog.Error("pipeline finished with error", "err", err)
	// 	}
	// }()

	counter := 0
	debounceDuration := 300 * time.Millisecond
	timer := time.NewTimer(debounceDuration)
	if !timer.Stop() {
		<-timer.C
	}

	for {
		select {
		case <-ctx.Done():
			pMu.Lock()
			if currentPipeline != nil {
				currentPipeline.Stop()
			}
			pMu.Unlock()
			return nil
		case ev, ok := <-watch.GetEvents():
			if !ok {
				return nil
			}
			slog.Debug("received", "event", ev)
			timer.Reset(debounceDuration)
		case <-timer.C:
			counter += 1
			slog.Info(fmt.Sprintf("[RELOADING (%d)]", counter))
			run()
		}
	}
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

	// w.Mu.Lock()
	// defer w.Mu.Unlock()
	fmt.Fprintf(w, "\r\033[K%s\n", formatCommandPreview(hlCode.String(), prefix))
}

func formatCommandPreview(str string, withPrefix string) string {
	sp := strings.Split(str, "\n")
	indent := strings.Repeat(" ", prefixDisplayWidth(withPrefix))
	rail := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Faint(true).Render("│")
	for i := range sp {
		if i == 0 {
			sp[i] = fmt.Sprintf("%s %s %s", writer.GetStyledPrefix(withPrefix), rail, sp[i])
			continue
		}
		sp[i] = indent + rail + " " + sp[i]
	}

	return strings.Join(sp, "\n")
}

func prefixDisplayWidth(prefix string) int {
	if prefix == "" {
		return 0
	}

	return lipgloss.Width("[" + prefix + "] ")
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
			if cycle := appendCycle(taskTrail, cmd.Text); cycle != nil {
				return nil, errors.New("Circular Task Dependency").KV("cycle", strings.Join(cycle, " -> "))
			}

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

func appendCycle(taskTrail []string, next string) []string {
	for i := range taskTrail {
		if taskTrail[i] != next {
			continue
		}

		cycle := append([]string{}, taskTrail[i:]...)
		cycle = append(cycle, next)
		return cycle
	}

	return nil
}
