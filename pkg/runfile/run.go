package runfile

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	fn "github.com/nxtcoder17/runfile/pkg/functions"
)

type runArgs struct {
	shell      []string
	env        []string // [key=value, key=value, ...]
	workingDir string

	cmd string

	stdout io.Writer
	stderr io.Writer
}

func runInShell(ctx context.Context, args runArgs) error {
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
	return c.Run()
}

func (r *RunFile) Run(ctx context.Context, taskName string) error {
	task, ok := r.Tasks[taskName]
	if !ok {
		return fmt.Errorf("task %s not found", taskName)
	}

	env := make([]string, len(task.Env))
	for k, v := range task.Env {
		switch v := v.(type) {
		case string:
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		case map[string]any:
			shcmd, ok := v["sh"]
			if !ok {
				return fmt.Errorf("env %s is not a string", k)
			}

			s, ok := shcmd.(string)
			if !ok {
				return fmt.Errorf("shell cmd is not a string")
			}

			value := new(bytes.Buffer)

			if err := runInShell(ctx, runArgs{
				shell:  task.Shell,
				env:    os.Environ(),
				cmd:    s,
				stdout: value,
			}); err != nil {
				return err
			}
			env = append(env, fmt.Sprintf("%s=%v", k, value.String()))
		default:
			panic(fmt.Sprintf("env %s is not a string (%T)", k, v))
		}
	}

	dotenvPaths := make([]string, len(task.DotEnv))
	for i, v := range task.DotEnv {
		dotenvPath := filepath.Join(filepath.Dir(r.attrs.RunfilePath), v)
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
	s, err := parseDotEnv(dotenvPaths...)
	if err != nil {
		return err
	}

	// INFO: keys from task.Env will override those coming from dotenv files, when duplicated
	env = append(s, env...)

	for _, cmd := range task.Commands {
		runInShell(ctx, runArgs{
			shell:      task.Shell,
			env:        append(os.Environ(), env...),
			cmd:        cmd,
			workingDir: fn.DefaultIfNil(task.Dir, fn.Must(os.Getwd())),
		})
	}
	return nil
}
