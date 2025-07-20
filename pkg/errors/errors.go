package errors

import (
	"errors"
	"fmt"
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
	if len(e.kv) > 0 {
		return fmt.Sprintf("%s %+v", e.err.Error(), e.kv)
	}

	return e.err.Error()
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
