package runfile

import (
	"os"
	"path/filepath"

	"github.com/nxtcoder17/runfile/pkg/errors"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"github.com/nxtcoder17/runfile/pkg/types"
	"sigs.k8s.io/yaml"
)

func ParseFromFile(ctx *types.Context, file string) (*ParsedRunfile, error) {
	var rf Runfile
	f, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.ErrReadRunfile(err).KV("file", file)
	}

	if err := yaml.Unmarshal(f, &rf); err != nil {
		return nil, errors.ErrParseRunfile(err).Msg("failed to unmarshal YAML into types.Runfile")
	}

	rf.Filepath = fn.Must(filepath.Abs(file))

	env, err := rf.resolveEnv(ctx)
	if err != nil {
		return nil, err
	}

	includedTasks, err := rf.resolveIncludedTasks(ctx)
	if err != nil {
		return nil, err
	}

	tasks := make(map[string]Task, len(rf.Tasks))
	for k, v := range rf.Tasks {
		v.Name = k
		v.ParentEnv = env
		v.Metadata.Namespace = ""
		tasks[k] = v
	}

	return &ParsedRunfile{
		Env:   env,
		Tasks: fn.MapMerge(tasks, includedTasks),
	}, nil
}

func (rf *Runfile) resolveEnv(ctx *types.Context) (map[string]string, error) {
	dotEnvFiles := make([]string, 0, len(rf.DotEnv))
	for i := range rf.DotEnv {
		de := rf.DotEnv[i]
		if !filepath.IsAbs(de) {
			de = filepath.Join(filepath.Dir(rf.Filepath), de)
		}
		dotEnvFiles = append(dotEnvFiles, de)
	}

	dotenvVars, err := ParseDotEnvFiles(dotEnvFiles...)
	if err != nil {
		return nil, err
	}

	envVars, err := ParseEnvVars(ctx, rf.Env, dotenvVars)
	if err != nil {
		return nil, err
	}

	return fn.MapMerge(dotenvVars, envVars), nil
}

func (rf *Runfile) resolveIncludedTasks(ctx *types.Context) (map[string]Task, error) {
	tasks := make(map[string]Task)

	for k, v := range rf.Includes {
		r, err := ParseFromFile(ctx, v.Runfile)
		if err != nil {
			return nil, errors.ErrParseIncludes(err).KV("include", v.Runfile)
		}

		for name, target := range r.Tasks {
			if v.Dir != "" {
				if target.Dir != nil {
					target.Dir = fn.Ptr(filepath.Join(v.Dir, *target.Dir))
				}
			}

			target.ParentEnv = r.Env
			target.Metadata.Namespace = k

			taskName := k + ":" + name
			target.Name = taskName
			tasks[taskName] = target
		}
	}

	return tasks, nil
}
