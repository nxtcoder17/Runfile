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
			return &types.ParsedCommandJson{Command: c}, nil
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

			if cj.Run == "" && cj.Command == "" {
				return nil, ferr(fmt.Errorf("key: 'run' or 'cmd', must be specified when setting command in json format"))
			}

			var pcj types.ParsedCommandJson
			pcj.Run = cj.Run
			pcj.Command = cj.Command

			if _, ok := prf.Tasks[cj.Run]; !ok {
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
