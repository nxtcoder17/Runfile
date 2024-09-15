package runfile

type RunFile struct {
	Version string

	Tasks map[string]TaskSpec `json:"tasks"`
}

type TaskSpec struct {
	Env      map[string]any `json:"env"`
	Commands []string       `json:"cmd"`
	Shell    []string       `json:"shell"`
}
