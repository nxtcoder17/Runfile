package errors

import (
	"fmt"
	"log/slog"
	"runtime"
)

type Error struct {
	msg string
	// kv is a slice of slog Attributes i.e. ("key", "value")
	keys []string
	// kv   map[string]any
	kv []any

	traces []string

	err error
}

// Error implements error.
func (e *Error) Error() string {
	return fmt.Sprintf("%v {%#v}", e.err, e.kv)
}

func (e *Error) Log() {
	args := make([]any, 0, len(e.kv))
	args = append(args, e.kv...)
	args = append(args, "traces", e.traces)
	slog.Error(e.msg, args...)
}

var _ error = (*Error)(nil)

func Err(msg string) *Error {
	return &Error{msg: msg}
}

func (e *Error) Wrap(err error) *Error {
	_, file, line, _ := runtime.Caller(1)
	e.traces = append(e.traces, fmt.Sprintf("%s:%d", file, line))
	e.err = err
	return e
}

func (e *Error) WrapStr(msg string) *Error {
	e.err = fmt.Errorf(msg)
	return e
}

func (e *Error) KV(kv ...any) *Error {
	// if e.kv == nil {
	// 	e.kv = make(map[string]any)
	// }

	// for i := 0; i < len(kv); i += 2 {
	// 	// e.keys = append(e.keys, kv[i].(string))
	// 	e.kv[kv[i].(string)] = kv[i+1]
	// }
	e.kv = append(e.kv, kv...)

	return e
}

func WithErr(err error) *Error {
	_, file, line, _ := runtime.Caller(1)
	err2, ok := err.(*Error)
	if !ok {
		err2 = &Error{err: err}
	}

	err2.traces = append(err2.traces, fmt.Sprintf("%s:%d", file, line))
	return err2
}

// ERROR constants
var (
	ErrReadRunfile  = Err("failed to read runfile")
	ErrParseRunfile = Err("failed to read runfile")

	ErrParseIncludes = Err("failed to parse includes")
	ErrParseDotEnv   = Err("failed to parse dotenv file")
	ErrInvalidDotEnv = Err("invalid dotenv file")

	ErrInvalidEnvVar       = Err("invalid env var")
	ErrRequiredEnvVar      = Err("required env var")
	ErrInvalidDefaultValue = Err("invalid default value for env var")

	ErrEvalEnvVarSh = Err("failed while executing env-var sh script")

	ErrTaskNotFound          = Err("task not found")
	ErrTaskFailed            = Err("task failed")
	ErrTaskParsingFailed     = Err("task parsing failed")
	ErrTaskRequirementNotMet = Err("task requirements not met")
	ErrTaskInvalidWorkingDir = Err("task invalid working directory")

	ErrTaskInvalidCommand = Err("task invalid command")
)
