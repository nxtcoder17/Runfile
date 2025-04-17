package parser

import (
	"encoding/json"
	"fmt"

	"github.com/nxtcoder17/runfile/errors"
	"github.com/nxtcoder17/runfile/types"
)

func parseCommand(ctx types.Context, prf *types.ParsedRunfile, taskEnv map[string]string, command any) (*types.ParsedCommandJson, error) {
	ferr := func(err error) error {
		return errors.ErrTaskInvalidCommand.Wrap(err).KV("command", command)
	}

	switch c := command.(type) {
	case string:
		{
			return &types.ParsedCommandJson{Command: &c}, nil
		}
	case map[string]any:
		{
			var cj types.CommandJson
			b, err := json.Marshal(c)
			if err != nil {
				return nil, ferr(err)
			}

			if err := json.Unmarshal(b, &cj); err != nil {
				return nil, ferr(err)
			}

			parsedEnv, err := parseEnvVars(ctx, cj.Env, evaluationParams{
				Env: taskEnv,
			})
			if err != nil {
				return nil, ferr(err)
			}

			pcj := types.ParsedCommandJson{
				Env: parsedEnv,
			}

			switch {
			case cj.Run != nil:
				{
					if ctx.TaskNamespace != "" {
						*cj.Run = ctx.TaskNamespace + ":" + *cj.Run
					}
					pcj.Run = cj.Run

					if _, ok := prf.Tasks[*cj.Run]; !ok {
						err := errors.ErrTaskNotFound.Wrap(fmt.Errorf("run target, not found")).KV("command", command, "run-target", cj.Run)
						return nil, err
					}
				}
			case cj.Command != nil:
				{
					pcj.Command = cj.Command
				}
			default:
				{
					return nil, fmt.Errorf("either 'run' or 'cmd' key, must be specified when setting command in json format")
				}
			}

			return &pcj, nil
		}
	default:
		{
			return nil, ferr(fmt.Errorf("invalid command type, must be either a string or an object"))
		}
	}
}
