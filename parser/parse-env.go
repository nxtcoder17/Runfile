package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/nxtcoder17/runfile/errors"
	fn "github.com/nxtcoder17/runfile/functions"
	"github.com/nxtcoder17/runfile/types"
)

type evaluationParams struct {
	Env map[string]string
}

/*
EnvVar can be provided in multiple forms:

> key1: "value1"
or,

> key1:
>   default: "value1"
or,

> key1:
>   default:
>     sh: "echo value1"
or,

> key1:
>   required: true
or,

> key1:
>   sh: "echo hi"
*/
func parseEnvVars(ctx context.Context, ev types.EnvVar, params evaluationParams) (map[string]string, error) {
	env := make(map[string]string, len(ev))
	for k, v := range ev {
		attr := []any{slog.Group("env", "key", k, "value", v)}
		switch v := v.(type) {
		case string:
			env[k] = v
		case map[string]any:
			if ev, ok := os.LookupEnv(k); ok {
				env[k] = ev
				continue
			}

			if s, ok := params.Env[k]; ok {
				env[k] = s
				continue
			}

			// CASE: not found

			// handle field: "required"
			if hasRequired, ok := v["required"]; ok {
				required, ok := hasRequired.(bool)
				if !ok {
					return nil, errors.ErrInvalidEnvVar.Wrap(fmt.Errorf("required field must be a boolean")).KV(attr...)
				}

				if required {
					return nil, errors.ErrRequiredEnvVar.KV(attr...)
				}
			}

			if defaultVal, ok := v["default"]; ok {
				pDefaults, err := parseEnvVars(ctx, types.EnvVar{k: defaultVal}, params)
				if err != nil {
					return nil, errors.ErrInvalidDefaultValue.Wrap(err).KV(attr...)
				}

				if dv, ok := pDefaults[k]; ok {
					env[k] = dv
					continue
				}
			}

			b, err := json.Marshal(v)
			if err != nil {
				return nil, errors.ErrInvalidEnvVar.Wrap(err).KV(attr...)
			}

			var specials struct {
				Sh *string `json:"sh"`
			}

			if err := json.Unmarshal(b, &specials); err != nil {
				return nil, errors.ErrInvalidEnvVar.Wrap(err).KV(attr...)
			}

			switch {
			case specials.Sh != nil:
				{
					value := new(bytes.Buffer)
					cmd := exec.CommandContext(ctx, "sh", "-c", *specials.Sh)
					cmd.Env = fn.ToEnviron(params.Env)
					cmd.Stdout = value
					if err := cmd.Run(); err != nil {
						return nil, errors.ErrEvalEnvVarSh.Wrap(err).KV(attr...)
					}

					env[k] = strings.TrimSpace(value.String())
				}
			default:
				{
					return nil, errors.ErrInvalidEnvVar.Wrap(fmt.Errorf("invalid env format")).KV(attr...)
				}
			}

		default:
			env[k] = fmt.Sprintf("%v", v)
		}
	}

	return env, nil
}
