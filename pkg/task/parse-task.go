package task

import (
	"maps"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nxtcoder17/runfile/pkg/errors"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
)

func (t *Task) ParseEnv(ctx *Context) (map[string]string, error) {
	// pt := ParsedTask{
	// 	Name:  t.Name,
	// 	Shell: t.Shell,
	// 	Dir: func() string {
	// 		if t.Dir == nil {
	// 			return fn.Must(os.Getwd())
	// 		}
	// 		return *t.Dir
	// 	}(),
	// 	Watch:       t.Watch,
	// 	Interactive: t.Interactive,
	// 	Parallel:    t.Parallel,
	//
	// 	Env:      make(map[string]string),
	// 	Commands: nil,
	// }

	env := make(map[string]string)

	maps.Copy(env, t.ParentEnv)

	dotEnvs := make([]string, 0, len(t.DotEnv))
	for i := range t.DotEnv {
		de := t.DotEnv[i]
		if !filepath.IsAbs(de) {
			result := filepath.Join(filepath.Dir(*t.Metadata.RunfilePath), de)
			de = result
		}

		dotEnvs = append(dotEnvs, de)
	}

	tdotenv, err := ParseDotEnvFiles(dotEnvs...)
	if err != nil {
		return nil, err
	}

	maps.Copy(env, tdotenv)

	tenv, err := ParseEnvVars(ctx.Context, t.Env, t.ParentEnv)
	if err != nil {
		return nil, err
	}

	maps.Copy(env, tenv)

	for _, requirement := range t.Requires {
		if requirement == nil {
			continue
		}

		if requirement.Sh != nil {
			cmd := exec.CommandContext(ctx.Context, "sh", "-c", *requirement.Sh)
			cmd.Env = fn.ToEnviron(env)
			cmd.Stdout = fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755))
			cmd.Stderr = fn.Must(os.OpenFile(os.DevNull, os.O_WRONLY, 0o755))
			cmd.Dir = func() string {
				if t.Dir != nil {
					return *t.Dir
				}
				return "."
			}()

			if err := cmd.Run(); err != nil {
				return nil, errors.ErrTaskRequirementNotMet(*requirement.Sh, err)
			}
			continue
		}
	}

	return env, nil
}
