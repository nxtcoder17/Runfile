package runfile

import (
	"encoding/json"
	"fmt"

	"github.com/nxtcoder17/runfile/pkg/errors"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
)

func parseCommand(_ *Context, command any, env map[string]string) (*ParsedCommandJson, error) {
	ferr := func(err error) error {
		return errors.ErrTaskInvalidCommand(command, err)
	}

	switch c := command.(type) {
	case string:
		{
			if c == "" {
				return nil, ferr(fmt.Errorf("empty command"))
			}

			return &ParsedCommandJson{Command: &c, Env: env}, nil
		}
	case map[string]any:
		{
			var cj CommandJson
			b, err := json.Marshal(c)
			if err != nil {
				return nil, ferr(err)
			}

			if err := json.Unmarshal(b, &cj); err != nil {
				return nil, ferr(err)
			}

			pcj := ParsedCommandJson{
				Env: fn.MapMerge(env, cj.Env),
			}

			switch {
			case cj.Run != nil:
				{
					if *cj.Run == "" {
						return nil, ferr(fmt.Errorf("empty run target"))
					}

					pcj.Run = cj.Run
				}
			case cj.Command != nil:
				{
					if *cj.Command == "" {
						return nil, ferr(fmt.Errorf("empty command"))
					}
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
