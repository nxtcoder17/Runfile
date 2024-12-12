package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

func runTask(ctx Context, prf *types.ParsedRunfile, args runTaskArgs) error {
	runfilePath := prf.Metadata.RunfilePath
	task := prf.Tasks[args.taskName]

	if task.Metadata.RunfilePath != nil {
		runfilePath = *task.Metadata.RunfilePath
	}

	trail := append(args.taskTrail, args.taskName)

	logger := ctx.With("task", args.taskName, "runfile", runfilePath)
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

	for _, command := range pt.Commands {
		logger.Debug("running command task", "command.run", command.Run, "parent.task", args.taskName)

		if command.If != nil && !*command.If {
			logger.Debug("skipping execution for failed `if`", "command", command.Run)
			continue
		}

		if command.Run != "" {
			if err := runTask(ctx, prf, runTaskArgs{
				taskTrail:    trail,
				taskName:     command.Run,
				envOverrides: pt.Env,
			}); err != nil {
				return errors.WithErr(err).KV("env-vars", prf.Env)
			}
			continue
		}

		// stdoutR, stdoutW := io.Pipe()
		// stderrR, stderrW := io.Pipe()

		// wg := sync.WaitGroup{}

		// [snippet source](https://rderik.com/blog/identify-if-output-goes-to-the-terminal-or-is-being-redirected-in-golang/)
		// stdout, _ := os.Stdout.Stat()
		// stderr, _ := os.Stderr.Stat()
		// isTTY := ((stdout.Mode() & os.ModeCharDevice) == os.ModeCharDevice) && ((stderr.Mode() & os.ModeCharDevice) == os.ModeCharDevice)
		//
		// if isTTY {
		// 	go func() {
		// 		defer wg.Done()
		// 		logPrefix := fmt.Sprintf("%s ", ctx.theme.TaskPrefixStyle.Render(fmt.Sprintf("[%s]", strings.Join(trail, "/"))))
		// 		processOutput(os.Stdout, stdoutR, &logPrefix)
		//
		// 		stderrPrefix := fmt.Sprintf("%s ", ctx.theme.TaskPrefixStyle.Render(fmt.Sprintf("[%s/stderr]", strings.Join(trail, "/"))))
		// 		processOutput(os.Stderr, stderrR, &stderrPrefix)
		// 	}()
		// } else {
		// 	wg.Add(1)
		// 	go func() {
		// 		defer wg.Done()
		// 		logPrefix := fmt.Sprintf("%s ", ctx.theme.TaskPrefixStyle.Render(fmt.Sprintf("[%s]", strings.Join(trail, "/"))))
		// 		processOutput(os.Stdout, stdoutR, &logPrefix)
		// 		// if pt.Interactive {
		// 		// 	processOutput(os.Stdout, stdoutR, &logPrefix)
		// 		// 	return
		// 		// }
		// 		// processOutputLineByLine(os.Stdout, stdoutR, &logPrefix)
		// 	}()
		//
		// 	wg.Add(1)
		// 	go func() {
		// 		defer wg.Done()
		// 		logPrefix := fmt.Sprintf("%s ", ctx.theme.TaskPrefixStyle.Render(fmt.Sprintf("[%s/stderr]", strings.Join(trail, "/"))))
		// 		processOutput(os.Stderr, stderrR, &logPrefix)
		// 		// if pt.Interactive {
		// 		// 	processOutput(os.Stderr, stderrR, &logPrefix)
		// 		// 	return
		// 		// }
		// 		// processOutputLineByLine(os.Stderr, stderrR, &logPrefix)
		// 	}()
		// }

		if isTTY() {
			borderColor := "#4388cc"
			if !isDarkTheme() {
				borderColor = "#3d5485"
			}
			s := lipgloss.NewStyle().BorderForeground(lipgloss.Color(borderColor)).PaddingLeft(1).PaddingRight(1).Border(lipgloss.RoundedBorder(), true, true, true, true)
			// labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor)).Blink(true)

			hlCode := new(bytes.Buffer)
			// choose colorschemes from `https://swapoff.org/chroma/playground/`
			colorscheme := "catppuccin-macchiato"
			if !isDarkTheme() {
				colorscheme = "monokailight"
			}
			_ = colorscheme
			// quick.Highlight(hlCode, strings.TrimSpace(command.Command), "bash", "terminal16m", colorscheme)

			cmdStr := strings.TrimSpace(command.Command)

			quick.Highlight(hlCode, cmdStr, "bash", "terminal16m", colorscheme)
			// cst := styles.Get("gruvbox")
			// fmt.Println("cst: ", cst.Name, styles.Fallback.Name, styles.Names())

			// fmt.Printf("%s\n", s.Render(args.taskName+" | "+hlCode.String()))
			fmt.Printf("%s\n", s.Render(padString(hlCode.String(), args.taskName)))
		}

		ex := executor.NewExecutor(executor.ExecutorArgs{
			Logger: logger,
			Command: func(c context.Context) *exec.Cmd {
				return CreateCommand(c, CmdArgs{
					Shell:       pt.Shell,
					Env:         fn.ToEnviron(pt.Env),
					Cmd:         command.Command,
					WorkingDir:  pt.WorkingDir,
					interactive: pt.Interactive,
					Stdout:      os.Stdout,
					Stderr:      os.Stderr,
				})
			},
		})

		if task.Watch.Enable {
			watch, err := watcher.NewWatcher(watcher.WatcherArgs{
				Logger:               logger,
				WatchDirs:            append(task.Watch.Dirs, pt.WorkingDir),
				OnlySuffixes:         pt.Watch.OnlySuffixes,
				IgnoreSuffixes:       pt.Watch.IgnoreSuffixes,
				ExcludeDirs:          pt.Watch.ExcludeDirs,
				UseDefaultIgnoreList: true,
			})
			if err != nil {
				return errors.WithErr(err)
			}

			go ex.Exec()

			go func() {
				<-ctx.Done()
				logger.Info("fwatcher is closing ...")
				<-time.After(200 * time.Millisecond)
				os.Exit(0)
			}()

			// if err := ex.Exec(); err != nil {
			// 	return errors.ErrTaskFailed.Wrap(err).KV("task", args.taskName)
			// }

			watch.WatchEvents(func(event watcher.Event, fp string) error {
				relPath, err := filepath.Rel(fn.Must(os.Getwd()), fp)
				if err != nil {
					return err
				}
				logger.Info(fmt.Sprintf("[RELOADING] due changes in %s", relPath))
				ex.Kill()
				<-time.After(100 * time.Millisecond)
				go ex.Exec()
				return nil
			})

			return nil
		}

		if err := ex.Exec(); err != nil {
			return errors.ErrTaskFailed.Wrap(err).KV("task", args.taskName)
		}
	}

	return nil
}
