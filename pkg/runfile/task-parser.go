package runfile

import (
	"bytes"
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
	Shell      []string          `json:"shell"`
	WorkingDir string            `json:"workingDir"`
	Env        map[string]string `json:"environ"`
	Commands   []CommandJson     `json:"commands"`
}

// func ParseTask(ctx Context, rf *Runfile, taskName string) (*ParsedTask, error) {
func ParseTask(ctx Context, rf *Runfile, task Task) (*ParsedTask, errors.Message) {
	attrs := []any{"task", task.Name, "runfile", rf.attrs.RunfilePath}
	for _, requirement := range task.Requires {
		if requirement == nil {
			continue
		}

		if requirement.Sh != nil {
			cmd := createCommand(ctx, cmdArgs{
				shell:      []string{"sh", "-c"},
				env:        nil,
				workingDir: filepath.Dir(rf.attrs.RunfilePath),
				cmd:        *requirement.Sh,
				stdout:     fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755)),
				stderr:     fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755)),
			})
			if err := cmd.Run(); err != nil {
				return nil, errors.TaskRequirementFailed.WithErr(err).WithMetadata("requirement", *requirement.Sh).WithMetadata(attrs)
			}
			continue
		}

		if requirement.GoTmpl != nil {
			t := template.New("requirement")
			t = t.Funcs(sprig.FuncMap())
			templateExpr := fmt.Sprintf(`{{ %s }}`, *requirement.GoTmpl)
			t, err := t.Parse(templateExpr)
			if err != nil {
				return nil, errors.TaskRequirementIncorrect.WithErr(err).WithMetadata("requirement", *requirement.GoTmpl).WithMetadata(attrs)
			}
			b := new(bytes.Buffer)
			if err := t.ExecuteTemplate(b, "requirement", map[string]string{}); err != nil {
				return nil, errors.TaskRequirementIncorrect.WithErr(err).WithMetadata("requirement", *requirement.GoTmpl).WithMetadata(attrs)
			}

			if b.String() != "true" {
				return nil, errors.TaskRequirementFailed.WithErr(fmt.Errorf("template must have evaluated to true")).WithMetadata("requirement", *requirement.GoTmpl).WithMetadata(attrs)
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

	fi, err2 := os.Stat(*task.Dir)
	if err2 != nil {
		return nil, errors.TaskWorkingDirectoryInvalid.WithErr(err2).WithMetadata("working-dir", *task.Dir).WithMetadata(attrs)
	}

	if !fi.IsDir() {
		return nil, errors.TaskWorkingDirectoryInvalid.WithErr(fmt.Errorf("path is not a directory")).WithMetadata("working-dir", *task.Dir).WithMetadata(attrs)
	}

	dotenvPaths, err := resolveDotEnvFiles(filepath.Dir(rf.attrs.RunfilePath), task.DotEnv...)
	if err != nil {
		return nil, err.WithMetadata(attrs)
	}

	dotenvVars, err := parseDotEnvFiles(dotenvPaths...)
	if err != nil {
		return nil, err.WithMetadata(attrs)
	}

	// INFO: keys from task.Env will override those coming from dotenv files, when duplicated
	envVars, err := parseEnvVars(ctx, task.Env, EvaluationArgs{
		Shell: task.Shell,
		Env:   dotenvVars,
	})
	if err != nil {
		return nil, err.WithMetadata(attrs)
	}

	commands := make([]CommandJson, 0, len(task.Commands))
	for i := range task.Commands {
		c2, err := parseCommand(rf, task.Commands[i])
		if err != nil {
			return nil, err.WithMetadata(attrs)
		}
		commands = append(commands, *c2)
	}

	return &ParsedTask{
		Shell:      task.Shell,
		WorkingDir: *task.Dir,
		Env:        fn.MapMerge(dotenvVars, envVars),
		Commands:   commands,
	}, nil
}

// returns absolute paths to dotenv files
func resolveDotEnvFiles(pwd string, dotEnvFiles ...string) ([]string, errors.Message) {
	paths := make([]string, 0, len(dotEnvFiles))

	for _, v := range dotEnvFiles {
		dotenvPath := v
		if !filepath.IsAbs(v) {
			dotenvPath = filepath.Join(pwd, v)
		}
		fi, err := os.Stat(dotenvPath)
		if err != nil {
			return nil, errors.DotEnvNotFound.WithErr(err).WithMetadata("dotenv", dotenvPath)
		}

		if fi.IsDir() {
			return nil, errors.DotEnvInvalid.WithErr(fmt.Errorf("dotenv path must be a file, but it is a directory")).WithMetadata("dotenv", dotenvPath)
		}

		paths = append(paths, dotenvPath)
	}

	return paths, nil
}

func parseCommand(rf *Runfile, command any) (*CommandJson, errors.Message) {
	switch c := command.(type) {
	case string:
		{
			return &CommandJson{Command: c}, nil
		}
	case map[string]any:
		{
			var cj CommandJson
			b, err := json.Marshal(c)
			if err != nil {
				return nil, errors.CommandInvalid.WithErr(err).WithMetadata("command", command)
			}

			if err := json.Unmarshal(b, &cj); err != nil {
				return nil, errors.CommandInvalid.WithErr(err).WithMetadata("command", command)
			}

			if cj.Run == "" {
				return nil, errors.CommandInvalid.WithErr(fmt.Errorf("key: 'run', must be specified when setting command in json format")).WithMetadata("command", command)
			}

			if _, ok := rf.Tasks[cj.Run]; !ok {
				return nil, errors.CommandInvalid.WithErr(fmt.Errorf("run target, not found")).WithMetadata("command", command, "run-target", cj.Run)
			}

			return &cj, nil
		}
	default:
		{
			return nil, errors.CommandInvalid.WithMetadata("command", command)
		}
	}
}
