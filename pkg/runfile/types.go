package runfile

import (
	"github.com/nxtcoder17/runfile/pkg/types"
)

type Context struct {
	*types.Context
	taskTrail []string
}

func NewContext(ctx *types.Context) *Context {
	return &Context{
		Context:   ctx,
		taskTrail: nil,
	}
}

// Only one of the fields must be set
type Requires struct {
	Sh     *string `json:"sh,omitempty"`
	GoTmpl *string `json:"gotmpl,omitempty"`
}

type Shell []string

type TaskMetadata struct {
	RunfilePath string `json:"-"`
	Description string `json:"description"`
}

type TaskWatch struct {
	Enable           *bool    `json:"enable,omitempty"`
	Dirs             []string `json:"dirs"`
	IgnoreDirs       []string `json:"ignoreDirs"`
	Extensions       []string `json:"extensions"`
	IgnoreExtensions []string `json:"ignoreExtensions"`
	SSE              *struct {
		Addr string `json:"addr"`
	} `json:"sse,omitempty"`
	// ExcludeDirs []string `json:"excludeDirs"`
}

type Task struct {
	Metadata struct {
		RunfilePath *string
		Namespace   string
	} `json:"-"`

	Name string `json:"-"`

	Shell any `json:"shell"`

	// load env vars from [.env](https://www.google.com/search?q=sample+dotenv+files&udm=2) files
	DotEnv []string `json:"dotenv"`

	// working directory for the task
	Dir *string `json:"dir,omitempty"`

	Env types.EnvExpr `json:"env,omitempty"`

	ParentEnv map[string]string `json:"-"`

	Watch *TaskWatch `json:"watch"`

	Requires []*Requires `json:"requires,omitempty"`

	Interactive bool `json:"interactive,omitempty"`

	// Parallel allows you to run commands
	Parallel bool `json:"parallel"`

	// List of commands to be executed in given shell (default: sh)
	// can take multiple forms
	//   - simple string
	//   - a json object with key
	//       `run`, signifying other tasks to run
	//       `if`, condition when to run this server
	Commands []any `json:"cmd"`
}

type CommandJson struct {
	Command *string `json:"cmd"`
	Run     *string `json:"run"`

	Env types.Env `json:"env"`

	// If is a go template expression, which must evaluate to true, for task to run
	If *string `json:"if,omitempty"`
}

//	type ParsedTask struct {
//		// Name should be resolved from key itself
//		Name string `json:"-"`
//
//		Shell       Shell
//		Dir         string
//		Watch       *TaskWatch
//		Env         map[string]string
//		Interactive bool
//
//		// Parallel allows you to run commands or run targets in parallel
//		Parallel bool
//
//		Commands []ParsedCommandJson
//
//		AllTasks *[]Task
//	}
type ParsedCommandJson struct {
	Command *string           `json:"cmd"`
	Run     *string           `json:"run"`
	Env     map[string]string `json:"env"`

	// If is a go template expression, which must evaluate to true, for task to run
	If *bool `json:"if"`
}
