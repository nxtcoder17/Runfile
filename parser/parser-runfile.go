package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nxtcoder17/runfile/errors"
	fn "github.com/nxtcoder17/runfile/functions"
	"github.com/nxtcoder17/runfile/types"
	"sigs.k8s.io/yaml"
)

func parse(file string) (*types.ParsedRunfile, error) {
	var runfile types.Runfile
	f, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.ErrReadRunfile.Wrap(err).KV("file", file)
	}

	if err = yaml.Unmarshal(f, &runfile); err != nil {
		return nil, errors.ErrParseRunfile.Wrap(err).KV("file", file)
	}

	var prf types.ParsedRunfile

	// prf.Metadata.RunfilePath = fn.Must(filepath.Rel(fn.Must(os.Getwd()), file))
	prf.Metadata.RunfilePath = file

	m, err := parseIncludes(runfile.Includes)
	if err != nil {
		return nil, err
	}

	prf.Tasks = runfile.Tasks
	for k, iprf := range m {
		for taskName, task := range iprf.Tasks {
			task.Metadata.RunfilePath = &iprf.Metadata.RunfilePath
			// task.Metadata.RunfilePath = fn.New(fn.Must(filepath.Rel(fn.Must(os.Getwd()), iprf.Metadata.RunfilePath)))
			prf.Tasks[fmt.Sprintf("%s:%s", k, taskName)] = task
		}
	}

	dotEnvFiles := make([]string, 0, len(runfile.DotEnv))
	for i := range runfile.DotEnv {
		dotEnvFiles = append(dotEnvFiles, fn.Must(filepath.Abs(runfile.DotEnv[i])))
	}

	// dotenvVars, err := parseDotEnvFiles(runfile.DotEnv...)
	dotenvVars, err := parseDotEnvFiles(dotEnvFiles...)
	if err != nil {
		return nil, err
	}

	envVars, err := parseEnvVars(context.TODO(), runfile.Env, evaluationParams{Env: dotenvVars})
	if err != nil {
		return nil, err
	}

	prf.Env = fn.MapMerge(dotenvVars, envVars)

	return &prf, nil
}

func Parse(file string) (*types.ParsedRunfile, error) {
	return parse(file)
}
