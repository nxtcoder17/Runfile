package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/nxtcoder17/runfile/errors"
	"github.com/nxtcoder17/runfile/logging"
	"github.com/nxtcoder17/runfile/runner"

	"github.com/nxtcoder17/runfile/parser"
	"github.com/urfave/cli/v3"
)

var Version string = fmt.Sprintf("nightly | %s", time.Now().Format(time.RFC3339))

var runfileNames []string = []string{
	"Runfile",
	"Runfile.yml",
	"Runfile.yaml",
}

//go:embed completions/run.fish
var shellCompletionFISH string

//go:embed completions/run.bash
var shellCompletionBASH string

//go:embed completions/run.zsh
var shellCompletionZSH string

//go:embed completions/run.ps
var shellCompletionPS string

func main() {
	cmd := cli.Command{
		Name:        "run",
		Version:     Version,
		Description: "A simple task runner",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "file",
				Aliases:   []string{"f"},
				TakesFile: true,
				Value:     "",
			},

			&cli.BoolFlag{
				Name:    "parallel",
				Aliases: []string{"p"},
				Value:   false,
			},

			&cli.BoolFlag{
				Name:    "watch",
				Aliases: []string{"w"},
				Value:   false,
			},

			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
			},

			&cli.BoolFlag{
				Name:    "list",
				Value:   false,
				Aliases: []string{"ls"},
			},

			&cli.BoolFlag{
				Name:  "debug-env",
				Value: false,
			},
		},

		// ShellCompletionCommandName: "completion:shell",
		EnableShellCompletion: true,

		// DefaultCommand:             "help",
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			if c.NArg() > 0 {
				return
			}

			runfilePath, err := locateRunfile(c)
			if err != nil {
				slog.Error("locating runfile", "err", err)
				panic(err)
			}

			generateShellCompletion(ctx, c.Root().Writer, runfilePath)
		},

		Commands: []*cli.Command{
			{
				Name:    "shell:completion",
				Suggest: true,
				Action: func(ctx context.Context, c *cli.Command) error {
					if c.NArg() != 2 {
						return fmt.Errorf("needs argument one of [bash,zsh,fish,ps]")
					}

					switch c.Args().Slice()[1] {
					case "fish":
						fmt.Fprint(c.Writer, shellCompletionFISH)
					case "bash":
						fmt.Fprint(c.Writer, shellCompletionBASH)
					case "zsh":
						fmt.Fprint(c.Writer, shellCompletionZSH)
					case "ps":
						fmt.Fprint(c.Writer, shellCompletionPS)
					}

					return nil
				},
			},
		},

		Suggest: true,
		Action: func(ctx context.Context, c *cli.Command) error {
			parallel := c.Bool("parallel")
			watch := c.Bool("watch")
			debug := c.Bool("debug")

			showList := c.Bool("list")
			if showList {
				runfilePath, err := locateRunfile(c)
				if err != nil {
					slog.Error("locating runfile, got", "err", err)
					return err
				}
				return generateShellCompletion(ctx, c.Root().Writer, runfilePath)
			}

			if c.NArg() == 0 {
				c.Command("help").Run(ctx, nil)
				return nil
			}

			runfilePath, err := locateRunfile(c)
			if err != nil {
				slog.Error("locating runfile, got", "err", err)
				return err
			}

			rf, err2 := parser.ParseRunfile(runfilePath)
			if err2 != nil {
				slog.Error("parsing runfile, got", "err", err2)
				panic(err2)
			}

			kv := make(map[string]string)

			// INFO: for supporting flags that have been suffixed post arguments
			args := make([]string, 0, len(c.Args().Slice()))
			for _, arg := range c.Args().Slice() {
				if arg == "-p" || arg == "--parallel" {
					parallel = true
					continue
				}

				if arg == "-w" || arg == "--watch" {
					watch = true
					continue
				}

				if arg == "--debug" {
					debug = true
					continue
				}

				sp := strings.SplitN(arg, "=", 2)
				if len(sp) == 2 {
					kv[sp[0]] = sp[1]
					continue
				}

				args = append(args, arg)
			}

			if parallel && watch {
				return fmt.Errorf("parallel and watch can't be set together")
			}

			logger := logging.New(logging.Options{
				ShowCaller:         false,
				SlogKeyAsPrefix:    "task",
				ShowDebugLogs:      debug,
				SetAsDefaultLogger: true,
			})

			return runner.Run(runner.NewContext(ctx, logger), rf, runner.RunArgs{
				Tasks:             args,
				ExecuteInParallel: parallel,
				Watch:             watch,
				Debug:             debug,
				KVs:               kv,
			})

			// return rf.Run(runfile.NewContext(ctx, logger), runfile.RunArgs{
			// 	Tasks:             args,
			// 	ExecuteInParallel: parallel,
			// 	Watch:             watch,
			// 	Debug:             debug,
			// 	KVs:               kv,
			// })
		},
	}

	ctx, cf := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGTERM)
	defer cf()

	go func() {
		<-ctx.Done()
		cf()
	}()

	if err := cmd.Run(ctx, os.Args); err != nil {
		errm, ok := err.(*errors.Error)
		slog.Debug("got", "err", err)
		if ok {
			if errm != nil {
				// errm.Error()
				// TODO: change it to a better logging
				// slog.Error("got", "err", errm)
				errm.Log()
			}
		} else {
			slog.Error("got", "err", err)
		}
	}
}

func locateRunfile(c *cli.Command) (string, error) {
	switch {
	case c.IsSet("file"):
		return c.String("file"), nil
	default:
		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}

		oldDir := ""

		for oldDir != dir {
			for _, fn := range runfileNames {
				if _, err := os.Stat(filepath.Join(dir, fn)); err != nil {
					if !os.IsNotExist(err) {
						return "", err
					}
					continue
				}

				return filepath.Join(dir, fn), nil
			}

			oldDir = dir
			dir = filepath.Dir(dir)
		}

		return "", fmt.Errorf("failed to locate your nearest Runfile")
	}
}
