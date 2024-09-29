package runfile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	sprig "github.com/go-task/slim-sprig/v3"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"github.com/nxtcoder17/runfile/pkg/runfile/errors"
)

type ParsedTask struct {
	Shell      []string      `json:"shell"`
	WorkingDir string        `json:"workingDir"`
	Environ    []string      `json:"environ"`
	Commands   []CommandJson `json:"commands"`
}

func ParseTask(ctx context.Context, rf *Runfile, taskName string) (*ParsedTask, error) {
	errctx := errors.Context{Task: taskName, Runfile: rf.attrs.RunfilePath}

	task, ok := rf.Tasks[taskName]
	if !ok {
		return nil, errors.TaskNotFound{Context: errctx}
	}

	for _, requirement := range task.Requires {
		if requirement == nil {
			continue
		}

		if requirement.Sh != nil {
			cmd := createCommand(ctx, cmdArgs{
				shell:      []string{"sh", "-c"},
				env:        os.Environ(),
				workingDir: filepath.Dir(rf.attrs.RunfilePath),
				cmd:        *requirement.Sh,
				stdout:     fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755)),
				stderr:     fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755)),
			})
			if err := cmd.Run(); err != nil {
				return nil, errors.ErrTaskFailedRequirements{Context: errctx.WithErr(err), Requirement: *requirement.Sh}
			}
			continue
		}

		if requirement.GoTmpl != nil {
			t := template.New("requirement")
			t = t.Funcs(sprig.FuncMap())
			templateExpr := fmt.Sprintf(`{{ %s }}`, *requirement.GoTmpl)
			t, err := t.Parse(templateExpr)
			if err != nil {
				return nil, err
			}
			b := new(bytes.Buffer)
			if err := t.ExecuteTemplate(b, "requirement", map[string]string{}); err != nil {
				return nil, err
			}

			if b.String() != "true" {
				return nil, errors.ErrTaskFailedRequirements{Context: errctx.WithErr(fmt.Errorf("`%s` evaluated to `%s` (wanted: `true`)", templateExpr, b.String())), Requirement: *requirement.GoTmpl}
			}

			continue
		}
	}

	if task.Shell == nil {
		task.Shell = []string{"sh", "-c"}
	}

	if task.Dir == nil {
		task.Dir = fn.New(fn.Must(os.Getwd()))
	}

	fi, err := os.Stat(*task.Dir)
	if err != nil {
		return nil, errors.InvalidWorkingDirectory{Context: errctx.WithErr(err)}
	}

	if !fi.IsDir() {
		return nil, errors.InvalidWorkingDirectory{Context: errctx.WithErr(fmt.Errorf("path (%s), is not a directory", *task.Dir))}
	}

	dotenvPaths, err := resolveDotEnvFiles(filepath.Dir(rf.attrs.RunfilePath), task.DotEnv...)
	if err != nil {
		return nil, errors.InvalidDotEnv{Context: errctx.WithErr(err).WithMessage("failed to resolve dotenv paths")}
	}

	dotenvVars, err := parseDotEnvFiles(dotenvPaths...)
	if err != nil {
		return nil, errors.InvalidDotEnv{Context: errctx.WithErr(err).WithMessage("failed to parse dotenv files")}
	}

	env := make([]string, 0, len(os.Environ())+len(dotenvVars))
	if !task.ignoreSystemEnv {
		env = append(env, os.Environ()...)
	}
	for k, v := range dotenvVars {
		env = append(env, fmt.Sprintf("%s=%v", k, v))
	}

	// INFO: keys from task.Env will override those coming from dotenv files, when duplicated
	envVars, err := parseEnvVars(ctx, task.Env, EvaluationArgs{
		Shell: task.Shell,
		Env:   dotenvVars,
	})
	if err != nil {
		return nil, errors.InvalidEnvVar{Context: errctx.WithErr(err).WithMessage("failed to parse/evaluate env vars")}
	}

	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%v", k, v))
	}

	commands := make([]CommandJson, 0, len(task.Commands))
	for i := range task.Commands {
		c2, err := parseCommand(rf, task.Commands[i])
		if err != nil {
			return nil, errors.IncorrectCommand{Context: errctx.WithErr(err).WithMessage(fmt.Sprintf("failed to parse command: %+v", task.Commands[i]))}
		}
		commands = append(commands, *c2)
	}

	return &ParsedTask{
		Shell:      task.Shell,
		WorkingDir: *task.Dir,
		Environ:    env,
		Commands:   commands,
	}, nil
}

// returns absolute paths to dotenv files
func resolveDotEnvFiles(pwd string, dotEnvFiles ...string) ([]string, error) {
	paths := make([]string, 0, len(dotEnvFiles))

	for _, v := range dotEnvFiles {
		dotenvPath := v
		if !filepath.IsAbs(v) {
			dotenvPath = filepath.Join(pwd, v)
		}
		fi, err := os.Stat(dotenvPath)
		if err != nil {
			return nil, err
		}

		if fi.IsDir() {
			return nil, fmt.Errorf("dotenv file must be a file, but %s is a directory", v)
		}

		paths = append(paths, dotenvPath)
	}

	return paths, nil
}

func parseCommand(rf *Runfile, command any) (*CommandJson, error) {
	switch c := command.(type) {
	case string:
		{
			return &CommandJson{
				Command: c,
			}, nil
		}
	case map[string]any:
		{
			var cj CommandJson
			b, err := json.Marshal(c)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(b, &cj); err != nil {
				return nil, err
			}

			if cj.Run == "" {
				return nil, fmt.Errorf("key: 'run', must be specified when setting command in json format")
			}

			if _, ok := rf.Tasks[cj.Run]; !ok {
				return nil, fmt.Errorf("[run target]: %s, not found in Runfile (%s)", cj.Run, rf.attrs.RunfilePath)
			}

			return &cj, nil
		}
	default:
		{
			return nil, fmt.Errorf("invalid command")
		}
	}
}
