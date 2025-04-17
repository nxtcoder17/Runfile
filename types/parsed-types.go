package types

type ParsedRunfile struct {
	Env      map[string]string
	Includes map[string]Task
	Tasks    map[string]Task

	Metadata struct {
		RunfilePath string
	}
}

type ParsedTask struct {
	// Namespace for a task is auto set, when it is imported under a name
	Namespace string `json:"-"`

	// Name should be resolved from key itself
	Name string `json:"-"`

	Shell       Shell             `json:"shell"`
	WorkingDir  string            `json:"workingDir"`
	Watch       *TaskWatch        `json:"watch,omitempty"`
	Env         map[string]string `json:"environ"`
	Interactive bool              `json:"interactive,omitempty"`

	// Parallel allows you to run commands or run targets in parallel
	Parallel bool `json:"parallel"`

	Commands []ParsedCommandJson `json:"commands"`
}

type ParsedCommandJson struct {
	Command *string           `json:"cmd"`
	Run     *string           `json:"run"`
	Env     map[string]string `json:"env"`

	// If is a go template expression, which must evaluate to true, for task to run
	If *bool `json:"if"`
}

type ParsedIncludeSpec struct {
	Runfile *Runfile
}
