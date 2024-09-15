package runfile

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type runArgs struct {
	shell []string
	env   map[string]string
	cmd   string

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

	environ := os.Environ()
	for k, v := range args.env {
		environ = append(environ, fmt.Sprintf("%s=%v", k, v))
	}

	// cargs := append(args.shell[1:], f.Name())
	cargs := append(args.shell[1:], args.cmd)
	c := exec.CommandContext(ctx, shell, cargs...)
	c.Env = environ
	c.Stdout = args.stdout
	c.Stderr = args.stderr
	return c.Run()
}

func (r *RunFile) Run(ctx context.Context, taskName string) error {
	task, ok := r.Tasks[taskName]
	if !ok {
		return fmt.Errorf("task %s not found", taskName)
	}

	env := make(map[string]string, len(task.Env))
	for k, v := range task.Env {
		switch v := v.(type) {
		case string:
			env[k] = v
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
				env:    env,
				cmd:    s,
				stdout: value,
			}); err != nil {
				return err
			}
			env[k] = value.String()
		default:
			panic(fmt.Sprintf("env %s is not a string (%T)", k, v))
		}
	}

	for _, cmd := range task.Commands {
		runInShell(ctx, runArgs{
			shell: task.Shell,
			env:   env,
			cmd:   cmd,
		})
	}
	return nil
}
