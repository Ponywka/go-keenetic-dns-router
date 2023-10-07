package contextedError

import (
	"errors"
	"runtime"
)

type ContextedError struct {
	error
	origin string
}

func (e *ContextedError) GetOrigin() string {
	return e.origin
}

func New(msg string) ContextedError {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	err := errors.New(msg)
	return NewFromExists(&err, fn.Name())
}

func NewFromExists(err *error, origin string) ContextedError {
	return ContextedError{*err, origin}
}
