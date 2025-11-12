package spec

type RunfileSpec struct {
	Includes map[string]IncludeSpec `yaml:"includes,omitempty"`
	Env      map[string]any         `yaml:"env,omitempty"`
	// load env vars from [dotenv](https://www.dotenv.org/docs/security/env) files
	// DotEnv can be in format `[".secrets/dev-setup.env"]`
	DotEnv []string            `yaml:"dotenv,omitempty"`
	Tasks  map[string]TaskSpec `yaml:"tasks"`
}

type IncludeSpec struct {
	Runfile string `yaml:"runfile"`
	Dir     string `yaml:"dir,omitempty"`
}

type TaskSpec struct {
	// Shell can take multiple forms
	//   - simple string like `bash`, `zsh` etc.
	//   - or, an array of strings like `[bash, -c]`
	Shell any `yaml:"shell,omitempty"`

	// load env vars from [dotenv](https://www.dotenv.org/docs/security/env) files
	// DotEnv can be in format `[".secrets/dev-setup.env"]`
	// DotEnv can be in format `[".secrets/dev-setup.env"]`
	DotEnv []string `yaml:"dotenv,omitempty"`

	// working directory for the task
	Dir string `yaml:"dir,omitempty"`

	Env map[string]any `yaml:"env,omitempty"`

	Watch *TaskWatchSpec `yaml:"watch"`

	// Requires []*Requires `yaml:"requires,omitempty"`

	// Interactive means the programs will be run with `stdin` linked to os.Stdout
	Interactive bool `yaml:"interactive,omitempty"`

	// Parallel allows you to run commands
	Parallel bool `yaml:"parallel"`

	// Commands can take multiple forms
	//   - simple string like `echo hello world`
	//   - a json object with key
	//       `run`, signifying other tasks to run
	//       `if`, condition when to run this server
	Commands []any `yaml:"cmd"`
}

type TaskWatchSpec struct {
	Enabled          bool     `yaml:"enabled,omitempty"`
	Dirs             []string `yaml:"dirs"`
	IgnoreDirs       []string `yaml:"ignoreDirs"`
	Extensions       []string `yaml:"extensions"`
	IgnoreExtensions []string `yaml:"ignoreExtensions"`
	SSE              *struct {
		Addr string `yaml:"addr"`
	} `yaml:"sse,omitempty"`
}
