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

func createCommands(ctx Context, prf *types.ParsedRunfile, pt *types.ParsedTask, args runTaskArgs) ([]executor.CommandGroup, error) {
	var cmds []executor.CommandGroup

	for _, cmd := range pt.Commands {
		switch {
		case cmd.Run != nil:
			{
				rt, ok := prf.Tasks[*cmd.Run]
				if !ok {
					return nil, fmt.Errorf("invalid run target")
				}

				rtp, err := parser.ParseTask(ctx, prf, rt)
				if err != nil {
					return nil, errors.WithErr(err).KV("env-vars", prf.Env)
				}

				rtCommands, err := createCommands(ctx, prf, rtp, args)
				if err != nil {
					return nil, errors.WithErr(err).KV("env-vars", prf.Env)
				}

				cg := executor.CommandGroup{
					Groups:   rtCommands,
					Parallel: rtp.Parallel,
				}
				//
				// logger := ctx.With("run", *cmd.Run)
				// logger.Info("got", "len(cg.Commands)", len(cg.Commands), "len(cg.Groups)", len(cg.Groups))
				cmds = append(cmds, cg)

				// cmds = append(cmds, rtCommands...)

				// ctx.Debug("HERE", "commands", len(cg.Commands), "run", *cmd.Run)
				// cg.Parallel = rtp.Parallel
			}

		case cmd.Command != nil:
			{
				cg := executor.CommandGroup{Parallel: pt.Parallel}

				cg.Commands = append(cg.Commands, func(c context.Context) *exec.Cmd {
					return CreateCommand(ctx, CmdArgs{
						Shell:       pt.Shell,
						Env:         fn.ToEnviron(pt.Env),
						Cmd:         *cmd.Command,
						WorkingDir:  pt.WorkingDir,
						interactive: pt.Interactive,
						Stdout:      os.Stdout,
						Stderr:      os.Stderr,
					})
				})

				ctx.Debug("HERE", "cmd", *cmd.Command, "parallel", pt.Parallel)

				cmds = append(cmds, cg)
			}
		}

		// cmds = append(cmds, cg)
	}

	return cmds, nil
}

// func runCommand(ctx Context, prf *types.ParsedRunfile, pt *types.ParsedTask, args runTaskArgs, command types.ParsedCommandJson) error {
// 	ctx.Debug("running command task", "command.run", command.Runs, "parent.task", args.taskName)
// 	if command.If != nil && !*command.If {
// 		ctx.Debug("skipping execution for failed `if`", "command", command.Runs)
// 		return nil
// 	}
//
// 	if command.Runs != nil {
// 		for _, run := range command.Runs {
// 			rt, ok := prf.Tasks[run]
// 			if !ok {
// 				return fmt.Errorf("invalid run target")
// 			}
//
// 			rtp, err := parser.ParseTask(ctx, prf, rt)
// 			if err != nil {
// 				return errors.WithErr(err).KV("env-vars", prf.Env)
// 			}
//
// 			if err := runTaskCommands(ctx, prf, rtp, args); err != nil {
// 				return errors.WithErr(err).KV("env-vars", prf.Env)
// 			}
// 			return nil
// 		}
// 	}
//
// 	// stdoutR, stdoutW := io.Pipe()
// 	// stderrR, stderrW := io.Pipe()
//
// 	// wg := sync.WaitGroup{}
//
// 	// [snippet source](https://rderik.com/blog/identify-if-output-goes-to-the-terminal-or-is-being-redirected-in-golang/)
// 	// stdout, _ := os.Stdout.Stat()
// 	// stderr, _ := os.Stderr.Stat()
// 	// isTTY := ((stdout.Mode() & os.ModeCharDevice) == os.ModeCharDevice) && ((stderr.Mode() & os.ModeCharDevice) == os.ModeCharDevice)
// 	//
// 	// if isTTY {
// 	// 	go func() {
// 	// 		defer wg.Done()
// 	// 		logPrefix := fmt.Sprintf("%s ", ctx.theme.TaskPrefixStyle.Render(fmt.Sprintf("[%s]", strings.Join(trail, "/"))))
// 	// 		processOutput(os.Stdout, stdoutR, &logPrefix)
// 	//
// 	// 		stderrPrefix := fmt.Sprintf("%s ", ctx.theme.TaskPrefixStyle.Render(fmt.Sprintf("[%s/stderr]", strings.Join(trail, "/"))))
// 	// 		processOutput(os.Stderr, stderrR, &stderrPrefix)
// 	// 	}()
// 	// } else {
// 	// 	wg.Add(1)
// 	// 	go func() {
// 	// 		defer wg.Done()
// 	// 		logPrefix := fmt.Sprintf("%s ", ctx.theme.TaskPrefixStyle.Render(fmt.Sprintf("[%s]", strings.Join(trail, "/"))))
// 	// 		processOutput(os.Stdout, stdoutR, &logPrefix)
// 	// 		// if pt.Interactive {
// 	// 		// 	processOutput(os.Stdout, stdoutR, &logPrefix)
// 	// 		// 	return
// 	// 		// }
// 	// 		// processOutputLineByLine(os.Stdout, stdoutR, &logPrefix)
// 	// 	}()
// 	//
// 	// 	wg.Add(1)
// 	// 	go func() {
// 	// 		defer wg.Done()
// 	// 		logPrefix := fmt.Sprintf("%s ", ctx.theme.TaskPrefixStyle.Render(fmt.Sprintf("[%s/stderr]", strings.Join(trail, "/"))))
// 	// 		processOutput(os.Stderr, stderrR, &logPrefix)
// 	// 		// if pt.Interactive {
// 	// 		// 	processOutput(os.Stderr, stderrR, &logPrefix)
// 	// 		// 	return
// 	// 		// }
// 	// 		// processOutputLineByLine(os.Stderr, stderrR, &logPrefix)
// 	// 	}()
// 	// }
//
// 	if isTTY() {
// 		borderColor := "#4388cc"
// 		if !isDarkTheme() {
// 			borderColor = "#3d5485"
// 		}
// 		s := lipgloss.NewStyle().BorderForeground(lipgloss.Color(borderColor)).PaddingLeft(1).PaddingRight(1).Border(lipgloss.RoundedBorder(), true, true, true, true)
// 		// labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor)).Blink(true)
//
// 		if args.DebugEnv {
// 			fmt.Printf("%s\n", s.Render(padString(fmt.Sprintf("%+v", prf.Env), "DEBUG: env")))
// 		}
//
// 		hlCode := new(bytes.Buffer)
// 		// choose colorschemes from `https://swapoff.org/chroma/playground/`
// 		colorscheme := "catppuccin-macchiato"
// 		if !isDarkTheme() {
// 			colorscheme = "monokailight"
// 		}
// 		// quick.Highlight(hlCode, strings.TrimSpace(command.Command), "bash", "terminal16m", colorscheme)
//
// 		for i := range command.Commands {
// 			cmdStr := strings.TrimSpace(command.Commands[i])
//
// 			quick.Highlight(hlCode, cmdStr, "bash", "terminal16m", colorscheme)
// 			// cst := styles.Get("gruvbox")
// 			// fmt.Println("cst: ", cst.Name, styles.Fallback.Name, styles.Names())
//
// 			// fmt.Printf("%s\n", s.Render(args.taskName+" | "+hlCode.String()))
// 			fmt.Printf("%s\n", s.Render(padString(hlCode.String(), args.taskName)))
// 		}
// 	}
//
// 	// logger2 := logging.New(logging.Options{
// 	// 	Prefix:          "[runfile]",
// 	// 	Writer:          os.Stderr,
// 	// 	SlogKeyAsPrefix: "task",
// 	// })
//
// 	// ex := executor.NewCmdExecutor(ctx, executor.CmdExecutorArgs{
// 	// 	Logger:      logger2,
// 	// 	Interactive: pt.Interactive,
// 	// 	Commands: func(c context.Context) []*exec.Cmd {
// 	// 		commands := make([]*exec.Cmd, 0, len(command.Commands))
// 	// 		for i := range command.Commands {
// 	// 			commands = append(commands, CreateCommand(c, CmdArgs{
// 	// 				Shell:       pt.Shell,
// 	// 				Env:         fn.ToEnviron(pt.Env),
// 	// 				Cmd:         command.Commands[i],
// 	// 				WorkingDir:  pt.WorkingDir,
// 	// 				interactive: pt.Interactive,
// 	// 				Stdout:      os.Stdout,
// 	// 				Stderr:      os.Stderr,
// 	// 			}))
// 	// 		}
// 	// 		return commands
// 	// 	},
// 	// })
// 	//
// 	// wg.Add(1)
// 	// go func() {
// 	// 	defer wg.Done()
// 	// 	<-ctx.Done()
// 	// 	ex.Stop()
// 	// }()
// 	//
// 	// if err := ex.Start(); err != nil {
// 	// 	return errors.ErrTaskFailed.Wrap(err).KV("task", args.taskName)
// 	// }
//
// 	return nil
// }

// func runTaskCommands(ctx Context, prf *types.ParsedRunfile, pt *types.ParsedTask, args runTaskArgs) error {
// 	for _, command := range pt.Commands {
// 		if err := runCommand(ctx, prf, pt, args, command); err != nil {
// 			return err
// 		}
// 	}
//
// 	return nil
// }

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

	execCommands, err := createCommands(ctx, prf, pt, args)
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
