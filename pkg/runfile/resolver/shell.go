package resolver

import (
	"fmt"

	"github.com/nxtcoder17/go.errors"
)

var shellAliasMap = map[string][]string{
	"sh":         {"sh", "-c"},
	"bash":       {"bash", "-c"},
	"zsh":        {"zsh", "-c"},
	"python":     {"python", "-c"},
	"node":       {"node", "-e"},
	"ruby":       {"ruby", "-e"},
	"perl":       {"perl", "-e"},
	"php":        {"php", "-r"},
	"rust":       {"cargo", "script", "-e"},
	"clojure":    {"clojure", "-e"},
	"lua":        {"lua", "-e"},
	"elixir":     {"elixir", "-e"},
	"powershell": {"powershell", "-Command"},
	"haskell":    {"runghc", "-e"},
}

func ParseShell(shell any) ([]string, error) {
	if shell == nil {
		return shellAliasMap["sh"], nil
	}

	switch v := shell.(type) {
	case string:
		{
			if v == "" {
				return shellAliasMap["sh"], nil
			}

			sh, ok := shellAliasMap[v]
			if !ok {
				return nil, errors.New(fmt.Sprintf("unsupported shell alias (%s), must specify shell in list format", v))
			}
			return sh, nil
		}
	case []any:
		if len(v) == 0 {
			return shellAliasMap["sh"], nil
		}

		result := make([]string, 0, len(v))
		for i := range v {
			str, ok := v[i].(string)
			if !ok {
				return nil, errors.New("invalid shell specs must consist of only strings")
			}
			result = append(result, str)
		}

		return result, nil
	default:
		return nil, errors.New("unknown shell format, must in a string or a list").KV("got", fmt.Sprintf("%v (type: %T)", v, v))
	}
}
