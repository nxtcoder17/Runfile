package types

import (
	"encoding/json"
	"fmt"
)

var shellAliasMap = map[string][]string{
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

type Shell []string

// UnmarshalJSON implements custom unmarshaling for Shell
func (s *Shell) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		// INFO: It means shell provided is a single string i.e. shell alias
		shell, ok := shellAliasMap[single]
		if !ok {
			return fmt.Errorf("invalid shell alias")
		}
		*s = Shell(shell)
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err == nil {
		// If it's a slice, assign it directly
		*s = Shell(multiple)
		return nil
	}

	return fmt.Errorf("invalid shell format")
}
