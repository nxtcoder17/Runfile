package parser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nxtcoder17/runfile/errors"
	fn "github.com/nxtcoder17/runfile/functions"
	"github.com/nxtcoder17/runfile/types"
)

func isAbsPath(p string) bool {
	j, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}
	return j == p
}

func ParseTask(ctx types.Context, prf *types.ParsedRunfile, task types.Task) (*types.ParsedTask, error) {
	taskCtx := ctx
	taskCtx.TaskName = task.Name
	if task.Metadata.RunfilePath == nil {
		task.Metadata.RunfilePath = &prf.Metadata.RunfilePath
	}
	workingDir := filepath.Dir(*task.Metadata.RunfilePath)

	taskEnv := prf.Env

	if taskEnv == nil {
		taskEnv = make(map[string]string)
	}

	dotEnvs := make([]string, 0, len(task.DotEnv))
	for i := range task.DotEnv {
		de := task.DotEnv[i]
		if !filepath.IsAbs(de) {
			result := filepath.Join(filepath.Dir(*task.Metadata.RunfilePath), de)
			de = result
		}

		dotEnvs = append(dotEnvs, de)
	}

	tdotenv, err := parseDotEnvFiles(dotEnvs...)
	if err != nil {
		return nil, err
	}

	for k, v := range tdotenv {
		taskEnv[k] = v
	}

	tenv, err := parseEnvVars(taskCtx, task.Env, evaluationParams{
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
			cmd := exec.CommandContext(taskCtx, "sh", "-c", *requirement.Sh)
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
		c2, err := parseCommand(ctx, prf, taskEnv, task.Commands[i])
		if err != nil {
			return nil, err
		}
		commands = append(commands, *c2)
	}

	watch := task.Watch
	if watch != nil {
		for i := range watch.Dirs {
			if !isAbsPath(watch.Dirs[i]) {
				watch.Dirs[i] = filepath.Join(*task.Dir, watch.Dirs[i])
			}
		}
	}

	return &types.ParsedTask{
		Name:        task.Name,
		Shell:       task.Shell,
		WorkingDir:  *task.Dir,
		Interactive: task.Interactive,
		Env:         taskEnv,
		Commands:    commands,
		Watch:       watch,
		Parallel:    task.Parallel,
	}, nil
}
