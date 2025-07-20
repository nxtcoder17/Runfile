package runfile

import (
	"maps"
	"os"
	"path/filepath"

	"github.com/nxtcoder17/runfile/pkg/errors"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
)

func (r *ParsedRunfile) ParseTaskEnv(ctx *Context, taskName string, parentEnv map[string]string) (map[string]string, error) {
	task, ok := r.Tasks[taskName]
	if !ok {
		return nil, errors.ErrTaskNotFound(taskName)
	}

	env := make(map[string]string)
	maps.Copy(env, parentEnv)

	maps.Copy(env, r.Env)

	dotEnvs := make([]string, 0, len(task.DotEnv))
	for i := range task.DotEnv {
		de := task.DotEnv[i]
		if !filepath.IsAbs(de) {
			result := filepath.Join(filepath.Dir(*task.Metadata.RunfilePath), de)
			de = result
		}

		dotEnvs = append(dotEnvs, de)
	}

	tdotenv, err := ParseDotEnvFiles(dotEnvs...)
	if err != nil {
		return nil, err
	}

	maps.Copy(env, tdotenv)

	tenv, err := ParseEnvVars(ctx.Context, task.Env, env)
	if err != nil {
		return nil, err
	}

	maps.Copy(env, tenv)

	for _, requirement := range task.Requires {
		if requirement != nil && requirement.Sh != nil {
			cmd := CreateCommand(ctx, CmdArgs{
				Shell:       shellAliasMap["sh"],
				Env:         fn.ToEnviron(env),
				WorkingDir:  task.Dir,
				Cmd:         *requirement.Sh,
				interactive: task.Interactive,
				Stdout:      fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755)),
				Stderr:      fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755)),
			})

			if err := cmd.Run(); err != nil {
				return nil, errors.ErrTaskRequirementNotMet(*requirement.Sh, err)
			}
		}
	}

	return env, nil
}
