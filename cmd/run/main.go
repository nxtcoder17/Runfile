package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/nxtcoder17/fwatcher/pkg/logging"
	"github.com/nxtcoder17/runfile/pkg/runfile"
	"github.com/urfave/cli/v3"
)

var (
	Version      string   = "0.0.1"
	runfileNames []string = []string{
		"Runfile",
		"Runfile.yml",
		"Runfile.yaml",
	}
)

func main() {
	logger := logging.NewSlogLogger(logging.SlogOptions{})

	cmd := cli.Command{
		Name:        "run",
		Version:     Version,
		Description: "A simple task runner",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "",
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
		},
		EnableShellCompletion: true,
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			if c.NArg() > 0 {
				return
			}

			runfilePath, err := locateRunfile(c)
			if err != nil {
				panic(err)
			}

			runfile, err := runfile.Parse(runfilePath)
			if err != nil {
				panic(err)
			}

			for k := range runfile.Tasks {
				fmt.Fprintf(c.Root().Writer, "%s\n", k)
			}

			m, err := runfile.ParseIncludes()
			if err != nil {
				panic(err)
			}

			for k, v := range m {
				for tn := range v.Runfile.Tasks {
					fmt.Fprintf(c.Root().Writer, "%s:%s\n", k, tn)
				}
			}
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			parallel := c.Bool("parallel")
			watch := c.Bool("watch")
			debug := c.Bool("debug")

			if c.NArg() == 0 {
				c.Command("help").Run(ctx, nil)
				return nil
			}

			runfilePath, err := locateRunfile(c)
			if err != nil {
				return err
			}

			rf, err := runfile.Parse(runfilePath)
			if err != nil {
				panic(err)
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

			logger := logging.NewSlogLogger(logging.SlogOptions{
				ShowCaller:         debug,
				ShowDebugLogs:      debug,
				SetAsDefaultLogger: true,
			})

			return rf.Run(runfile.NewContext(ctx, logger), runfile.RunArgs{
				Tasks:             args,
				ExecuteInParallel: parallel,
				Watch:             watch,
				Debug:             debug,
				KVs:               kv,
			})
		},
	}

	ctx, cf := context.WithCancel(context.TODO())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n\rcanceling...")
		cf()
		os.Exit(1)
	}()

	if err := cmd.Run(ctx, os.Args); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
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
