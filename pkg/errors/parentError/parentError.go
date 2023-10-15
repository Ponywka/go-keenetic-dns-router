package parentError

import (
	"errors"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/errors/contextedError"
	"runtime"
)

type ParentError struct {
	contextedError.ContextedError
	child error
}

func (e *ParentError) GetChild() error {
	return e.child
}

func New(msg string, child *error) ParentError {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	parentErr := errors.New(msg)
	contextedErr := contextedError.NewFromExists(&parentErr, fn.Name())
	return ParentError{contextedErr, *child}
}
