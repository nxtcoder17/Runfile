package runfile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/template"

	sprig "github.com/go-task/slim-sprig/v3"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"github.com/nxtcoder17/runfile/pkg/runfile/errors"
)

type Context struct {
	context.Context
	*slog.Logger

	RunfilePath string
	Taskname    string
}

func NewContext(ctx context.Context, logger *slog.Logger) Context {
	lgr := logger
	if lgr == nil {
		lgr = slog.Default()
	}

	return Context{Context: ctx, Logger: lgr}
}

type EvaluationArgs struct {
	Shell []string
	Env   map[string]string
}

func ToEnviron(m map[string]string) []string {
	results := os.Environ()
	for k, v := range m {
		results = append(results, fmt.Sprintf("%s=%v", k, v))
	}
	return results
}

type EnvKV struct {
	Key string

	Value  *string `json:"value"`
	Sh     *string `json:"sh"`
	GoTmpl *string `json:"gotmpl"`
}

func (ejv EnvKV) Parse(ctx Context, args EvaluationArgs) (*string, errors.Message) {
	switch {
	case ejv.Value != nil:
		{
			return ejv.Value, nil
		}
	case ejv.Sh != nil:
		{
			value := new(bytes.Buffer)

			cmd := createCommand(ctx, cmdArgs{
				shell:  args.Shell,
				env:    ToEnviron(args.Env),
				cmd:    *ejv.Sh,
				stdout: value,
			})
			if err := cmd.Run(); err != nil {
				return nil, errors.TaskEnvCommandFailed.WithErr(err)
			}

			return fn.New(strings.TrimSpace(value.String())), nil
		}
	case ejv.GoTmpl != nil:
		{
			t := template.New(ejv.Key).Funcs(sprig.FuncMap())
			t, err := t.Parse(fmt.Sprintf(`{{ %s }}`, *ejv.GoTmpl))
			if err != nil {
				return nil, errors.TaskEnvGoTmplFailed.WithErr(err)
			}

			value := new(bytes.Buffer)
			if err := t.ExecuteTemplate(value, ejv.Key, map[string]string{}); err != nil {
				return nil, errors.TaskEnvGoTmplFailed.WithErr(err)
			}

			return fn.New(strings.TrimSpace(value.String())), nil
		}
	default:
		{
			return nil, errors.TaskEnvInvalid.WithErr(fmt.Errorf("failed to parse, unknown format, one of [value, sh, gotmpl] must be set"))
		}
	}
}

func parseEnvVars(ctx Context, ev EnvVar, args EvaluationArgs) (map[string]string, errors.Message) {
	env := make(map[string]string, len(ev))
	for k, v := range ev {
		attr := []any{slog.Group("env", "key", k, "value", v)}
		switch v := v.(type) {
		case string:
			env[k] = v
		case map[string]any:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, errors.TaskEnvInvalid.WithErr(err).WithMetadata(attr)
			}

			var envAsJson struct {
				*EnvKV
				Required bool
				Default  *EnvKV
			}

			if err := json.Unmarshal(b, &envAsJson); err != nil {
				return nil, errors.TaskEnvInvalid.WithErr(err).WithMetadata(attr)
			}

			switch {
			case envAsJson.Required:
				{
					isDefined := false
					if _, ok := os.LookupEnv(k); ok {
						isDefined = true
					}

					if !isDefined {
						if _, ok := args.Env[k]; ok {
							isDefined = true
						}
					}

					if !isDefined {
						return nil, errors.TaskEnvInvalid.WithErr(fmt.Errorf("env required, but not provided")).WithMetadata(attr)
					}
				}

			case envAsJson.EnvKV != nil:
				{
					envAsJson.Key = k
					s, err := envAsJson.EnvKV.Parse(ctx, args)
					if err != nil {
						return nil, err.WithMetadata(attr)
					}
					env[k] = *s
				}

			case envAsJson.Default != nil:
				{
					envAsJson.Default.Key = k
					s, err := envAsJson.Default.Parse(ctx, args)
					if err != nil {
						return nil, err.WithMetadata(attr)
					}
					env[k] = *s
				}
			default:
				{
					return nil, errors.TaskEnvInvalid.WithErr(fmt.Errorf("either required, value, sh, gotmpl or default, must be defined")).WithMetadata(attr)
				}
			}

		default:
			env[k] = fmt.Sprintf("%v", v)
		}
	}

	return env, nil
}
