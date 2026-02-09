package resolver

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nxtcoder17/go.errors"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
)

func parseDotEnvFilesInto(store map[string]string, files []string) error {
	for _, file := range files {
		if !filepath.IsAbs(file) {
			return errors.New("dotenv file must have absolute paths").KV("dotenv.file", file)
		}

		f, err := os.Open(file)
		if err != nil {
			return errors.New("failed to open dotenv file").Wrap(err).KV("dotenv.file", file)
		}

		m, err := godotenv.Parse(f)
		if err != nil {
			return errors.New("failed to parse dotenv file").Wrap(err).KV("dotenv.file", file)
		}

		if err := f.Close(); err != nil {
			return errors.New("failed to close dotenv file").Wrap(err).KV("dotenv.file", file)
		}

		maps.Copy(store, m)
	}

	return nil
}

func parseEnvInto(ctx context.Context, envStore map[string]string, envMap map[string]any) error {
	lookupEnv := func(key string) (string, bool) {
		if v, ok := os.LookupEnv(key); ok {
			return v, true
		}
		if v, ok := envStore[key]; ok {
			return v, true
		}

		return "", false
	}

	lazyEvalMap := make(map[string]*exec.Cmd)

	for k, v := range envMap {
		// INFO: Skip if key already exists (in OS env or envStore)
		if _, ok := lookupEnv(k); ok {
			continue
		}

		switch value := v.(type) {
		case string:
			envStore[k] = value
		case map[string]any:
			{

				if requiredVal, ok := value["required"]; ok {
					isRequired, ok := requiredVal.(bool)
					if !ok {
						return errors.New("ENV-EXPRESSION: value field `required` must be a boolean").KV("env.key", k, "env.value", value)
					}
					if isRequired {
						if _, exists := lookupEnv(k); !exists {
							return errors.New(fmt.Sprintf("ENV-EXPRESSION: env var '%s' is required, it must be provided", k)).KV("env.key", k, "env.value", value)
						}
					}
				}

				for optKey, optVal := range value {
					if shell, ok := shellAliasMap[optKey]; ok {
						envEvalScript, ok := optVal.(string)
						if !ok {
							return errors.New(fmt.Sprintf("ENV-EXPRESSION: value field `%s`, must have a string value", optKey)).KV("env.key", k, "env.value", value)
						}

					// #nosec G204 - This is intentional: env vars with sh/bash keys are meant to
					// execute shell commands defined in the Runfile. The Runfile is trusted
					// user-provided configuration, similar to Makefiles or shell scripts.
					lazyEvalMap[k] = exec.CommandContext(ctx, shell[0], append(shell[1:], envEvalScript)...)
						break
					}
				}
			}
		default:
			envStore[k] = fmt.Sprint(v)
		}
	}

	cmdEnv := fn.ToEnviron(envStore)

	for k, cmd := range lazyEvalMap {
		cmd.Env = cmdEnv
		stdout := new(bytes.Buffer)
		cmd.Stdout = stdout

		stderr := new(bytes.Buffer)
		cmd.Stderr = stderr

		if err := cmd.Run(); err != nil {
			return errors.New("ENV-EXPRESSION: evaluation failed").Wrap(err).KV("env.key", k, "eval.cmd", cmd.String(), "eval.stderr", stderr.String())
		}

		envStore[k] = strings.TrimSpace(stdout.String())
	}

	return nil
}
