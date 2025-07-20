package task

import (
	"fmt"

	fn "github.com/nxtcoder17/runfile/pkg/functions"
)

var shellAliasMap = map[string][]string{
	"sh":         {"sh", "-c"},
	"bash":       {"bash", "-c"},
	"python":     {"python", "-c"},
	"go":         {"go", "run", "-e"},
	"node":       {"node", "-e"},
	"ruby":       {"ruby", "-e"},
	"perl":       {"perl", "-e"},
	"php":        {"php", "-r"},
	"rust":       {"cargo", "script", "-e"},
	"clojure":    {"closure", "-e"},
	"lua":        {"lua", "-e"},
	"exlixir":    {"exlixir", "-e"},
	"powershell": {"powershell", "-Command"},
	"haskell":    {"runghc", "-e"},
}

var shellAliasKeys = fn.MapKeys(shellAliasMap)

func (t *Task) ParseShell() (Shell, error) {
	if t.Shell == nil {
		return Shell{"sh", "-c"}, nil
	}

	switch val := t.Shell.(type) {
	case string:
		shell, ok := shellAliasMap[val]
		if !ok {
			return nil, fmt.Errorf("invalid shell alias, must be one of %#v", shellAliasKeys)
		}
		return shell, nil
	case []string:
		return val, nil
	default:
		return nil, fmt.Errorf("shell must be a string or []string")
	}
}
