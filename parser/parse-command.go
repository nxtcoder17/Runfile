package parser

import (
	"encoding/json"
	"fmt"

	"github.com/nxtcoder17/runfile/errors"
	"github.com/nxtcoder17/runfile/types"
)

func parseCommand(prf *types.ParsedRunfile, command any) (*types.ParsedCommandJson, error) {
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

			pcj := types.ParsedCommandJson{
				Env: cj.Env,
			}

			switch {
			case cj.Run != nil:
				{
					pcj.Run = cj.Run
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

			if _, ok := prf.Tasks[*cj.Run]; !ok {
				return nil, errors.ErrTaskNotFound.Wrap(fmt.Errorf("run target, not found")).KV("command", command, "run-target", cj.Run)
			}

			return &pcj, nil
		}
	default:
		{
			return nil, ferr(fmt.Errorf("invalid command type, must be either a string or an object"))
		}
	}
}
