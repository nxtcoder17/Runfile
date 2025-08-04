package errors

import (
	"errors"
	"fmt"
	"slices"
)

type Error struct {
	err error

	msg *string
	kv  map[string]any
}

var _ error = (*Error)(nil)

func (e *Error) Msg(msg string) *Error {
	if e.err != nil {
		e.msg = &msg
	}
	return e
}

func (e *Error) KV(kv ...any) *Error {
	if e.kv == nil {
		e.kv = make(map[string]any)
	}
	if e != nil {
		for i := 1; i < len(kv); i++ {
			k, v := kv[i-1], kv[i]
			if ks, ok := k.(string); ok {
				e.kv[ks] = v
			}
		}
	}
	return e
}

// Error implements error.
func (e *Error) Error() string {
	if e.msg != nil {
		return fmt.Sprintf("%s [%s]", e.err.Error(), *e.msg)
	}

	return e.err.Error()
}

// Unwrap returns the underlying error for error chain compatibility
func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) GetMsg() string {
	if e.msg != nil {
		return *e.msg
	}
	return ""
}

func (e *Error) SlogAttrs() []any {
	keys := make([]string, 0, len(e.kv))
	for k := range e.kv {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	result := make([]any, 0, len(e.kv)*2)

	for _, k := range keys {
		result = append(result, k, e.kv[k])
	}

	return result
}

func WrapErr(err error) *Error {
	return &Error{
		err: err,
		msg: nil,
		kv:  nil,
	}
}

func WrapStr(v string) *Error {
	return &Error{
		err: errors.New(v),
		msg: nil,
		kv:  nil,
	}
}

// New creates a new error with the given message
// This is an alias for WrapStr for consistency
func New(msg string) *Error {
	return WrapStr(msg)
}
