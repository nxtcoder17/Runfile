package errors

import (
	"fmt"
)

func ErrReadRunfile(err error) *Error {
	return WrapErr(err).Msg("failed to read runfile")
}

func ErrParseRunfile(err error) *Error {
	return WrapErr(err).Msg("failed to parse runfile")
}

func ErrParseIncludes(err error) *Error {
	return WrapErr(err).Msg("failed to parse includes")
}

func ErrParseDotEnv(err error) *Error {
	return WrapErr(err).Msg("failed to parse dotenv file")
}

func ErrInvalidDotEnv(err error) *Error {
	return WrapErr(err).Msg("invalid dotenv file")
}

func ErrInvalidEnvVar(key string, err error) *Error {
	return WrapErr(err).Msg("invalid env var (" + key + ")")
}

func ErrRequiredEnvVar(key string) *Error {
	return WrapStr("required env var (" + key + ")")
}

func ErrInvalidDefaultValue(err error, key string, value any) *Error {
	return WrapErr(err).Msg("invalid default value for env var (" + key + "), default: " + fmt.Sprint(value))
}

func ErrEvalEnvVarSh(err error) *Error {
	return WrapErr(err).Msg("failed while executing env-var sh script")
}

func ErrTaskNotFound(taskName string) *Error {
	return WrapStr("task not found").KV("task", taskName)
}

func ErrTaskFailed(err error) *Error {
	return WrapErr(err).Msg("task failed")
}

func ErrTaskParsingFailed(err error) *Error {
	return WrapErr(err).Msg("task parsing failed")
}

func ErrTaskRequirementNotMet(requirement string, err error) *Error {
	return WrapErr(err).Msg("task requirements not met").KV("requirement", requirement)
}

func ErrTaskInvalidWorkingDir(workingDir string, err error) *Error {
	return WrapErr(err).Msg("task invalid working directory").KV("working-dir", workingDir)
}

func ErrTaskInvalidCommand(command any, err error) *Error {
	return WrapErr(err).Msg("task invalid command").KV("command", command)
}

func ErrInvalidShellAlias(alias string) *Error {
	return WrapStr("invalid shell alias").KV("alias", alias)
}
