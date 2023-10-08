package contextedError

import (
	"errors"
	"reflect"
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

func NewFromFunc(err *error, i interface{}) ContextedError {
	return ContextedError{*err, runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()}
}

func NewFromExists(err *error, origin string) ContextedError {
	return ContextedError{*err, origin}
}
