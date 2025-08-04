package runfile

import (
	"context"
	"io"

	"github.com/nxtcoder17/fastlog"
)

type (
	EnvExpr map[string]any
	Env     map[string]string
)

type Context struct {
	context.Context
	logger    *fastlog.Logger
	taskTrail []string
}

func NewContext(ctx context.Context, logger *fastlog.Logger) *Context {
	return &Context{
		Context:   ctx,
		logger:    logger,
		taskTrail: nil,
	}
}

// NewTestContext creates a new Context suitable for testing
func NewTestContext() *Context {
	return NewContext(context.TODO(), fastlog.New(fastlog.Options{Writer: io.Discard}))
}

// Logger returns the logger instance
func (c *Context) Logger() *fastlog.Logger {
	return c.logger
}

// Debug logs a debug message
func (c *Context) Debug(msg string, args ...any) {
	c.logger.Debug(msg, args...)
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

	Env EnvExpr `json:"env,omitempty"`

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

	Env Env `json:"env"`

	// If is a go template expression, which must evaluate to true, for task to run
	If *string `json:"if,omitempty"`
}

type ParsedCommandJson struct {
	Command *string           `json:"cmd"`
	Run     *string           `json:"run"`
	Env     map[string]string `json:"env"`

	// If is a go template expression, which must evaluate to true, for task to run
	If *bool `json:"if"`
}

type Runfile struct {
	Filepath string `json:"-"`

	Version  string                 `json:"version,omitempty"`
	Includes map[string]IncludeSpec `json:"includes"`
	Env      EnvExpr                `json:"env,omitempty"`
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
