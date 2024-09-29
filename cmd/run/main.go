package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/nxtcoder17/fwatcher/pkg/logging"
	"github.com/nxtcoder17/runfile/pkg/runfile"
	"github.com/urfave/cli/v3"
)

var Version string = "0.0.1"

func main() {
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
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			parallel := c.Bool("parallel")
			watch := c.Bool("watch")
			debug := c.Bool("debug")

			logging.NewSlogLogger(logging.SlogOptions{
				ShowCaller:         debug,
				ShowDebugLogs:      debug,
				SetAsDefaultLogger: true,
			})

			if c.Args().Len() < 1 {
				return fmt.Errorf("missing argument, at least one argument is required")
			}

			runfilePath, err := locateRunfile(c)
			if err != nil {
				return err
			}

			rf, err := runfile.Parse(runfilePath)
			if err != nil {
				panic(err)
			}

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

				args = append(args, arg)
			}

			if parallel && watch {
				return fmt.Errorf("parallel and watch can't be set together")
			}

			return rf.Run(ctx, runfile.RunArgs{
				Tasks:             args,
				ExecuteInParallel: parallel,
				Watch:             watch,
				Debug:             debug,
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
		log.Fatal(err)
	}
}

func locateRunfile(c *cli.Command) (string, error) {
	var runfilePath string
	switch {
	case c.IsSet("file"):
		runfilePath = c.String("file")
	default:
		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		for {
			_, err := os.Stat(filepath.Join(dir, "Runfile"))
			if err != nil {
				dir = filepath.Dir(dir)
				continue
			}
			runfilePath = filepath.Join(dir, "Runfile")
			break
		}
	}
	return runfilePath, nil
}
