package runfile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	fn "github.com/nxtcoder17/runfile/pkg/functions"
)

type ParsedTask struct {
	Shell      []string      `json:"shell"`
	WorkingDir string        `json:"workingDir"`
	Environ    []string      `json:"environ"`
	Commands   []CommandJson `json:"commands"`
}

type ErrTask struct {
	ErrTaskNotFound
	Message string
	Err     error
}

func (e ErrTask) Error() string {
	return fmt.Sprintf("[task] %s (Runfile: %s), says %s, got %v", e.TaskName, e.RunfilePath, e.Message, e.Err)
}

func NewErrTask(runfilePath, taskName, message string, err error) ErrTask {
	return ErrTask{
		ErrTaskNotFound: ErrTaskNotFound{
			RunfilePath: runfilePath,
			TaskName:    taskName,
		},
		Message: message,
		Err:     err,
	}
}

func ParseTask(ctx context.Context, rf *Runfile, taskName string) (*ParsedTask, error) {
	newErr := func(message string, err error) ErrTask {
		return NewErrTask(rf.attrs.RunfilePath, taskName, message, err)
	}

	task, ok := rf.Tasks[taskName]
	if !ok {
		return nil, ErrTaskNotFound{TaskName: taskName, RunfilePath: rf.attrs.RunfilePath}
	}

	if task.Shell == nil {
		task.Shell = []string{"sh", "-c"}
	}

	if task.Dir == nil {
		task.Dir = fn.New(fn.Must(os.Getwd()))
	}

	fi, err := os.Stat(*task.Dir)
	if err != nil {
		return nil, newErr("invalid directory", err)
	}

	if !fi.IsDir() {
		return nil, newErr("invalid working directory", fmt.Errorf("path (%s), is not a directory", *task.Dir))
	}

	dotenvPaths, err := resolveDotEnvFiles(filepath.Dir(rf.attrs.RunfilePath), task.DotEnv...)
	if err != nil {
		return nil, newErr("while resolving dotenv paths", err)
	}

	dotenvVars, err := parseDotEnvFiles(dotenvPaths...)
	if err != nil {
		return nil, newErr("while parsing dotenv files", err)
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
		Shell:   task.Shell,
		Environ: env,
	})
	if err != nil {
		return nil, newErr("while evaluating task env vars", err)
	}

	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%v", k, v))
	}

	commands := make([]CommandJson, 0, len(task.Commands))
	for i := range task.Commands {
		c2, err := parseCommand(rf, task.Commands[i])
		if err != nil {
			return nil, newErr("while parsing command", err)
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
