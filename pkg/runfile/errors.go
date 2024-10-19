package runfile

import (
	"github.com/nxtcoder17/runfile/pkg/errors"
)

type Error struct {
	TaskName string
	Runfile  string

	*errors.Message
}

var _ error = (*Error)(nil)

func NewError(taskName string, runfile string) *Error {
	return &Error{
		TaskName: taskName,
		Runfile:  runfile,
		Message:  &errors.Message{},
	}
}

func (e *Error) WithTask(task string) *Error {
	e.TaskName = task
	e.Message = e.Message.WithMetadata("task", task)
	return e
}

func (e *Error) WithRunfile(rf string) *Error {
	e.Runfile = rf
	e.Message = e.Message.WithMetadata("runfile", rf)
	return e
}

func (e *Error) WithErr(err error) *Error {
	e.Message = e.Message.WithErr(err)
	return e
}

func (e Error) WithMetadata(attrs ...any) *Error {
	e.Message = e.Message.WithMetadata(attrs...)
	return &e
}

var (
	RunfileReadFailed    = &Error{Message: errors.New("Runfile Read Failed", nil)}
	RunfileParsingFailed = &Error{Message: errors.New("Runfile Parsing Failed", nil)}
)

var (
	TaskNotFound = &Error{Message: errors.New("Task Not Found", nil)}
	TaskFailed   = &Error{Message: errors.New("Task Failed", nil)}

	TaskWorkingDirectoryInvalid = &Error{Message: errors.New("Task Working Directory Invalid", nil)}

	TaskRequirementFailed    = &Error{Message: errors.New("Task Requirement Failed", nil)}
	TaskRequirementIncorrect = &Error{Message: errors.New("Task Requirement Incorrect", nil)}

	TaskEnvInvalid       = &Error{Message: errors.New("Task Env is invalid", nil)}
	TaskEnvRequired      = &Error{Message: errors.New("Task Env Required, but not provided", nil)}
	TaskEnvCommandFailed = &Error{Message: errors.New("Task Env command failed", nil)}
	TaskEnvGoTmplFailed  = &Error{Message: errors.New("Task Env GoTemplate failed", nil)}
)

var (
	DotEnvNotFound      = &Error{Message: errors.New("DotEnv Not Found", nil)}
	DotEnvInvalid       = &Error{Message: errors.New("Dotenv Invalid", nil)}
	DotEnvParsingFailed = &Error{Message: errors.New("DotEnv Parsing Failed", nil)}
)

var (
	CommandFailed  = &Error{Message: errors.New("Command Failed", nil)}
	CommandInvalid = &Error{Message: errors.New("Command Invalid", nil)}
)
