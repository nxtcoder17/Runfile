package runfile

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"golang.org/x/sync/errgroup"
)

type cmdArgs struct {
	shell      []string
	env        []string // [key=value, key=value, ...]
	workingDir string

	cmd string

	stdout io.Writer
	stderr io.Writer
}

func createCommand(ctx context.Context, args cmdArgs) *exec.Cmd {
	if args.shell == nil {
		args.shell = []string{"sh", "-c"}
	}

	if args.stdout == nil {
		args.stdout = os.Stdout
	}

	if args.stderr == nil {
		args.stderr = os.Stderr
	}

	shell := args.shell[0]
	// f, err := os.CreateTemp(os.TempDir(), "runfile-")
	// if err != nil {
	// 	return err
	// }
	// f.WriteString(args.cmd)
	// f.Close()

	// cargs := append(args.shell[1:], f.Name())
	cargs := append(args.shell[1:], args.cmd)
	c := exec.CommandContext(ctx, shell, cargs...)
	c.Dir = args.workingDir
	c.Env = args.env
	c.Stdout = args.stdout
	c.Stderr = args.stderr
	return c
}

func (rf *Runfile) runTask(ctx context.Context, task Task) error {
	shell := task.Shell
	if shell == nil {
		shell = []string{"sh", "-c"}
	}

	dotenvPaths := make([]string, len(task.DotEnv))
	for i, v := range task.DotEnv {
		dotenvPath := filepath.Join(filepath.Dir(rf.attrs.RunfilePath), v)
		fi, err := os.Stat(dotenvPath)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return fmt.Errorf("dotenv file must be a file, but %s is a directory", v)
		}

		dotenvPaths[i] = dotenvPath
	}

	// parsing dotenv
	dotEnvVars, err := parseDotEnvFiles(dotenvPaths...)
	if err != nil {
		return err
	}

	env := make([]string, 0, len(os.Environ())+len(dotEnvVars))
	env = append(env, os.Environ()...)
	for k, v := range dotEnvVars {
		env = append(env, fmt.Sprintf("%s=%v", k, v))
	}

	// INFO: keys from task.Env will override those coming from dotenv files, when duplicated
	envVars, err := parseEnvVars(ctx, task.Env, EvaluationArgs{
		Shell:   task.Shell,
		Environ: env,
	})
	if err != nil {
		return err
	}

	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%v", k, v))
	}

	script := make([]string, 0, len(task.Commands))

	for _, command := range task.Commands {
		script = append(script, command)
	}

	cmd := createCommand(ctx, cmdArgs{
		shell:      task.Shell,
		env:        env,
		cmd:        strings.Join(script, "\n"),
		workingDir: fn.DefaultIfNil(task.Dir, fn.Must(os.Getwd())),
	})

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

type RunArgs struct {
	Tasks             []string
	ExecuteInParallel bool
	Watch             bool
	Debug             bool
}

func (rf *Runfile) Run(ctx context.Context, args RunArgs) error {
	for _, v := range args.Tasks {
		if _, ok := rf.Tasks[v]; !ok {
			return ErrTaskNotFound{TaskName: v, RunfilePath: rf.attrs.RunfilePath}
		}
	}

	if args.ExecuteInParallel {
		slog.Default().Debug("running in parallel mode", "tasks", args.Tasks)
		g := new(errgroup.Group)

		for _, tn := range args.Tasks {
			g.Go(func() error {
				return rf.runTask(ctx, rf.Tasks[tn])
			})
		}

		// Wait for all tasks to finish
		if err := g.Wait(); err != nil {
			return err
		}

		return nil
	}

	for _, tn := range args.Tasks {
		if err := rf.runTask(ctx, rf.Tasks[tn]); err != nil {
			return err
		}
	}

	return nil
}
