package errors

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/nxtcoder17/runfile/types"
)

type Error struct {
	msg string

	taskName string

	kv []any

	traces []string

	err error
}

func (e *Error) GetWrappedErrorString() string {
	if e.err == nil {
		return ""
	}

	return "\nfailed with error:\n" + e.err.Error()
}

func (e *Error) resolveTaskName() string {
	if e.taskName != "" {
		return e.taskName
	}
	for i := 0; i < len(e.kv)-1; i += 2 { // assume kv is key/value pairs
		if key, ok := e.kv[i].(string); ok && key == "task" {
			if val, ok := e.kv[i+1].(string); ok {
				return val
			}
		}
	}
	return ""
}

// Error implements error.
func (e *Error) Error() string {
	return e.err.Error()
	// return fmt.Sprintf("%v {%#v}", e.err, e.kv)
}

func (e *Error) WithTaskName(tn string) *Error {
	e.taskName = tn
	return e
}

func (e *Error) WithCtx(ctx types.Context) *Error {
	return e.WithTaskName(ctx.TaskName)
}

func (e *Error) GetTaskName() string {
	return e.taskName
}

func (e *Error) Log() {
	fmt.Fprintf(os.Stderr, "%s%s%s\n", types.GetErrorStyledPrefix(e.resolveTaskName()), e.msg, e.GetWrappedErrorString())
	if os.Getenv("RUNFILE_DEBUG") == "true" {
		e.InspectLog()
	}
}

func (e *Error) InspectLog() {
	args := make([]any, 0, len(e.kv))
	args = append(args, e.kv...)

	// b, _ := json.MarshalIndent(e.traces, "", "  ")
	// args = append(args, "traces", string(b))

	args = append(args, "traces", e.traces)
	slog.Error(e.msg, args...)
}

var _ error = (*Error)(nil)

func Err(msg string) *Error {
	_, file, line, _ := runtime.Caller(1)
	e := &Error{msg: msg}
	e.traces = append(e.traces, fmt.Sprintf("%s:%d", file, line))
	return e
}

func (e *Error) Wrap(err error) *Error {
	_, file, line, _ := runtime.Caller(1)
	e.traces = append(e.traces, fmt.Sprintf("%s:%d", file, line))
	if e.err != nil {
		e.err = errors.Join(e.err, err)
	} else {
		e.err = err
	}
	return e
}

func (e *Error) WrapStr(msg string) *Error {
	if e.err != nil {
		e.err = errors.Join(e.err, fmt.Errorf(msg))
	} else {
		e.err = fmt.Errorf(msg)
	}
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
	if len(kv) == 0 {
		return e
	}
	_, file, line, _ := runtime.Caller(1)
	e.kv = append(e.kv, kv...)
	e.traces = append(e.traces, fmt.Sprintf("%s:%d", file, line))

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

	ErrInvalidEnvVar = func(k string) *Error {
		return Err(fmt.Sprintf("invalid env var (%s)", k))
	}

	ErrRequiredEnvVar = func(k string) *Error {
		return Err(fmt.Sprintf("required env var (%s)", k))
	}

	ErrInvalidDefaultValue = func(k string, v any) *Error {
		return Err(fmt.Sprintf("invalid default value for env var (%s),default: %v", k, v))
	}

	ErrEvalEnvVarSh = Err("failed while executing env-var sh script")

	ErrTaskNotFound          = Err("task not found")
	ErrTaskFailed            = Err("task failed")
	ErrTaskParsingFailed     = Err("task parsing failed")
	ErrTaskRequirementNotMet = Err("task requirements not met")
	ErrTaskInvalidWorkingDir = Err("task invalid working directory")

	ErrTaskInvalidCommand = Err("task invalid command")
)
