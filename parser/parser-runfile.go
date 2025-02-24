package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nxtcoder17/go.pkgs/log"
	"github.com/nxtcoder17/runfile/errors"
	fn "github.com/nxtcoder17/runfile/functions"
	"github.com/nxtcoder17/runfile/types"
	"sigs.k8s.io/yaml"
)

func parseRunfile(ctx types.Context, runfile *types.Runfile) (*types.ParsedRunfile, error) {
	prf := &types.ParsedRunfile{
		Env:   make(map[string]string),
		Tasks: make(map[string]types.Task),
	}

	for k, task := range runfile.Tasks {
		task.Name = k
		prf.Tasks[k] = task
	}

	m, err := parseIncludes(ctx, runfile.Includes)
	if err != nil {
		return nil, err
	}

	for k, iprf := range m {
		for taskName, task := range iprf.Tasks {
			task.Name = k
			task.Metadata.RunfilePath = &iprf.Metadata.RunfilePath
			prf.Tasks[fmt.Sprintf("%s:%s", k, taskName)] = task
		}

		for k, v := range iprf.Env {
			prf.Env[k] = v
		}
	}

	dotEnvFiles := make([]string, 0, len(runfile.DotEnv))
	for i := range runfile.DotEnv {
		de := runfile.DotEnv[i]
		if !filepath.IsAbs(de) {
			result := filepath.Join(filepath.Dir(runfile.Filepath), de)
			// fmt.Println("HERE", "runfilepath", prf.Metadata.RunfilePath, "dotenv", de, "result", result)
			de = result
		}
		dotEnvFiles = append(dotEnvFiles, de)
	}

	// dotenvVars, err := parseDotEnvFiles(runfile.DotEnv...)
	dotenvVars, err := parseDotEnvFiles(dotEnvFiles...)
	if err != nil {
		return nil, err
	}

	envVars, err := parseEnvVars(types.Context{Context: context.TODO(), Logger: log.New()}, runfile.Env, evaluationParams{Env: dotenvVars})
	if err != nil {
		return nil, err
	}

	prf.Env = fn.MapMerge(prf.Env, dotenvVars, envVars)

	return prf, nil
}

func parseRunfileFromFile(ctx types.Context, file string) (*types.ParsedRunfile, error) {
	var runfile types.Runfile

	f, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.ErrReadRunfile.Wrap(err).KV("file", file)
	}

	if err := yaml.Unmarshal(f, &runfile); err != nil {
		return nil, errors.ErrParseRunfile.Wrap(err)
	}

	runfile.Filepath = fn.Must(filepath.Abs(file))

	prf, err := parseRunfile(ctx, &runfile)
	if err != nil {
		return nil, err
	}

	// prf.Metadata.RunfilePath = file
	prf.Metadata.RunfilePath = fn.Must(filepath.Abs(file))
	return prf, nil
}

func ParseRunfile(ctx types.Context, file string) (*types.ParsedRunfile, error) {
	return parseRunfileFromFile(ctx, file)
}
