package runfile

import (
	"fmt"
)

type Task struct {
	// load env vars from [.env](https://www.google.com/search?q=sample+dotenv+files&udm=2) files
	DotEnv []string `json:"dotenv"`

	// working directory for the task
	Dir *string `json:"dir,omitempty"`

	Env map[string]any `json:"env"`

	// List of commands to be executed in given shell (default: sh)
	Commands []string `json:"cmd"`

	// Shell in which above commands will be executed
	// Default: ["sh", "-c"]
	/* Common Usecases could be:
	   - ["bash", "-c"]
	   - ["python", "-c"]
	   - ["node", "-e"]
	*/
	Shell []string `json:"shell"`
}

type ErrTaskNotFound struct {
	TaskName    string
	RunfilePath string
}

func (e ErrTaskNotFound) Error() string {
	return fmt.Sprintf("[task] %s, not found in [Runfile] at %s", e.TaskName, e.RunfilePath)
}
