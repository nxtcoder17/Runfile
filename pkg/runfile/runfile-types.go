package runfile

import (
	"github.com/nxtcoder17/runfile/pkg/types"
)

type Runfile struct {
	Filepath string `json:"-"`

	Version  string                 `json:"version,omitempty"`
	Includes map[string]IncludeSpec `json:"includes"`
	Env      types.EnvExpr          `json:"env,omitempty"`
	DotEnv   []string               `json:"dotEnv,omitempty"`
	Tasks    map[string]Task        `json:"tasks"`
}

type IncludeSpec struct {
	Runfile string `json:"runfile"`
	Dir     string `json:"dir,omitempty"`
}

type ParsedRunfile struct {
	Env   map[string]string
	Tasks map[string]Task
}
