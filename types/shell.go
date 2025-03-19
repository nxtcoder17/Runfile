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
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("invalid shell format: %w", err)
	}

	switch val := v.(type) {
	case string:
		shell, ok := shellAliasMap[val]
		if !ok {
			return fmt.Errorf("invalid shell alias")
		}
		*s = Shell(shell)
	case []any:
		// INFO: json unmarshals array as []any
		var shells []string
		for _, item := range val {
			str, ok := item.(string)
			if !ok {
				return fmt.Errorf("invalid shell values, must be an []string")
			}
			shells = append(shells, str)
		}
		*s = Shell(shells)
	default:
		return fmt.Errorf("unexpected JSON type for shell")
	}

	return nil
}

// func (s *Shell) UnmarshalJSON(data []byte) error {
// 	var single string
// 	if err := json.Unmarshal(data, &single); err == nil {
// 		// INFO: It means shell provided is a single string i.e. shell alias
// 		shell, ok := shellAliasMap[single]
// 		if !ok {
// 			return fmt.Errorf("invalid shell alias")
// 		}
// 		*s = Shell(shell)
// 		return nil
// 	}
//
// 	var multiple []string
// 	if err := json.Unmarshal(data, &multiple); err == nil {
// 		// If it's a slice, assign it directly
// 		*s = Shell(multiple)
// 		return nil
// 	}
//
// 	return fmt.Errorf("invalid shell format")
// }
