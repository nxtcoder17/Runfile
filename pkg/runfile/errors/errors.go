package errors

import (
	"encoding/json"
	"fmt"
)

type Context struct {
	Task    string
	Runfile string

	message *string
	err     error
}

func (c Context) WithMessage(msg string) Context {
	c.message = &msg
	return c
}

func (c Context) WithErr(err error) Context {
	c.err = err
	return c
}

func (c Context) ToString() string {
	m := map[string]string{
		"task":    c.Task,
		"runfile": c.Runfile,
	}
	if c.message != nil {
		m["message"] = *c.message
	}

	if c.err != nil {
		m["err"] = c.err.Error()
	}

	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return string(b)
}

func (e Context) Error() string {
	return e.ToString()
}

type (
	ErrTaskInvalid struct{ Context }
)

type ErrTaskFailedRequirements struct {
	Context
	Requirement string
}

func (e ErrTaskFailedRequirements) Error() string {
	if e.message == nil {
		e.Context = e.Context.WithMessage(fmt.Sprintf("failed (requirement: %q)", e.Requirement))
	}
	return e.Context.Error()
}

type TaskNotFound struct {
	Context
}

func (e TaskNotFound) Error() string {
	// return fmt.Sprintf("[task] %s, not found in [Runfile] at %s", e.TaskName, e.RunfilePath)
	if e.message == nil {
		e.Context = e.Context.WithMessage("Not Found")
	}
	return e.Context.Error()
}

type ErrTaskGeneric struct {
	Context
}

type InvalidWorkingDirectory struct {
	Context
}

func (e InvalidWorkingDirectory) Error() string {
	// return fmt.Sprintf("[task] %s, not found in [Runfile] at %s", e.TaskName, e.RunfilePath)
	if e.message == nil {
		e.Context = e.Context.WithMessage("Invalid Working Directory")
	}
	return e.Context.Error()
}

type InvalidDotEnv struct {
	Context
}

func (e InvalidDotEnv) Error() string {
	if e.message == nil {
		e.Context = e.Context.WithMessage("invalid dotenv")
	}
	return e.Context.Error()
}

type InvalidEnvVar struct {
	Context
}

func (e InvalidEnvVar) Error() string {
	if e.message == nil {
		e.Context = e.Context.WithMessage("invalid dotenv")
	}
	return e.Context.Error()
}

type IncorrectCommand struct {
	Context
}

func (e IncorrectCommand) Error() string {
	if e.message == nil {
		e.Context = e.Context.WithMessage("incorrect command")
	}
	return e.Context.Error()
}
