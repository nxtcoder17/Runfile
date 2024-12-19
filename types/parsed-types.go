package types

type ParsedRunfile struct {
	Env   map[string]string `json:"env,omitempty"`
	Tasks map[string]Task   `json:"tasks"`

	Metadata struct {
		RunfilePath string
	} `json:"-"`
}

type ParsedTask struct {
	Shell       []string            `json:"shell"`
	WorkingDir  string              `json:"workingDir"`
	Watch       *TaskWatch          `json:"watch,omitempty"`
	Env         map[string]string   `json:"environ"`
	Interactive bool                `json:"interactive,omitempty"`
	Commands    []ParsedCommandJson `json:"commands"`
}
