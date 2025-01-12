package types

type Runfile struct {
	Version  string                 `json:"version,omitempty"`
	Includes map[string]IncludeSpec `json:"includes"`
	Env      EnvVar                 `json:"env,omitempty"`
	DotEnv   []string               `json:"dotEnv,omitempty"`
	Tasks    map[string]Task        `json:"tasks"`
}

type IncludeSpec struct {
	Runfile string `json:"runfile"`
	Dir     string `json:"dir,omitempty"`
}

// Only one of the fields must be set
type Requires struct {
	Sh     *string `json:"sh,omitempty"`
	GoTmpl *string `json:"gotmpl,omitempty"`
}

/*
// EnvVar Values could take multiple forms:
- my_key: "value"
or
  - my_key:
    sh: "echo hello hi"

Object values with `sh` key, such that the output of this command will be the value of the top-level key
*/
type EnvVar map[string]any

type TaskMetadata struct {
	RunfilePath string `json:"-"`
	Description string `json:"description"`
}

type TaskWatch struct {
	Enable     *bool    `json:"enable,omitempty"`
	Dirs       []string `json:"dirs"`
	Extensions []string `json:"extensions"`
	SSE        *struct {
		Addr string `json:"addr"`
	} `json:"sse,omitempty"`
	// ExcludeDirs []string `json:"excludeDirs"`
}

type Task struct {
	Metadata struct {
		RunfilePath *string
	}

	Name string `json:"-"`
	// Shell in which above commands will be executed
	// Default: ["sh", "-c"]
	/* Common Usecases could be:
	   - ["bash", "-c"]
	   - ["python", "-c"]
	   - ["node", "-e"]
	*/
	Shell []string `json:"shell"`

	// load env vars from [.env](https://www.google.com/search?q=sample+dotenv+files&udm=2) files
	DotEnv []string `json:"dotenv"`

	// working directory for the task
	Dir *string `json:"dir,omitempty"`

	Env EnvVar `json:"env,omitempty"`

	Watch *TaskWatch `json:"watch"`

	Requires []*Requires `json:"requires,omitempty"`

	Interactive bool `json:"interactive,omitempty"`

	// List of commands to be executed in given shell (default: sh)
	// can take multiple forms
	//   - simple string
	//   - a json object with key
	//       `run`, signifying other tasks to run
	//       `if`, condition when to run this server
	Commands []any `json:"cmd"`
}

type CommandJson struct {
	Command string `json:"cmd"`
	Run     string `json:"run"`
	Env     string `json:"env"`

	// If is a go template expression, which must evaluate to true, for task to run
	If *string `json:"if,omitempty"`
}

type ParsedCommandJson struct {
	Command string `json:"cmd"`
	Run     string `json:"run"`
	Env     string `json:"env"`

	// If is a go template expression, which must evaluate to true, for task to run
	If *bool `json:"if"`
}

type ParsedIncludeSpec struct {
	Runfile *Runfile
}
