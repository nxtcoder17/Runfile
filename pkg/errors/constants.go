package errors

import (
	"fmt"

	"github.com/nxtcoder17/go.errors"
)

func ErrReadRunfile(err error) *errors.Error {
	return errors.New("failed to read runfile").Wrap(err)
}

func ErrParseRunfile(err error) *errors.Error {
	return errors.New("failed to parse runfile").Wrap(err)
}

func ErrParseIncludes(err error) *errors.Error {
	return errors.New("failed to parse includes").Wrap(err)
}

func ErrParseDotEnv(err error) *errors.Error {
	return errors.New("failed to parse dotenv file").Wrap(err)
}

func ErrInvalidDotEnv(err error) *errors.Error {
	return errors.New("invalid dotenv file").Wrap(err)
}

func ErrInvalidEnvVar(key string, err error) *errors.Error {
	return errors.New("invalid env var (" + key + ")").Wrap(err)
}

func ErrRequiredEnvVar(key string) *errors.Error {
	return errors.New("required env var (" + key + ")")
}

func ErrInvalidDefaultValue(err error, key string, value any) *errors.Error {
	return errors.New("invalid default value for env var (" + key + "), default: " + fmt.Sprint(value)).Wrap(err)
}

func ErrEvalEnvVarSh(err error) *errors.Error {
	return errors.New("failed while executing env-var sh script").Wrap(err)
}

func ErrTaskNotFound(taskName string) *errors.Error {
	return errors.New("task not found").KV("task", taskName)
}

func ErrTaskFailed(err error) *errors.Error {
	return errors.New("task failed").Wrap(err)
}

func ErrTaskParsingFailed(err error) *errors.Error {
	return errors.New("task parsing failed").Wrap(err)
}

func ErrTaskRequirementNotMet(requirement string, err error) *errors.Error {
	return errors.New("task requirements not met").Wrap(err).KV("requirement", requirement)
}

func ErrTaskInvalidWorkingDir(workingDir string, err error) *errors.Error {
	return errors.New("task invalid working directory").Wrap(err).KV("working-dir", workingDir)
}

func ErrTaskInvalidCommand(command any, err error) *errors.Error {
	return errors.New("task invalid command").Wrap(err).KV("command", command)
}

func ErrInvalidShellAlias(alias string) *errors.Error {
	return errors.New("invalid shell alias").KV("alias", alias)
}

func ErrCircularDependency(taskName string) *errors.Error {
	return errors.New("invalid shell alias").KV("task", taskName)
}
