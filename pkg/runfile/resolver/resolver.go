package resolver

import (
	"context"
	"io"
	"maps"
	"os"
	"path/filepath"

	"github.com/nxtcoder17/go.errors"
	"github.com/nxtcoder17/runfile/pkg/runfile/spec"
	"gopkg.in/yaml.v3"
)

type extendedTaskSpec struct {
	spec.TaskSpec

	IsImported bool
	// ImportPrefix needs to be set only if IsImported is true
	ImportPrefix string
}

type Resolver struct {
	Env   map[string]string
	Tasks map[string]extendedTaskSpec
}

func loadRunfile(file string) (*spec.RunfileSpec, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.New("failed to open file").Wrap(err).KV("filepath", file)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.New("failed to read file").Wrap(err).KV("filepath", file)
	}

	var rf spec.RunfileSpec
	if err := yaml.Unmarshal(b, &rf); err != nil {
		return nil, errors.New("failed to unmarshal file content into Runfile Spec").Wrap(err).KV("filepath", file)
	}

	return &rf, nil
}

func Load(ctx context.Context, file string) (*Resolver, error) {
	rf, err := loadRunfile(file)
	if err != nil {
		return nil, err
	}

	runfileDir := filepath.Dir(file)

	dotenv := make([]string, 0, len(rf.DotEnv))
	for _, f := range rf.DotEnv {
		if !filepath.IsAbs(f) {
			f = filepath.Join(runfileDir, f)
		}
		dotenv = append(dotenv, f)
	}

	env := rf.Env

	tasks := make(map[string]extendedTaskSpec, len(rf.Tasks))

	for k, task := range rf.Tasks {
		if task.Dir == "" {
			task.Dir = "."
		}

		task.Dir = filepath.Join(runfileDir, task.Dir)
		for i := range task.DotEnv {
			if !filepath.IsAbs(task.DotEnv[i]) {
				task.DotEnv[i] = filepath.Join(runfileDir, task.DotEnv[i])
			}
		}
		tasks[k] = extendedTaskSpec{TaskSpec: task, IsImported: false}
	}

	for includeKey, includeSpec := range rf.Includes {
		includedFile := includeSpec.Runfile
		if !filepath.IsAbs(includedFile) {
			includedFile = filepath.Join(runfileDir, includedFile)
		}

		includedDir := filepath.Dir(includedFile)

		included, err := loadRunfile(includedFile)
		if err != nil {
			return nil, err
		}

		for _, f := range included.DotEnv {
			if !filepath.IsAbs(f) {
				f = filepath.Join(includedDir, f)
			}
			dotenv = append(dotenv, f)
		}

		maps.Copy(env, included.Env)

		for tn, task := range included.Tasks {
			if task.Dir == "" {
				task.Dir = "."
			}

			for i := range task.DotEnv {
				if !filepath.IsAbs(task.DotEnv[i]) {
					task.DotEnv[i] = filepath.Join(runfileDir, task.DotEnv[i])
				}
			}

			task.Dir = filepath.Join(includedDir, task.Dir)
			tasks[includeKey+":"+tn] = extendedTaskSpec{
				TaskSpec:     task,
				IsImported:   true,
				ImportPrefix: includeKey,
			}
		}
	}

	envStore := make(map[string]string)
	if err := parseDotEnvFilesInto(envStore, dotenv); err != nil {
		return nil, err
	}
	if err := parseEnvInto(ctx, envStore, env); err != nil {
		return nil, err
	}

	return &Resolver{
		Env:   envStore,
		Tasks: tasks,
	}, nil
}
