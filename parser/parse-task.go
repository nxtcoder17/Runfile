package parser

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nxtcoder17/runfile/errors"
	fn "github.com/nxtcoder17/runfile/functions"
	"github.com/nxtcoder17/runfile/types"
)

func ParseTask(ctx context.Context, prf *types.ParsedRunfile, task types.Task) (*types.ParsedTask, error) {
	workingDir := filepath.Dir(prf.Metadata.RunfilePath)
	if task.Metadata.RunfilePath != nil {
		workingDir = filepath.Dir(*task.Metadata.RunfilePath)
	}

	taskEnv := prf.Env

	if taskEnv == nil {
		taskEnv = make(map[string]string)
	}

	tdotenv, err := parseDotEnvFiles(task.DotEnv...)
	if err != nil {
		return nil, err
	}

	for k, v := range tdotenv {
		taskEnv[k] = v
	}

	tenv, err := parseEnvVars(ctx, task.Env, evaluationParams{
		Env: prf.Env,
	})
	if err != nil {
		return nil, errors.WithErr(err)
	}

	for k, v := range tenv {
		taskEnv[k] = v
	}

	for _, requirement := range task.Requires {
		if requirement == nil {
			continue
		}

		if requirement.Sh != nil {
			cmd := exec.CommandContext(ctx, "sh", "-c", *requirement.Sh)
			cmd.Env = fn.ToEnviron(taskEnv)
			cmd.Stdout = fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755))
			cmd.Stderr = fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755))
			cmd.Dir = workingDir
			if err := cmd.Run(); err != nil {
				return nil, errors.ErrTaskRequirementNotMet.Wrap(err).KV("requirement", *requirement.Sh)
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
		return nil, errors.ErrTaskInvalidWorkingDir.Wrap(err).KV("working-dir", *task.Dir)
	}

	if !fi.IsDir() {
		return nil, errors.ErrTaskInvalidWorkingDir.Wrap(fmt.Errorf("path is not a directory")).KV("working-dir", *task.Dir)
	}

	commands := make([]types.ParsedCommandJson, 0, len(task.Commands))
	for i := range task.Commands {
		c2, err := parseCommand(prf, task.Commands[i])
		if err != nil {
			return nil, err
		}
		commands = append(commands, *c2)
	}

	return &types.ParsedTask{
		Shell:       task.Shell,
		WorkingDir:  *task.Dir,
		Interactive: task.Interactive,
		Env:         taskEnv,
		Commands:    commands,
		Watch:       task.Watch,
	}, nil
}
