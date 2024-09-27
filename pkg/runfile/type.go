package runfile

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

/*
EnvVar Values could take multiple forms:

- my_key: "value"

or

  - my_key:
    "sh": "echo hello hi"

Object values with `sh` key, such that the output of this command will be the value of the top-level key
*/
type EnvVar map[string]any

type EvaluationArgs struct {
	Shell   []string
	Environ []string
}

func parseEnvVars(ctx context.Context, ev EnvVar, args EvaluationArgs) (map[string]string, error) {
	env := make(map[string]string, len(ev))
	for k, v := range ev {
		switch v := v.(type) {
		case string:
			env[k] = v
		case map[string]any:
			shcmd, ok := v["sh"]
			if !ok {
				return nil, fmt.Errorf("sh key is missing")
			}

			s, ok := shcmd.(string)
			if !ok {
				return nil, fmt.Errorf("shell cmd is not a string")
			}

			value := new(bytes.Buffer)

			cmd := createCommand(ctx, cmdArgs{
				shell:  args.Shell,
				env:    args.Environ,
				cmd:    s,
				stdout: value,
			})
			if err := cmd.Run(); err != nil {
				return nil, err
			}
			env[k] = strings.TrimSpace(value.String())
		default:
			env[k] = fmt.Sprintf("%v", v)
		}
	}

	return env, nil
}
