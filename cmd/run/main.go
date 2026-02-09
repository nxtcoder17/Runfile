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

	"github.com/nxtcoder17/fastlog"
	"github.com/nxtcoder17/go.errors"
	"github.com/nxtcoder17/runfile/pkg/runfile"
	"github.com/urfave/cli/v3"
)

var Version string

//go:embed completions/run.fish
var shellCompletionFISH string

//go:embed completions/run.bash
var shellCompletionBASH string

//go:embed completions/run.zsh
var shellCompletionZSH string

//go:embed completions/run.ps
var shellCompletionPS string

func main() {
	if Version == "" {
		Version = fmt.Sprintf("nightly | %s", time.Now().Format(time.RFC3339))
	}

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
				Action: func(ctx context.Context, c *cli.Command, b bool) error {
					v := "false"
					if b {
						v = "true"
					}
					os.Setenv("RUNFILE_DEBUG", v)
					return nil
				},
			},

			&cli.BoolFlag{
				Name:    "list",
				Value:   false,
				Aliases: []string{"l"},
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

			// runfilePath, err := locateRunfile(c)
			// if err != nil {
			// 	slog.Error("locating runfile", "err", err)
			// 	panic(err)
			// }

			// generateShellCompletion(ctx, c.Root().Writer, runfilePath)
		},

		Commands: []*cli.Command{
			{
				Name:                  "shell:completion",
				Usage:                 "[shell]",
				EnableShellCompletion: false,
				Action: func(ctx context.Context, c *cli.Command) error {
					if c.NArg() == 0 {
						for _, shell := range []string{"fish", "bash", "zsh", "powershell"} {
							fmt.Fprintf(c.Writer, "%s\n", shell)
						}
						return nil
					}

					switch c.Args().First() {
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
			{
				Name:                  "init",
				EnableShellCompletion: false,
				Action: func(ctx context.Context, c *cli.Command) error {
					dir, err := os.Getwd()
					if err != nil {
						return err
					}

					_, err = getRunfilePath(dir)
					if err == nil {
						slog.Info("Runfile already exists in current directory")
						return nil
					}

					// TODO: implement init command to create a sample Runfile
					slog.Info("init command not yet implemented")
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
				// runfilePath, err := locateRunfile(c)
				// if err != nil {
				// 	slog.Error("locating runfile, got", "err", err)
				// 	return err
				// }
				// return generateShellCompletion(ctx, c.Root().Writer, runfilePath)
			}

			if c.NArg() == 0 {
				c.Command("help").Run(ctx, nil)
				return nil
			}

			kv := make(map[string]string)

			// INFO: for supporting flags that have been suffixed post arguments
			args := make([]string, 0, len(c.Args().Slice()))
			for _, arg := range c.Args().Slice() {
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

			logger := fastlog.New(fastlog.Console(), fastlog.ShowDebugLogs(debug), fastlog.WithoutTimestamp())
			slog.SetDefault(logger.Slog())

			runfilePath, err := locateRunfile(c)
			if err != nil {
				slog.Error("locating runfile, got", "err", err)
				return err
			}

			if err := runfile.RunTask(ctx, runfilePath, args[0], kv); err != nil {
				return err
			}

			return nil
		},
	}

	ctx, cf := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGTERM)
	defer cf()

	go func() {
		<-ctx.Done()
		cf()
	}()

	if err := cmd.Run(ctx, os.Args); err != nil {
		if err2, ok := err.(*errors.Error); ok {
			slog.Error("failed to run task", err2.AsKeyValues()...)
			return
		}
		slog.Error("failed to run task", "err", err)
	}
}

var ErrRunfileNotFound = fmt.Errorf("failed to locate your nearest Runfile")

func getRunfilePath(dir string) (string, error) {
	runfileNames := []string{
		"Runfile",
		"Runfile.yml",
		"Runfile.yaml",
	}

	for _, f := range runfileNames {
		stat, err := os.Stat(filepath.Join(dir, f))
		if err != nil {
			if !os.IsNotExist(err) {
				return "", err
			}
			continue
		}

		if stat.IsDir() {
			return "", fmt.Errorf("%s is a directory", filepath.Join(dir, f))
		}

		return filepath.Join(dir, f), nil
	}

	return "", ErrRunfileNotFound
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
			fp, err := getRunfilePath(dir)
			if err != nil {
				if errors.Is(err, ErrRunfileNotFound) {
					// Not found in this dir, try parent
					oldDir = dir
					dir = filepath.Dir(dir)
					continue
				}
				// Some other error
				return "", err
			}

			return fp, nil
		}

		return "", ErrRunfileNotFound
	}
}
