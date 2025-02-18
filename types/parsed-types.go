package types

type ParsedRunfile struct {
	Env   map[string]string `json:"env,omitempty"`
	Tasks map[string]Task   `json:"tasks"`

	Metadata struct {
		RunfilePath string
	} `json:"-"`
}

type ParsedTask struct {
	// Name should be resolved from key itself
	Name string `json:"-"`

	Shell       []string          `json:"shell"`
	WorkingDir  string            `json:"workingDir"`
	Watch       *TaskWatch        `json:"watch,omitempty"`
	Env         map[string]string `json:"environ"`
	Interactive bool              `json:"interactive,omitempty"`

	// Parallel allows you to run commands or run targets in parallel
	Parallel bool `json:"parallel"`

	Commands []ParsedCommandJson `json:"commands"`
}

type ParsedCommandJson struct {
	Command *string `json:"cmd"`
	Run     *string `json:"run"`
	Env     string  `json:"env"`

	// If is a go template expression, which must evaluate to true, for task to run
	If *bool `json:"if"`
}

type ParsedIncludeSpec struct {
	Runfile *Runfile
}
