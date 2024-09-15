package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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

			runfile, err := runfile.ParseRunFile(runfilePath)
			if err != nil {
				panic(err)
			}

			for k := range runfile.Tasks {
				fmt.Fprintf(c.Root().Writer, "%s\n", k)
			}
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() > 1 {
				return fmt.Errorf("too many arguments")
			}
			if c.Args().Len() != 1 {
				return fmt.Errorf("missing argument")
			}

			runfilePath, err := locateRunfile(c)
			if err != nil {
				return err
			}

			runfile, err := runfile.ParseRunFile(runfilePath)
			if err != nil {
				panic(err)
			}

			s := c.Args().First()
			return runfile.Run(ctx, s)
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
