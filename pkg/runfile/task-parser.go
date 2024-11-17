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
)

type ParsedTask struct {
	Shell       []string            `json:"shell"`
	WorkingDir  string              `json:"workingDir"`
	Env         map[string]string   `json:"environ"`
	Interactive bool                `json:"interactive,omitempty"`
	Commands    []ParsedCommandJson `json:"commands"`
}

func evalGoTemplateCondition(tpl string) (bool, *Error) {
	t := template.New("requirement")
	t = t.Funcs(sprig.FuncMap())
	templateExpr := fmt.Sprintf(`{{ %s }}`, tpl)
	t, err := t.Parse(templateExpr)
	if err != nil {
		return false, TaskRequirementIncorrect.WithErr(err).WithMetadata("requirement", tpl)
	}
	b := new(bytes.Buffer)
	if err := t.ExecuteTemplate(b, "requirement", map[string]string{}); err != nil {
		return false, TaskRequirementIncorrect.WithErr(err).WithMetadata("requirement", tpl)
	}

	if b.String() != "true" {
		return false, TaskRequirementFailed.WithErr(fmt.Errorf("template must have evaluated to true")).WithMetadata("requirement", tpl)
	}

	return true, nil
}

func ParseTask(ctx Context, rf *Runfile, task Task) (*ParsedTask, *Error) {
	globalEnv := make(map[string]string)

	if rf.Env != nil {
		genv, err := parseEnvVars(ctx, rf.Env, EvaluationArgs{
			Shell: nil,
			Env:   nil,
		})
		if err != nil {
			return nil, err
		}
		for k, v := range genv {
			globalEnv[k] = v
		}
	}

	if rf.DotEnv != nil {
		dotEnvPaths, err := resolveDotEnvFiles(filepath.Dir(rf.attrs.RunfilePath), rf.DotEnv...)
		if err != nil {
			return nil, err
		}
		m, err := parseDotEnvFiles(dotEnvPaths...)
		if err != nil {
			return nil, err
		}
		for k, v := range m {
			globalEnv[k] = v
		}
	}

	for _, requirement := range task.Requires {
		if requirement == nil {
			continue
		}

		if requirement.Sh != nil {
			cmd := createCommand(ctx, cmdArgs{
				shell:      []string{"sh", "-c"},
				env:        ToEnviron(globalEnv),
				workingDir: filepath.Dir(rf.attrs.RunfilePath),
				cmd:        *requirement.Sh,
				stdout:     fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755)),
				stderr:     fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755)),
			})
			if err := cmd.Run(); err != nil {
				return nil, TaskRequirementFailed.WithErr(err).WithMetadata("requirement", *requirement.Sh)
			}
			continue
		}

		if requirement.GoTmpl != nil {
			if _, err := evalGoTemplateCondition(*requirement.GoTmpl); err != nil {
				return nil, err
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
		return nil, TaskWorkingDirectoryInvalid.WithErr(err2).WithMetadata("working-dir", *task.Dir)
	}

	if !fi.IsDir() {
		return nil, TaskWorkingDirectoryInvalid.WithErr(fmt.Errorf("path is not a directory")).WithMetadata("working-dir", *task.Dir)
	}

	taskDotEnvPaths, err := resolveDotEnvFiles(filepath.Dir(rf.attrs.RunfilePath), task.DotEnv...)
	if err != nil {
		return nil, err
	}

	taskDotenvVars, err := parseDotEnvFiles(taskDotEnvPaths...)
	if err != nil {
		return nil, err
	}

	// INFO: keys from task.Env will override those coming from dotenv files, when duplicated
	taskEnvVars, err := parseEnvVars(ctx, task.Env, EvaluationArgs{
		Shell: task.Shell,
		Env:   fn.MapMerge(globalEnv, taskDotenvVars),
	})
	if err != nil {
		return nil, err
	}

	commands := make([]ParsedCommandJson, 0, len(task.Commands))
	for i := range task.Commands {
		c2, err := parseCommand(rf, task.Commands[i])
		if err != nil {
			return nil, err
		}
		commands = append(commands, *c2)
	}

	return &ParsedTask{
		Shell:       task.Shell,
		WorkingDir:  *task.Dir,
		Interactive: task.Interactive,
		Env:         fn.MapMerge(globalEnv, taskDotenvVars, taskEnvVars),
		Commands:    commands,
	}, nil
}

// returns absolute paths to dotenv files
func resolveDotEnvFiles(pwd string, dotEnvFiles ...string) ([]string, *Error) {
	paths := make([]string, 0, len(dotEnvFiles))

	for _, v := range dotEnvFiles {
		dotenvPath := v
		if !filepath.IsAbs(v) {
			dotenvPath = filepath.Join(pwd, v)
		}
		fi, err := os.Stat(dotenvPath)
		if err != nil {
			return nil, DotEnvNotFound.WithErr(err).WithMetadata("dotenv", dotenvPath)
		}

		if fi.IsDir() {
			return nil, DotEnvInvalid.WithErr(fmt.Errorf("dotenv path must be a file, but it is a directory")).WithMetadata("dotenv", dotenvPath)
		}

		paths = append(paths, dotenvPath)
	}

	return paths, nil
}

func parseCommand(rf *Runfile, command any) (*ParsedCommandJson, *Error) {
	switch c := command.(type) {
	case string:
		{
			return &ParsedCommandJson{Command: c}, nil
		}
	case map[string]any:
		{
			var cj CommandJson
			b, err := json.Marshal(c)
			if err != nil {
				return nil, CommandInvalid.WithErr(err).WithMetadata("command", command)
			}

			if err := json.Unmarshal(b, &cj); err != nil {
				return nil, CommandInvalid.WithErr(err).WithMetadata("command", command)
			}

			if cj.Run == "" && cj.Command == "" {
				return nil, CommandInvalid.WithErr(fmt.Errorf("key: 'run'/'cmd', must be specified when setting command in json format")).WithMetadata("command", cj)
			}

			var pcj ParsedCommandJson
			pcj.Run = cj.Run
			pcj.Command = cj.Command

			if cj.If != nil {
				ok, _ := evalGoTemplateCondition(*cj.If)
				// if err != nil {
				// 	return nil, err
				// }
				pcj.If = &ok
			}

			// if _, ok := rf.Tasks[cj.Run]; !ok {
			// 	return nil, CommandInvalid.WithErr(fmt.Errorf("run target, not found")).WithMetadata("command", command, "run-target", cj.Run)
			// }

			return &pcj, nil
		}
	default:
		{
			return nil, CommandInvalid.WithMetadata("command", command)
		}
	}
}
