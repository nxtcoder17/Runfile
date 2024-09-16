package runfile

type RunFile struct {
	Version string

	Tasks map[string]TaskSpec `json:"tasks"`
}

type TaskSpec struct {
	// load env vars from [.env](https://www.google.com/search?q=sample+dotenv+files&udm=2) files
	DotEnv   []string       `json:"dotenv"`
	Env      map[string]any `json:"env"`
	Commands []string       `json:"cmd"`
	Shell    []string       `json:"shell"`
}
