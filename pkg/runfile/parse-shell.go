package runfile

import (
	"github.com/nxtcoder17/runfile/pkg/errors"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
)

var shellAliasMap = map[string][]string{
	"sh":         {"sh", "-c"},
	"bash":       {"bash", "-c"},
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

var shellAliasKeys = fn.MapKeys(shellAliasMap)

func (r *ParsedRunfile) ParseTaskShell(taskName string) (Shell, error) {
	task, ok := r.Tasks[taskName]
	if !ok {
		return nil, errors.ErrTaskNotFound(taskName)
	}

	if task.Shell == nil {
		return shellAliasMap["sh"], nil
	}

	switch val := task.Shell.(type) {
	case string:
		shell, ok := shellAliasMap[val]
		if !ok {
			return nil, errors.ErrInvalidShellAlias(val).KV("available", shellAliasKeys)
		}
		return shell, nil
	case []string:
		return val, nil
	default:
		return nil, errors.WrapStr("shell must be a string or []string")
	}
}
