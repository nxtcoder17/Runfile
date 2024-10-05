package errors

// runfile
var (
	RunfileReadFailed    = New("Runfile Read Failed", nil)
	RunfileParsingFailed = New("Runfile Parsing Failed", nil)
)

var (
	TaskNotFound = New("Task Not Found", nil)
	TaskFailed   = New("Task Failed", nil)

	TaskWorkingDirectoryInvalid = New("Task Working Directory Invalid", nil)

	TaskRequirementFailed    = New("Task Requirement Failed", nil)
	TaskRequirementIncorrect = New("Task Requirement Incorrect", nil)

	TaskEnvInvalid       = New("Task Env is invalid", nil)
	TaskEnvRequired      = New("Task Env is Required", nil)
	TaskEnvCommandFailed = New("Task Env command failed", nil)
	TaskEnvGoTmplFailed  = New("Task Env GoTemplate failed", nil)
)

var (
	DotEnvNotFound      = New("DotEnv Not Found", nil)
	DotEnvInvalid       = New("Dotenv Invalid", nil)
	DotEnvParsingFailed = New("DotEnv Parsing Failed", nil)
)

var (
	CommandFailed  = New("Command Failed", nil)
	CommandInvalid = New("Command Invalid", nil)
)
